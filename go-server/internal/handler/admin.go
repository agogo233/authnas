package handler

import (
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userService       *service.UserService
	groupService      *service.GroupService
	clientService     *service.ClientService
	invitationService *service.InvitationService
	proxyAuthService  *service.ProxyAuthService
}

func NewAdminHandler(
	userService *service.UserService,
	groupService *service.GroupService,
	clientService *service.ClientService,
	invitationService *service.InvitationService,
	proxyAuthService *service.ProxyAuthService,
) *AdminHandler {
	return &AdminHandler{
		userService:       userService,
		groupService:      groupService,
		clientService:     clientService,
		invitationService: invitationService,
		proxyAuthService:  proxyAuthService,
	}
}

func (h *AdminHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/users/count", h.CountUsers)
	r.GET("/users", h.ListUsers)
	r.POST("/users", h.CreateUser)
	r.GET("/users/:id", h.GetUser)
	r.PUT("/users/:id", h.UpdateUser)
	r.DELETE("/users/:id", h.DeleteUser)
	r.POST("/users/:id/approve", h.ApproveUser)
	r.POST("/users/:id/reset-password", h.ResetPassword)

	r.GET("/groups", h.ListGroups)
	r.POST("/groups", h.CreateGroup)
	r.GET("/groups/:id", h.GetGroup)
	r.PUT("/groups/:id", h.UpdateGroup)
	r.DELETE("/groups/:id", h.DeleteGroup)

	r.GET("/clients", h.ListClients)
	r.POST("/clients", h.CreateClient)
	r.GET("/clients/:id", h.GetClient)
	r.PUT("/clients/:id", h.UpdateClient)
	r.DELETE("/clients/:id", h.DeleteClient)

	r.GET("/invitations", h.ListInvitations)
	r.POST("/invitations", h.CreateInvitation)
	r.GET("/invitations/:id", h.GetInvitation)
	r.DELETE("/invitations/:id", h.DeleteInvitation)

	r.GET("/proxyauth", h.ListProxyAuth)
	r.POST("/proxyauth", h.CreateProxyAuth)
	r.GET("/proxyauth/:id", h.GetProxyAuth)
	r.PUT("/proxyauth/:id", h.UpdateProxyAuth)
	r.DELETE("/proxyauth/:id", h.DeleteProxyAuth)
}
