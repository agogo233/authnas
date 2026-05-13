package handler

import (
	"net/http"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProxyAuthHandler struct {
	cfg           *config.Config
	authService   *service.AuthService
	userRepo      *repository.UserRepository
	keyRepo       *repository.KeyRepository
	groupRepo     *repository.GroupRepository
}

func NewProxyAuthHandler(
	cfg *config.Config,
	authService *service.AuthService,
	userRepo *repository.UserRepository,
	keyRepo *repository.KeyRepository,
	groupRepo *repository.GroupRepository,
) *ProxyAuthHandler {
	return &ProxyAuthHandler{
		cfg:         cfg,
		authService: authService,
		userRepo:    userRepo,
		keyRepo:     keyRepo,
		groupRepo:   groupRepo,
	}
}

func (h *ProxyAuthHandler) ForwardAuth(c *gin.Context) {
	token := c.GetHeader("X-Auth-Token")
	if token == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	user, err := h.authService.ValidateToken(token)
	if err != nil || user == nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	c.Header("X-Auth-User-ID", user.UserID)
	c.Header("X-Auth-Username", user.Username)
	c.Status(http.StatusOK)
}

func (h *ProxyAuthHandler) AuthRequest(c *gin.Context) {
	response.Success(c, gin.H{
		"authURL": h.cfg.App.URL + "/login",
	})
}

type ProxyAuthListResponse struct {
	ProxyAuths []ProxyAuthListItem `json:"proxyauths"`
	Total      int64               `json:"total"`
}

type ProxyAuthListItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ProxyURL  string `json:"proxyUrl"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"createdAt"`
}

type CreateProxyAuthRequest struct {
	Name       string  `json:"name" binding:"required"`
	ProxyURL   string  `json:"proxyUrl" binding:"required"`
	HeaderName string  `json:"headerName" binding:"required"`
	Scopes     *string `json:"scopes"`
	GroupID    *string `json:"groupId"`
	Enabled    *bool   `json:"enabled"`
}

type UpdateProxyAuthRequest struct {
	Name       *string `json:"name"`
	ProxyURL   *string `json:"proxyUrl"`
	HeaderName *string `json:"headerName"`
	Enabled    *bool   `json:"enabled"`
}

func (h *AdminHandler) ListProxyAuth(c *gin.Context) {
	proxyAuths, total, err := h.proxyAuthService.List(0, 100)
	if err != nil {
		response.InternalServerError(c, "failed to list proxy auth")
		return
	}

	var items []ProxyAuthListItem
	for _, pa := range proxyAuths {
		items = append(items, ProxyAuthListItem{
			ID:        pa.ID,
			Name:      pa.Name,
			ProxyURL:  pa.ProxyURL,
			Enabled:   pa.Enabled,
			CreatedAt: pa.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response.SuccessList(c, items, total)
}

func (h *AdminHandler) GetProxyAuth(c *gin.Context) {
	id := c.Param("id")

	proxyAuth, err := h.proxyAuthService.GetByID(id)
	if err != nil || proxyAuth == nil {
		response.NotFound(c, "proxy auth not found")
		return
	}

	response.Success(c, gin.H{
		"id":         proxyAuth.ID,
		"name":       proxyAuth.Name,
		"proxyUrl":   proxyAuth.ProxyURL,
		"headerName": proxyAuth.HeaderName,
		"enabled":    proxyAuth.Enabled,
		"createdAt":  proxyAuth.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) CreateProxyAuth(c *gin.Context) {
	var req CreateProxyAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	proxyAuth := &model.ProxyAuth{
		ID:         uuid.New().String(),
		Name:       req.Name,
		ProxyURL:   req.ProxyURL,
		HeaderName: req.HeaderName,
		Scopes:     req.Scopes,
		GroupID:    req.GroupID,
		Enabled:    enabled,
	}

	if err := h.proxyAuthService.Create(proxyAuth); err != nil {
		response.InternalServerError(c, "failed to create proxy auth")
		return
	}

	response.Success(c, ProxyAuthListItem{
		ID:        proxyAuth.ID,
		Name:      proxyAuth.Name,
		ProxyURL:  proxyAuth.ProxyURL,
		Enabled:   proxyAuth.Enabled,
		CreatedAt: proxyAuth.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) UpdateProxyAuth(c *gin.Context) {
	id := c.Param("id")

	var req UpdateProxyAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	proxyAuth, err := h.proxyAuthService.GetByID(id)
	if err != nil || proxyAuth == nil {
		response.NotFound(c, "proxy auth not found")
		return
	}

	if req.Name != nil {
		proxyAuth.Name = *req.Name
	}
	if req.ProxyURL != nil {
		proxyAuth.ProxyURL = *req.ProxyURL
	}
	if req.HeaderName != nil {
		proxyAuth.HeaderName = *req.HeaderName
	}
	if req.Enabled != nil {
		proxyAuth.Enabled = *req.Enabled
	}

	if err := h.proxyAuthService.Update(proxyAuth); err != nil {
		response.InternalServerError(c, "failed to update proxy auth")
		return
	}

	response.SuccessWithMessage(c, "proxy auth updated successfully")
}

func (h *AdminHandler) DeleteProxyAuth(c *gin.Context) {
	id := c.Param("id")

	existingProxyAuth, err := h.proxyAuthService.GetByID(id)
	if err != nil || existingProxyAuth == nil {
		response.NotFound(c, "proxy auth not found")
		return
	}

	err = h.proxyAuthService.Delete(id)
	if err != nil {
		response.InternalServerError(c, "failed to delete proxy auth")
		return
	}

	response.SuccessWithMessage(c, "proxy auth deleted successfully")
}
