package repository

import (
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/model"
)

func TestEmailVerificationRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewEmailVerificationRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-ev-create", Username: "evuser", Email: StringPtr("evuser@example.com")}
	userRepo.Create(user)

	ev := &model.EmailVerification{
		ID:        "ev-1",
		UserID:    "user-ev-create",
		Email:     "evuser@example.com",
		Code:      "CODE123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := repo.Create(ev); err != nil {
		t.Fatalf("Failed to create EmailVerification: %v", err)
	}

	retrieved, err := repo.GetByID("ev-1")
	if err != nil {
		t.Fatalf("Failed to get EmailVerification: %v", err)
	}

	if retrieved.Email != "evuser@example.com" {
		t.Errorf("Expected email 'evuser@example.com', got '%s'", retrieved.Email)
	}
}

func TestEmailVerificationRepository_GetByCode(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewEmailVerificationRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-ev-code", Username: "evcodeuser", Email: StringPtr("evcode@example.com")}
	userRepo.Create(user)

	ev := &model.EmailVerification{
		ID:        "ev-code",
		UserID:    "user-ev-code",
		Email:     "evcode@example.com",
		Code:      "UNIQUECODE",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	repo.Create(ev)

	retrieved, err := repo.GetByCode("UNIQUECODE")
	if err != nil {
		t.Fatalf("Failed to get EmailVerification by code: %v", err)
	}

	if retrieved.ID != "ev-code" {
		t.Errorf("Expected ID 'ev-code', got '%s'", retrieved.ID)
	}
}

func TestEmailVerificationRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewEmailVerificationRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-ev-delete", Username: "evdeleteuser", Email: StringPtr("evdelete@example.com")}
	userRepo.Create(user)

	ev := &model.EmailVerification{
		ID:        "ev-delete",
		UserID:    "user-ev-delete",
		Email:     "evdelete@example.com",
		Code:      "DELETECODE",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	repo.Create(ev)

	if err := repo.Delete("ev-delete"); err != nil {
		t.Fatalf("Failed to delete EmailVerification: %v", err)
	}

	_, err := repo.GetByID("ev-delete")
	if err == nil {
		t.Error("Expected error when getting deleted EmailVerification")
	}
}

func TestEmailVerificationRepository_DeleteByUserID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewEmailVerificationRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-ev-delete-by-uid", Username: "evdeluiduser", Email: StringPtr("evdeluid@example.com")}
	userRepo.Create(user)

	ev := &model.EmailVerification{
		ID:        "ev-delete-by-uid",
		UserID:    "user-ev-delete-by-uid",
		Email:     "evdeluid@example.com",
		Code:      "DELBYUID",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	repo.Create(ev)

	if err := repo.DeleteByUserID("user-ev-delete-by-uid"); err != nil {
		t.Fatalf("Failed to delete EmailVerification by UserID: %v", err)
	}

	_, err := repo.GetByID("ev-delete-by-uid")
	if err == nil {
		t.Error("Expected error when getting deleted EmailVerification by UserID")
	}
}

func TestPasswordResetRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewPasswordResetRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-pr-create", Username: "pruser", Email: StringPtr("pruser@example.com")}
	userRepo.Create(user)

	pr := &model.PasswordReset{
		ID:        "pr-1",
		UserID:    "user-pr-create",
		Code:      "RESETCODE123",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := repo.Create(pr); err != nil {
		t.Fatalf("Failed to create PasswordReset: %v", err)
	}

	retrieved, err := repo.GetByID("pr-1")
	if err != nil {
		t.Fatalf("Failed to get PasswordReset: %v", err)
	}

	if retrieved.Code != "RESETCODE123" {
		t.Errorf("Expected code 'RESETCODE123', got '%s'", retrieved.Code)
	}
}

func TestPasswordResetRepository_GetByCode(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewPasswordResetRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-pr-code", Username: "prcodeuser", Email: StringPtr("prcode@example.com")}
	userRepo.Create(user)

	pr := &model.PasswordReset{
		ID:        "pr-code",
		UserID:    "user-pr-code",
		Code:      "UNIQUEPRCODE",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	repo.Create(pr)

	retrieved, err := repo.GetByCode("UNIQUEPRCODE")
	if err != nil {
		t.Fatalf("Failed to get PasswordReset by code: %v", err)
	}

	if retrieved.UserID != "user-pr-code" {
		t.Errorf("Expected UserID 'user-pr-code', got '%s'", retrieved.UserID)
	}
}

func TestPasswordResetRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewPasswordResetRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-pr-delete", Username: "prdeleteuser", Email: StringPtr("prdelete@example.com")}
	userRepo.Create(user)

	pr := &model.PasswordReset{
		ID:        "pr-delete",
		UserID:    "user-pr-delete",
		Code:      "PRDELETECODE",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	repo.Create(pr)

	if err := repo.Delete("pr-delete"); err != nil {
		t.Fatalf("Failed to delete PasswordReset: %v", err)
	}

	_, err := repo.GetByID("pr-delete")
	if err == nil {
		t.Error("Expected error when getting deleted PasswordReset")
	}
}

func TestPasswordResetRepository_DeleteByUserID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewPasswordResetRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-pr-delete-by-uid", Username: "prdeluiduser", Email: StringPtr("prdeluid@example.com")}
	userRepo.Create(user)

	pr := &model.PasswordReset{
		ID:        "pr-delete-by-uid",
		UserID:    "user-pr-delete-by-uid",
		Code:      "PRDELBYUID",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	repo.Create(pr)

	if err := repo.DeleteByUserID("user-pr-delete-by-uid"); err != nil {
		t.Fatalf("Failed to delete PasswordReset by UserID: %v", err)
	}

	_, err := repo.GetByID("pr-delete-by-uid")
	if err == nil {
		t.Error("Expected error when getting deleted PasswordReset by UserID")
	}
}
