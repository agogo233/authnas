package middleware

import (
	"fmt"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/gin-gonic/gin"
)

func SecurityHeaders(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		if cfg.Server.HSTSMaxAge > 0 {
			hstsValue := ""
			if cfg.Server.HSTSIncludeSubDomains {
				hstsValue = "max-age=%d; includeSubDomains"
			} else {
				hstsValue = "max-age=%d"
			}
			if cfg.Server.HSTSPreload {
				hstsValue += "; preload"
			}
			c.Header("Strict-Transport-Security", fmt.Sprintf(hstsValue, cfg.Server.HSTSMaxAge))
		}

		if cfg.Server.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", cfg.Server.ContentSecurityPolicy)
		}

		c.Next()
	}
}
