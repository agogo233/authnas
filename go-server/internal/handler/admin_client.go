package handler

import (
	"errors"
	"strings"

	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type ClientListResponse struct {
	Clients []ClientListItem `json:"clients"`
	Total   int64            `json:"total"`
}

type ClientListItem struct {
	ID        string  `json:"id"`
	ClientID  string  `json:"clientId"`
	Name      string  `json:"name"`
	LogoURI   *string `json:"logoUri"`
	CreatedAt string  `json:"createdAt"`
}

type CreateClientRequest struct {
	ClientID     string  `json:"clientId" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	LogoURI      *string `json:"logoUri"`
	RedirectURIs *string `json:"redirectUris"`
	Scopes       *string `json:"scopes"`
}

type UpdateClientRequest struct {
	Name         *string `json:"name"`
	LogoURI      *string `json:"logoUri"`
	RedirectURIs *string `json:"redirectUris"`
	Scopes       *string `json:"scopes"`
}

func (h *AdminHandler) ListClients(c *gin.Context) {
	clients, total, err := h.clientService.List(0, 100)
	if err != nil {
		response.InternalServerError(c, "failed to list clients")
		return
	}

	var items []ClientListItem
	for _, cl := range clients {
		items = append(items, ClientListItem{
			ID:        cl.ID,
			ClientID:  cl.ClientID,
			Name:      cl.Name,
			LogoURI:   cl.LogoURI,
			CreatedAt: cl.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response.SuccessList(c, items, total)
}

func (h *AdminHandler) GetClient(c *gin.Context) {
	id := c.Param("id")

	client, err := h.clientService.GetByID(id)
	if err != nil || client == nil {
		response.NotFound(c, "client not found")
		return
	}

	response.Success(c, gin.H{
		"id":           client.ID,
		"clientId":     client.ClientID,
		"name":         client.Name,
		"logoUri":      client.LogoURI,
		"redirectUris": client.RedirectURIs,
		"scopes":       client.Scopes,
		"createdAt":    client.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) CreateClient(c *gin.Context) {
	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	client := &model.Client{
		ID:       uuid.New().String(),
		ClientID: req.ClientID,
		Name:     req.Name,
		LogoURI:  req.LogoURI,
	}
	if req.RedirectURIs != nil {
		client.RedirectURIs = *req.RedirectURIs
	}
	if req.Scopes != nil {
		client.Scopes = req.Scopes
	} else {
		client.Scopes = strPtr("openid profile email")
	}

	if err := h.clientService.Create(client); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || (err.Error() != "" && strings.Contains(err.Error(), "UNIQUE")) {
			response.BadRequest(c, "client with this ID already exists")
			return
		}
		response.InternalServerError(c, "failed to create client")
		return
	}

	response.Success(c, ClientListItem{
		ID:        client.ID,
		ClientID:  client.ClientID,
		Name:      client.Name,
		LogoURI:   client.LogoURI,
		CreatedAt: client.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) UpdateClient(c *gin.Context) {
	id := c.Param("id")

	var req UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	client, err := h.clientService.GetByID(id)
	if err != nil || client == nil {
		response.NotFound(c, "client not found")
		return
	}

	if req.Name != nil {
		client.Name = *req.Name
	}
	if req.LogoURI != nil {
		client.LogoURI = req.LogoURI
	}
	if req.RedirectURIs != nil {
		client.RedirectURIs = *req.RedirectURIs
	}
	if req.Scopes != nil {
		client.Scopes = req.Scopes
	}

	if err := h.clientService.Update(client); err != nil {
		response.InternalServerError(c, "failed to update client")
		return
	}

	response.SuccessWithMessage(c, "client updated successfully")
}

func (h *AdminHandler) DeleteClient(c *gin.Context) {
	id := c.Param("id")

	existingClient, err := h.clientService.GetByID(id)
	if err != nil || existingClient == nil {
		response.NotFound(c, "client not found")
		return
	}

	err = h.clientService.Delete(id)
	if err != nil {
		response.InternalServerError(c, "failed to delete client")
		return
	}

	response.SuccessWithMessage(c, "client deleted successfully")
}
