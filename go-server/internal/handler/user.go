package handler

import (
	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService    *service.UserService
	totpService    *service.TOTPService
	passkeyService *service.PasskeyService
}

func NewUserHandler(userService *service.UserService, totpService *service.TOTPService, passkeyService *service.PasskeyService) *UserHandler {
	return &UserHandler{
		userService:    userService,
		totpService:    totpService,
		passkeyService: passkeyService,
	}
}

func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/me", h.GetMe)
	r.PUT("/me", h.UpdateMe)
	r.PUT("/me/password", h.UpdatePassword)
	r.GET("/me/sessions", h.ListSessions)
	r.DELETE("/me/sessions", h.DeleteAllSessions)
	r.DELETE("/me/sessions/:id", h.DeleteSession)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	hasTotp := false
	if h.totpService != nil {
		hasTotp = h.totpService.HasTOTP(user.ID)
	}

	hasPasskeys := false
	if h.passkeyService != nil {
		hasPasskeys = h.passkeyService.HasPasskeys(user.ID)
	}

	resp := response.UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		Name:          user.Name,
		IsAdmin:       user.IsAdmin,
		EmailVerified: user.EmailVerified,
		Approved:      user.Approved,
		MFARequired:   user.MFARequired,
		HasTotp:       hasTotp,
		HasPasskeys:   hasPasskeys,
		HasPassword:   user.PasswordHash != nil && *user.PasswordHash != "",
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if !user.UpdatedAt.IsZero() {
		updStr := user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.UpdatedAt = &updStr
	}

	response.Success(c, resp)
}

type UpdateMeRequest struct {
	Email *string `json:"email"`
	Name  *string `json:"name"`
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	var req UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Email != nil {
		if !emailRegex.MatchString(*req.Email) {
			response.BadRequest(c, "invalid email format")
			return
		}
		user.Email = req.Email
	}
	if req.Name != nil {
		user.Name = req.Name
	}

	err := h.userService.Update(user)
	if err != nil {
		response.InternalServerError(c, "failed to update user")
		return
	}

	hasTotp := false
	if h.totpService != nil {
		hasTotp = h.totpService.HasTOTP(user.ID)
	}

	hasPasskeys := false
	if h.passkeyService != nil {
		hasPasskeys = h.passkeyService.HasPasskeys(user.ID)
	}

	resp := response.UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		Name:          user.Name,
		IsAdmin:       user.IsAdmin,
		EmailVerified: user.EmailVerified,
		Approved:      user.Approved,
		MFARequired:   user.MFARequired,
		HasTotp:       hasTotp,
		HasPasskeys:   hasPasskeys,
		HasPassword:   user.PasswordHash != nil && *user.PasswordHash != "",
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if !user.UpdatedAt.IsZero() {
		updStr := user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.UpdatedAt = &updStr
	}

	response.Success(c, resp)
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword" binding:"required"`
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.userService.UpdatePassword(user.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "password updated successfully")
}

func (h *UserHandler) DeleteAllSessions(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	err := h.userService.RevokeAllSessions(user.ID)
	if err != nil {
		response.InternalServerError(c, "failed to revoke sessions")
		return
	}

	response.SuccessWithMessage(c, "all sessions revoked")
}

type SessionResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	ClientID  string `json:"clientId"`
	CreatedAt string `json:"createdAt"`
	ExpiresAt string `json:"expiresAt"`
	UserAgent string `json:"userAgent,omitempty"`
}

func (h *UserHandler) ListSessions(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	sessions, err := h.userService.GetUserSessions(user.ID)
	if err != nil {
		response.InternalServerError(c, "failed to get sessions")
		return
	}

	resp := make([]SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		resp = append(resp, SessionResponse{
			ID:        s.ID,
			UserID:    s.UserID,
			ClientID:  s.ClientID,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			ExpiresAt: s.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			UserAgent: s.UserAgent,
		})
	}

	response.Success(c, resp)
}

func (h *UserHandler) DeleteSession(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		response.BadRequest(c, "session id is required")
		return
	}

	err := h.userService.RevokeSession(user.ID, sessionID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "session deleted")
}
