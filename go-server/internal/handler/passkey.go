package handler

import (
	"bytes"

	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
)

type PasskeyHandler struct {
	passkeyService *service.PasskeyService
}

func NewPasskeyHandler(passkeyService *service.PasskeyService) *PasskeyHandler {
	return &PasskeyHandler{passkeyService: passkeyService}
}

func (h *PasskeyHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/registration/start", h.RegistrationStart)
	r.POST("/registration/end", h.RegistrationEnd)
	r.GET("", h.GetPasskeys)
	r.DELETE("/:id", h.DeletePasskey)
}

type RegistrationStartResponse struct {
	Challenge string `json:"challenge,omitempty"`
	Options   string `json:"options,omitempty"`
}

func (h *PasskeyHandler) RegistrationStart(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	opts, err := h.passkeyService.GenerateRegistrationOptions(user.ID, user.Username)
	if err != nil {
		response.InternalServerError(c, "failed to generate options")
		return
	}

	response.Success(c, gin.H{
		"challenge": opts.Challenge,
		"options":   string(opts.Options),
	})
}

type RegistrationEndRequest struct {
	Challenge string `json:"challenge" binding:"required"`
	Options   string `json:"options" binding:"required"`
	Name      string `json:"name"`
}

func (h *PasskeyHandler) RegistrationEnd(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	var req RegistrationEndRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	parsedData, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader([]byte(req.Options)))
	if err != nil {
		response.BadRequest(c, "invalid response")
		return
	}

	passkey, err := h.passkeyService.CreateCredentialFromResponse(user.ID, user.Username, req.Name, parsedData)
	if err != nil {
		response.InternalServerError(c, "failed to create passkey")
		return
	}

	response.Success(c, gin.H{
		"id":           passkey.ID,
		"name":         passkey.Name,
		"credentialId": passkey.CredentialID,
		"createdAt":    passkey.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updatedAt":    passkey.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

type PasskeyResponse struct {
	ID              string  `json:"id"`
	Name            *string `json:"name,omitempty"`
	CredentialID    string  `json:"credentialId"`
	AttestationType *string `json:"attestationType,omitempty"`
	LastUsedAt      *string `json:"lastUsedAt,omitempty"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

func (h *PasskeyHandler) GetPasskeys(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	passkeys, err := h.passkeyService.GetByUserID(user.ID)
	if err != nil {
		response.InternalServerError(c, "failed to get passkeys")
		return
	}

	var resp []PasskeyResponse
	for _, pk := range passkeys {
		var lastUsed *string
		if pk.LastUsedAt != nil {
			t := pk.LastUsedAt.Format("2006-01-02T15:04:05Z07:00")
			lastUsed = &t
		}
		createdAt := pk.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		updatedAt := ""
		if !pk.UpdatedAt.IsZero() {
			updatedAt = pk.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		resp = append(resp, PasskeyResponse{
			ID:              pk.ID,
			Name:            pk.Name,
			CredentialID:    pk.CredentialID,
			AttestationType: pk.AttestationType,
			LastUsedAt:      lastUsed,
			CreatedAt:       createdAt,
			UpdatedAt:       updatedAt,
		})
	}

	response.Success(c, resp)
}

func (h *PasskeyHandler) DeletePasskey(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	passkeyID := c.Param("id")
	if passkeyID == "" {
		response.BadRequest(c, "passkey id is required")
		return
	}

	passkey, err := h.passkeyService.GetByID(passkeyID)
	if err != nil || passkey == nil {
		response.NotFound(c, "passkey not found")
		return
	}

	if passkey.UserID != user.ID {
		response.Forbidden(c, "not authorized")
		return
	}

	err = h.passkeyService.Delete(passkeyID)
	if err != nil {
		response.InternalServerError(c, "failed to delete passkey")
		return
	}

	response.SuccessWithMessage(c, "passkey deleted successfully")
}
