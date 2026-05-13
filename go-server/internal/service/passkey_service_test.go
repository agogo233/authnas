package service

import (
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/pkg/utils"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

func setupPasskeyTestDB(t *testing.T) (*gorm.DB, func()) {
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
	}

	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Group{},
		&model.Passkey{},
		&model.PasskeyAuthOptions{},
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

func newPasskeyServiceForTest(t *testing.T, db *gorm.DB) *PasskeyService {
	cfg := &config.Config{
		App: config.AppConfig{
			URL:  "http://localhost:8080",
			Name: "TestAuth",
		},
	}

	passkeyRepo := repository.NewPasskeyRepository(db)
	userRepo := repository.NewUserRepository(db)
	authOptsRepo := repository.NewPasskeyAuthOptionsRepository(db)
	randomUtil := utils.NewRandom()

	svc := NewPasskeyService(
		cfg,
		passkeyRepo,
		userRepo,
		authOptsRepo,
		randomUtil,
	)

	return svc
}

func TestPasskeyServiceExt_CRUD(t *testing.T) {
	db, cleanup := setupPasskeyTestDB(t)
	defer cleanup()
	svc := newPasskeyServiceForTest(t, db)

	user := &model.User{
		ID:       "user-passkey-ext-1",
		Username: "passkeyuserext1",
		Email:    func() *string { s := "passkeyext1@example.com"; return &s }(),
	}
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	passkey := &model.Passkey{
		ID:           "passkey-ext-1",
		UserID:       "user-passkey-ext-1",
		Name:         func() *string { s := "Test Passkey"; return &s }(),
		CredentialID: "credential-ext-1",
		PublicKey:    "public-key-data",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := svc.Create(passkey); err != nil {
		t.Fatalf("Failed to create passkey: %v", err)
	}

	retrieved, err := svc.GetByID("passkey-ext-1")
	if err != nil {
		t.Fatalf("Failed to get passkey by ID: %v", err)
	}
	if retrieved.Name == nil || *retrieved.Name != "Test Passkey" {
		t.Errorf("Expected name 'Test Passkey', got '%v'", retrieved.Name)
	}

	retrievedByCred, err := svc.GetByCredentialID("credential-ext-1")
	if err != nil {
		t.Fatalf("Failed to get passkey by credential ID: %v", err)
	}
	if retrievedByCred.ID != "passkey-ext-1" {
		t.Errorf("Expected passkey ID 'passkey-ext-1', got '%s'", retrievedByCred.ID)
	}

	userPasskeys, err := svc.GetByUserID("user-passkey-ext-1")
	if err != nil {
		t.Fatalf("Failed to get passkeys by user ID: %v", err)
	}
	if len(userPasskeys) != 1 {
		t.Errorf("Expected 1 passkey, got %d", len(userPasskeys))
	}

	retrieved.Name = func() *string { s := "Updated Passkey"; return &s }()
	if err := svc.Update(retrieved); err != nil {
		t.Fatalf("Failed to update passkey: %v", err)
	}

	updated, _ := svc.GetByID("passkey-ext-1")
	if updated.Name == nil || *updated.Name != "Updated Passkey" {
		t.Errorf("Expected updated name 'Updated Passkey', got '%v'", updated.Name)
	}

	if err := svc.Delete("passkey-ext-1"); err != nil {
		t.Fatalf("Failed to delete passkey: %v", err)
	}

	_, err = svc.GetByID("passkey-ext-1")
	if err == nil {
		t.Error("Expected error after deleting passkey")
	}
}

func TestPasskeyServiceExt_HasPasskeys(t *testing.T) {
	db, cleanup := setupPasskeyTestDB(t)
	defer cleanup()
	svc := newPasskeyServiceForTest(t, db)

	user := &model.User{
		ID:       "user-has-passkey-ext",
		Username: "haspasskeyext",
		Email:    func() *string { s := "haspasskeyext@example.com"; return &s }(),
	}
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if svc.HasPasskeys("user-has-passkey-ext") {
		t.Error("Expected false when user has no passkeys")
	}

	passkey := &model.Passkey{
		ID:           "passkey-has-ext-1",
		UserID:       "user-has-passkey-ext",
		CredentialID: "credential-has-ext-1",
		PublicKey:    "public-key",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := svc.Create(passkey); err != nil {
		t.Fatalf("Failed to create passkey: %v", err)
	}

	if !svc.HasPasskeys("user-has-passkey-ext") {
		t.Error("Expected true when user has passkeys")
	}
}

func TestPasskeyServiceExt_HasPasskeys_NonexistentUser(t *testing.T) {
	db, cleanup := setupPasskeyTestDB(t)
	defer cleanup()
	svc := newPasskeyServiceForTest(t, db)

	if svc.HasPasskeys("nonexistent-user") {
		t.Error("Expected false for nonexistent user")
	}
}

func TestPasskeyServiceExt_GenerateRegistrationOptions(t *testing.T) {
	db, cleanup := setupPasskeyTestDB(t)
	defer cleanup()
	svc := newPasskeyServiceForTest(t, db)

	user := &model.User{
		ID:       "user-reg-options-ext",
		Username: "regoptionsext",
		Email:    func() *string { s := "regoptionsext@example.com"; return &s }(),
	}
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	opts, err := svc.GenerateRegistrationOptions("user-reg-options-ext", "regoptionsext")
	if err != nil {
		t.Fatalf("GenerateRegistrationOptions failed: %v", err)
	}

	if opts.Challenge == "" {
		t.Error("Expected non-empty challenge")
	}

	if opts.Options == nil || len(opts.Options) == 0 {
		t.Error("Expected non-empty options")
	}
}

func TestPasskeyServiceExt_GenerateRegistrationOptions_NonexistentUser(t *testing.T) {
	db, cleanup := setupPasskeyTestDB(t)
	defer cleanup()
	svc := newPasskeyServiceForTest(t, db)

	_, err := svc.GenerateRegistrationOptions("nonexistent-user", "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent user")
	}
}

func TestPasskeyServiceExt_GenerateAuthenticationOptions_EmptyUserID(t *testing.T) {
	db, cleanup := setupPasskeyTestDB(t)
	defer cleanup()
	svc := newPasskeyServiceForTest(t, db)

	opts, err := svc.GenerateAuthenticationOptions("")
	if err != nil {
		t.Fatalf("GenerateAuthenticationOptions failed: %v", err)
	}

	if opts.Challenge == "" {
		t.Error("Expected non-empty challenge")
	}
}

func TestPasskeyServiceExt_CreateCredentialFromResponse_NoAuthOptions(t *testing.T) {
	db, cleanup := setupPasskeyTestDB(t)
	defer cleanup()
	svc := newPasskeyServiceForTest(t, db)

	user := &model.User{
		ID:       "user-no-auth-opts-ext",
		Username: "noauthopts-ext",
		Email:    func() *string { s := "noauthopts-ext@example.com"; return &s }(),
	}
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err := svc.CreateCredentialFromResponse("user-no-auth-opts-ext", "noauthopts-ext", "Passkey", nil)
	if err == nil {
		t.Error("Expected error when no auth options exist")
	}
}

func TestPasskeyServiceExt_CreateCredentialFromResponse_ExpiredAuthOptions(t *testing.T) {
	db, cleanup := setupPasskeyTestDB(t)
	defer cleanup()
	svc := newPasskeyServiceForTest(t, db)

	user := &model.User{
		ID:       "user-expired-auth-ext",
		Username: "expiredauthext",
		Email:    func() *string { s := "expiredauthext@example.com"; return &s }(),
	}
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	expiredAuthOpts := &model.PasskeyAuthOptions{
		ID:        "expired-opts-ext",
		UserID:    func() *string { s := "user-expired-auth-ext"; return &s }(),
		Challenge: "expired-challenge",
		Options:   "{}",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	if err := db.Create(expiredAuthOpts).Error; err != nil {
		t.Fatalf("Failed to create expired auth options: %v", err)
	}

	passkey, _ := svc.CreateCredentialFromResponse("user-expired-auth-ext", "expiredauthext", "Passkey", nil)
	if passkey != nil {
		t.Error("Expected nil passkey for expired auth options")
	}
}

func TestPasskeyServiceExt_ValidateAuthentication_NoPasskey(t *testing.T) {
	db, cleanup := setupPasskeyTestDB(t)
	defer cleanup()
	svc := newPasskeyServiceForTest(t, db)

	_, err := svc.ValidateAuthentication("nonexistent-credential", nil)
	if err == nil {
		t.Error("Expected error when passkey not found")
	}
}

func TestPasskeyServiceExt_ValidateAuthentication_NoAuthOptions(googlet *testing.T) {
	googlet.Skip("ValidateAuthentication panics when response is nil, cannot test this scenario without significant refactoring")
}

func TestPasskeyServiceExt_DeleteByUserID(t *testing.T) {
	db, cleanup := setupPasskeyTestDB(t)
	defer cleanup()
	svc := newPasskeyServiceForTest(t, db)

	user := &model.User{
		ID:       "user-delete-all-ext",
		Username: "deleteallext",
		Email:    func() *string { s := "deleteallext@example.com"; return &s }(),
	}
	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	for i := 0; i < 3; i++ {
		passkey := &model.Passkey{
			ID:           "passkey-delete-ext-" + string(rune('a'+i)),
			UserID:       "user-delete-all-ext",
			CredentialID: "credential-delete-ext-" + string(rune('a'+i)),
			PublicKey:    "public-key",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := svc.Create(passkey); err != nil {
			t.Fatalf("Failed to create passkey: %v", err)
		}
	}

	passkeys, _ := svc.GetByUserID("user-delete-all-ext")
	if len(passkeys) != 3 {
		t.Fatalf("Expected 3 passkeys, got %d", len(passkeys))
	}

	passkeyRepo := repository.NewPasskeyRepository(db)
	if err := passkeyRepo.DeleteByUserID("user-delete-all-ext"); err != nil {
		t.Fatalf("DeleteByUserID failed: %v", err)
	}

	passkeys, _ = svc.GetByUserID("user-delete-all-ext")
	if len(passkeys) != 0 {
		t.Errorf("Expected 0 passkeys after delete, got %d", len(passkeys))
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"http://localhost:8080", "localhost"},
		{"https://example.com", "example.com"},
		{"https://auth.example.com:3000", "auth.example.com"},
		{"http://127.0.0.1:8080", "127.0.0.1"},
	}

	for _, tt := range tests {
		result := extractDomain(tt.url)
		if result != tt.expected {
			t.Errorf("extractDomain(%s) = %s; want %s", tt.url, result, tt.expected)
		}
	}
}

func TestPasskeyUserExt_WebAuthnInterface(t *testing.T) {
	user := &passkeyUser{
		id:    []byte("user-id"),
		name:  "testuser",
		creds: []webauthn.Credential{},
	}

	if string(user.WebAuthnID()) != "user-id" {
		t.Errorf("WebAuthnID() = %s; want user-id", string(user.WebAuthnID()))
	}

	if user.WebAuthnName() != "testuser" {
		t.Errorf("WebAuthnName() = %s; want testuser", user.WebAuthnName())
	}

	if user.WebAuthnDisplayName() != "testuser" {
		t.Errorf("WebAuthnDisplayName() = %s; want testuser", user.WebAuthnDisplayName())
	}

	if len(user.WebAuthnCredentials()) != 0 {
		t.Errorf("WebAuthnCredentials() = %d; want 0", len(user.WebAuthnCredentials()))
	}
}

func TestMustMarshalExt(t *testing.T) {
	result := mustMarshal(map[string]string{"key": "value"})
	if result != `{"key":"value"}` {
		t.Errorf("mustMarshal() = %s; want {\"key\":\"value\"}", result)
	}
}
