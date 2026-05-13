package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/authnas/authnas/go-server/internal/config"
)

func Config(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("config", cfg)
		c.Next()
	}
}
