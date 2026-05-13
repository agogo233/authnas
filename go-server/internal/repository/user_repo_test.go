package repository

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/model"
)

func TestUserRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewUserRepository(db)

	user := &model.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    StringPtr("test@example.com"),
	}

	if err := repo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := repo.GetByID("user-1")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrieved.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", retrieved.Username)
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewUserRepository(db)

	user := &model.User{
		ID:       "user-get-by-id",
		Username: "getbyid",
		Email:    StringPtr("getbyid@example.com"),
	}

	repo.Create(user)

	retrieved, err := repo.GetByID("user-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if retrieved.Username != "getbyid" {
		t.Errorf("Expected username 'getbyid', got '%s'", retrieved.Username)
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewUserRepository(db)

	user := &model.User{
		ID:       "user-get-by-username",
		Username: "uniqueusername",
		Email:    StringPtr("unique@example.com"),
	}

	repo.Create(user)

	retrieved, err := repo.GetByUsername("uniqueusername")
	if err != nil {
		t.Fatalf("Failed to get user by username: %v", err)
	}

	if retrieved.ID != "user-get-by-username" {
		t.Errorf("Expected ID 'user-get-by-username', got '%s'", retrieved.ID)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewUserRepository(db)

	user := &model.User{
		ID:       "user-get-by-email",
		Username: "emailuser",
		Email:    StringPtr("uniqueemail@example.com"),
	}

	repo.Create(user)

	retrieved, err := repo.GetByEmail("uniqueemail@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if retrieved.ID != "user-get-by-email" {
		t.Errorf("Expected ID 'user-get-by-email', got '%s'", retrieved.ID)
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewUserRepository(db)

	user := &model.User{
		ID:       "user-update",
		Username: "updateuser",
		Email:    StringPtr("update@example.com"),
	}

	repo.Create(user)

	user.Name = StringPtr("Updated Name")
	if err := repo.Update(user); err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	retrieved, _ := repo.GetByID("user-update")
	if retrieved.Name == nil || *retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%v'", retrieved.Name)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewUserRepository(db)

	user := &model.User{
		ID:       "user-delete",
		Username: "deleteuser",
		Email:    StringPtr("delete@example.com"),
	}

	repo.Create(user)

	if err := repo.Delete("user-delete"); err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	_, err := repo.GetByID("user-delete")
	if err == nil {
		t.Error("Expected error when getting deleted user")
	}
}

func TestUserRepository_List(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewUserRepository(db)

	for i := 0; i < 5; i++ {
		user := &model.User{
			ID:       "user-list-" + string(rune('a'+i)),
			Username: "listuser" + string(rune('a'+i)),
			Email:    StringPtr("list" + string(rune('a'+i)) + "@example.com"),
		}
		repo.Create(user)
	}

	users, total, err := repo.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}
}

func TestUserRepository_Search(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewUserRepository(db)

	users := []struct {
		id       string
		username string
		email    string
	}{
		{"search-1", "alice", "alice@example.com"},
		{"search-2", "bob", "bob@example.com"},
		{"search-3", "alice_smith", "alice_smith@example.com"},
	}

	for _, u := range users {
		user := &model.User{
			ID:       u.id,
			Username: u.username,
			Email:    StringPtr(u.email),
		}
		repo.Create(user)
	}

	results, total, err := repo.Search("alice", 0, 10)
	if err != nil {
		t.Fatalf("Failed to search users: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 results for 'alice', got %d", total)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 users, got %d", len(results))
	}
}

func TestUserRepository_IncrementTokenVersion(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)
	repo := NewUserRepository(db)

	user := &model.User{
		ID:           "user-token-version",
		Username:     "tokenuser",
		Email:        StringPtr("token@example.com"),
		TokenVersion: 0,
	}

	repo.Create(user)

	if err := repo.IncrementTokenVersion("user-token-version"); err != nil {
		t.Fatalf("Failed to increment token version: %v", err)
	}

	retrieved, _ := repo.GetByID("user-token-version")
	if retrieved.TokenVersion != 1 {
		t.Errorf("Expected TokenVersion 1, got %d", retrieved.TokenVersion)
	}
}
