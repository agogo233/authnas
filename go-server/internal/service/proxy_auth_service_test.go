package service

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"gorm.io/gorm"
)

func setupProxyAuthServiceTestDB(t *testing.T) (*gorm.DB, func()) {
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

	if err := db.AutoMigrate(&model.User{}, &model.ProxyAuth{}); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func newTestProxyAuthService(db *gorm.DB) *ProxyAuthService {
	cfg := &config.Config{}
	proxyAuthRepo := repository.NewProxyAuthRepository(db)
	userRepo := repository.NewUserRepository(db)
	return NewProxyAuthService(cfg, proxyAuthRepo, userRepo)
}

func TestProxyAuthService_Create(t *testing.T) {
	db, cleanup := setupProxyAuthServiceTestDB(t)
	defer cleanup()

	svc := newTestProxyAuthService(db)

	email := "test@example.com"
	user := &model.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    &email,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	pa := &model.ProxyAuth{
		ID:         "proxy-auth-1",
		Name:       "Test Proxy",
		ProxyURL:   "https://proxy.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
	}

	err := svc.Create(pa)
	if err != nil {
		t.Fatalf("Failed to create proxy auth: %v", err)
	}

	if pa.ID == "" {
		t.Error("ProxyAuth ID should be set")
	}
}

func TestProxyAuthService_GetByID(t *testing.T) {
	db, cleanup := setupProxyAuthServiceTestDB(t)
	defer cleanup()

	svc := newTestProxyAuthService(db)

	email := "test@example.com"
	user := &model.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    &email,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	pa := &model.ProxyAuth{
		ID:         "proxy-auth-1",
		Name:       "Test Proxy",
		ProxyURL:   "https://proxy.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
	}
	svc.Create(pa)

	found, err := svc.GetByID(pa.ID)
	if err != nil {
		t.Fatalf("Failed to get proxy auth by ID: %v", err)
	}

	if found.ID != pa.ID {
		t.Errorf("Expected ID '%s', got '%s'", pa.ID, found.ID)
	}
}

func TestProxyAuthService_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupProxyAuthServiceTestDB(t)
	defer cleanup()

	svc := newTestProxyAuthService(db)

	_, err := svc.GetByID("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent proxy auth")
	}
}

func TestProxyAuthService_GetEnabled(t *testing.T) {
	db, cleanup := setupProxyAuthServiceTestDB(t)
	defer cleanup()

	svc := newTestProxyAuthService(db)

	pa1 := &model.ProxyAuth{
		ID:         "proxy-auth-1",
		Name:       "Enabled Proxy",
		ProxyURL:   "https://proxy1.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
	}
	pa2 := &model.ProxyAuth{
		ID:         "proxy-auth-2",
		Name:       "Another Proxy",
		ProxyURL:   "https://proxy2.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
	}
	svc.Create(pa1)
	svc.Create(pa2)

	enabled, err := svc.GetEnabled()
	if err != nil {
		t.Fatalf("Failed to get enabled proxy auth: %v", err)
	}

	if len(enabled) != 2 {
		t.Errorf("Expected 2 proxy auth entries, got %d", len(enabled))
	}
}

func TestProxyAuthService_Update(t *testing.T) {
	db, cleanup := setupProxyAuthServiceTestDB(t)
	defer cleanup()

	svc := newTestProxyAuthService(db)

	pa := &model.ProxyAuth{
		ID:         "proxy-auth-1",
		Name:       "Original Name",
		ProxyURL:   "https://proxy.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
	}
	svc.Create(pa)

	pa.Name = "Updated Name"
	err := svc.Update(pa)
	if err != nil {
		t.Fatalf("Failed to update proxy auth: %v", err)
	}

	updated, _ := svc.GetByID(pa.ID)
	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
	}
}

func TestProxyAuthService_Delete(t *testing.T) {
	db, cleanup := setupProxyAuthServiceTestDB(t)
	defer cleanup()

	svc := newTestProxyAuthService(db)

	pa := &model.ProxyAuth{
		ID:         "proxy-auth-1",
		Name:       "Test Proxy",
		ProxyURL:   "https://proxy.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
	}
	svc.Create(pa)

	err := svc.Delete(pa.ID)
	if err != nil {
		t.Fatalf("Failed to delete proxy auth: %v", err)
	}

	_, err = svc.GetByID(pa.ID)
	if err == nil {
		t.Error("Expected error after deleting proxy auth")
	}
}

func TestProxyAuthService_List(t *testing.T) {
	db, cleanup := setupProxyAuthServiceTestDB(t)
	defer cleanup()

	svc := newTestProxyAuthService(db)

	for i := 0; i < 3; i++ {
		pa := &model.ProxyAuth{
			ID:         "proxy-auth-" + string(rune('a'+i)),
			Name:       "Proxy " + string(rune('A'+i)),
			ProxyURL:   "https://proxy.example.com",
			HeaderName: "X-User-ID",
			Enabled:    true,
		}
		svc.Create(pa)
	}

	list, total, err := svc.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list proxy auth: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 proxy auth entries, got %d", len(list))
	}
}

func TestProxyAuthService_List_Pagination(t *testing.T) {
	db, cleanup := setupProxyAuthServiceTestDB(t)
	defer cleanup()

	svc := newTestProxyAuthService(db)

	for i := 0; i < 5; i++ {
		pa := &model.ProxyAuth{
			ID:         "proxy-auth-" + string(rune('a'+i)),
			Name:       "Proxy " + string(rune('A'+i)),
			ProxyURL:   "https://proxy.example.com",
			HeaderName: "X-User-ID",
			Enabled:    true,
		}
		svc.Create(pa)
	}

	list, total, err := svc.List(0, 2)
	if err != nil {
		t.Fatalf("Failed to list proxy auth: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 proxy auth entries, got %d", len(list))
	}
}

func TestProxyAuthService_List_Empty(t *testing.T) {
	db, cleanup := setupProxyAuthServiceTestDB(t)
	defer cleanup()

	svc := newTestProxyAuthService(db)

	list, total, err := svc.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list proxy auth: %v", err)
	}

	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}

	if len(list) != 0 {
		t.Errorf("Expected 0 proxy auth entries, got %d", len(list))
	}
}
