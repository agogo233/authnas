package middleware

import (
	"strings"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	cfg         *config.Config
	userRepo    *repository.UserRepository
	keyRepo     *repository.KeyRepository
	authService *service.AuthService
}

func NewAuthMiddleware(cfg *config.Config, userRepo *repository.UserRepository, keyRepo *repository.KeyRepository, authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		cfg:         cfg,
		userRepo:    userRepo,
		keyRepo:     keyRepo,
		authService: authService,
	}
}

type Claims struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	TokenVersion int    `json:"token_version"`
	IsAdmin      bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func (am *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
				return []byte(am.cfg.Security.StorageKey), nil
			}
			if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
				if am.authService != nil {
					pubKey := am.authService.GetPublicKey()
					if pubKey != nil {
						return pubKey, nil
					}
				}
				return nil, jwt.ErrSignatureInvalid
			}
			return nil, jwt.ErrSignatureInvalid
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		user, err := am.userRepo.GetByID(claims.UserID)
		if err != nil {
			response.Unauthorized(c, "user not found")
			c.Abort()
			return
		}

		if user.TokenVersion != claims.TokenVersion {
			response.Unauthorized(c, "session revoked")
			c.Abort()
			return
		}

		if user.MustChangePassword {
			path := c.Request.URL.Path
			if path != "/api/user/change-password" && path != "/api/user/password" {
				response.ErrorWithCode(c, 403, "password change required", "MUST_CHANGE_PASSWORD")
				c.Abort()
				return
			}
		}

		c.Set("user", user)
		c.Set("claims", claims)
		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) *model.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*model.User)
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		claims, exists := c.Get("claims")
		if !exists {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		if !claims.(*Claims).IsAdmin {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
