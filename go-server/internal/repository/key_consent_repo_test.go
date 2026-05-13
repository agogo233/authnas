package repository

import (
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/model"
)

func TestKeyRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewKeyRepository(db)

	user := &model.User{ID: "key-user-1", Username: "keyuser1"}
	db.Create(user)

	key := &model.Key{
		ID:        "key-1",
		UserID:    "key-user-1",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if err := repo.Create(key); err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	retrieved, err := repo.GetByID("key-1")
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}

	if retrieved.UserID != "key-user-1" {
		t.Errorf("Expected UserID 'key-user-1', got '%s'", retrieved.UserID)
	}
}

func TestKeyRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewKeyRepository(db)

	user := &model.User{ID: "key-user-get", Username: "keyuserget"}
	db.Create(user)

	key := &model.Key{
		ID:        "key-get-by-id",
		UserID:    "key-user-get",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	repo.Create(key)

	retrieved, err := repo.GetByID("key-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get key by ID: %v", err)
	}

	if retrieved.ID != "key-get-by-id" {
		t.Errorf("Expected ID 'key-get-by-id', got '%s'", retrieved.ID)
	}
}

func TestKeyRepository_GetByID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewKeyRepository(db)

	_, err := repo.GetByID("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent key")
	}
}

func TestKeyRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewKeyRepository(db)

	key := &model.Key{
		ID:        "key-delete",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	repo.Create(key)

	err := repo.Delete("key-delete")
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	_, err = repo.GetByID("key-delete")
	if err == nil {
		t.Error("Expected error after deleting key")
	}
}

func TestKeyRepository_DeleteByUserID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewKeyRepository(db)

	user1 := &model.User{ID: "key-user-to-delete", Username: "keyusertodelete1"}
	user2 := &model.User{ID: "key-other-user", Username: "keyotheruser1"}
	db.Create(user1)
	db.Create(user2)

	key1 := &model.Key{ID: "key-ud-1", UserID: "key-user-to-delete", RefreshTokenHash: "hash1", ExpiresAt: time.Now().Add(time.Hour)}
	key2 := &model.Key{ID: "key-ud-2", UserID: "key-user-to-delete", RefreshTokenHash: "hash2", ExpiresAt: time.Now().Add(time.Hour)}
	key3 := &model.Key{ID: "key-ud-3", UserID: "key-other-user", RefreshTokenHash: "hash3", ExpiresAt: time.Now().Add(time.Hour)}
	repo.Create(key1)
	repo.Create(key2)
	repo.Create(key3)

	err := repo.DeleteByUserID("key-user-to-delete")
	if err != nil {
		t.Fatalf("Failed to delete keys by user ID: %v", err)
	}

	_, err = repo.GetByID("key-ud-1")
	if err == nil {
		t.Error("Expected error for deleted key")
	}

	_, err = repo.GetByID("key-ud-3")
	if err != nil {
		t.Error("Key for other user should still exist")
	}
}

func TestKeyRepository_DeleteByClientID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewKeyRepository(db)

	user1 := &model.User{ID: "key-user-c1", Username: "keyuserc11"}
	user2 := &model.User{ID: "key-user-c2", Username: "keyuserc21"}
	user3 := &model.User{ID: "key-user-c3", Username: "keyuserc31"}
	db.Create(user1)
	db.Create(user2)
	db.Create(user3)

	key1 := &model.Key{ID: "key-cd-1", UserID: "key-user-c1", ClientID: "client-to-delete", RefreshTokenHash: "hashc1", ExpiresAt: time.Now().Add(time.Hour)}
	key2 := &model.Key{ID: "key-cd-2", UserID: "key-user-c2", ClientID: "client-to-delete", RefreshTokenHash: "hashc2", ExpiresAt: time.Now().Add(time.Hour)}
	key3 := &model.Key{ID: "key-cd-3", UserID: "key-user-c3", ClientID: "other-client", RefreshTokenHash: "hashc3", ExpiresAt: time.Now().Add(time.Hour)}
	repo.Create(key1)
	repo.Create(key2)
	repo.Create(key3)

	err := repo.DeleteByClientID("client-to-delete")
	if err != nil {
		t.Fatalf("Failed to delete keys by client ID: %v", err)
	}

	_, err = repo.GetByID("key-cd-1")
	if err == nil {
		t.Error("Expected error for deleted key")
	}

	_, err = repo.GetByID("key-cd-3")
	if err != nil {
		t.Error("Key for other client should still exist")
	}
}

func TestKeyRepository_GetByUserID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewKeyRepository(db)

	user1 := &model.User{ID: "key-user-keys", Username: "keyuserkeys1"}
	user2 := &model.User{ID: "key-other-user-keys", Username: "keyotheruserkeys1"}
	db.Create(user1)
	db.Create(user2)

	key1 := &model.Key{ID: "key-gbu-1", UserID: "key-user-keys", RefreshTokenHash: "hashgu1", ExpiresAt: time.Now().Add(time.Hour)}
	key2 := &model.Key{ID: "key-gbu-2", UserID: "key-user-keys", RefreshTokenHash: "hashgu2", ExpiresAt: time.Now().Add(time.Hour)}
	key3 := &model.Key{ID: "key-gbu-3", UserID: "key-other-user-keys", RefreshTokenHash: "hashgu3", ExpiresAt: time.Now().Add(time.Hour)}
	repo.Create(key1)
	repo.Create(key2)
	repo.Create(key3)

	keys, err := repo.GetByUserID("key-user-keys")
	if err != nil {
		t.Fatalf("Failed to get keys by user ID: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
}

func TestConsentRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewConsentRepository(db)

	user := &model.User{ID: "consent-user-1", Username: "consentuser1"}
	db.Create(user)

	consent := &model.Consent{
		ID:       "consent-1",
		UserID:   "consent-user-1",
		ClientID: "client-1",
		Scopes:   "openid profile",
	}

	if err := repo.Create(consent); err != nil {
		t.Fatalf("Failed to create consent: %v", err)
	}

	retrieved, err := repo.GetByID("consent-1")
	if err != nil {
		t.Fatalf("Failed to get consent: %v", err)
	}

	if retrieved.UserID != "consent-user-1" {
		t.Errorf("Expected UserID 'consent-user-1', got '%s'", retrieved.UserID)
	}
}

func TestConsentRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewConsentRepository(db)

	user := &model.User{ID: "consent-user-get", Username: "consentuserget"}
	db.Create(user)

	consent := &model.Consent{
		ID:       "consent-get-by-id",
		UserID:   "consent-user-get",
		ClientID: "client-1",
		Scopes:   "openid profile",
	}
	repo.Create(consent)

	retrieved, err := repo.GetByID("consent-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get consent by ID: %v", err)
	}

	if retrieved.ID != "consent-get-by-id" {
		t.Errorf("Expected ID 'consent-get-by-id', got '%s'", retrieved.ID)
	}
}

func TestConsentRepository_GetByID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewConsentRepository(db)

	_, err := repo.GetByID("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent consent")
	}
}

func TestConsentRepository_GetByUserAndClient(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewConsentRepository(db)

	user := &model.User{ID: "consent-user-ubc", Username: "consentuserubc"}
	db.Create(user)

	consent := &model.Consent{
		ID:       "consent-ubc",
		UserID:   "consent-user-ubc",
		ClientID: "client-ubc",
		Scopes:   "openid profile email",
	}
	repo.Create(consent)

	retrieved, err := repo.GetByUserAndClient("consent-user-ubc", "client-ubc")
	if err != nil {
		t.Fatalf("Failed to get consent by user and client: %v", err)
	}

	if retrieved.Scopes != "openid profile email" {
		t.Errorf("Expected Scopes 'openid profile email', got '%s'", retrieved.Scopes)
	}
}

func TestConsentRepository_GetByUserAndClient_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewConsentRepository(db)

	_, err := repo.GetByUserAndClient("nonexistent-user", "nonexistent-client")
	if err == nil {
		t.Error("Expected error for nonexistent consent")
	}
}

func TestConsentRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewConsentRepository(db)

	user := &model.User{ID: "consent-user-update", Username: "consentuserupdate"}
	db.Create(user)

	consent := &model.Consent{
		ID:       "consent-update",
		UserID:   "consent-user-update",
		ClientID: "client-update",
		Scopes:   "openid",
	}
	repo.Create(consent)

	consent.Scopes = "openid profile email"
	if err := repo.Update(consent); err != nil {
		t.Fatalf("Failed to update consent: %v", err)
	}

	retrieved, _ := repo.GetByID("consent-update")
	if retrieved.Scopes != "openid profile email" {
		t.Errorf("Expected updated Scopes, got '%s'", retrieved.Scopes)
	}
}

func TestConsentRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewConsentRepository(db)

	user := &model.User{ID: "consent-user-delete", Username: "consentuserdelete"}
	db.Create(user)

	consent := &model.Consent{
		ID:       "consent-delete",
		UserID:   "consent-user-delete",
		ClientID: "client-delete",
		Scopes:   "openid",
	}
	repo.Create(consent)

	err := repo.Delete("consent-delete")
	if err != nil {
		t.Fatalf("Failed to delete consent: %v", err)
	}

	_, err = repo.GetByID("consent-delete")
	if err == nil {
		t.Error("Expected error after deleting consent")
	}
}

func TestConsentRepository_DeleteByUserAndClient(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewConsentRepository(db)

	user1 := &model.User{ID: "consent-user-dubc", Username: "consentuserdubc"}
	user2 := &model.User{ID: "consent-other-user-dubc", Username: "consentotheruserdubc"}
	db.Create(user1)
	db.Create(user2)

	consent1 := &model.Consent{ID: "consent-dubc-1", UserID: "consent-user-dubc", ClientID: "client-dubc", Scopes: "openid"}
	consent2 := &model.Consent{ID: "consent-dubc-2", UserID: "consent-user-dubc", ClientID: "other-client", Scopes: "openid"}
	consent3 := &model.Consent{ID: "consent-dubc-3", UserID: "consent-other-user-dubc", ClientID: "client-dubc", Scopes: "openid"}
	repo.Create(consent1)
	repo.Create(consent2)
	repo.Create(consent3)

	err := repo.DeleteByUserAndClient("consent-user-dubc", "client-dubc")
	if err != nil {
		t.Fatalf("Failed to delete consent by user and client: %v", err)
	}

	_, err = repo.GetByID("consent-dubc-1")
	if err == nil {
		t.Error("Expected error for deleted consent")
	}

	_, err = repo.GetByID("consent-dubc-2")
	if err != nil {
		t.Error("Consent for other client should still exist")
	}

	_, err = repo.GetByID("consent-dubc-3")
	if err != nil {
		t.Error("Consent for other user should still exist")
	}
}

func TestConsentRepository_DeleteByUserID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewConsentRepository(db)

	user1 := &model.User{ID: "user-dbu", Username: "userdbu"}
	user2 := &model.User{ID: "other-user-dbu", Username: "otheruserdbu"}
	db.Create(user1)
	db.Create(user2)

	consent1 := &model.Consent{ID: "consent-dbu-1", UserID: "user-dbu", ClientID: "client-1", Scopes: "openid"}
	consent2 := &model.Consent{ID: "consent-dbu-2", UserID: "user-dbu", ClientID: "client-2", Scopes: "openid"}
	consent3 := &model.Consent{ID: "consent-dbu-3", UserID: "other-user-dbu", ClientID: "client-3", Scopes: "openid"}
	repo.Create(consent1)
	repo.Create(consent2)
	repo.Create(consent3)

	err := repo.DeleteByUserID("user-dbu")
	if err != nil {
		t.Fatalf("Failed to delete consent by user ID: %v", err)
	}

	_, err = repo.GetByID("consent-dbu-1")
	if err == nil {
		t.Error("Expected error for deleted consent")
	}

	_, err = repo.GetByID("consent-dbu-3")
	if err != nil {
		t.Error("Consent for other user should still exist")
	}
}
