package middleware

import (
	"net/url"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewCORS returns a CORS middleware configured with the given frontend origin.
// In production, only the specified origin is allowed. In development, any
// localhost/127.0.0.1 origin is also permitted so `next dev` works against a
// Docker-backed API.
func NewCORS(frontendURL, env string) gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if env == "production" {
		cfg.AllowOrigins = []string{frontendURL}
	} else {
		cfg.AllowOriginFunc = func(origin string) bool {
			if origin == frontendURL {
				return true
			}
			return isLocalDevOrigin(origin)
		}
	}

	return cors.New(cfg)
}

func isLocalDevOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "http" {
		return false
	}

	host := u.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		return false
	}
	return u.Port() != ""
}
