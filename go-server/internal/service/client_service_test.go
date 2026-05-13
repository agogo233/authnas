package service

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"gorm.io/gorm"
)

func setupClientServiceTestDB(t *testing.T) (*gorm.DB, func()) {
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

	if err := db.AutoMigrate(&model.Client{}); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func newTestClientService(db *gorm.DB) *ClientService {
	cfg := &config.Config{}
	clientRepo := repository.NewClientRepository(db)
	return NewClientService(cfg, clientRepo)
}

func TestClientService_Create(t *testing.T) {
	db, cleanup := setupClientServiceTestDB(t)
	defer cleanup()

	svc := newTestClientService(db)

	redirectURIs := "https://example.com/callback"
	client := &model.Client{
		ID:           "client-1",
		ClientID:     "my-client-id",
		Name:         "Test Client",
		RedirectURIs: redirectURIs,
	}

	err := svc.Create(client)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.ID == "" {
		t.Error("Client ID should be set")
	}
}

func TestClientService_GetByID(t *testing.T) {
	db, cleanup := setupClientServiceTestDB(t)
	defer cleanup()

	svc := newTestClientService(db)

	redirectURIs := "https://example.com/callback"
	client := &model.Client{
		ID:           "client-1",
		ClientID:     "my-client-id",
		Name:         "Test Client",
		RedirectURIs: redirectURIs,
	}
	svc.Create(client)

	found, err := svc.GetByID(client.ID)
	if err != nil {
		t.Fatalf("Failed to get client by ID: %v", err)
	}

	if found.ID != client.ID {
		t.Errorf("Expected ID '%s', got '%s'", client.ID, found.ID)
	}
}

func TestClientService_GetByClientID(t *testing.T) {
	db, cleanup := setupClientServiceTestDB(t)
	defer cleanup()

	svc := newTestClientService(db)

	redirectURIs := "https://example.com/callback"
	client := &model.Client{
		ID:           "client-1",
		ClientID:     "unique-client-id",
		Name:         "Test Client",
		RedirectURIs: redirectURIs,
	}
	svc.Create(client)

	found, err := svc.GetByClientID("unique-client-id")
	if err != nil {
		t.Fatalf("Failed to get client by clientID: %v", err)
	}

	if found.ClientID != "unique-client-id" {
		t.Errorf("Expected ClientID 'unique-client-id', got '%s'", found.ClientID)
	}
}

func TestClientService_GetByClientID_NotFound(t *testing.T) {
	db, cleanup := setupClientServiceTestDB(t)
	defer cleanup()

	svc := newTestClientService(db)

	_, err := svc.GetByClientID("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent client")
	}
}

func TestClientService_Update(t *testing.T) {
	db, cleanup := setupClientServiceTestDB(t)
	defer cleanup()

	svc := newTestClientService(db)

	redirectURIs := "https://example.com/callback"
	client := &model.Client{
		ID:           "client-1",
		ClientID:     "my-client-id",
		Name:         "Original Name",
		RedirectURIs: redirectURIs,
	}
	svc.Create(client)

	client.Name = "Updated Name"
	err := svc.Update(client)
	if err != nil {
		t.Fatalf("Failed to update client: %v", err)
	}

	updated, _ := svc.GetByID(client.ID)
	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
	}
}

func TestClientService_Delete(t *testing.T) {
	db, cleanup := setupClientServiceTestDB(t)
	defer cleanup()

	svc := newTestClientService(db)

	redirectURIs := "https://example.com/callback"
	client := &model.Client{
		ID:           "client-1",
		ClientID:     "my-client-id",
		Name:         "Test Client",
		RedirectURIs: redirectURIs,
	}
	svc.Create(client)

	err := svc.Delete(client.ID)
	if err != nil {
		t.Fatalf("Failed to delete client: %v", err)
	}

	_, err = svc.GetByID(client.ID)
	if err == nil {
		t.Error("Expected error after deleting client")
	}
}

func TestClientService_List(t *testing.T) {
	db, cleanup := setupClientServiceTestDB(t)
	defer cleanup()

	svc := newTestClientService(db)

	for i := 0; i < 3; i++ {
		redirectURIs := "https://example.com/callback"
		client := &model.Client{
			ID:           "client-" + string(rune('a'+i)),
			ClientID:     "client-id-" + string(rune('a'+i)),
			Name:         "Client " + string(rune('A'+i)),
			RedirectURIs: redirectURIs,
		}
		svc.Create(client)
	}

	clients, total, err := svc.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list clients: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}

	if len(clients) != 3 {
		t.Errorf("Expected 3 clients, got %d", len(clients))
	}
}

func TestClientService_List_Pagination(t *testing.T) {
	db, cleanup := setupClientServiceTestDB(t)
	defer cleanup()

	svc := newTestClientService(db)

	for i := 0; i < 5; i++ {
		redirectURIs := "https://example.com/callback"
		client := &model.Client{
			ID:           "client-" + string(rune('a'+i)),
			ClientID:     "client-id-" + string(rune('a'+i)),
			Name:         "Client " + string(rune('A'+i)),
			RedirectURIs: redirectURIs,
		}
		svc.Create(client)
	}

	clients, total, err := svc.List(0, 2)
	if err != nil {
		t.Fatalf("Failed to list clients: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(clients))
	}
}

func TestClientService_List_Empty(t *testing.T) {
	db, cleanup := setupClientServiceTestDB(t)
	defer cleanup()

	svc := newTestClientService(db)

	clients, total, err := svc.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list clients: %v", err)
	}

	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}

	if len(clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(clients))
	}
}
