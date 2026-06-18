package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/notfixingit3/echostate/internal/auth"
	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/middleware"
)

func (h *Handler) authConfig(c *gin.Context) {
	required, _ := h.auth.AuthRequired(c.Request.Context())
	settings := auth.SettingsFromConfig(config.GetSettings(), h.config.FrontendURL)
	c.JSON(http.StatusOK, gin.H{
		"auth_required":               required,
		"auth_enabled":              settings.AuthEnabled,
		"enrollment_code_ttl_hours": settings.EnrollmentCodeTTLHours,
		"enrollment_code_length":    settings.EnrollmentCodeLength,
		"recovery_code_length":      settings.RecoveryCodeLength,
		"session_ttl_hours":         settings.SessionTTLHours,
		"max_code_attempts":         settings.MaxCodeAttempts,
		"code_attempt_window_minutes": settings.CodeAttemptWindowMinutes,
		"webauthn_rp_id":            settings.WebAuthnRPID,
		"webauthn_rp_origin":        settings.WebAuthnRPOrigin,
	})
}

func (h *Handler) authSession(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
		return
	}
	creds, _ := h.auth.ListCredentials(c.Request.Context(), user.ID)
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user":          user,
		"credentials":   creds,
	})
}

func (h *Handler) enrollVerify(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.auth.VerifyEnrollmentCode(c.Request.Context(), req.Code, c.ClientIP())
	if err != nil {
		status := http.StatusUnauthorized
		if err == auth.ErrTooManyAttempts {
			status = http.StatusTooManyRequests
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	token, expires, err := h.auth.CreateEnrollmentSession(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start enrollment"})
		return
	}

	creds, _ := h.auth.ListCredentials(c.Request.Context(), user.ID)
	needsPasskey := len(creds) == 0

	setEnrollCookie(c, token, expires, h.config.Env == "production")
	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"needs_passkey": needsPasskey,
		"expires_at":    expires,
	})
}

func (h *Handler) webauthnRegisterBegin(c *gin.Context) {
	userID, ok := h.enrollOrSessionUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "enrollment or session required"})
		return
	}

	options, challengeID, err := h.auth.BeginRegistration(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"options":      options,
		"challenge_id": challengeID,
	})
}

func (h *Handler) webauthnRegisterFinish(c *gin.Context) {
	userID, ok := h.enrollOrSessionUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "enrollment or session required"})
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	var req struct {
		ChallengeID string `json:"challenge_id"`
		Nickname    string `json:"nickname"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.ChallengeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "challenge_id required"})
		return
	}

	cred, err := h.auth.FinishRegistration(c.Request.Context(), userID, req.ChallengeID, body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.auth.StoreCredential(c.Request.Context(), userID, *cred, req.Nickname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store credential"})
		return
	}

	if token, err := c.Cookie(auth.EnrollCookieName); err == nil && token != "" {
		_ = h.auth.DeleteEnrollmentSession(c.Request.Context(), token)
		clearEnrollCookie(c, h.config.Env == "production")
	}

	sessionToken, expires, err := h.auth.CreateSession(c.Request.Context(), userID, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	setSessionCookie(c, sessionToken, expires, h.config.Env == "production")

	user, _ := h.auth.GetUser(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{"user": user, "registered": true})
}

func (h *Handler) webauthnLoginBegin(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id"`
	}
	_ = c.ShouldBindJSON(&req)

	var userID uuid.UUID
	if id, ok := h.enrollOrSessionUserID(c); ok {
		userID = id
	} else if strings.TrimSpace(req.UserID) != "" {
		parsed, err := uuid.Parse(req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		userID = parsed
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id or enrollment required"})
		return
	}

	options, challengeID, err := h.auth.BeginLogin(c.Request.Context(), userID)
	if err != nil {
		if err == auth.ErrNoCredentials {
			token, expires, err2 := h.auth.CreateEnrollmentSession(c.Request.Context(), userID)
			if err2 != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start enrollment"})
				return
			}
			setEnrollCookie(c, token, expires, h.config.Env == "production")
			c.JSON(http.StatusOK, gin.H{"needs_passkey": true})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"options":      options,
		"challenge_id": challengeID,
		"user_id":      userID,
	})
}

func (h *Handler) webauthnLoginFinish(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	var req struct {
		ChallengeID string    `json:"challenge_id"`
		UserID      uuid.UUID `json:"user_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.ChallengeID == "" || req.UserID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "challenge_id and user_id required"})
		return
	}

	if _, err := h.auth.FinishLogin(c.Request.Context(), req.UserID, req.ChallengeID, body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionToken, expires, err := h.auth.CreateSession(c.Request.Context(), req.UserID, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	setSessionCookie(c, sessionToken, expires, h.config.Env == "production")

	user, _ := h.auth.GetUser(c.Request.Context(), req.UserID)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) authLogout(c *gin.Context) {
	if token, err := c.Cookie(auth.SessionCookieName); err == nil {
		_ = h.auth.DeleteSession(c.Request.Context(), token)
	}
	clearSessionCookie(c, h.config.Env == "production")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) issueDeviceCode(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	result, err := h.auth.IssueEnrollmentCode(c.Request.Context(), user.ID, &user.ID, auth.PurposeDevice, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue code"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) enrollOrSessionUserID(c *gin.Context) (uuid.UUID, bool) {
	if token, err := c.Cookie(auth.EnrollCookieName); err == nil && token != "" {
		user, err := h.auth.EnrollmentUser(c.Request.Context(), token)
		if err == nil {
			return user.ID, true
		}
	}
	if user := middleware.GetUser(c); user != nil {
		return user.ID, true
	}
	return uuid.Nil, false
}

func setSessionCookie(c *gin.Context, token string, expires time.Time, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.SessionCookieName, token, int(time.Until(expires).Seconds()), "/", "", secure, true)
}

func clearSessionCookie(c *gin.Context, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.SessionCookieName, "", -1, "/", "", secure, true)
}

func setEnrollCookie(c *gin.Context, token string, expires time.Time, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.EnrollCookieName, token, int(time.Until(expires).Seconds()), "/", "", secure, true)
}

func clearEnrollCookie(c *gin.Context, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.EnrollCookieName, "", -1, "/", "", secure, true)
}