package handler

import (
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminSettingsHandler struct {
	settingService *service.SystemSettingService
	emailService   *service.EmailService
}

func NewAdminSettingsHandler(settingService *service.SystemSettingService, emailService *service.EmailService) *AdminSettingsHandler {
	return &AdminSettingsHandler{
		settingService: settingService,
		emailService:   emailService,
	}
}

func (h *AdminSettingsHandler) GetGeneral(c *gin.Context) {
	settings, err := h.settingService.GetGeneral()
	if err != nil {
		settings = service.GeneralSettings{
			AppName: "AuthNas",
			AppURL:  "http://localhost:8080",
		}
	}
	response.Success(c, settings)
}

func (h *AdminSettingsHandler) SetGeneral(c *gin.Context) {
	var req service.GeneralSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	if err := h.settingService.SetGeneral(req); err != nil {
		response.InternalServerError(c, "failed to save settings")
		return
	}
	response.Success(c, req)
}

func (h *AdminSettingsHandler) GetSecurity(c *gin.Context) {
	settings, err := h.settingService.GetSecurity()
	if err != nil {
		settings = service.SecuritySettings{
			PasswordMinLength: 8,
			PasswordStrength:  3,
		}
	}
	response.Success(c, settings)
}

func (h *AdminSettingsHandler) SetSecurity(c *gin.Context) {
	var req service.SecuritySettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	if err := h.settingService.SetSecurity(req); err != nil {
		response.InternalServerError(c, "failed to save settings")
		return
	}
	response.Success(c, req)
}

func (h *AdminSettingsHandler) GetEmail(c *gin.Context) {
	settings, err := h.settingService.GetEmail()
	if err != nil {
		settings = service.EmailSettings{
			SMTPPort: 587,
			FromName: "AuthNas",
		}
	}
	settings.SMTPPass = ""
	response.Success(c, settings)
}

func (h *AdminSettingsHandler) SetEmail(c *gin.Context) {
	var req service.EmailSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	existing, _ := h.settingService.GetEmail()
	if req.SMTPPass == "" && existing.SMTPPass != "" {
		req.SMTPPass = existing.SMTPPass
	}
	if err := h.settingService.SetEmail(req); err != nil {
		response.InternalServerError(c, "failed to save settings")
		return
	}
	req.SMTPPass = ""
	response.Success(c, req)
}

func (h *AdminSettingsHandler) GetSession(c *gin.Context) {
	settings, err := h.settingService.GetSession()
	if err != nil {
		settings = service.SessionSettings{
			AccessTokenExpiry:  15,
			RefreshTokenExpiry: 7,
			MaxSessionsPerUser: 5,
		}
	}
	response.Success(c, settings)
}

func (h *AdminSettingsHandler) SetSession(c *gin.Context) {
	var req service.SessionSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	if err := h.settingService.SetSession(req); err != nil {
		response.InternalServerError(c, "failed to save settings")
		return
	}
	response.Success(c, req)
}

func (h *AdminSettingsHandler) GetRateLimit(c *gin.Context) {
	settings, err := h.settingService.GetRateLimit()
	if err != nil {
		settings = service.RateLimitSettings{
			Enabled: true,
		}
	}
	response.Success(c, settings)
}

func (h *AdminSettingsHandler) SetRateLimit(c *gin.Context) {
	var req service.RateLimitSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	if err := h.settingService.SetRateLimit(req); err != nil {
		response.InternalServerError(c, "failed to save settings")
		return
	}
	response.Success(c, req)
}

func (h *AdminSettingsHandler) TestEmail(c *gin.Context) {
	var req struct {
		To string `json:"to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	if err := h.emailService.SendTestEmail(req.To); err != nil {
		response.InternalServerError(c, "failed to send test email: "+err.Error())
		return
	}
	response.Success(c, gin.H{"message": "test email sent"})
}
