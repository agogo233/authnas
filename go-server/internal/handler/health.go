package handler

import (
	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
	})
}

type PublicConfig struct {
	AppName                string `json:"app_name"`
	SignupRequiresApproval bool   `json:"signup_requires_approval"`
	EmailVerification      bool   `json:"email_verification"`
	MFARequired            bool   `json:"mfa_required"`
	PasswordMinLength      int    `json:"password_min_length"`
	PasswordStrength       int    `json:"password_strength"`
	DefaultRedirect        string `json:"default_redirect"`
	ContactEmail           string `json:"contact_email"`
}

func GetPublicConfig(c *gin.Context) {
	cfg := c.MustGet("config").(*config.Config)
	response.Success(c, PublicConfig{
		AppName:                cfg.App.Name,
		SignupRequiresApproval: cfg.Security.SignupRequiresApproval,
		EmailVerification:      cfg.Security.EmailVerification,
		MFARequired:            cfg.Security.MFARequired,
		PasswordMinLength:      cfg.Security.PasswordMinLength,
		PasswordStrength:       cfg.Security.PasswordStrength,
		DefaultRedirect:        cfg.App.URL,
		ContactEmail:           cfg.Email.FromAddress,
	})
}
