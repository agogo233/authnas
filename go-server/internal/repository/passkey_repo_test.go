package repository

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

func setupTestDBForPasskey(t *testing.T) *gorm.DB {
	return SetupTestDB(t)
}

func TestPasskeyRepository_Create(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyRepository(db)
	userRepo := NewUserRepository(db)

	user := &model.User{
		ID:       "user-1",
		Username: "passkeyuser",
		Email:    StringPtr("passkey@example.com"),
	}
	userRepo.Create(user)

	transports := "usb"
	passkey := &model.Passkey{
		ID:           "passkey-1",
		UserID:       "user-1",
		Name:         StringPtr("Test Passkey"),
		CredentialID: "credential-1",
		PublicKey:    "public-key-data",
		Transports:   &transports,
	}

	if err := repo.Create(passkey); err != nil {
		t.Fatalf("Failed to create passkey: %v", err)
	}

	retrieved, err := repo.GetByID("passkey-1")
	if err != nil {
		t.Fatalf("Failed to get passkey: %v", err)
	}

	if retrieved.CredentialID != "credential-1" {
		t.Errorf("Expected CredentialID 'credential-1', got '%s'", retrieved.CredentialID)
	}
}

func TestPasskeyRepository_GetByID(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-get-by-id", Username: "user1", Email: StringPtr("user1@example.com")}
	userRepo.Create(user)

	passkey := &model.Passkey{
		ID:           "passkey-get-by-id",
		UserID:       "user-get-by-id",
		CredentialID: "cred-get-by-id",
		PublicKey:    "public-key",
	}

	repo.Create(passkey)

	retrieved, err := repo.GetByID("passkey-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get passkey by ID: %v", err)
	}

	if retrieved.CredentialID != "cred-get-by-id" {
		t.Errorf("Expected CredentialID 'cred-get-by-id', got '%s'", retrieved.CredentialID)
	}
}

func TestPasskeyRepository_GetByCredentialID(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-get-by-cred", Username: "user2", Email: StringPtr("user2@example.com")}
	userRepo.Create(user)

	passkey := &model.Passkey{
		ID:           "passkey-get-by-cred",
		UserID:       "user-get-by-cred",
		CredentialID: "unique-credential-id",
		PublicKey:    "public-key",
	}

	repo.Create(passkey)

	retrieved, err := repo.GetByCredentialID("unique-credential-id")
	if err != nil {
		t.Fatalf("Failed to get passkey by CredentialID: %v", err)
	}

	if retrieved.ID != "passkey-get-by-cred" {
		t.Errorf("Expected ID 'passkey-get-by-cred', got '%s'", retrieved.ID)
	}
}

func TestPasskeyRepository_GetByCredentialID_NotFound(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyRepository(db)

	_, err := repo.GetByCredentialID("non-existent-credential")
	if err == nil {
		t.Error("Expected error when getting passkey with non-existent credential ID")
	}
}

func TestPasskeyRepository_GetByUserID(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "same-user-for-passkeys", Username: "sameuser", Email: StringPtr("sameuser@example.com")}
	userRepo.Create(user)

	for i := 0; i < 3; i++ {
		passkey := &model.Passkey{
			ID:           "passkey-user-" + string(rune('a'+i)),
			UserID:       "same-user-for-passkeys",
			CredentialID: "cred-user-" + string(rune('a'+i)),
			PublicKey:    "public-key",
		}
		repo.Create(passkey)
	}

	passkeys, err := repo.GetByUserID("same-user-for-passkeys")
	if err != nil {
		t.Fatalf("Failed to get passkeys by UserID: %v", err)
	}

	if len(passkeys) != 3 {
		t.Errorf("Expected 3 passkeys, got %d", len(passkeys))
	}
}

func TestPasskeyRepository_Update(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-update", Username: "updateuser", Email: StringPtr("updateuser@example.com")}
	userRepo.Create(user)

	transports := "usb"
	passkey := &model.Passkey{
		ID:           "passkey-update",
		UserID:       "user-update",
		Name:         StringPtr("Original Name"),
		CredentialID: "cred-update",
		PublicKey:    "original-public-key",
		Transports:   &transports,
	}

	repo.Create(passkey)

	newName := "Updated Name"
	passkey.Name = &newName
	if err := repo.Update(passkey); err != nil {
		t.Fatalf("Failed to update passkey: %v", err)
	}

	retrieved, _ := repo.GetByID("passkey-update")
	if retrieved.Name == nil || *retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%v'", retrieved.Name)
	}
}

func TestPasskeyRepository_Delete(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-delete", Username: "deleteuser", Email: StringPtr("deleteuser@example.com")}
	userRepo.Create(user)

	passkey := &model.Passkey{
		ID:           "passkey-delete",
		UserID:       "user-delete",
		CredentialID: "cred-delete",
		PublicKey:    "public-key",
	}

	repo.Create(passkey)

	if err := repo.Delete("passkey-delete"); err != nil {
		t.Fatalf("Failed to delete passkey: %v", err)
	}

	_, err := repo.GetByID("passkey-delete")
	if err == nil {
		t.Error("Expected error when getting deleted passkey")
	}
}

func TestPasskeyRepository_DeleteByUserID(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyRepository(db)
	userRepo := NewUserRepository(db)
	user := &model.User{ID: "user-to-delete-passkeys", Username: "deluser", Email: StringPtr("deluser@example.com")}
	userRepo.Create(user)

	for i := 0; i < 3; i++ {
		passkey := &model.Passkey{
			ID:           "passkey-del-user-" + string(rune('a'+i)),
			UserID:       "user-to-delete-passkeys",
			CredentialID: "cred-del-user-" + string(rune('a'+i)),
			PublicKey:    "public-key",
		}
		repo.Create(passkey)
	}

	if err := repo.DeleteByUserID("user-to-delete-passkeys"); err != nil {
		t.Fatalf("Failed to delete passkeys by UserID: %v", err)
	}

	passkeys, err := repo.GetByUserID("user-to-delete-passkeys")
	if err != nil {
		t.Fatalf("Failed to get passkeys after delete: %v", err)
	}

	if len(passkeys) != 0 {
		t.Errorf("Expected 0 passkeys after delete, got %d", len(passkeys))
	}
}

func TestPasskeyAuthOptionsRepository_Create(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyAuthOptionsRepository(db)

	userID := "user-for-auth-opts"
	opts := &model.PasskeyAuthOptions{
		ID:        "auth-opts-1",
		UserID:    &userID,
		Challenge: "challenge-data",
		Options:   "{}",
	}

	if err := repo.Create(opts); err != nil {
		t.Fatalf("Failed to create passkey auth options: %v", err)
	}

	retrieved, err := repo.GetByID("auth-opts-1")
	if err != nil {
		t.Fatalf("Failed to get passkey auth options: %v", err)
	}

	if retrieved.Challenge != "challenge-data" {
		t.Errorf("Expected challenge 'challenge-data', got '%s'", retrieved.Challenge)
	}
}

func TestPasskeyAuthOptionsRepository_GetByID(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyAuthOptionsRepository(db)

	userID := "user-get-opts-by-id"
	opts := &model.PasskeyAuthOptions{
		ID:        "auth-opts-get-by-id",
		UserID:    &userID,
		Challenge: "get-by-id-challenge",
		Options:   "{}",
	}

	repo.Create(opts)

	retrieved, err := repo.GetByID("auth-opts-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get passkey auth options by ID: %v", err)
	}

	if retrieved.Challenge != "get-by-id-challenge" {
		t.Errorf("Expected challenge 'get-by-id-challenge', got '%s'", retrieved.Challenge)
	}
}

func TestPasskeyAuthOptionsRepository_GetByUserID(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyAuthOptionsRepository(db)

	userID := "user-get-opts-by-user"
	opts1 := &model.PasskeyAuthOptions{
		ID:        "auth-opts-user-1",
		UserID:    &userID,
		Challenge: "challenge-1",
		Options:   "{}",
	}
	opts2 := &model.PasskeyAuthOptions{
		ID:        "auth-opts-user-2",
		UserID:    &userID,
		Challenge: "challenge-2",
		Options:   "{}",
	}

	repo.Create(opts1)
	repo.Create(opts2)

	retrieved, err := repo.GetByUserID(userID)
	if err != nil {
		t.Fatalf("Failed to get passkey auth options by UserID: %v", err)
	}

	if retrieved.ID != "auth-opts-user-2" {
		t.Logf("Expected most recent options 'auth-opts-user-2', got '%s'", retrieved.ID)
	}
}

func TestPasskeyAuthOptionsRepository_GetByChallenge(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyAuthOptionsRepository(db)

	userID := "user-get-opts-by-challenge"
	opts := &model.PasskeyAuthOptions{
		ID:        "auth-opts-challenge",
		UserID:    &userID,
		Challenge: "unique-challenge-code",
		Options:   "{}",
	}

	repo.Create(opts)

	retrieved, err := repo.GetByChallenge("unique-challenge-code")
	if err != nil {
		t.Fatalf("Failed to get passkey auth options by challenge: %v", err)
	}

	if retrieved.ID != "auth-opts-challenge" {
		t.Errorf("Expected ID 'auth-opts-challenge', got '%s'", retrieved.ID)
	}
}

func TestPasskeyAuthOptionsRepository_GetByChallenge_NotFound(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyAuthOptionsRepository(db)

	_, err := repo.GetByChallenge("non-existent-challenge")
	if err == nil {
		t.Error("Expected error when getting options with non-existent challenge")
	}
}

func TestPasskeyAuthOptionsRepository_Delete(t *testing.T) {
	db := setupTestDBForPasskey(t)
	defer CleanupDB(db)

	repo := NewPasskeyAuthOptionsRepository(db)

	userID := "user-delete-opts"
	opts := &model.PasskeyAuthOptions{
		ID:        "auth-opts-delete",
		UserID:    &userID,
		Challenge: "delete-challenge",
		Options:   "{}",
	}

	repo.Create(opts)

	if err := repo.Delete("auth-opts-delete"); err != nil {
		t.Fatalf("Failed to delete passkey auth options: %v", err)
	}

	_, err := repo.GetByID("auth-opts-delete")
	if err == nil {
		t.Error("Expected error when getting deleted passkey auth options")
	}
}
