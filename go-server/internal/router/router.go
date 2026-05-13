package router

import (
	"github.com/authnas/authnas/go-server/internal/handler"
	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/gin-gonic/gin"
)

type Routers struct {
	AuthHandler          *handler.AuthHandler
	UserHandler          *handler.UserHandler
	PasskeyHandler       *handler.PasskeyHandler
	TOTPHandler          *handler.TOTPHandler
	AdminHandler         *handler.AdminHandler
	AdminSettingsHandler *handler.AdminSettingsHandler
	OIDCHandler          *handler.OIDCHandler
	ProxyAuthHandler     *handler.ProxyAuthHandler
	AuthMiddleware       *middleware.AuthMiddleware
	CSRFService          *service.CSRFService
}

func (r *Routers) RegisterPublicRoutes(api *gin.RouterGroup) {
	api.GET("/health", handler.HealthCheck)
	api.GET("/config", r.AuthHandler.GetPublicConfig)
	r.AuthHandler.RegisterRoutes(api)
}

func (r *Routers) RegisterUserRoutes(api *gin.RouterGroup) {
	user := api.Group("/user")
	user.Use(r.AuthMiddleware.Authenticate())
	user.Use(middleware.CSRFValidation(r.CSRFService))
	r.UserHandler.RegisterRoutes(user)
}

func (r *Routers) RegisterPasskeyRoutes(api *gin.RouterGroup) {
	passkey := api.Group("/passkey")
	passkey.Use(r.AuthMiddleware.Authenticate())
	passkey.Use(middleware.CSRFValidation(r.CSRFService))
	r.PasskeyHandler.RegisterRoutes(passkey)
}

func (r *Routers) RegisterTOTPRoutes(api *gin.RouterGroup) {
	totpRoutes := api.Group("/totp")
	totpRoutes.Use(r.AuthMiddleware.Authenticate())
	totpRoutes.Use(middleware.CSRFValidation(r.CSRFService))
	r.TOTPHandler.RegisterRoutes(totpRoutes)
}

func (r *Routers) RegisterAdminRoutes(api *gin.RouterGroup) {
	admin := api.Group("/admin")
	admin.Use(r.AuthMiddleware.Authenticate())
	admin.Use(middleware.RequireAdmin())
	r.AdminHandler.RegisterRoutes(admin)

	settings := admin.Group("/settings")
	settings.Use(middleware.CSRFValidation(r.CSRFService))
	settings.GET("/general", r.AdminSettingsHandler.GetGeneral)
	settings.POST("/general", r.AdminSettingsHandler.SetGeneral)
	settings.GET("/security", r.AdminSettingsHandler.GetSecurity)
	settings.POST("/security", r.AdminSettingsHandler.SetSecurity)
	settings.GET("/email", r.AdminSettingsHandler.GetEmail)
	settings.POST("/email", r.AdminSettingsHandler.SetEmail)
	settings.GET("/session", r.AdminSettingsHandler.GetSession)
	settings.POST("/session", r.AdminSettingsHandler.SetSession)
	settings.GET("/ratelimit", r.AdminSettingsHandler.GetRateLimit)
	settings.POST("/ratelimit", r.AdminSettingsHandler.SetRateLimit)
	settings.POST("/email/test", r.AdminSettingsHandler.TestEmail)
}

func (r *Routers) RegisterOIDCRoutes(router *gin.Engine) {
	r.OIDCHandler.RegisterRoutes(router)
}

func (r *Routers) RegisterProxyAuthRoutes(api *gin.RouterGroup) {
	authz := api.Group("/authz")
	{
		authz.GET("/forward-auth", r.ProxyAuthHandler.ForwardAuth)
		authz.GET("/auth-request", r.ProxyAuthHandler.AuthRequest)
	}
}

func SetupAPIRoutes(api *gin.RouterGroup, routers *Routers) {
	routers.RegisterPublicRoutes(api)
	routers.RegisterUserRoutes(api)
	routers.RegisterPasskeyRoutes(api)
	routers.RegisterTOTPRoutes(api)
	routers.RegisterAdminRoutes(api)
	routers.RegisterProxyAuthRoutes(api)
}
