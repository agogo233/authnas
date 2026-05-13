package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/pkg/utils"
	"gorm.io/gorm"
)

func setupOIDCTestDB(t *testing.T) (*gorm.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
		App: config.AppConfig{
			URL:  "http://localhost:8080",
			Name: "TestAuth",
		},
		OIDC: config.OIDCConfig{
			Issuer: "http://localhost:8080",
		},
		JWT: config.JWTConfig{
			AccessTokenExpiry:  "1h",
			RefreshTokenExpiry: "168h",
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
		&model.OIDCPayload{},
		&model.Key{},
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

func newOIDCServiceForTest(t *testing.T, db *gorm.DB) *OIDCService {
	cfg := &config.Config{
		App: config.AppConfig{
			URL:  "http://localhost:8080",
			Name: "TestAuth",
		},
		OIDC: config.OIDCConfig{
			Issuer: "http://localhost:8080",
		},
		JWT: config.JWTConfig{
			AccessTokenExpiry:  "1h",
			RefreshTokenExpiry: "168h",
		},
	}

	clientRepo := repository.NewClientRepository(db)
	consentRepo := repository.NewConsentRepository(db)
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	keyRepo := repository.NewKeyRepository(db)

	randomUtil := utils.NewRandom()

	svc := NewOIDCService(
		cfg,
		db,
		clientRepo,
		consentRepo,
		oidcPayloadRepo,
		userRepo,
		groupRepo,
		keyRepo,
		nil,
		randomUtil,
	)

	return svc
}

func TestOIDCServiceExt_ValidateAuthorizationRequest_OpenIDScopeAlwaysAllowed(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	client := &model.Client{
		ID:           "client-ext-1",
		ClientID:     "test-client-ext",
		Name:         "Test Client",
		RedirectURIs: "https://example.com/callback",
		Scopes:       func() *string { s := "openid profile"; return &s }(),
	}
	clientRepo := repository.NewClientRepository(db)
	if err := clientRepo.Create(client); err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err := svc.ValidateAuthorizationRequest("test-client-ext", "https://example.com/callback", "code", "openid custom_scope")
	if err == nil {
		t.Error("custom_scope should not be allowed when client only has openid profile")
	}

	_, err = svc.ValidateAuthorizationRequest("test-client-ext", "https://example.com/callback", "code", "openid profile")
	if err != nil {
		t.Fatalf("openid profile should be allowed, got error: %v", err)
	}
}

func TestOIDCServiceExt_GetAuthorizationSession_Expired(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	payload := &model.OIDCPayload{
		ID:        "payload-expired",
		UID:       "expired-uid",
		Payload:   `{"client_id":"test","user_id":"user","redirect_uri":"https://example.com","scope":"openid","state":"state","nonce":"nonce","code_challenge":"","auth_time":123}`,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	if err := db.Create(payload).Error; err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	_, err := svc.GetAuthorizationSession("expired-uid")
	if err == nil {
		t.Error("Expected error for expired session")
	}
}

func TestOIDCServiceExt_DeleteAuthorizationSession(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	req := &AuthorizationRequest{
		ClientID:    "test-client",
		RedirectURI: "https://example.com/callback",
		Scope:       "openid",
	}

	_, uid, err := svc.CreateAuthorizationSession(req, "user-123")
	if err != nil {
		t.Fatalf("CreateAuthorizationSession failed: %v", err)
	}

	if err := svc.DeleteAuthorizationSession(uid); err != nil {
		t.Fatalf("DeleteAuthorizationSession failed: %v", err)
	}

	_, err = svc.GetAuthorizationSession(uid)
	if err == nil {
		t.Error("Expected error after deleting session")
	}
}

func TestOIDCServiceExt_CreateAuthorizationCode(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	session := &AuthorizationSession{
		ClientID:    "test-client",
		UserID:      "user-123",
		RedirectURI: "https://example.com/callback",
		Scope:       "openid profile",
		State:       "state",
		Nonce:       "nonce",
		AuthTime:    time.Now().Unix(),
	}

	code, err := svc.CreateAuthorizationCode(session)
	if err != nil {
		t.Fatalf("CreateAuthorizationCode failed: %v", err)
	}

	if code == "" {
		t.Error("Expected non-empty authorization code")
	}
}

func TestOIDCServiceExt_GetPublicKey(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	pubKey := svc.GetPublicKey()

	if pubKey == nil {
		t.Fatal("Expected non-nil public key")
	}

	if pubKey != nil && pubKey.Size() == 0 {
		t.Error("Expected RSA public key with non-zero size")
	}
}

func TestOIDCServiceExt_BuildRedirectURL(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	params := map[string]string{
		"code":  "abc123",
		"state": "xyz789",
	}

	url := svc.BuildRedirectURL("https://example.com/callback", params)

	if url == "" {
		t.Fatal("Expected non-empty URL")
	}

	if url != "https://example.com/callback?code=abc123&state=xyz789" && url != "https://example.com/callback?state=xyz789&code=abc123" {
		t.Errorf("Unexpected redirect URL: %s", url)
	}
}

func TestOIDCServiceExt_HasValidConsent_Expired(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	user := &model.User{
		ID:       "user-expired",
		Username: "expireduser",
		Email:    func() *string { s := "expired@example.com"; return &s }(),
	}
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	expiredTime := time.Now().Add(-1 * time.Hour)
	consent := &model.Consent{
		ID:        "consent-expired",
		UserID:    "user-expired",
		ClientID:  "client-expired",
		Scopes:    "openid",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: &expiredTime,
	}
	if err := db.Create(consent).Error; err != nil {
		t.Fatalf("Failed to create consent: %v", err)
	}

	if svc.HasValidConsent("user-expired", "client-expired", "openid") {
		t.Error("Expected false when consent is expired")
	}
}

func TestOIDCServiceExt_SaveConsent(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	user := &model.User{
		ID:       "user-new",
		Username: "newuser",
		Email:    func() *string { s := "new@example.com"; return &s }(),
	}
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if err := svc.SaveConsent("user-new", "client-new", "openid profile"); err != nil {
		t.Fatalf("SaveConsent failed: %v", err)
	}

	if !svc.HasValidConsent("user-new", "client-new", "openid profile") {
		t.Error("Expected consent to exist after saving")
	}

	if err := svc.SaveConsent("user-new", "client-new", "openid profile email"); err != nil {
		t.Fatalf("SaveConsent update failed: %v", err)
	}

	consentRepo := repository.NewConsentRepository(db)
	consent, _ := consentRepo.GetByUserAndClient("user-new", "client-new")
	if consent.Scopes != "openid profile email" {
		t.Errorf("Expected updated scopes, got '%s'", consent.Scopes)
	}
}

func TestOIDCServiceExt_RevokeToken(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	if err := svc.RevokeToken("nonexistent-token"); err != nil {
		t.Fatalf("RevokeToken should not fail for nonexistent token: %v", err)
	}
}

func TestOIDCServiceExt_RevokeTokensByClientID(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	if err := svc.RevokeTokensByClientID("nonexistent-client"); err != nil {
		t.Fatalf("RevokeTokensByClientID failed: %v", err)
	}
}

func TestOIDCServiceExt_ValidatePostLogoutRedirectURI(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	if err := svc.ValidatePostLogoutRedirectURI("", ""); err != nil {
		t.Error("Empty post_logout_uri should be valid")
	}

	if err := svc.ValidatePostLogoutRedirectURI("javascript:alert(1)", ""); err == nil {
		t.Error("javascript: protocol should not be allowed")
	}

	if err := svc.ValidatePostLogoutRedirectURI("ftp://example.com", ""); err == nil {
		t.Error("Only http/https should be allowed")
	}

	if err := svc.ValidatePostLogoutRedirectURI("https://example.com/logout", "nonexistent-client"); err == nil {
		t.Error("Should fail for nonexistent client")
	}
}

func TestOIDCServiceExt_ValidatePostLogoutRedirectURI_WithClient(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	postLogoutURI := "https://example.com/logout"
	client := &model.Client{
		ID:                     "client-postlogout",
		ClientID:               "postlogout-client",
		Name:                   "PostLogout Client",
		RedirectURIs:           "https://example.com/callback",
		PostLogoutRedirectURIs: &postLogoutURI,
	}
	clientRepo := repository.NewClientRepository(db)
	if err := clientRepo.Create(client); err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if err := svc.ValidatePostLogoutRedirectURI("https://example.com/logout", "postlogout-client"); err != nil {
		t.Errorf("Valid post_logout_redirect_uri should pass: %v", err)
	}

	if err := svc.ValidatePostLogoutRedirectURI("https://evil.com/logout", "postlogout-client"); err == nil {
		t.Error("Invalid post_logout_redirect_uri should fail")
	}
}

func TestOIDCServiceExt_ValidateBackChannelLogoutToken(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	if err := svc.ValidateBackChannelLogoutToken("invalid-token"); err == nil {
		t.Error("Invalid token should fail validation")
	}
}

func TestOIDCServiceExt_Stop(t *testing.T) {
	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()
	svc := newOIDCServiceForTest(t, db)

	svc.Stop()
}

func TestAuthorizationSessionExt_Payload(t *testing.T) {
	session := &AuthorizationSession{
		ClientID:      "test-client",
		UserID:        "user-123",
		RedirectURI:   "https://example.com",
		Scope:         "openid",
		State:         "state",
		Nonce:         "nonce",
		CodeChallenge: "challenge",
		AuthTime:      1234567890,
	}

	payload := session.Payload()

	if payload == "" {
		t.Fatal("Expected non-empty payload")
	}

	var parsed AuthorizationSession
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		t.Fatalf("Failed to parse payload: %v", err)
	}

	if parsed.ClientID != session.ClientID {
		t.Errorf("Expected ClientID '%s', got '%s'", session.ClientID, parsed.ClientID)
	}

	if parsed.UserID != session.UserID {
		t.Errorf("Expected UserID '%s', got '%s'", session.UserID, parsed.UserID)
	}
}

func TestOIDCServiceExt_HasValidConsent_NilConsentRepo(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			URL:  "http://localhost:8080",
			Name: "TestAuth",
		},
		OIDC: config.OIDCConfig{
			Issuer: "http://localhost:8080",
		},
		JWT: config.JWTConfig{
			AccessTokenExpiry:  "1h",
			RefreshTokenExpiry: "168h",
		},
	}

	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()

	clientRepo := repository.NewClientRepository(db)
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	keyRepo := repository.NewKeyRepository(db)

	randomUtil := utils.NewRandom()

	svc := &OIDCService{
		cfg:             cfg,
		db:              db,
		clientRepo:      clientRepo,
		consentRepo:     nil,
		oidcPayloadRepo: oidcPayloadRepo,
		userRepo:        userRepo,
		groupRepo:       groupRepo,
		keyRepo:         keyRepo,
		authService:     nil,
		random:          randomUtil,
		stopCleanup:     make(chan struct{}),
	}

	if svc.HasValidConsent("user-1", "client-1", "openid") {
		t.Error("Expected false when consentRepo is nil")
	}
}

func TestOIDCServiceExt_SaveConsent_NilConsentRepo(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			URL:  "http://localhost:8080",
			Name: "TestAuth",
		},
		OIDC: config.OIDCConfig{
			Issuer: "http://localhost:8080",
		},
		JWT: config.JWTConfig{
			AccessTokenExpiry:  "1h",
			RefreshTokenExpiry: "168h",
		},
	}

	db, cleanup := setupOIDCTestDB(t)
	defer cleanup()

	clientRepo := repository.NewClientRepository(db)
	oidcPayloadRepo := repository.NewOIDCPayloadRepository(db)
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	keyRepo := repository.NewKeyRepository(db)

	randomUtil := utils.NewRandom()

	svc := &OIDCService{
		cfg:             cfg,
		db:              db,
		clientRepo:      clientRepo,
		consentRepo:     nil,
		oidcPayloadRepo: oidcPayloadRepo,
		userRepo:        userRepo,
		groupRepo:       groupRepo,
		keyRepo:         keyRepo,
		authService:     nil,
		random:          randomUtil,
		stopCleanup:     make(chan struct{}),
	}

	if err := svc.SaveConsent("user-1", "client-1", "openid"); err != nil {
		t.Errorf("SaveConsent should not fail when consentRepo is nil: %v", err)
	}
}
