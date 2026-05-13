package repository

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

func setupTestDBForTOTP(t *testing.T) *gorm.DB {
	return SetupTestDB(t)
}

func TestTOTPRepository_Create(t *testing.T) {
	db := setupTestDBForTOTP(t)
	defer CleanupDB(db)

	repo := NewTOTPRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-1", Username: "user1", Email: StringPtr("user1@example.com")}
	userRepo.Create(user)

	totp := &model.TOTP{
		ID:     "totp-1",
		UserID: "user-1",
		Secret: "JBSWY3DPEHPK3PXP",
		Issuer: "AuthNas",
	}

	if err := repo.Create(totp); err != nil {
		t.Fatalf("Failed to create TOTP: %v", err)
	}

	retrieved, err := repo.GetByID("totp-1")
	if err != nil {
		t.Fatalf("Failed to get TOTP: %v", err)
	}

	if retrieved.Secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("Expected secret 'JBSWY3DPEHPK3PXP', got '%s'", retrieved.Secret)
	}
}

func TestTOTPRepository_GetByID(t *testing.T) {
	db := setupTestDBForTOTP(t)
	defer CleanupDB(db)

	repo := NewTOTPRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-get-by-id", Username: "user2", Email: StringPtr("user2@example.com")}
	userRepo.Create(user)

	totp := &model.TOTP{
		ID:     "totp-get-by-id",
		UserID: "user-get-by-id",
		Secret: "SECRET123",
		Issuer: "AuthNas",
	}

	repo.Create(totp)

	retrieved, err := repo.GetByID("totp-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get TOTP by ID: %v", err)
	}

	if retrieved.UserID != "user-get-by-id" {
		t.Errorf("Expected UserID 'user-get-by-id', got '%s'", retrieved.UserID)
	}
}

func TestTOTPRepository_GetByUserID(t *testing.T) {
	db := setupTestDBForTOTP(t)
	defer CleanupDB(db)

	repo := NewTOTPRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "unique-user-for-totp", Username: "totpuser", Email: StringPtr("totpuser@example.com")}
	userRepo.Create(user)

	totp := &model.TOTP{
		ID:     "totp-get-by-user",
		UserID: "unique-user-for-totp",
		Secret: "USERSECRET",
		Issuer: "AuthNas",
	}

	repo.Create(totp)

	retrieved, err := repo.GetByUserID("unique-user-for-totp")
	if err != nil {
		t.Fatalf("Failed to get TOTP by UserID: %v", err)
	}

	if retrieved.ID != "totp-get-by-user" {
		t.Errorf("Expected ID 'totp-get-by-user', got '%s'", retrieved.ID)
	}
}

func TestTOTPRepository_GetByUserID_NotFound(t *testing.T) {
	db := setupTestDBForTOTP(t)
	defer CleanupDB(db)

	repo := NewTOTPRepository(db)

	_, err := repo.GetByUserID("non-existent-user")
	if err == nil {
		t.Error("Expected error when getting TOTP for non-existent user")
	}
}

func TestTOTPRepository_Update(t *testing.T) {
	db := setupTestDBForTOTP(t)
	defer CleanupDB(db)

	repo := NewTOTPRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-update", Username: "updateuser", Email: StringPtr("updateuser@example.com")}
	userRepo.Create(user)

	totp := &model.TOTP{
		ID:     "totp-update",
		UserID: "user-update",
		Secret: "OLDSECRET",
		Issuer: "AuthNas",
	}

	repo.Create(totp)

	totp.Secret = "NEWSECRET"
	if err := repo.Update(totp); err != nil {
		t.Fatalf("Failed to update TOTP: %v", err)
	}

	retrieved, _ := repo.GetByID("totp-update")
	if retrieved.Secret != "NEWSECRET" {
		t.Errorf("Expected secret 'NEWSECRET', got '%s'", retrieved.Secret)
	}
}

func TestTOTPRepository_Delete(t *testing.T) {
	db := setupTestDBForTOTP(t)
	defer CleanupDB(db)

	repo := NewTOTPRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-delete", Username: "deleteuser", Email: StringPtr("deleteuser@example.com")}
	userRepo.Create(user)

	totp := &model.TOTP{
		ID:     "totp-delete",
		UserID: "user-delete",
		Secret: "DELETESECRET",
		Issuer: "AuthNas",
	}

	repo.Create(totp)

	if err := repo.Delete("totp-delete"); err != nil {
		t.Fatalf("Failed to delete TOTP: %v", err)
	}

	_, err := repo.GetByID("totp-delete")
	if err == nil {
		t.Error("Expected error when getting deleted TOTP")
	}
}

func TestTOTPRepository_DeleteByUserID(t *testing.T) {
	db := setupTestDBForTOTP(t)
	defer CleanupDB(db)

	repo := NewTOTPRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-to-delete-totp", Username: "deluser", Email: StringPtr("deluser@example.com")}
	userRepo.Create(user)

	totp := &model.TOTP{
		ID:     "totp-delete-by-user",
		UserID: "user-to-delete-totp",
		Secret: "DELETESECRET",
		Issuer: "AuthNas",
	}

	repo.Create(totp)

	if err := repo.DeleteByUserID("user-to-delete-totp"); err != nil {
		t.Fatalf("Failed to delete TOTP by UserID: %v", err)
	}

	_, err := repo.GetByUserID("user-to-delete-totp")
	if err == nil {
		t.Error("Expected error when getting deleted TOTP by UserID")
	}
}
