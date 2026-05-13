package middleware

import (
	"net/http"
	"strings"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/gin-gonic/gin"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin == "" {
			c.Next()
			return
		}

		if !isOriginAllowed(origin, cfg.CORS.AllowedOrigins, cfg.CORS.AllowCredentials) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		if cfg.CORS.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.CORS.AllowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.CORS.AllowedMethods, ", "))

		if c.Request.Method == http.MethodOptions {
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isOriginAllowed(origin string, allowedOrigins []string, allowCredentials bool) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			if allowCredentials {
				return false
			}
			return true
		}
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			suffix := "." + domain
			if strings.HasSuffix(origin, suffix) || origin == domain {
				return true
			}
		}
		if allowed == origin {
			return true
		}
	}
	return false
}
