package service

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/pkg/email"
	"github.com/authnas/authnas/go-server/pkg/utils"
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

func setupTestConfig() *config.Config {
	random := utils.NewRandom()
	keyBytes, _ := random.GenerateRandomBytes(32)
	storageKey := base64.StdEncoding.EncodeToString(keyBytes)
	return &config.Config{
		App: config.AppConfig{
			URL:         "http://localhost:8080",
			Name:        "AuthNas Test",
			Environment: "test",
		},
		Security: config.SecurityConfig{
			StorageKey:       storageKey,
			PasswordStrength: 3,
			MFARequired:      false,
		},
		JWT: config.JWTConfig{
			AccessTokenExpiry:  "15m",
			RefreshTokenExpiry: "168h",
		},
		OIDC: config.OIDCConfig{
			Issuer: "http://localhost:8080",
		},
	}
}

func setupTestConfigWithDomain() *config.Config {
	random := utils.NewRandom()
	keyBytes, _ := random.GenerateRandomBytes(32)
	storageKey := base64.StdEncoding.EncodeToString(keyBytes)
	return &config.Config{
		App: config.AppConfig{
			URL:         "https://example.com",
			Name:        "AuthNas Test",
			Environment: "test",
		},
		Security: config.SecurityConfig{
			StorageKey:       storageKey,
			PasswordStrength: 3,
			MFARequired:      false,
		},
		JWT: config.JWTConfig{
			AccessTokenExpiry:  "15m",
			RefreshTokenExpiry: "168h",
		},
		OIDC: config.OIDCConfig{
			Issuer: "https://example.com",
		},
	}
}

func strPtr(s string) *string {
	return &s
}

func TestAuthService_GenerateAndValidateToken(t *testing.T) {
	cfg := setupTestConfig()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	keyRepo := repository.NewKeyRepository(db)

	randomUtil := utils.NewRandom()
	authService := &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		keyRepo:  keyRepo,
		random:   randomUtil,
	}

	user := &model.User{
		ID:            "test-user-1",
		Username:      "testuser",
		Email:         strPtr("test@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, err := authService.GenerateTokenPair(user, "")
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
	if tokenPair.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}

	claims, err := authService.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}
	if claims.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, claims.Username)
	}
	if claims.TokenVersion != user.TokenVersion {
		t.Errorf("Expected token version %d, got %d", user.TokenVersion, claims.TokenVersion)
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	cfg := setupTestConfig()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	keyRepo := repository.NewKeyRepository(db)

	randomUtil := utils.NewRandom()
	authService := &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		keyRepo:  keyRepo,
		random:   randomUtil,
	}

	user := &model.User{
		ID:            "test-user-2",
		Username:      "testuser2",
		Email:         strPtr("test2@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair, err := authService.GenerateTokenPair(user, "")
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	newTokenPair, err := authService.RefreshToken(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if newTokenPair.AccessToken == "" {
		t.Error("New access token should not be empty")
	}
	if newTokenPair.RefreshToken == "" {
		t.Error("New refresh token should not be empty")
	}

	oldTokensValid, _ := keyRepo.GetByRefreshTokenHash(tokenPair.RefreshToken)
	if oldTokensValid != nil {
		t.Error("Old refresh token should have been invalidated")
	}
}

func TestAuthService_RevokeAllSessions(t *testing.T) {
	cfg := setupTestConfig()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	keyRepo := repository.NewKeyRepository(db)

	randomUtil := utils.NewRandom()
	authService := &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		keyRepo:  keyRepo,
		random:   randomUtil,
	}

	user := &model.User{
		ID:            "test-user-3",
		Username:      "testuser3",
		Email:         strPtr("test3@example.com"),
		TokenVersion:  0,
		EmailVerified: true,
		Approved:      true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tokenPair1, _ := authService.GenerateTokenPair(user, "")
	tokenPair2, _ := authService.GenerateTokenPair(user, "")

	if err := authService.RevokeAllSessions(user.ID); err != nil {
		t.Fatalf("Failed to revoke all sessions: %v", err)
	}

	updatedUser, _ := userRepo.GetByID(user.ID)
	if updatedUser.TokenVersion != 1 {
		t.Errorf("Expected token version 1, got %d", updatedUser.TokenVersion)
	}

	_, err := authService.ValidateToken(tokenPair1.AccessToken)
	if err == nil {
		t.Error("First token should be invalid after revocation")
	}

	_, err = authService.ValidateToken(tokenPair2.AccessToken)
	if err == nil {
		t.Error("Second token should be invalid after revocation")
	}

	_, err = authService.RefreshToken(tokenPair1.RefreshToken)
	if err == nil {
		t.Error("First refresh token should be invalid after revocation")
	}

	_, err = authService.RefreshToken(tokenPair2.RefreshToken)
	if err == nil {
		t.Error("Second refresh token should be invalid after revocation")
	}
}

func TestPasswordService_CheckStrength(t *testing.T) {
	cfg := setupTestConfig()
	passwordService := NewPasswordService(cfg)

	tests := []struct {
		password string
		minScore int
	}{
		{"password", 0},
		{"123456", 0},
		{"abcdef", 0},
		{"Tr0ub4dor&3", 3},
		{"correct-horse-battery-staple", 4},
	}

	for _, tt := range tests {
		result := passwordService.CheckStrength(tt.password)
		if result.Score < tt.minScore {
			t.Errorf("Password %q: expected score >= %d, got %d (strength: %s)",
				tt.password, tt.minScore, result.Score, result.Strength)
		}
	}
}

func TestPasswordService_IsStrongEnough(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			PasswordStrength: 3,
		},
	}
	passwordService := NewPasswordService(cfg)

	if passwordService.IsStrongEnough("weak") {
		t.Error("Weak password should not pass")
	}

	if !passwordService.IsStrongEnough("Tr0ub4dor&3") {
		t.Error("Strong password should pass")
	}
}

func TestTOTPService_Generate(t *testing.T) {
	cfg := setupTestConfig()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	totpRepo := repository.NewTOTPRepository(db)

	user := &model.User{
		ID:       "totp-user-1",
		Username: "totpuser",
		Email:    strPtr("totp@example.com"),
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	totpService := NewTOTPService(cfg, totpRepo, userRepo)

	result, err := totpService.Generate(user)
	if err != nil {
		t.Fatalf("Failed to generate TOTP: %v", err)
	}

	if result.Secret == "" {
		t.Error("TOTP secret should not be empty")
	}
	if result.QRCodeURI == "" {
		t.Error("TOTP QR code URI should not be empty")
	}
}

func TestPasskeyService_GenerateAuthenticationOptions(t *testing.T) {
	cfg := setupTestConfigWithDomain()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	passkeyRepo := repository.NewPasskeyRepository(db)
	authOptsRepo := repository.NewPasskeyAuthOptionsRepository(db)

	user := &model.User{
		ID:       "passkey-user-1",
		Username: "passkeyuser",
		Email:    strPtr("passkey@example.com"),
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	randomUtil := utils.NewRandom()
	passkeyService := NewPasskeyService(cfg, passkeyRepo, userRepo, authOptsRepo, randomUtil)

	authOpts, err := passkeyService.GenerateAuthenticationOptions(user.ID)
	if err != nil {
		t.Fatalf("Failed to generate authentication options: %v", err)
	}

	if authOpts.Challenge == "" {
		t.Error("Challenge should not be empty")
	}
	if authOpts.Options == nil {
		t.Error("Options should not be nil")
	}
}

func TestCleanupService_CleanupExpired(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	keyRepo := repository.NewKeyRepository(db)
	passkeyAuthOptsRepo := repository.NewPasskeyAuthOptionsRepository(db)
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)
	emailVerificationRepo := repository.NewEmailVerificationRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	emailLogRepo := repository.NewEmailLogRepository(db)

	user := &model.User{
		ID:       "cleanup-user-1",
		Username: "cleanupuser",
		Email:    strPtr("cleanup@example.com"),
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	cleanupService := NewCleanupService(
		keyRepo,
		passkeyAuthOptsRepo,
		oidcPayloadRepo,
		emailVerificationRepo,
		passwordResetRepo,
		emailLogRepo,
		nil,
	)

	expiredKey := &model.Key{
		ID:               "expired-key-1",
		UserID:           "cleanup-user-1",
		TokenVersion:     1,
		RefreshTokenHash: "expired-hash-1",
		ExpiresAt:        time.Now().Add(-1 * time.Hour),
		CreatedAt:        time.Now().Add(-2 * time.Hour),
	}
	if err := keyRepo.Create(expiredKey); err != nil {
		t.Fatalf("Failed to create expired key: %v", err)
	}

	validKey := &model.Key{
		ID:               "valid-key-1",
		UserID:           "cleanup-user-1",
		TokenVersion:     1,
		RefreshTokenHash: "valid-hash-1",
		ExpiresAt:        time.Now().Add(1 * time.Hour),
		CreatedAt:        time.Now(),
	}
	if err := keyRepo.Create(validKey); err != nil {
		t.Fatalf("Failed to create valid key: %v", err)
	}

	result, err := cleanupService.CleanupExpired()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	t.Logf("Cleanup result: %+v", result)

	keys, _ := keyRepo.GetByUserID("cleanup-user-1")
	validKeyFound := false
	for _, k := range keys {
		if k.ID == "valid-key-1" {
			validKeyFound = true
			break
		}
	}
	if !validKeyFound {
		t.Error("Valid key should still exist after cleanup")
	}
}

func TestOIDCService_Discovery(t *testing.T) {
	cfg := setupTestConfig()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	clientRepo := repository.NewClientRepository(db)
	consentRepo := repository.NewConsentRepository(db)
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	keyRepo := repository.NewKeyRepository(db)

	randomUtil := utils.NewRandom()
	authService := &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		keyRepo:  keyRepo,
		random:   randomUtil,
	}

	oidcService := NewOIDCService(
		cfg,
		db,
		clientRepo,
		consentRepo,
		oidcPayloadRepo,
		userRepo,
		groupRepo,
		keyRepo,
		authService,
		randomUtil,
	)

	discovery := oidcService.Discovery()

	if discovery.Issuer != cfg.OIDC.Issuer {
		t.Errorf("Expected issuer %s, got %s", cfg.OIDC.Issuer, discovery.Issuer)
	}
	if discovery.AuthorizationEndpoint == "" {
		t.Error("Authorization endpoint should not be empty")
	}
	if discovery.TokenEndpoint == "" {
		t.Error("Token endpoint should not be empty")
	}
	if discovery.UserInfoEndpoint == "" {
		t.Error("UserInfo endpoint should not be empty")
	}
}

func TestOIDCService_CreateAuthorizationSession(t *testing.T) {
	cfg := setupTestConfig()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	clientRepo := repository.NewClientRepository(db)
	consentRepo := repository.NewConsentRepository(db)
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	keyRepo := repository.NewKeyRepository(db)

	randomUtil := utils.NewRandom()
	authService := &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		keyRepo:  keyRepo,
		random:   randomUtil,
	}

	oidcService := NewOIDCService(
		cfg,
		db,
		clientRepo,
		consentRepo,
		oidcPayloadRepo,
		userRepo,
		groupRepo,
		keyRepo,
		authService,
		randomUtil,
	)

	client := &model.Client{
		ID:            "client-1",
		ClientID:      "test-client-id",
		Name:          "Test Client",
		RedirectURIs:  "https://example.com/callback",
		ResponseTypes: strPtr("code id_token"),
		Scopes:        strPtr("openid profile email"),
		GrantTypes:    strPtr("authorization_code"),
	}
	if err := clientRepo.Create(client); err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req := &AuthorizationRequest{
		ClientID:      "test-client-id",
		RedirectURI:   "https://example.com/callback",
		ResponseType:  "code",
		Scope:         "openid profile",
		State:         "test-state",
		Nonce:         "test-nonce",
		CodeChallenge: "test-challenge",
	}

	session, uid, err := oidcService.CreateAuthorizationSession(req, "")
	if err != nil {
		t.Fatalf("Failed to create authorization session: %v", err)
	}

	if uid == "" {
		t.Error("UID should not be empty")
	}
	if session.ClientID != req.ClientID {
		t.Errorf("Expected client ID %s, got %s", req.ClientID, session.ClientID)
	}
	if session.RedirectURI != req.RedirectURI {
		t.Errorf("Expected redirect URI %s, got %s", req.RedirectURI, session.RedirectURI)
	}
	if session.Scope != req.Scope {
		t.Errorf("Expected scope %s, got %s", req.Scope, session.Scope)
	}
}

func TestOIDCService_PKCEValidation(t *testing.T) {
	codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	expectedChallenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	actualChallenge := sha256Hash(codeVerifier)
	if actualChallenge != expectedChallenge {
		t.Errorf("PKCE hash mismatch: expected %s, got %s", expectedChallenge, actualChallenge)
	}
}

func TestOIDCService_ValidateAuthorizationRequest(t *testing.T) {
	cfg := setupTestConfig()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	clientRepo := repository.NewClientRepository(db)
	consentRepo := repository.NewConsentRepository(db)
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	keyRepo := repository.NewKeyRepository(db)

	randomUtil := utils.NewRandom()
	authService := &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		keyRepo:  keyRepo,
		random:   randomUtil,
	}

	oidcService := NewOIDCService(
		cfg,
		db,
		clientRepo,
		consentRepo,
		oidcPayloadRepo,
		userRepo,
		groupRepo,
		keyRepo,
		authService,
		randomUtil,
	)

	client := &model.Client{
		ID:            "client-2",
		ClientID:      "test-client-2",
		Name:          "Test Client 2",
		RedirectURIs:  "https://example.com/callback",
		ResponseTypes: strPtr("code id_token"),
		Scopes:        strPtr("openid profile email"),
		GrantTypes:    strPtr("authorization_code"),
	}
	if err := clientRepo.Create(client); err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	validClient, err := oidcService.ValidateAuthorizationRequest(
		"test-client-2",
		"https://example.com/callback",
		"code",
		"openid profile",
	)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
	if validClient.ClientID != "test-client-2" {
		t.Errorf("Expected client ID test-client-2, got %s", validClient.ClientID)
	}

	_, err = oidcService.ValidateAuthorizationRequest(
		"invalid-client",
		"https://example.com/callback",
		"code",
		"openid",
	)
	if err == nil {
		t.Error("Should fail for invalid client")
	}

	_, err = oidcService.ValidateAuthorizationRequest(
		"test-client-2",
		"https://wrong.com/callback",
		"code",
		"openid",
	)
	if err == nil {
		t.Error("Should fail for invalid redirect URI")
	}
}

func TestCleanupService_OldEmailLogs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	keyRepo := repository.NewKeyRepository(db)
	passkeyAuthOptsRepo := repository.NewPasskeyAuthOptionsRepository(db)
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)
	emailVerificationRepo := repository.NewEmailVerificationRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	emailLogRepo := repository.NewEmailLogRepository(db)

	cleanupService := NewCleanupService(
		keyRepo,
		passkeyAuthOptsRepo,
		oidcPayloadRepo,
		emailVerificationRepo,
		passwordResetRepo,
		emailLogRepo,
		nil,
	)

	oldEmailLog := &model.EmailLog{
		ID:        "old-email-log",
		Recipient: "test@example.com",
		Subject:   "Test",
		Template:  "test",
		Status:    "sent",
		CreatedAt: time.Now().Add(-35 * 24 * time.Hour),
	}
	if err := emailLogRepo.Create(oldEmailLog); err != nil {
		t.Fatalf("Failed to create old email log: %v", err)
	}

	newEmailLog := &model.EmailLog{
		ID:        "new-email-log",
		Recipient: "test@example.com",
		Subject:   "Test",
		Template:  "test",
		Status:    "sent",
		CreatedAt: time.Now(),
	}
	if err := emailLogRepo.Create(newEmailLog); err != nil {
		t.Fatalf("Failed to create new email log: %v", err)
	}

	deleted, err := cleanupService.CleanupOldEmailLogs(30)
	if err != nil {
		t.Fatalf("CleanupOldEmailLogs failed: %v", err)
	}

	t.Logf("Deleted %d old email logs", deleted)
}

func TestEmailService_SendVerificationEmail(t *testing.T) {
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
	emailService := NewEmailService(
		cfg,
		emailVerificationRepo,
		passwordResetRepo,
		emailLogRepo,
		emailSender,
	)

	user := &model.User{
		ID:       "email-user-1",
		Username: "emailuser",
		Email:    strPtr("email@example.com"),
	}

	err := emailService.SendVerificationEmail(user, "test-code-123")
	if err != nil {
		t.Fatalf("Failed to send verification email: %v", err)
	}
}

func TestEmailService_RenderTemplate(t *testing.T) {
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

	emailService := NewEmailService(
		cfg,
		emailVerificationRepo,
		passwordResetRepo,
		emailLogRepo,
		nil,
	)

	html, err := emailService.renderTemplate("verification", EmailTemplateData{
		AppName:  "AuthNas",
		UserName: "testuser",
		Link:     "http://localhost:8080/verify?code=123",
		Code:     "123",
	})
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if html == "" {
		t.Error("Rendered HTML should not be empty")
	}
	if !contains(html, "Verify Your Email") {
		t.Error("HTML should contain email verification title")
	}
	if !contains(html, "testuser") {
		t.Error("HTML should contain username")
	}
	if !contains(html, "http://localhost:8080/verify?code=123") {
		t.Error("HTML should contain verification link")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
