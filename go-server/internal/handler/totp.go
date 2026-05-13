package handler

import (
	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/gin-gonic/gin"
)

type TOTPHandler struct {
	totpService *service.TOTPService
	userService *service.UserService
}

func NewTOTPHandler(totpService *service.TOTPService, userService *service.UserService) *TOTPHandler {
	return &TOTPHandler{totpService: totpService, userService: userService}
}

func (h *TOTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/registration", h.Register)
	r.POST("/verify", h.Verify)
	r.DELETE("", h.Delete)
}

type RegisterTOTPResponse struct {
	Secret    string `json:"secret"`
	QRCodeURI string `json:"qr_code_uri"`
}

type VerifyTOTPRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *TOTPHandler) Register(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	result, err := h.totpService.Generate(user)
	if err != nil {
		response.InternalServerError(c, "failed to generate TOTP")
		return
	}

	response.Success(c, RegisterTOTPResponse{
		Secret:    result.Secret,
		QRCodeURI: result.QRCodeURI,
	})
}

func (h *TOTPHandler) Verify(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	if h.totpService.IsLockedOut(user.ID) {
		response.TooManyRequests(c, "too many failed attempts, please try again later")
		return
	}

	var req VerifyTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if !h.totpService.Validate(user.ID, req.Token) {
		response.Unauthorized(c, "invalid token")
		return
	}

	response.SuccessWithMessage(c, "TOTP verified and enabled")
}

type DeleteTOTPRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *TOTPHandler) Delete(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	var req DeleteTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if !h.totpService.Validate(user.ID, req.Token) {
		response.Unauthorized(c, "invalid token")
		return
	}

	err := h.totpService.Delete(user.ID)
	if err != nil {
		response.InternalServerError(c, "failed to delete TOTP")
		return
	}

	if h.userService != nil {
		fullUser, err := h.userService.GetByID(user.ID)
		if err == nil && fullUser != nil {
			fullUser.MFARequired = false
			if err := h.userService.Update(fullUser); err != nil {
				response.InternalServerError(c, "failed to update user MFA setting")
				return
			}
		}
	}

	response.SuccessWithMessage(c, "TOTP deleted successfully")
}
