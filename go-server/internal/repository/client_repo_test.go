package repository

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

func setupTestDBForClient(t *testing.T) *gorm.DB {
	db := SetupTestDB(t)
	return db
}

func TestClientRepository_Create(t *testing.T) {
	db := setupTestDBForClient(t)
	defer CleanupDB(db)

	repo := NewClientRepository(db)

	client := &model.Client{
		ID:       "client-1",
		ClientID: "test-client-id",
		Name:     "Test Client",
	}

	if err := repo.Create(client); err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	retrieved, err := repo.GetByID("client-1")
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}

	if retrieved.ClientID != "test-client-id" {
		t.Errorf("Expected ClientID 'test-client-id', got '%s'", retrieved.ClientID)
	}
}

func TestClientRepository_GetByID(t *testing.T) {
	db := setupTestDBForClient(t)
	defer CleanupDB(db)

	repo := NewClientRepository(db)

	client := &model.Client{
		ID:       "client-get-by-id",
		ClientID: "get-by-id-client",
		Name:     "Get By ID Client",
	}

	repo.Create(client)

	retrieved, err := repo.GetByID("client-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get client by ID: %v", err)
	}

	if retrieved.ClientID != "get-by-id-client" {
		t.Errorf("Expected ClientID 'get-by-id-client', got '%s'", retrieved.ClientID)
	}
}

func TestClientRepository_GetByClientID(t *testing.T) {
	db := setupTestDBForClient(t)
	defer CleanupDB(db)

	repo := NewClientRepository(db)

	client := &model.Client{
		ID:       "client-get-by-client-id",
		ClientID: "unique-client-id",
		Name:     "Get By Client ID Client",
	}

	repo.Create(client)

	retrieved, err := repo.GetByClientID("unique-client-id")
	if err != nil {
		t.Fatalf("Failed to get client by ClientID: %v", err)
	}

	if retrieved.ID != "client-get-by-client-id" {
		t.Errorf("Expected ID 'client-get-by-client-id', got '%s'", retrieved.ID)
	}
}

func TestClientRepository_GetByClientID_NotFound(t *testing.T) {
	db := setupTestDBForClient(t)
	defer CleanupDB(db)

	repo := NewClientRepository(db)

	_, err := repo.GetByClientID("non-existent-client")
	if err == nil {
		t.Error("Expected error when getting non-existent client")
	}
}

func TestClientRepository_Update(t *testing.T) {
	db := setupTestDBForClient(t)
	defer CleanupDB(db)

	repo := NewClientRepository(db)

	client := &model.Client{
		ID:       "client-update",
		ClientID: "update-client-id",
		Name:     "Original Name",
	}

	repo.Create(client)

	client.Name = "Updated Name"
	if err := repo.Update(client); err != nil {
		t.Fatalf("Failed to update client: %v", err)
	}

	retrieved, _ := repo.GetByID("client-update")
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", retrieved.Name)
	}
}

func TestClientRepository_Delete(t *testing.T) {
	db := setupTestDBForClient(t)
	defer CleanupDB(db)

	repo := NewClientRepository(db)

	client := &model.Client{
		ID:       "client-delete",
		ClientID: "delete-client-id",
		Name:     "Delete Client",
	}

	repo.Create(client)

	if err := repo.Delete("client-delete"); err != nil {
		t.Fatalf("Failed to delete client: %v", err)
	}

	_, err := repo.GetByID("client-delete")
	if err == nil {
		t.Error("Expected error when getting deleted client")
	}
}

func TestClientRepository_List(t *testing.T) {
	db := setupTestDBForClient(t)
	defer CleanupDB(db)

	repo := NewClientRepository(db)

	for i := 0; i < 5; i++ {
		client := &model.Client{
			ID:       "client-list-" + string(rune('a'+i)),
			ClientID: "list-client-" + string(rune('a'+i)),
			Name:     "List Client " + string(rune('a'+i)),
		}
		repo.Create(client)
	}

	clients, total, err := repo.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list clients: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(clients) != 5 {
		t.Errorf("Expected 5 clients, got %d", len(clients))
	}
}

func TestClientRepository_List_WithPagination(t *testing.T) {
	db := setupTestDBForClient(t)
	defer CleanupDB(db)

	repo := NewClientRepository(db)

	for i := 0; i < 10; i++ {
		client := &model.Client{
			ID:       "client-page-" + string(rune('a'+i)),
			ClientID: "page-client-" + string(rune('a'+i)),
			Name:     "Page Client " + string(rune('a'+i)),
		}
		repo.Create(client)
	}

	clients, total, err := repo.List(0, 3)
	if err != nil {
		t.Fatalf("Failed to list clients with pagination: %v", err)
	}

	if total != 10 {
		t.Errorf("Expected total 10, got %d", total)
	}

	if len(clients) != 3 {
		t.Errorf("Expected 3 clients, got %d", len(clients))
	}
}
