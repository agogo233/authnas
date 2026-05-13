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

func setupInvitationServiceTestDB(t *testing.T) (*gorm.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Invitation{}); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	creatorEmail := "creator@example.com"
	creator := &model.User{
		ID:       "test-creator-id",
		Username: "creator",
		Email:    &creatorEmail,
	}
	if err := db.Create(creator).Error; err != nil {
		t.Fatalf("Failed to create creator user: %v", err)
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func newTestInvitationService(db *gorm.DB) *InvitationService {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}
	invitationRepo := repository.NewInvitationRepository(db)
	userRepo := repository.NewUserRepository(db)
	return NewInvitationService(cfg, invitationRepo, userRepo)
}

func TestInvitationService_ValidateInvitation_NotFound(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	result, err := svc.ValidateInvitation("non-existent", "any-code")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for non-existent invitation")
	}

	if result.ErrorMessage != "invitation not found" {
		t.Errorf("Expected 'invitation not found', got '%s'", result.ErrorMessage)
	}
}

func TestInvitationService_ValidateInvitation_InvalidCode(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	username := "testuser"
	invitation := &model.Invitation{
		ID:        "invitation-1",
		Email:     "test@example.com",
		Username:  &username,
		Code:      "correct-code",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := db.Create(invitation).Error; err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	result, err := svc.ValidateInvitation(invitation.ID, "wrong-code")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for wrong code")
	}

	if result.ErrorMessage != "invalid invitation code" {
		t.Errorf("Expected 'invalid invitation code', got '%s'", result.ErrorMessage)
	}
}

func TestInvitationService_ValidateInvitation_Expired(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	username := "testuser"
	invitation := &model.Invitation{
		ID:        "invitation-1",
		Email:     "test@example.com",
		Username:  &username,
		Code:      "correct-code",
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	if err := db.Create(invitation).Error; err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	result, err := svc.ValidateInvitation(invitation.ID, invitation.Code)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for expired invitation")
	}

	if result.ErrorMessage != "invitation has expired" {
		t.Errorf("Expected 'invitation has expired', got '%s'", result.ErrorMessage)
	}
}

func TestInvitationService_ValidateInvitation_MaxUsesReached(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	username := "testuser"
	maxUses := 2
	invitation := &model.Invitation{
		ID:        "invitation-1",
		Email:     "test@example.com",
		Username:  &username,
		Code:      "correct-code",
		ExpiresAt: time.Now().Add(time.Hour),
		MaxUses:   &maxUses,
		UsedCount: 2,
	}
	if err := db.Create(invitation).Error; err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	result, err := svc.ValidateInvitation(invitation.ID, invitation.Code)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for max uses reached")
	}

	if result.ErrorMessage != "invitation has reached maximum uses" {
		t.Errorf("Expected 'invitation has reached maximum uses', got '%s'", result.ErrorMessage)
	}
}

func TestInvitationService_ValidateInvitation_Valid(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	username := "testuser"
	invitation := &model.Invitation{
		ID:        "invitation-1",
		Email:     "test@example.com",
		Username:  &username,
		Code:      "valid-code",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := db.Create(invitation).Error; err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	result, err := svc.ValidateInvitation(invitation.ID, invitation.Code)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid result")
	}

	if result.Invitation == nil {
		t.Error("Expected invitation object in result")
	}
}

func TestInvitationService_Create(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	invitation, err := svc.Create("test@example.com", "testuser", 7*24*time.Hour, nil, nil, nil, "test-creator-id")
	if err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	if invitation.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", invitation.Email)
	}

	if invitation.Username == nil || *invitation.Username != "testuser" {
		t.Error("Username mismatch")
	}

	if invitation.Code == "" {
		t.Error("Code should be generated")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestInvitationService_Create_DefaultExpiry(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	invitation, err := svc.Create("test@example.com", "testuser", 0, nil, nil, nil, "test-creator-id")
	if err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	if invitation.ExpiresAt.Before(expectedExpiry.Add(-time.Minute)) || invitation.ExpiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Error("Default expiry should be 7 days")
	}
}

func TestInvitationService_GetByID(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	created, err := svc.Create("test@example.com", "testuser", 7*24*time.Hour, nil, nil, nil, "test-creator-id")
	if err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	found, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get invitation by ID: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, found.ID)
	}
}

func TestInvitationService_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	_, err := svc.GetByID("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent invitation")
	}
}

func TestInvitationService_GetByCode(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	created, err := svc.Create("test@example.com", "testuser", 7*24*time.Hour, nil, nil, nil, "test-creator-id")
	if err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	found, err := svc.GetByCode(created.Code)
	if err != nil {
		t.Fatalf("Failed to get invitation by code: %v", err)
	}

	if found.Code != created.Code {
		t.Errorf("Expected Code '%s', got '%s'", created.Code, found.Code)
	}
}

func TestInvitationService_GetByCode_NotFound(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	_, err := svc.GetByCode("non-existent-code")
	if err == nil {
		t.Error("Expected error for non-existent code")
	}
}

func TestInvitationService_List(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	svc.Create("test1@example.com", "user1", 7*24*time.Hour, nil, nil, nil, "test-creator-id")
	svc.Create("test2@example.com", "user2", 7*24*time.Hour, nil, nil, nil, "test-creator-id")

	invitations, total, err := svc.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list invitations: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}

	if len(invitations) != 2 {
		t.Errorf("Expected 2 invitations, got %d", len(invitations))
	}
}

func TestInvitationService_List_Pagination(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	for i := 0; i < 5; i++ {
		svc.Create("test@example.com", "user", 7*24*time.Hour, nil, nil, nil, "test-creator-id")
	}

	invitations, total, err := svc.List(0, 2)
	if err != nil {
		t.Fatalf("Failed to list invitations: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(invitations) != 2 {
		t.Errorf("Expected 2 invitations, got %d", len(invitations))
	}
}

func TestInvitationService_Delete(t *testing.T) {
	db, cleanup := setupInvitationServiceTestDB(t)
	defer cleanup()

	svc := newTestInvitationService(db)

	created, err := svc.Create("test@example.com", "testuser", 7*24*time.Hour, nil, nil, nil, "test-creator-id")
	if err != nil {
		t.Fatalf("Failed to create invitation: %v", err)
	}

	err = svc.Delete(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete invitation: %v", err)
	}

	_, err = svc.GetByID(created.ID)
	if err == nil {
		t.Error("Expected error after deleting invitation")
	}
}

func TestInvitationValidation_Structure(t *testing.T) {
	validation := &InvitationValidation{
		Valid:        true,
		Invitation:   nil,
		ErrorMessage: "",
	}

	if validation.Valid != true {
		t.Error("Valid should be true")
	}

	validation.Valid = false
	validation.ErrorMessage = "test error"

	if validation.Valid {
		t.Error("Valid should be false")
	}
	if validation.ErrorMessage != "test error" {
		t.Error("ErrorMessage mismatch")
	}
}
