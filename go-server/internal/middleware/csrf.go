package middleware

import (
	"strings"

	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/gin-gonic/gin"
)

func CSRFValidation(csrfService *service.CSRFService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			token = c.PostForm("csrf_token")
		}

		if token == "" {
			cookieToken, err := c.Cookie("csrf_token")
			if err == nil {
				token = cookieToken
			}
		}

		if token == "" {
			response.Forbidden(c, "csrf_token_required")
			c.Abort()
			return
		}

		userID := ""
		if currentUser := GetCurrentUser(c); currentUser != nil {
			userID = currentUser.ID
		}

		if !csrfService.ValidateToken(token, userID) {
			response.Forbidden(c, "csrf_token_invalid")
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetBearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
