package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/handler"
	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/internal/router"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/authnas/authnas/go-server/pkg/email"
	"github.com/authnas/authnas/go-server/pkg/utils"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.New(cfg)
	if err != nil {
		return err
	}

	if err := database.RunMigrations(db); err != nil {
		return err
	}

	emailSender := email.NewSender(cfg)
	randomUtil := utils.NewRandom()
	timeUtil := utils.NewTime()

	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	clientRepo := repository.NewClientRepository(db)
	consentRepo := repository.NewConsentRepository(db)
	passkeyRepo := repository.NewPasskeyRepository(db)
	totpRepo := repository.NewTOTPRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	keyRepo := repository.NewKeyRepository(db)
	emailVerificationRepo := repository.NewEmailVerificationRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	proxyAuthRepo := repository.NewProxyAuthRepository(db)
	emailLogRepo := repository.NewEmailLogRepository(db)

	authService := service.NewAuthService(cfg, userRepo, keyRepo, totpRepo, passkeyRepo, emailVerificationRepo, randomUtil)
	passkeyService := service.NewPasskeyService(cfg, passkeyRepo, userRepo, repository.NewPasskeyAuthOptionsRepository(db), randomUtil)
	totpService := service.NewTOTPService(cfg, totpRepo, userRepo)
	emailService := service.NewEmailService(cfg, emailVerificationRepo, passwordResetRepo, emailLogRepo, emailSender)
	invitationService := service.NewInvitationService(cfg, invitationRepo, userRepo)
	userService := service.NewUserService(cfg, userRepo, groupRepo, keyRepo, totpRepo, passkeyRepo, emailVerificationRepo, passwordResetRepo, consentRepo, emailService, randomUtil, timeUtil)
	userService.SetInvitationService(invitationService)
	userService.SetDB(db)
	groupService := service.NewGroupService(cfg, groupRepo, userRepo)
	clientService := service.NewClientService(cfg, clientRepo)
	proxyAuthService := service.NewProxyAuthService(cfg, proxyAuthRepo, userRepo)
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)
	oidcService := service.NewOIDCService(cfg, db, clientRepo, consentRepo, oidcPayloadRepo, userRepo, groupRepo, keyRepo, authService, randomUtil)
	csrfService := service.NewCSRFService(randomUtil, 60)
	cleanupService := service.NewCleanupService(
		keyRepo,
		repository.NewPasskeyAuthOptionsRepository(db),
		oidcPayloadRepo,
		emailVerificationRepo,
		passwordResetRepo,
		emailLogRepo,
		csrfService,
	)
	go cleanupService.StartCleanupScheduler(15 * time.Minute)

	if err := userService.EnsureInitialAdmin(
		cfg.Security.InitialAdminUsername,
		cfg.Security.InitialAdminEmail,
		cfg.Security.InitialAdminPassword,
	); err != nil {
		log.Printf("Warning: failed to ensure initial admin: %v", err)
	}

	settingService := service.NewSystemSettingService(db)
	if err := settingService.InitializeDefaults(); err != nil {
		log.Printf("Warning: failed to initialize system settings defaults: %v", err)
	}
	authMiddleware := middleware.NewAuthMiddleware(cfg, userRepo, keyRepo, authService)
	authHandler := handler.NewAuthHandler(cfg, authService, userService, totpService, passkeyService, invitationService, emailService, settingService, csrfService, authMiddleware)
	userHandler := handler.NewUserHandler(userService, totpService, passkeyService)
	passkeyHandler := handler.NewPasskeyHandler(passkeyService)
	totpHandler := handler.NewTOTPHandler(totpService, userService)
	adminHandler := handler.NewAdminHandler(userService, groupService, clientService, invitationService, proxyAuthService)
	adminSettingsHandler := handler.NewAdminSettingsHandler(settingService, emailService)
	oidcHandler := handler.NewOIDCHandler(cfg, oidcService, clientService, userService, csrfService)
	proxyAuthHandler := handler.NewProxyAuthHandler(cfg, authService, userRepo, keyRepo, groupRepo)

	routers := &router.Routers{
		AuthHandler:          authHandler,
		UserHandler:          userHandler,
		PasskeyHandler:       passkeyHandler,
		TOTPHandler:          totpHandler,
		AdminHandler:         adminHandler,
		AdminSettingsHandler: adminSettingsHandler,
		OIDCHandler:          oidcHandler,
		ProxyAuthHandler:     proxyAuthHandler,
		AuthMiddleware:       authMiddleware,
		CSRFService:          csrfService,
	}

	r := gin.Default()

	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.RateLimit(cfg))
	middleware.InitTrustedProxies(cfg)
	r.Use(middleware.GetRealClientIP())
	r.Use(middleware.SecurityHeaders(cfg))
	r.Use(middleware.Config(cfg))

	spaHandler := func(c *gin.Context) {
		c.File("./static/index.html")
		c.Abort()
	}

	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/oidc") {
			c.Next()
			return
		}

		if path == "/" {
			spaHandler(c)
			return
		}

		if strings.Contains(path, ".") {
			c.Next()
			return
		}

		spaHandler(c)
	})

	r.Use(static.Serve("/", static.LocalFile("./static", false)))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/oidc") {
			c.Status(404)
			return
		}
		c.File("./static/index.html")
	})

	api := r.Group("/api")
	router.SetupAPIRoutes(api, routers)
	routers.RegisterOIDCRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	oidcService.Stop()
	cleanupService.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server exited gracefully")
	return nil
}
