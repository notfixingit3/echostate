package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/notfixingit3/echostate/internal/config"
)

// APIKeyAuth protects write endpoints when an API key is configured.
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := configuredAPIKey()
		if expected == "" {
			c.Next()
			return
		}

		provided := strings.TrimSpace(c.GetHeader("X-API-Key"))
		if provided == "" {
			auth := strings.TrimSpace(c.GetHeader("Authorization"))
			if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
				provided = strings.TrimSpace(auth[7:])
			}
		}

		if provided == "" || provided != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing API key"})
			return
		}
		c.Next()
	}
}

func configuredAPIKey() string {
	if value := strings.TrimSpace(os.Getenv("ECHOSTATE_API_KEY")); value != "" {
		return value
	}
	return strings.TrimSpace(config.GetSettings().APIKey)
}
