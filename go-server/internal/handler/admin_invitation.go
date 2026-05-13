package handler

import (
	"time"

	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/gin-gonic/gin"
)

type InvitationListResponse struct {
	Invitations []InvitationListItem `json:"invitations"`
	Total       int64                `json:"total"`
}

type InvitationListItem struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Username  *string `json:"username,omitempty"`
	Code      string  `json:"code,omitempty"`
	ExpiresAt string  `json:"expiresAt"`
	CreatedAt string  `json:"createdAt"`
}

type CreateInvitationRequest struct {
	Email     string  `json:"email" binding:"required"`
	Username  string  `json:"username"`
	Scopes    *string `json:"scopes"`
	GroupID   *string `json:"groupId"`
	MaxUses   *int    `json:"maxUses"`
	ExpiresAt *string `json:"expiresAt"`
}

func (h *AdminHandler) ListInvitations(c *gin.Context) {
	invitations, total, err := h.invitationService.List(0, 100)
	if err != nil {
		response.InternalServerError(c, "failed to list invitations")
		return
	}

	var items []InvitationListItem
	for _, inv := range invitations {
		items = append(items, InvitationListItem{
			ID:        inv.ID,
			Email:     inv.Email,
			Username:  inv.Username,
			Code:      inv.Code,
			ExpiresAt: inv.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt: inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response.SuccessList(c, items, total)
}

func (h *AdminHandler) GetInvitation(c *gin.Context) {
	id := c.Param("id")

	invitation, err := h.invitationService.GetByID(id)
	if err != nil || invitation == nil {
		response.NotFound(c, "invitation not found")
		return
	}

	response.Success(c, gin.H{
		"id":        invitation.ID,
		"email":     invitation.Email,
		"username":  invitation.Username,
		"code":      invitation.Code,
		"expiresAt": invitation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		"createdAt": invitation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) CreateInvitation(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	expiresIn := time.Duration(0)
	if req.ExpiresAt != nil {
		expiresTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			response.BadRequest(c, "invalid expiresAt format, use RFC3339")
			return
		}
		expiresIn = time.Until(expiresTime)
		if expiresIn < 0 {
			response.BadRequest(c, "expiresAt must be in the future")
			return
		}
	}

	invitation, err := h.invitationService.Create(req.Email, req.Username, expiresIn, req.GroupID, req.Scopes, req.MaxUses, currentUser.ID)
	if err != nil {
		response.InternalServerError(c, "failed to create invitation")
		return
	}

	response.Success(c, InvitationListItem{
		ID:        invitation.ID,
		Email:     invitation.Email,
		Username:  invitation.Username,
		Code:      invitation.Code,
		ExpiresAt: invitation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt: invitation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) DeleteInvitation(c *gin.Context) {
	id := c.Param("id")

	existingInvitation, err := h.invitationService.GetByID(id)
	if err != nil || existingInvitation == nil {
		response.NotFound(c, "invitation not found")
		return
	}

	err = h.invitationService.Delete(id)
	if err != nil {
		response.InternalServerError(c, "failed to delete invitation")
		return
	}

	response.SuccessWithMessage(c, "invitation deleted successfully")
}
