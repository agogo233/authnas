package repository

import (
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/model"
)

func TestOIDCPayloadRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewOIDCPayloadRepository(db)

	payload := &model.OIDCPayload{
		ID:        "oidc-1",
		UID:       "user-uid-1",
		Payload:   `{"sub":"user123","name":"Test User"}`,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := repo.Create(payload); err != nil {
		t.Fatalf("Failed to create OIDCPayload: %v", err)
	}

	retrieved, err := repo.GetByID("oidc-1")
	if err != nil {
		t.Fatalf("Failed to get OIDCPayload: %v", err)
	}

	if retrieved.UID != "user-uid-1" {
		t.Errorf("Expected UID 'user-uid-1', got '%s'", retrieved.UID)
	}
}

func TestOIDCPayloadRepository_GetByUID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewOIDCPayloadRepository(db)

	payload := &model.OIDCPayload{
		ID:        "oidc-get-by-uid",
		UID:       "unique-uid-for-test",
		Payload:   `{"sub":"user456"}`,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	repo.Create(payload)

	retrieved, err := repo.GetByUID("unique-uid-for-test")
	if err != nil {
		t.Fatalf("Failed to get OIDCPayload by UID: %v", err)
	}

	if retrieved.ID != "oidc-get-by-uid" {
		t.Errorf("Expected ID 'oidc-get-by-uid', got '%s'", retrieved.ID)
	}
}

func TestOIDCPayloadRepository_GetByUID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewOIDCPayloadRepository(db)

	_, err := repo.GetByUID("non-existent-uid")
	if err == nil {
		t.Error("Expected error when getting OIDCPayload for non-existent UID")
	}
}

func TestOIDCPayloadRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewOIDCPayloadRepository(db)

	payload := &model.OIDCPayload{
		ID:        "oidc-update",
		UID:       "uid-for-update",
		Payload:   `{"sub":"user789","old":"data"}`,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	repo.Create(payload)

	payload.Payload = `{"sub":"user789","new":"data"}`
	if err := repo.Update(payload); err != nil {
		t.Fatalf("Failed to update OIDCPayload: %v", err)
	}

	retrieved, _ := repo.GetByID("oidc-update")
	if retrieved.Payload != `{"sub":"user789","new":"data"}` {
		t.Errorf("Expected updated payload, got '%s'", retrieved.Payload)
	}
}

func TestOIDCPayloadRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewOIDCPayloadRepository(db)

	payload := &model.OIDCPayload{
		ID:        "oidc-delete",
		UID:       "uid-for-delete",
		Payload:   `{"sub":"user-del"}`,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	repo.Create(payload)

	if err := repo.Delete("oidc-delete"); err != nil {
		t.Fatalf("Failed to delete OIDCPayload: %v", err)
	}

	_, err := repo.GetByID("oidc-delete")
	if err == nil {
		t.Error("Expected error when getting deleted OIDCPayload")
	}
}

func TestOIDCPayloadRepository_DeleteByUID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewOIDCPayloadRepository(db)

	payload := &model.OIDCPayload{
		ID:        "oidc-delete-by-uid",
		UID:       "uid-to-delete-by-uid",
		Payload:   `{"sub":"user-del-by-uid"}`,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	repo.Create(payload)

	if err := repo.DeleteByUID("uid-to-delete-by-uid"); err != nil {
		t.Fatalf("Failed to delete OIDCPayload by UID: %v", err)
	}

	_, err := repo.GetByUID("uid-to-delete-by-uid")
	if err == nil {
		t.Error("Expected error when getting deleted OIDCPayload by UID")
	}
}

func TestOIDCPayloadRepository_GetAndDeleteByUID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewOIDCPayloadRepository(db)

	payload := &model.OIDCPayload{
		ID:        "oidc-get-and-delete",
		UID:       "uid-get-and-delete",
		Payload:   `{"sub":"user-get-and-delete"}`,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	repo.Create(payload)

	retrieved, err := repo.GetAndDeleteByUID("uid-get-and-delete")
	if err != nil {
		t.Fatalf("Failed to get and delete OIDCPayload: %v", err)
	}

	if retrieved.ID != "oidc-get-and-delete" {
		t.Errorf("Expected ID 'oidc-get-and-delete', got '%s'", retrieved.ID)
	}

	_, err = repo.GetByUID("uid-get-and-delete")
	if err == nil {
		t.Error("Expected error after get and delete")
	}
}

func TestOIDCPayloadRepository_DeleteExpired(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewOIDCPayloadRepository(db)

	expiredPayload := &model.OIDCPayload{
		ID:        "oidc-expired",
		UID:       "uid-expired",
		Payload:   `{"sub":"expired-user"}`,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	validPayload := &model.OIDCPayload{
		ID:        "oidc-valid",
		UID:       "uid-valid",
		Payload:   `{"sub":"valid-user"}`,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	repo.Create(expiredPayload)
	repo.Create(validPayload)

	deleted, err := repo.DeleteExpired()
	if err != nil {
		t.Fatalf("Failed to delete expired OIDCPayloads: %v", err)
	}

	if deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", deleted)
	}

	_, err = repo.GetByID("oidc-expired")
	if err == nil {
		t.Error("Expired payload should have been deleted")
	}

	_, err = repo.GetByID("oidc-valid")
	if err != nil {
		t.Error("Valid payload should still exist")
	}
}
