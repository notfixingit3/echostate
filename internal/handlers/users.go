package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/auth"
	"github.com/notfixingit3/echostate/internal/middleware"
)

func (h *Handler) listUsers(c *gin.Context) {
	users, err := h.auth.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	if users == nil {
		users = []auth.User{}
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) createUser(c *gin.Context) {
	var req struct {
		DisplayName string `json:"display_name" binding:"required"`
		Role        string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.auth.CreateUser(c.Request.Context(), req.DisplayName, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) issueUserCode(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req struct {
		Purpose  string `json:"purpose"`
		Recovery bool   `json:"recovery"`
	}
	_ = c.ShouldBindJSON(&req)

	actor := middleware.GetUser(c)
	var createdBy *uuid.UUID
	if actor != nil {
		createdBy = &actor.ID
	}

	purpose := strings.TrimSpace(req.Purpose)
	if purpose == "" {
		if req.Recovery {
			purpose = auth.PurposeRecovery
		} else {
			purpose = auth.PurposeInitial
		}
	}

	result, err := h.auth.IssueEnrollmentCode(c.Request.Context(), userID, createdBy, purpose, req.Recovery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue code"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) listUserCredentials(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	actor := middleware.GetUser(c)
	if actor != nil && actor.Role != auth.RoleAdmin && actor.ID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	creds, err := h.auth.ListCredentials(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list credentials"})
		return
	}
	c.JSON(http.StatusOK, creds)
}

func (h *Handler) deleteUserCredential(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	credID, err := uuid.Parse(c.Param("credId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential id"})
		return
	}
	actor := middleware.GetUser(c)
	if actor != nil && actor.Role != auth.RoleAdmin && actor.ID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	if err := h.auth.DeleteCredential(c.Request.Context(), userID, credID); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete credential"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) renameUserCredential(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	credID, err := uuid.Parse(c.Param("credId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential id"})
		return
	}
	var req struct {
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actor := middleware.GetUser(c)
	if actor != nil && actor.Role != auth.RoleAdmin && actor.ID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	if err := h.auth.RenameCredential(c.Request.Context(), userID, credID, req.Nickname); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rename credential"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": true})
}

func (h *Handler) cliIssueAdminCode(c *gin.Context) {
	secret := strings.TrimSpace(os.Getenv("ECHOSTATE_BREAK_GLASS_SECRET"))
	if secret == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": auth.ErrBreakGlassSecret.Error()})
		return
	}
	provided := strings.TrimSpace(c.GetHeader("X-Break-Glass-Secret"))
	if provided == "" || provided != secret {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid break-glass secret"})
		return
	}
	admin, err := h.auth.FindAdminUser(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin user not found"})
		return
	}
	result, err := h.auth.IssueEnrollmentCode(c.Request.Context(), admin.ID, nil, auth.PurposeRecovery, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue recovery code"})
		return
	}
	c.JSON(http.StatusOK, result)
}