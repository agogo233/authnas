package service

import (
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"gorm.io/gorm"
)

func setupConsentServiceTestDB(t *testing.T) (*gorm.DB, func()) {
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

	if err := db.AutoMigrate(&model.User{}, &model.Client{}, &model.Consent{}); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func newTestConsentService(db *gorm.DB) *ConsentService {
	cfg := &config.Config{}
	userRepo := repository.NewUserRepository(db)
	clientRepo := repository.NewClientRepository(db)
	consentRepo := repository.NewConsentRepository(db)
	return NewConsentService(cfg, consentRepo, clientRepo, userRepo)
}

func createTestUserForConsent(t *testing.T, db *gorm.DB, username string) *model.User {
	email := username + "@example.com"
	user := &model.User{
		ID:       "user-" + username,
		Username: username,
		Email:    &email,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	return user
}

func createTestClientForConsent(t *testing.T, db *gorm.DB, clientID string) *model.Client {
	client := &model.Client{
		ID:           "client-" + clientID,
		ClientID:     clientID,
		Name:         "Test Client " + clientID,
		RedirectURIs: "https://example.com/callback",
	}
	if err := db.Create(client).Error; err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

func TestConsentService_Create(t *testing.T) {
	db, cleanup := setupConsentServiceTestDB(t)
	defer cleanup()

	user := createTestUserForConsent(t, db, "testuser")
	client := createTestClientForConsent(t, db, "testclient")

	svc := newTestConsentService(db)

	expiresAt := time.Now().Add(time.Hour)
	consent := &model.Consent{
		ID:        "consent-1",
		UserID:    user.ID,
		ClientID:  client.ID,
		Scopes:    "openid profile",
		ExpiresAt: &expiresAt,
	}

	err := svc.Create(consent)
	if err != nil {
		t.Fatalf("Failed to create consent: %v", err)
	}

	if consent.ID == "" {
		t.Error("Consent ID should be set")
	}
}

func TestConsentService_GetByID(t *testing.T) {
	db, cleanup := setupConsentServiceTestDB(t)
	defer cleanup()

	user := createTestUserForConsent(t, db, "testuser")
	client := createTestClientForConsent(t, db, "testclient")

	svc := newTestConsentService(db)

	expiresAt := time.Now().Add(time.Hour)
	consent := &model.Consent{
		ID:        "consent-1",
		UserID:    user.ID,
		ClientID:  client.ID,
		Scopes:    "openid profile",
		ExpiresAt: &expiresAt,
	}
	svc.Create(consent)

	found, err := svc.GetByID(consent.ID)
	if err != nil {
		t.Fatalf("Failed to get consent by ID: %v", err)
	}

	if found.ID != consent.ID {
		t.Errorf("Expected ID '%s', got '%s'", consent.ID, found.ID)
	}
}

func TestConsentService_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupConsentServiceTestDB(t)
	defer cleanup()

	svc := newTestConsentService(db)

	_, err := svc.GetByID("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent consent")
	}
}

func TestConsentService_GetByUserAndClient(t *testing.T) {
	db, cleanup := setupConsentServiceTestDB(t)
	defer cleanup()

	user := createTestUserForConsent(t, db, "testuser")
	client := createTestClientForConsent(t, db, "testclient")

	svc := newTestConsentService(db)

	expiresAt := time.Now().Add(time.Hour)
	consent := &model.Consent{
		ID:        "consent-1",
		UserID:    user.ID,
		ClientID:  client.ID,
		Scopes:    "openid profile",
		ExpiresAt: &expiresAt,
	}
	svc.Create(consent)

	found, err := svc.GetByUserAndClient(user.ID, client.ID)
	if err != nil {
		t.Fatalf("Failed to get consent by user and client: %v", err)
	}

	if found.UserID != user.ID {
		t.Errorf("Expected UserID '%s', got '%s'", user.ID, found.UserID)
	}
	if found.ClientID != client.ID {
		t.Errorf("Expected ClientID '%s', got '%s'", client.ID, found.ClientID)
	}
}

func TestConsentService_GetByUserAndClient_NotFound(t *testing.T) {
	db, cleanup := setupConsentServiceTestDB(t)
	defer cleanup()

	user := createTestUserForConsent(t, db, "testuser")
	client := createTestClientForConsent(t, db, "testclient")

	svc := newTestConsentService(db)

	_, err := svc.GetByUserAndClient(user.ID, client.ID)
	if err == nil {
		t.Error("Expected error when consent not found")
	}
}

func TestConsentService_Update(t *testing.T) {
	db, cleanup := setupConsentServiceTestDB(t)
	defer cleanup()

	user := createTestUserForConsent(t, db, "testuser")
	client := createTestClientForConsent(t, db, "testclient")

	svc := newTestConsentService(db)

	expiresAt := time.Now().Add(time.Hour)
	consent := &model.Consent{
		ID:        "consent-1",
		UserID:    user.ID,
		ClientID:  client.ID,
		Scopes:    "openid",
		ExpiresAt: &expiresAt,
	}
	svc.Create(consent)

	consent.Scopes = "openid profile email"
	err := svc.Update(consent)
	if err != nil {
		t.Fatalf("Failed to update consent: %v", err)
	}

	updated, _ := svc.GetByID(consent.ID)
	if updated.Scopes != "openid profile email" {
		t.Errorf("Expected scope 'openid profile email', got '%s'", updated.Scopes)
	}
}

func TestConsentService_Delete(t *testing.T) {
	db, cleanup := setupConsentServiceTestDB(t)
	defer cleanup()

	user := createTestUserForConsent(t, db, "testuser")
	client := createTestClientForConsent(t, db, "testclient")

	svc := newTestConsentService(db)

	expiresAt := time.Now().Add(time.Hour)
	consent := &model.Consent{
		ID:        "consent-1",
		UserID:    user.ID,
		ClientID:  client.ID,
		Scopes:    "openid profile",
		ExpiresAt: &expiresAt,
	}
	svc.Create(consent)

	err := svc.Delete(consent.ID)
	if err != nil {
		t.Fatalf("Failed to delete consent: %v", err)
	}

	_, err = svc.GetByID(consent.ID)
	if err == nil {
		t.Error("Expected error after deleting consent")
	}
}
