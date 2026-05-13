// Package e2e provides end-to-end tests for AuthNas Go server.
//
// IMPORTANT: These tests directly test the Go server (localhost:8080).
// Vite is only used as a build tool to compile frontend resources to static files.
// The Go server serves both API endpoints and frontend static resources.
// Tests do NOT require or use the Vite development server.
//
// Architecture:
//
//	web/src/ (Vue/TypeScript source)
//	       ↓ npm run build (Vite compiles)
//	go-server/static/ (compiled HTML/CSS/JS)
//	       ↓ Go Server reads and serves
//	Test http://localhost:8080 (all API + pages)
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/handler"
	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/internal/router"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/authnas/authnas/go-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type E2ETestServer struct {
	t            *testing.T
	server       *http.Server
	url          string
	db           *gorm.DB
	client       *http.Client
	cfg          *config.Config
	csrfService  *service.CSRFService
}

// NewE2ETestServerWithCookies creates a test server with an HTTP cookie jar
// that automatically stores and sends cookies (including HttpOnly).
func NewE2ETestServerWithCookies(t *testing.T) *E2ETestServer {
	serverMu.Lock()
	port := 18080 + serverIndex
	serverIndex++
	serverMu.Unlock()

	s := setupE2ETestServer(t, port)
	jar, _ := cookiejar.New(nil)
	s.client.Jar = jar
	return s
}

var (
	serverIndex int
	serverMu    sync.Mutex
)

func NewE2ETestServer(t *testing.T) *E2ETestServer {
	serverMu.Lock()
	port := 18080 + serverIndex
	serverIndex++
	serverMu.Unlock()

	return setupE2ETestServer(t, port)
}

func setupE2ETestServer(t *testing.T, port int) *E2ETestServer {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/e2e_test.db"

	cfg := &config.Config{
		App: config.AppConfig{
			URL:         fmt.Sprintf("http://localhost:%d", port),
			Name:        "AuthNas E2E Test",
			Environment: "test",
		},
		Security: config.SecurityConfig{
			StorageKey:             "test-storage-key-at-least-32-characters-long-for-e2e",
			PasswordStrength:       0,
			MFARequired:            false,
			EmailVerification:      false,
			SignupRequiresApproval: false,
		},
		JWT: config.JWTConfig{
			AccessTokenExpiry:  "15m",
			RefreshTokenExpiry: "168h",
		},
		OIDC: config.OIDCConfig{
			Issuer: fmt.Sprintf("http://localhost:%d", port),
		},
		Email: config.EmailConfig{
			Enabled: false,
		},
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
		RateLimit: config.RateLimitConfig{
			Enabled:           false,
			RequestsPerMinute: 1000,
		},
	}

	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	db.Config.Logger = logger.Default.LogMode(logger.Silent)

	if err := db.AutoMigrate(
		&model.User{},
		&model.Group{},
		&model.UserGroup{},
		&model.Client{},
		&model.Consent{},
		&model.Passkey{},
		&model.PasskeyAuthOptions{},
		&model.TOTP{},
		&model.Key{},
		&model.Invitation{},
		&model.OIDCPayload{},
		&model.EmailVerification{},
		&model.PasswordReset{},
		&model.ProxyAuth{},
		&model.EmailLog{},
		&model.SystemSetting{},
	); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	r, csrfService := setupRouter(db, cfg)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.ListenAndServe()
	}()

	time.Sleep(100 * time.Millisecond)

	return &E2ETestServer{
		t:            t,
		server:       server,
		url:          fmt.Sprintf("http://localhost:%d", port),
		db:           db,
		client:       &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		cfg:          cfg,
		csrfService:  csrfService,
	}
}

func (s *E2ETestServer) Close() {
	if s.server != nil {
		s.server.Shutdown(context.Background())
	}
	if s.db != nil {
		if sqlDB, err := s.db.DB(); err == nil {
			sqlDB.Close()
		}
	}
}

func (s *E2ETestServer) Cleanup() {
	middleware.ResetEmailVerifyRateLimit()

	tables := []string{
		"email_logs",
		"password_resets",
		"email_verifications",
		"oidc_payloads",
		"consents",
		"invitations",
		"passkey_auth_options",
		"passkeys",
		"totp",
		"user_group",
		"proxy_auth",
		"clients",
		"_group",
		"keys",
		"user",
	}

	for _, table := range tables {
		s.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

func (s *E2ETestServer) DoRequest(method, path string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	var reqBody []byte
	contentType := "application/json"
	if body != nil {
		switch b := body.(type) {
		case []byte:
			reqBody = b
		case io.Reader:
			reqBody, _ = io.ReadAll(b)
		default:
			reqBody, _ = json.Marshal(body)
		}
	}
	if headers != nil {
		if ct, ok := headers["Content-Type"]; ok {
			contentType = ct
		}
	}

	req, err := http.NewRequest(method, s.url+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	for k, v := range headers {
		if k != "Content-Type" {
			req.Header.Set(k, v)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody []byte
	if resp.ContentLength >= 0 {
		respBody = make([]byte, 0, resp.ContentLength)
	}
	respBody, _ = io.ReadAll(resp.Body)

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Header:     resp.Header,
	}, nil
}

func (s *E2ETestServer) AuthenticatedRequest(method, path string, body interface{}, accessToken string) (*HTTPResponse, error) {
	headers := map[string]string{}
	if accessToken != "" {
		headers["Authorization"] = "Bearer " + accessToken
	}
	if strings.HasPrefix(path, "/api/totp") || strings.HasPrefix(path, "/api/passkey") || strings.HasPrefix(path, "/api/user") || strings.HasPrefix(path, "/api/admin") {
		csrfToken, err := s.GetCSRFToken(accessToken)
		if err != nil {
			return nil, err
		}
		headers["X-CSRF-Token"] = csrfToken
	}
	return s.DoRequest(method, path, body, headers)
}

func (s *E2ETestServer) GetCSRFToken(accessToken string) (string, error) {
	headers := map[string]string{}
	if accessToken != "" {
		headers["Authorization"] = "Bearer " + accessToken
	}
	resp, err := s.DoRequest("GET", "/api/auth/csrf", nil, headers)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get CSRF token: %d", resp.StatusCode)
	}
	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return "", err
	}
	return result.Data.Token, nil
}

func (s *E2ETestServer) AuthenticatedCSRFRequest(method, path string, body interface{}, accessToken string) (*HTTPResponse, error) {
	csrfToken, err := s.GetCSRFToken(accessToken)
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		"X-CSRF-Token": csrfToken,
	}
	if accessToken != "" {
		headers["Authorization"] = "Bearer " + accessToken
	}
	return s.DoRequest(method, path, body, headers)
}

func (s *E2ETestServer) GetCurrentUser(accessToken string) (*model.User, error) {
	resp, err := s.AuthenticatedRequest("GET", "/api/user/me", nil, accessToken)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(resp.Body))
	}

	var wrapped struct {
		Success bool       `json:"success"`
		Data    model.User `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &wrapped); err != nil {
		return nil, err
	}
	return &wrapped.Data, nil
}

func extractRefreshToken(header http.Header) string {
	cookies := header["Set-Cookie"]
	for _, cookie := range cookies {
		if strings.HasPrefix(cookie, "auth_refresh_token=") {
			parts := strings.SplitN(cookie, "=", 2)
			if len(parts) == 2 {
				value := parts[1]
				if idx := strings.Index(value, ";"); idx != -1 {
					value = value[:idx]
				}
				return value
			}
		}
	}
	return ""
}

type TestUser struct {
	ID           string
	Username     string
	Email        string
	Password     string `json:"-"`
	AccessToken  string
	RefreshToken string
}

func (s *E2ETestServer) RegisterUser(username, email, password string) (*TestUser, error) {
	resp, err := s.DoRequest("POST", "/api/auth/register", map[string]interface{}{
		"username": username,
		"email":    email,
		"password": password,
	}, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registration failed: %d, body: %s", resp.StatusCode, string(resp.Body))
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	user := &TestUser{
		Username:     username,
		Email:        email,
		Password:     password,
		AccessToken:  result.Data.AccessToken,
		RefreshToken: extractRefreshToken(resp.Header),
	}

	u, _ := s.GetCurrentUser(user.AccessToken)
	if u != nil {
		user.ID = u.ID
	}

	return user, nil
}

func (s *E2ETestServer) LoginUser(input, password string) (*TestUser, error) {
	resp, err := s.DoRequest("POST", "/api/auth/login", map[string]interface{}{
		"input":    input,
		"password": password,
	}, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed: %d, body: %s", resp.StatusCode, string(resp.Body))
	}

	var loginResp struct {
		Success     bool `json:"success"`
		MFARequired bool `json:"mfa_required"`
		Data        struct {
			AccessToken string `json:"accessToken"`
		} `json:"data,omitempty"`
	}
	if err := json.Unmarshal(resp.Body, &loginResp); err != nil {
		return nil, err
	}

	return &TestUser{
		AccessToken:  loginResp.Data.AccessToken,
		RefreshToken: extractRefreshToken(resp.Header),
		Password:     password,
	}, nil
}

func (s *E2ETestServer) CreateAdminUser(username, email, password string) (*TestUser, error) {
	userRepo := repository.NewUserRepository(s.db)
	keyRepo := repository.NewKeyRepository(s.db)
	totpRepo := repository.NewTOTPRepository(s.db)
	passkeyRepo := repository.NewPasskeyRepository(s.db)
	emailVerificationRepo := repository.NewEmailVerificationRepository(s.db)
	randomUtil := utils.NewRandom()

	authService := service.NewAuthService(s.cfg, userRepo, keyRepo, totpRepo, passkeyRepo, emailVerificationRepo, randomUtil)

	user := &model.User{
		ID:            fmt.Sprintf("admin-%s", username),
		Username:      username,
		Email:         &email,
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}

	hashedPassword, _ := authService.HashPassword(password)
	user.PasswordHash = &hashedPassword

	if err := userRepo.Create(user); err != nil {
		return nil, err
	}

	user, err := userRepo.GetByID(user.ID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("failed to get created user")
	}

	tokens, err := authService.GenerateTokenPair(user, "")
	if err != nil {
		return nil, err
	}

	return &TestUser{
		ID:           user.ID,
		Username:     username,
		Email:        email,
		Password:     password,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func setupRouter(db *gorm.DB, cfg *config.Config) (*gin.Engine, *service.CSRFService) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.Config(cfg))

	service.ResetLoginAttempts()
	middleware.ResetPasswordResetRateLimit()

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
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)

	authService := service.NewAuthService(cfg, userRepo, keyRepo, totpRepo, passkeyRepo, emailVerificationRepo, randomUtil)
	passkeyService := service.NewPasskeyService(cfg, passkeyRepo, userRepo, repository.NewPasskeyAuthOptionsRepository(db), randomUtil)
	totpService := service.NewTOTPService(cfg, totpRepo, userRepo)
	invitationService := service.NewInvitationService(cfg, invitationRepo, userRepo)
	emailService := service.NewEmailService(cfg, emailVerificationRepo, passwordResetRepo, emailLogRepo, nil)
	userService := service.NewUserService(cfg, userRepo, groupRepo, keyRepo, totpRepo, passkeyRepo, emailVerificationRepo, passwordResetRepo, consentRepo, emailService, randomUtil, timeUtil)
	userService.SetInvitationService(invitationService)
	userService.SetDB(db)
	groupService := service.NewGroupService(cfg, groupRepo, userRepo)
	clientService := service.NewClientService(cfg, clientRepo)
	proxyAuthService := service.NewProxyAuthService(cfg, proxyAuthRepo, userRepo)
	oidcService := service.NewOIDCService(cfg, db, clientRepo, consentRepo, oidcPayloadRepo, userRepo, groupRepo, keyRepo, authService, randomUtil)
	csrfService := service.NewCSRFService(randomUtil, 60)

	authMiddleware := middleware.NewAuthMiddleware(cfg, userRepo, keyRepo, authService)
	authHandler := handler.NewAuthHandler(cfg, authService, userService, totpService, passkeyService, invitationService, emailService, csrfService, authMiddleware)
	userHandler := handler.NewUserHandler(userService, totpService, passkeyService)
	passkeyHandler := handler.NewPasskeyHandler(passkeyService)
	totpHandler := handler.NewTOTPHandler(totpService, userService)
	adminHandler := handler.NewAdminHandler(userService, groupService, clientService, invitationService, proxyAuthService)
	settingService := service.NewSystemSettingService(db)
	settingService.InitializeDefaults()
	adminSettingsHandler := handler.NewAdminSettingsHandler(settingService, emailService)
	oidcHandler := handler.NewOIDCHandler(cfg, oidcService, clientService, userService, csrfService)

	routers := &router.Routers{
		AuthHandler:          authHandler,
		UserHandler:          userHandler,
		PasskeyHandler:       passkeyHandler,
		TOTPHandler:          totpHandler,
		AdminHandler:         adminHandler,
		AdminSettingsHandler: adminSettingsHandler,
		OIDCHandler:          oidcHandler,
		AuthMiddleware:       authMiddleware,
		CSRFService:          csrfService,
	}

	api := r.Group("/api")
	router.SetupAPIRoutes(api, routers)
	routers.RegisterOIDCRoutes(r)

	return r, csrfService
}

func assertResponseOK(t *testing.T, resp *HTTPResponse, expectedBody interface{}) {
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
		return
	}
	if expectedBody != nil {
		if err := json.Unmarshal(resp.Body, expectedBody); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}
	}
}

func assertResponseStatus(t *testing.T, resp *HTTPResponse, expectedStatus int) {
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d. Body: %s", expectedStatus, resp.StatusCode, string(resp.Body))
	}
}

func dumpRequest(req *http.Request) string {
	dump, _ := httputil.DumpRequest(req, true)
	return string(dump)
}

func dumpResponse(resp *HTTPResponse) string {
	return fmt.Sprintf("Status: %d\nHeaders: %v\nBody: %s", resp.StatusCode, resp.Header, string(resp.Body))
}
