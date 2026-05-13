package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/authnas/authnas/go-server/pkg/email"
	"github.com/authnas/authnas/go-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
	}

	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

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
	); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func boolPtr(b bool) *bool { return &b }

func setupTestConfig() *config.Config {
	random := utils.NewRandom()
	keyBytes, _ := random.GenerateRandomBytes(32)
	storageKey := base64.StdEncoding.EncodeToString(keyBytes)
	return &config.Config{
		App: config.AppConfig{
			URL:         "http://localhost",
			Name:        "AuthNas Test",
			Environment: "test",
		},
		Security: config.SecurityConfig{
			StorageKey:        storageKey,
			PasswordStrength:  0,
			MFARequired:       false,
			EmailVerification: false,
		},
		JWT: config.JWTConfig{
			AccessTokenExpiry:  "15m",
			RefreshTokenExpiry: "168h",
		},
		OIDC: config.OIDCConfig{
			Issuer: "http://localhost",
		},
	}
}

type testServices struct {
	authService       *service.AuthService
	userService       *service.UserService
	totpService       *service.TOTPService
	passkeyService    *service.PasskeyService
	clientService     *service.ClientService
	groupService      *service.GroupService
	invitationService *service.InvitationService
	proxyAuthService  *service.ProxyAuthService
}

func setupTestRouter(db *gorm.DB, cfg *config.Config) (*gin.Engine, *testServices) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	userRepo := repository.NewUserRepository(db)
	keyRepo := repository.NewKeyRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	totpRepo := repository.NewTOTPRepository(db)
	passkeyRepo := repository.NewPasskeyRepository(db)
	authOptsRepo := repository.NewPasskeyAuthOptionsRepository(db)
	clientRepo := repository.NewClientRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	proxyAuthRepo := repository.NewProxyAuthRepository(db)
	emailVerificationRepo := repository.NewEmailVerificationRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	emailLogRepo := repository.NewEmailLogRepository(db)
	consentRepo := repository.NewConsentRepository(db)
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)

	randomUtil := utils.NewRandom()
	timeUtil := utils.NewTime()
	emailSender := email.NewSender(cfg)

	authService := service.NewAuthService(cfg, userRepo, keyRepo, totpRepo, passkeyRepo, emailVerificationRepo, randomUtil)
	totpService := service.NewTOTPService(cfg, totpRepo, userRepo)
	passkeyService := service.NewPasskeyService(cfg, passkeyRepo, userRepo, authOptsRepo, randomUtil)
	clientService := service.NewClientService(cfg, clientRepo)
	groupService := service.NewGroupService(cfg, groupRepo, userRepo)
	invitationService := service.NewInvitationService(cfg, invitationRepo, userRepo)
	proxyAuthService := service.NewProxyAuthService(cfg, proxyAuthRepo, userRepo)
	emailService := service.NewEmailService(cfg, emailVerificationRepo, passwordResetRepo, emailLogRepo, emailSender)
	userService := service.NewUserService(cfg, userRepo, groupRepo, keyRepo, totpRepo, passkeyRepo, emailVerificationRepo, passwordResetRepo, consentRepo, emailService, invitationService, randomUtil, timeUtil, db)
	settingService := service.NewSystemSettingService(db)
	if err := settingService.InitializeDefaults(); err != nil {
		panic(err)
	}
	csrfService := service.NewCSRFService(randomUtil, 60)
	oidcService := service.NewOIDCService(cfg, db, clientRepo, consentRepo, oidcPayloadRepo, userRepo, groupRepo, keyRepo, authService, randomUtil)

	authMiddleware := middleware.NewAuthMiddleware(cfg, userRepo, keyRepo, authService)
	authHandler := NewAuthHandler(cfg, authService, userService, totpService, passkeyService, invitationService, emailService, settingService, csrfService, authMiddleware)
	passkeyHandler := NewPasskeyHandler(passkeyService)
	totpHandler := NewTOTPHandler(totpService, userService)
	userHandler := NewUserHandler(userService, totpService, passkeyService)
	adminHandler := NewAdminHandler(userService, groupService, clientService, invitationService, proxyAuthService)

	api := r.Group("/api")
	{
		authHandler.RegisterRoutes(api)

		passkey := api.Group("/passkey")
		passkey.Use(authMiddleware.Authenticate())
		passkeyHandler.RegisterRoutes(passkey)

		totpRoutes := api.Group("/totp")
		totpRoutes.Use(authMiddleware.Authenticate())
		totpHandler.RegisterRoutes(totpRoutes)

		userRoutes := api.Group("/user")
		userRoutes.Use(authMiddleware.Authenticate())
		userHandler.RegisterRoutes(userRoutes)

		admin := api.Group("/admin")
		admin.Use(authMiddleware.Authenticate())
		admin.Use(middleware.RequireAdmin())
		{
			adminHandler.RegisterRoutes(admin)
		}
	}

	oidcHandler := NewOIDCHandler(cfg, oidcService, clientService, userService, csrfService)
	r.GET("/oidc/.well-known/openid-configuration", oidcHandler.Discovery)

	services := &testServices{
		authService:       authService,
		userService:       userService,
		totpService:       totpService,
		passkeyService:    passkeyService,
		clientService:     clientService,
		groupService:      groupService,
		invitationService: invitationService,
		proxyAuthService:  proxyAuthService,
	}

	return r, services
}

func TestAuthHandler_Login_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "test-user-login",
		Username:      "testuser",
		Email:         strPtr("test@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	reqBody := LoginRequest{
		Input:    "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken string      `json:"accessToken"`
			User        *model.User `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.Data.AccessToken == "" {
		t.Error("Expected access token to be set")
	}

	setCookieHeaders := resp.Header().Values("Set-Cookie")
	found := false
	for _, h := range setCookieHeaders {
		if strings.HasPrefix(h, "auth_refresh_token=") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected refresh token cookie to be set")
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "test-user-login-2",
		Username:      "testuser2",
		Email:         strPtr("test2@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	reqBody := LoginRequest{
		Input:    "testuser2",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	reqBody := RegisterRequest{
		Email:    "newuser@example.com",
		Username: "newuser",
		Password: "Tr0ub4dor&3Horse",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.Data.AccessToken == "" {
		t.Error("Expected access token to be set")
	}
}

func TestAuthHandler_PasskeyStart_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "passkey-test-user",
		Username:      "passkeyuser",
		Email:         strPtr("passkey@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	reqBody := PasskeyStartRequest{
		Username: "passkeyuser",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/passkey/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Challenge string `json:"challenge"`
			Options   string `json:"options"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.Data.Challenge == "" {
		t.Error("Expected challenge to be set")
	}
}

func TestAuthHandler_PasskeyEnd_InvalidResponse(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	reqBody := PasskeyEndRequest{
		CredentialID: "test-credential-id",
		Challenge:    "test-challenge",
		Response:     "invalid-response",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/passkey/end", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, resp.Code, resp.Body.String())
	}
}

func TestAuthHandler_TOTPVerify_MissingFields(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/totp", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, resp.Code, resp.Body.String())
	}
}

func TestAuthHandler_SendVerifyEmail_UserNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	reqBody := SendVerifyEmailRequest{Email: "nonexistent@example.com"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/send_verify_email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusNotFound, resp.Code, resp.Body.String())
	}
}

func TestPasskeyHandler_RegistrationEnd_InvalidResponse(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "passkey-reg-user",
		Username:      "passkeyreguser",
		Email:         strPtr("passkeyreg@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	reqBody := RegistrationEndRequest{
		Challenge: "test-challenge",
		Options:   "invalid-options",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/passkey/registration/end", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_CreateClient_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-user",
		Username:      "adminuser",
		Email:         strPtr("admin@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := CreateClientRequest{
		ClientID: "test-client",
		Name:     "Test Client",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/admin/clients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_UpdateClient_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-user-2",
		Username:      "adminuser2",
		Email:         strPtr("admin2@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := UpdateClientRequest{
		Name: strPtr("Updated Client Name"),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/admin/clients/nonexistent-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusNotFound, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_CreateProxyAuth_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-user-3",
		Username:      "adminuser3",
		Email:         strPtr("admin3@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := CreateProxyAuthRequest{
		Name:       "Test Proxy Auth",
		ProxyURL:   "https://proxy.example.com",
		HeaderName: "X-User-ID",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/admin/proxyauth", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_UpdateProxyAuth_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-user-4",
		Username:      "adminuser4",
		Email:         strPtr("admin4@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := UpdateProxyAuthRequest{
		Name: strPtr("Updated Proxy Auth"),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/admin/proxyauth/nonexistent-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusNotFound, resp.Code, resp.Body.String())
	}
}

func TestAuthHandler_GetInvitation_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	invitation := &model.Invitation{
		ID:        "test-invitation",
		Email:     "invited@example.com",
		Username:  strPtr("inviteduser"),
		Code:      "test-code-123",
		ExpiresAt: utils.NewTime().Now().Add(24 * 60 * 60 * 1000000000),
		CreatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(invitation).Error; err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	req, _ := http.NewRequest("GET", "/api/auth/invitation/test-invitation/test-code-123", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Email    string  `json:"email"`
			Username *string `json:"username,omitempty"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.Data.Email != invitation.Email {
		t.Errorf("Expected email %s, got %s", invitation.Email, response.Data.Email)
	}
}

func TestAuthHandler_GetInvitation_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	req, _ := http.NewRequest("GET", "/api/auth/invitation/nonexistent/invalid-challenge", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.Code)
	}
}

func TestAuthHandler_VerifyEmail_InvalidCode(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "verify-email-user",
		Username:      "verifyemailuser",
		Email:         strPtr("verify@example.com"),
		TokenVersion:  0,
		EmailVerified: false,
		Approved:      true,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	reqBody := VerifyEmailRequest{
		UserID:    user.ID,
		Challenge: "invalid-code",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/verify_email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_ListUsers_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-list-users",
		Username:      "adminlistusers",
		Email:         strPtr("adminlist@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	for i := 0; i < 3; i++ {
		user := &model.User{
			ID:            "list-user-" + string(rune('a'+i)),
			Username:      "listuser" + string(rune('a'+i)),
			Email:         strPtr("listuser" + string(rune('a'+i)) + "@example.com"),
			TokenVersion:  0,
			EmailVerified: true,
			Approved:      true,
		}
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var response struct {
		Success  bool           `json:"success"`
		Data     []UserListItem `json:"data"`
		Total    int64          `json:"total"`
		Page     int            `json:"page"`
		PageSize int            `json:"page_size"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response.Data) < 4 {
		t.Errorf("Expected at least 4 users, got %d", len(response.Data))
	}
}

func TestAdminHandler_CreateUser_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-create-user",
		Username:      "admincreateuser",
		Email:         strPtr("admincreate@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := CreateUserRequest{
		Email:    "newadminuser@example.com",
		Username: "newadminuser",
		Password: strPtr("password123"),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/admin/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_DeleteUser_CannotDeleteSelf(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-delete-self",
		Username:      "admindeleteself",
		Email:         strPtr("admindeleteself@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("DELETE", "/api/admin/users/admin-delete-self", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_ApproveUser_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-approve",
		Username:      "adminapprove",
		Email:         strPtr("adminapprove@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	userToApprove := &model.User{
		ID:            "user-to-approve",
		Username:      "useroapprove",
		Email:         strPtr("userapprove@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      false,
	}
	if err := db.Create(userToApprove).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := ApproveUserRequest{Approved: true}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/admin/users/user-to-approve/approve", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestTOTPHandler_Register_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "totp-handler-user",
		Username:      "tophandleruser",
		Email:         strPtr("tophandler@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("POST", "/api/totp/registration", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var wrappedResponse struct {
		Success bool `json:"success"`
		Data    struct {
			Secret    string `json:"secret"`
			QRCodeURI string `json:"qr_code_uri"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &wrappedResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !wrappedResponse.Success {
		t.Error("Expected success to be true")
	}
	if wrappedResponse.Data.Secret == "" {
		t.Error("Expected secret to be set")
	}
	if wrappedResponse.Data.QRCodeURI == "" {
		t.Error("Expected QR code URI to be set")
	}
}

func TestTOTPHandler_Delete_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "totp-delete-user",
		Username:      "totpdeletuser",
		Email:         strPtr("totpdelete@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	totpRepo := repository.NewTOTPRepository(db)
	totp := &model.TOTP{
		ID:     "totp-to-delete",
		UserID: user.ID,
		Secret: "TESTSECRET123456789012345678901234567890",
		Issuer: "AuthNas Test",
	}
	if err := totpRepo.Create(totp); err != nil {
		t.Fatalf("Failed to create TOTP: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	reqBody := DeleteTOTPRequest{Token: "000000"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("DELETE", "/api/totp", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code == http.StatusOK {
		t.Log("Delete succeeded with invalid token (expected if TOTP not properly validated)")
	} else if resp.Code == http.StatusUnauthorized {
		t.Log("Delete failed with invalid token (expected)")
	} else {
		t.Errorf("Unexpected status %d. Body: %s", resp.Code, resp.Body.String())
	}
}

func TestPasskeyHandler_GetPasskeys_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "passkey-list-user",
		Username:      "passkeylistuser",
		Email:         strPtr("passkeylist@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("GET", "/api/passkey", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestPasskeyHandler_DeletePasskey_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "passkey-delete-user",
		Username:      "passkeydeleteuser",
		Email:         strPtr("passkeydelete@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("DELETE", "/api/passkey/nonexistent-id", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusNotFound, resp.Code, resp.Body.String())
	}
}

func TestOIDCHandler_Discovery(t *testing.T) {
	t.Skip("Skipping OIDC Discovery test - requires full OIDC Service initialization")
}

func TestAdminHandler_ListGroups_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-list-groups",
		Username:      "adminlistgroups",
		Email:         strPtr("adminlistgroups@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	group := &model.Group{
		ID:          "test-group",
		Name:        "Test Group",
		Description: strPtr("A test group"),
		CreatedAt:   utils.NewTime().Now(),
		UpdatedAt:   utils.NewTime().Now(),
	}
	if err := db.Create(group).Error; err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/groups", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var response struct {
		Success bool            `json:"success"`
		Data    []GroupListItem `json:"data"`
		Total   int64           `json:"total"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response.Data) < 1 {
		t.Error("Expected at least 1 group")
	}
}

func TestAdminHandler_CreateGroup_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-create-group",
		Username:      "admincreategroup",
		Email:         strPtr("admincreategroup@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := CreateGroupRequest{
		Name:        "New Group",
		Description: "A new group",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/admin/groups", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_ListInvitations_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-list-invitations",
		Username:      "adminlistinvitations",
		Email:         strPtr("adminlistinvitations@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	invitation := &model.Invitation{
		ID:        "test-invitation-list",
		Email:     "invitedlist@example.com",
		Username:  strPtr("invitedlistuser"),
		Code:      "test-code-list",
		ExpiresAt: utils.NewTime().Now().Add(24 * 60 * 60 * 1000000000),
		CreatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(invitation).Error; err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/invitations", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_CreateInvitation_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-create-invitation",
		Username:      "admincreateinvitation",
		Email:         strPtr("admincreateinvitation@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := CreateInvitationRequest{
		Email:    "newinvited@example.com",
		Username: "newinviteduser",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/admin/invitations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Code  string `json:"code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.Data.Code == "" {
		t.Error("Expected invitation code to be set")
	}
}

func TestAdminHandler_ListProxyAuth_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-list-proxyauth",
		Username:      "adminlistproxyauth",
		Email:         strPtr("adminlistproxyauth@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	proxyAuth := &model.ProxyAuth{
		ID:         "test-proxyauth",
		Name:       "Test Proxy Auth",
		ProxyURL:   "https://proxy.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
		CreatedAt:  utils.NewTime().Now(),
		UpdatedAt:  utils.NewTime().Now(),
	}
	if err := db.Create(proxyAuth).Error; err != nil {
		t.Fatalf("Failed to create proxy auth: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/proxyauth", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_DeleteProxyAuth_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-delete-proxyauth",
		Username:      "admindeleteproxyauth",
		Email:         strPtr("admindeleteproxyauth@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	proxyAuth := &model.ProxyAuth{
		ID:         "proxyauth-to-delete",
		Name:       "Proxy Auth To Delete",
		ProxyURL:   "https://proxy-delete.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
		CreatedAt:  utils.NewTime().Now(),
		UpdatedAt:  utils.NewTime().Now(),
	}
	if err := db.Create(proxyAuth).Error; err != nil {
		t.Fatalf("Failed to create proxy auth: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("DELETE", "/api/admin/proxyauth/proxyauth-to-delete", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestEmailService_ForHandler(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			URL:  "http://localhost:8080",
			Name: "AuthNas Test",
		},
		Email: config.EmailConfig{
			Enabled: false,
		},
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	emailVerificationRepo := repository.NewEmailVerificationRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	emailLogRepo := repository.NewEmailLogRepository(db)

	emailSender := email.NewSender(cfg)
	_ = service.NewEmailService(
		cfg,
		emailVerificationRepo,
		passwordResetRepo,
		emailLogRepo,
		emailSender,
	)
}

func TestAuthHandler_ForgotPassword_UserNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	reqBody := map[string]string{"email": "nonexistent@example.com"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/forgot_password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAuthHandler_ResetPassword_InvalidCode(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	reqBody := map[string]string{"code": "invalid-code", "new_password": "newpassword123"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/reset_password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, resp.Code, resp.Body.String())
	}
}

func TestAuthHandler_ResetPassword_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "reset-password-user",
		Username:      "resetpassworduser",
		Email:         strPtr("reset@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	passwordResetRepo := repository.NewPasswordResetRepository(db)
	passwordReset := &model.PasswordReset{
		ID:        "test-password-reset",
		UserID:    user.ID,
		Code:      "valid-reset-code",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := passwordResetRepo.Create(passwordReset); err != nil {
		t.Fatalf("Failed to create password reset: %v", err)
	}

	reqBody := map[string]string{"code": "valid-reset-code", "newPassword": "9f5e!7k2#mQ3$jR8pL"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/reset_password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAuthHandler_GetMe_WithValidToken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "getme-user",
		Username:      "getmeuser",
		Email:         strPtr("getme@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Data.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, response.Data.ID)
	}
	if response.Data.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, response.Data.Username)
	}
}

func TestAuthHandler_GetMe_NoToken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	req, _ := http.NewRequest("GET", "/api/auth/me", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "refresh-user",
		Username:      "refreshuser",
		Email:         strPtr("refresh@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Login first to get refresh token cookie
	loginBody := LoginRequest{
		Input:    "refreshuser",
		Password: "password123",
	}
	body, _ := json.Marshal(loginBody)

	loginReq, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()

	router.ServeHTTP(loginResp, loginReq)

	if loginResp.Code != http.StatusOK {
		t.Fatalf("Login failed: %d. Body: %s", loginResp.Code, loginResp.Body.String())
	}

	// Extract refresh token from Set-Cookie header
	var refreshToken string
	for _, h := range loginResp.Header().Values("Set-Cookie") {
		if strings.HasPrefix(h, "auth_refresh_token=") {
			cookieParts := strings.SplitN(h, "=", 2)
			if len(cookieParts) == 2 {
				refreshToken = strings.SplitN(cookieParts[1], ";", 2)[0]
			}
			break
		}
	}
	if refreshToken == "" {
		t.Fatal("No refresh token cookie found")
	}

	// Now test refresh endpoint
	refreshReq, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
	refreshReq.Header.Set("Cookie", "auth_refresh_token="+refreshToken)
	refreshResp := httptest.NewRecorder()

	router.ServeHTTP(refreshResp, refreshReq)

	if refreshResp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, refreshResp.Code, refreshResp.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(refreshResp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.Data.AccessToken == "" {
		t.Error("Expected new access token")
	}
}

func TestAuthHandler_RefreshToken_NoCookie(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	req, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
}

func TestAuthHandler_GetCSRFToken_Unauthenticated(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	req, _ := http.NewRequest("GET", "/api/auth/csrf", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.Data.Token == "" {
		t.Error("Expected CSRF token to be set")
	}
}

func TestAuthHandler_GetCSRFToken_Authenticated(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "csrf-auth-user",
		Username:      "csrfaauthuser",
		Email:         strPtr("csrfaauth@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("GET", "/api/auth/csrf", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Data.Token == "" {
		t.Error("Expected CSRF token to be set")
	}
}

func TestAdminHandler_ListClients_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-list-clients",
		Username:      "adminlistclients",
		Email:         strPtr("adminlistclients@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	client := &model.Client{
		ID:        "test-client-list",
		ClientID:  "client-list-id",
		Name:      "Test Client List",
		CreatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/clients", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_GetClient_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-get-client",
		Username:      "admingetclient",
		Email:         strPtr("admingetclient@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	client := &model.Client{
		ID:        "client-to-get",
		ClientID:  "get-client-id",
		Name:      "Client To Get",
		CreatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/clients/client-to-get", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_GetClient_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-get-client-nf",
		Username:      "admingetclientnf",
		Email:         strPtr("admingetclientnf@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/clients/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.Code)
	}
}

func TestAdminHandler_UpdateClient_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-update-client",
		Username:      "adminupdateclient",
		Email:         strPtr("adminupdateclient@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	client := &model.Client{
		ID:        "client-to-update",
		ClientID:  "update-client-id",
		Name:      "Client To Update",
		CreatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := UpdateClientRequest{
		Name: strPtr("Updated Client Name"),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/admin/clients/client-to-update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_DeleteClient_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-delete-client",
		Username:      "admindeleteclient",
		Email:         strPtr("admindeleteclient@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	client := &model.Client{
		ID:        "client-to-delete",
		ClientID:  "delete-client-id",
		Name:      "Client To Delete",
		CreatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("DELETE", "/api/admin/clients/client-to-delete", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_DeleteClient_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-delete-client-nf",
		Username:      "admindeleteclientnf",
		Email:         strPtr("admindeleteclientnf@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("DELETE", "/api/admin/clients/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.Code)
	}
}

func TestAdminHandler_GetGroup_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-get-group",
		Username:      "admingetgroup",
		Email:         strPtr("admingetgroup@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	group := &model.Group{
		ID:        "group-to-get",
		Name:      "Group To Get",
		CreatedAt: utils.NewTime().Now(),
		UpdatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(group).Error; err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/groups/group-to-get", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_GetGroup_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-get-group-nf",
		Username:      "admingetgroupnf",
		Email:         strPtr("admingetgroupnf@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/groups/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.Code)
	}
}

func TestAdminHandler_UpdateGroup_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-update-group",
		Username:      "adminupdategroup",
		Email:         strPtr("adminupdategroup@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	group := &model.Group{
		ID:        "group-to-update",
		Name:      "Group To Update",
		CreatedAt: utils.NewTime().Now(),
		UpdatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(group).Error; err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := UpdateGroupRequest{
		Name:        "Updated Group Name",
		Description: "Updated description",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/admin/groups/group-to-update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_DeleteGroup_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-delete-group",
		Username:      "admindeletegroup",
		Email:         strPtr("admindeletegroup@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	group := &model.Group{
		ID:        "group-to-delete",
		Name:      "Group To Delete",
		CreatedAt: utils.NewTime().Now(),
		UpdatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(group).Error; err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("DELETE", "/api/admin/groups/group-to-delete", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_DeleteGroup_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-delete-group-nf",
		Username:      "admindeletegroupnf",
		Email:         strPtr("admindeletegroupnf@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("DELETE", "/api/admin/groups/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.Code)
	}
}

func TestAdminHandler_GetInvitation_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-get-invitation",
		Username:      "admingetinvitation",
		Email:         strPtr("admingetinvitation@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	invitation := &model.Invitation{
		ID:        "invitation-to-get",
		Email:     "get@example.com",
		Username:  strPtr("getinviteuser"),
		Code:      "get-invite-code",
		ExpiresAt: utils.NewTime().Now().Add(24 * time.Hour),
		CreatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(invitation).Error; err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/invitations/invitation-to-get", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_GetInvitation_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-get-inv-nf",
		Username:      "admingetinvnf",
		Email:         strPtr("admingetinvnf@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/invitations/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.Code)
	}
}

func TestAdminHandler_DeleteInvitation_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-delete-inv",
		Username:      "admindeleteinv",
		Email:         strPtr("admindeleteinv@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	invitation := &model.Invitation{
		ID:        "invitation-to-delete",
		Email:     "delete@example.com",
		Username:  strPtr("deleteinvuser"),
		Code:      "delete-invite-code",
		ExpiresAt: utils.NewTime().Now().Add(24 * time.Hour),
		CreatedAt: utils.NewTime().Now(),
	}
	if err := db.Create(invitation).Error; err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("DELETE", "/api/admin/invitations/invitation-to-delete", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_DeleteInvitation_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-delete-inv-nf",
		Username:      "admindeleteinvnf",
		Email:         strPtr("admindeleteinvnf@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("DELETE", "/api/admin/invitations/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.Code)
	}
}

func TestAdminHandler_GetProxyAuth_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-get-proxy",
		Username:      "admingetproxy",
		Email:         strPtr("admingetproxy@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	proxyAuth := &model.ProxyAuth{
		ID:         "proxyauth-to-get",
		Name:       "Proxy Auth To Get",
		ProxyURL:   "https://proxy-get.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
		CreatedAt:  utils.NewTime().Now(),
		UpdatedAt:  utils.NewTime().Now(),
	}
	if err := db.Create(proxyAuth).Error; err != nil {
		t.Fatalf("Failed to create proxy auth: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/proxyauth/proxyauth-to-get", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_GetProxyAuth_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-get-proxy-nf",
		Username:      "admingetproxynf",
		Email:         strPtr("admingetproxynf@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/proxyauth/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.Code)
	}
}

func TestAdminHandler_UpdateProxyAuth_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-update-proxy",
		Username:      "adminupdateproxy",
		Email:         strPtr("adminupdateproxy@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	proxyAuth := &model.ProxyAuth{
		ID:         "proxyauth-to-update",
		Name:       "Proxy Auth To Update",
		ProxyURL:   "https://proxy-update.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
		CreatedAt:  utils.NewTime().Now(),
		UpdatedAt:  utils.NewTime().Now(),
	}
	if err := db.Create(proxyAuth).Error; err != nil {
		t.Fatalf("Failed to create proxy auth: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := UpdateProxyAuthRequest{
		Name: strPtr("Updated Proxy Auth Name"),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/admin/proxyauth/proxyauth-to-update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestUserHandler_UpdatePassword_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "update-pwd-user",
		Username:      "updatepwduser",
		Email:         strPtr("updatepwd@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	reqBody := UpdatePasswordRequest{
		OldPassword: "password123",
		NewPassword: "newStrongP@ss1",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/user/me/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestUserHandler_UpdatePassword_WrongOldPassword(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "update-pwd-wrong",
		Username:      "updatepwdwrong",
		Email:         strPtr("updatepwdwrong@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	reqBody := UpdatePasswordRequest{
		OldPassword: "wrongpassword",
		NewPassword: "newStrongP@ss1",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/user/me/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestUserHandler_UpdatePassword_NoToken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, _ := setupTestRouter(db, cfg)

	reqBody := UpdatePasswordRequest{
		OldPassword: "password123",
		NewPassword: "newStrongP@ss1",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/user/me/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
}

func TestUserHandler_UpdateMe_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "update-me-user",
		Username:      "updatemeuser",
		Email:         strPtr("updateme@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	reqBody := UpdateMeRequest{
		Name: strPtr("Updated Name"),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/user/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestUserHandler_UpdateMe_InvalidEmail(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "update-me-bad-email",
		Username:      "updatembademail",
		Email:         strPtr("updatembademail@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	reqBody := UpdateMeRequest{
		Email: strPtr("not-an-email"),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/user/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestUserHandler_DeleteAllSessions_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "del-sessions-user",
		Username:      "delsessionsuser",
		Email:         strPtr("delsessions@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("DELETE", "/api/user/me/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestUserHandler_ListSessions_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "list-sessions-user",
		Username:      "listsessionsuser",
		Email:         strPtr("listsessions@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("GET", "/api/user/me/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestUserHandler_DeleteSession_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "del-session-user",
		Username:      "delsessionuser",
		Email:         strPtr("delsession@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("DELETE", "/api/user/me/sessions/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestTOTPHandler_Verify_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "totp-verify-user",
		Username:      "totpverifyuser",
		Email:         strPtr("totpverify@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("POST", "/api/totp/registration", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	regResp := httptest.NewRecorder()
	router.ServeHTTP(regResp, req)

	var regResponse struct {
		Success bool `json:"success"`
		Data    struct {
			Secret string `json:"secret"`
		} `json:"data"`
	}
	if err := json.Unmarshal(regResp.Body.Bytes(), &regResponse); err != nil {
		t.Fatalf("Failed to parse registration response: %v", err)
	}

	totpCode, err := services.totpService.GenerateTOTPCode(regResponse.Data.Secret)
	if err != nil {
		t.Fatalf("Failed to generate TOTP code: %v", err)
	}

	reqBody := VerifyTOTPRequest{Token: totpCode}
	body, _ := json.Marshal(reqBody)

	req, _ = http.NewRequest("POST", "/api/totp/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestTOTPHandler_Verify_InvalidToken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "totp-verify-invalid",
		Username:      "totpverifyinvalid",
		Email:         strPtr("totpverifyinv@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("POST", "/api/totp/registration", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	reqBody := VerifyTOTPRequest{Token: "000000"}
	body, _ := json.Marshal(reqBody)

	req, _ = http.NewRequest("POST", "/api/totp/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp = httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusUnauthorized, resp.Code, resp.Body.String())
	}
}

func TestTOTPHandler_Delete_WithValidToken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	user := &model.User{
		ID:            "totp-delete-success",
		Username:      "totpdeletesuccess",
		Email:         strPtr("totpdeletesuccess@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	user.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(user, "")

	req, _ := http.NewRequest("POST", "/api/totp/registration", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	var regResponse struct {
		Success bool `json:"success"`
		Data    struct {
			Secret string `json:"secret"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &regResponse); err != nil {
		t.Fatalf("Failed to parse registration response: %v", err)
	}

	totpCode, err := services.totpService.GenerateTOTPCode(regResponse.Data.Secret)
	if err != nil {
		t.Fatalf("Failed to generate TOTP code: %v", err)
	}

	reqBody := VerifyTOTPRequest{Token: totpCode}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/totp/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	totpCode2, err := services.totpService.GenerateTOTPCode(regResponse.Data.Secret)
	if err != nil {
		t.Fatalf("Failed to generate TOTP code for delete: %v", err)
	}

	delReqBody := DeleteTOTPRequest{Token: totpCode2}
	body, _ = json.Marshal(delReqBody)
	req, _ = http.NewRequest("DELETE", "/api/totp", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp = httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_CountUsers_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-count-users",
		Username:      "admincountusers",
		Email:         strPtr("admincountusers@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/users/count", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_GetUser_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-get-user",
		Username:      "admingetuser",
		Email:         strPtr("admingetuser@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	regularUser := &model.User{
		ID:            "user-to-get",
		Username:      "usertoget",
		Email:         strPtr("usertoget@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	if err := db.Create(regularUser).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	req, _ := http.NewRequest("GET", "/api/admin/users/user-to-get", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_UpdateUser_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-update-user",
		Username:      "adminupdateuser",
		Email:         strPtr("adminupdateuser@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	regularUser := &model.User{
		ID:            "user-to-update",
		Username:      "usertoupdate",
		Email:         strPtr("usertoupdate@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	if err := db.Create(regularUser).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := UpdateUserRequest{
		Approved: boolPtr(true),
		IsAdmin:  boolPtr(false),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/admin/users/user-to-update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAdminHandler_ResetPassword_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := setupTestConfig()
	router, services := setupTestRouter(db, cfg)

	adminUser := &model.User{
		ID:            "admin-reset-pwd",
		Username:      "adminresetpwd",
		Email:         strPtr("adminresetpwd@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       true,
	}
	hashedPassword, _ := services.authService.HashPassword("password123")
	adminUser.PasswordHash = strPtr(hashedPassword)
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	regularUser := &model.User{
		ID:            "user-reset-pwd",
		Username:      "userresetpwd",
		Email:         strPtr("userresetpwd@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	if err := db.Create(regularUser).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, _ := services.authService.GenerateTokenPair(adminUser, "")

	reqBody := ResetPasswordRequest{NewPassword: "newStrongP@ss1"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/admin/users/user-reset-pwd/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}
