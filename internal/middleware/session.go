package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/notfixingit3/echostate/internal/auth"
)

const contextUserKey = "echostate_user"

func SessionAuth(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(auth.SessionCookieName)
		if err != nil || token == "" {
			c.Next()
			return
		}
		user, err := svc.SessionUser(c.Request.Context(), token)
		if err != nil {
			c.Next()
			return
		}
		c.Set(contextUserKey, user)
		c.Next()
	}
}

func RequireAuth(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		required, err := svc.AuthRequired(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth check failed"})
			return
		}
		if !required {
			c.Next()
			return
		}
		if GetUser(c) != nil {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
	}
}

func RequireRole(minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUser(c)
		if user == nil {
			c.Next()
			return
		}
		if !auth.CanAccess(user.Role, minRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func RequireAuthRole(svc *auth.Service, minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		required, err := svc.AuthRequired(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth check failed"})
			return
		}
		if !required {
			if minRole == auth.RoleAdmin {
				if !APIKeySatisfied(c) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing API key"})
					return
				}
			}
			c.Next()
			return
		}
		user := GetUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		if !auth.CanAccess(user.Role, minRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func GetUser(c *gin.Context) *auth.User {
	value, ok := c.Get(contextUserKey)
	if !ok {
		return nil
	}
	user, ok := value.(*auth.User)
	if !ok {
		return nil
	}
	return user
}

func APIKeySatisfied(c *gin.Context) bool {
	expected := configuredAPIKey()
	if expected == "" {
		return true
	}
	provided := extractAPIKey(c)
	return provided != "" && provided == expected
}

func extractAPIKey(c *gin.Context) string {
	provided := c.GetHeader("X-API-Key")
	if provided != "" {
		return provided
	}
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}