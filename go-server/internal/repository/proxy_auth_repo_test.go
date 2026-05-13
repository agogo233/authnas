package repository

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/model"
)

func TestProxyAuthRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewProxyAuthRepository(db)

	pa := &model.ProxyAuth{
		ID:         "proxy-1",
		Name:       "Test Proxy",
		ProxyURL:   "https://proxy.example.com",
		HeaderName: "X-User-ID",
		Enabled:    true,
	}

	if err := repo.Create(pa); err != nil {
		t.Fatalf("Failed to create ProxyAuth: %v", err)
	}

	retrieved, err := repo.GetByID("proxy-1")
	if err != nil {
		t.Fatalf("Failed to get ProxyAuth: %v", err)
	}

	if retrieved.Name != "Test Proxy" {
		t.Errorf("Expected name 'Test Proxy', got '%s'", retrieved.Name)
	}
}

func TestProxyAuthRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewProxyAuthRepository(db)

	pa := &model.ProxyAuth{
		ID:         "proxy-get-by-id",
		Name:       "Get By ID Proxy",
		ProxyURL:   "https://get.example.com",
		HeaderName: "X-Proxy-ID",
		Enabled:    true,
	}

	repo.Create(pa)

	retrieved, err := repo.GetByID("proxy-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get ProxyAuth by ID: %v", err)
	}

	if retrieved.ProxyURL != "https://get.example.com" {
		t.Errorf("Expected ProxyURL 'https://get.example.com', got '%s'", retrieved.ProxyURL)
	}
}

func TestProxyAuthRepository_GetEnabled(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewProxyAuthRepository(db)

	enabledProxy := &model.ProxyAuth{
		ID:         "proxy-enabled",
		Name:       "Enabled Proxy",
		ProxyURL:   "https://enabled.example.com",
		HeaderName: "X-Enabled",
		Enabled:    true,
	}

	if err := repo.db.Create(enabledProxy).Error; err != nil {
		t.Fatalf("Failed to create enabled proxy: %v", err)
	}

	disabledProxy := &model.ProxyAuth{
		ID:         "proxy-disabled",
		Name:       "Disabled Proxy",
		ProxyURL:   "https://disabled.example.com",
		HeaderName: "X-Disabled",
		Enabled:    false,
	}

	if err := repo.db.Exec(
		"INSERT INTO proxy_auth (id, name, proxy_url, header_name, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))",
		disabledProxy.ID, disabledProxy.Name, disabledProxy.ProxyURL, disabledProxy.HeaderName, false,
	).Error; err != nil {
		t.Fatalf("Failed to create disabled proxy: %v", err)
	}

	enabled, err := repo.GetEnabled()
	if err != nil {
		t.Fatalf("Failed to get enabled ProxyAuth: %v", err)
	}

	if len(enabled) != 1 {
		t.Errorf("Expected 1 enabled ProxyAuth, got %d", len(enabled))
	}

	if enabled[0].ID != "proxy-enabled" {
		t.Errorf("Expected enabled ProxyAuth ID 'proxy-enabled', got '%s'", enabled[0].ID)
	}
}

func TestProxyAuthRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewProxyAuthRepository(db)

	pa := &model.ProxyAuth{
		ID:         "proxy-update",
		Name:       "Old Name",
		ProxyURL:   "https://old.example.com",
		HeaderName: "X-Old",
		Enabled:    true,
	}

	repo.Create(pa)

	pa.Name = "New Name"
	pa.ProxyURL = "https://new.example.com"
	if err := repo.Update(pa); err != nil {
		t.Fatalf("Failed to update ProxyAuth: %v", err)
	}

	retrieved, _ := repo.GetByID("proxy-update")
	if retrieved.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", retrieved.Name)
	}
}

func TestProxyAuthRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewProxyAuthRepository(db)

	pa := &model.ProxyAuth{
		ID:         "proxy-delete",
		Name:       "Delete Proxy",
		ProxyURL:   "https://delete.example.com",
		HeaderName: "X-Delete",
		Enabled:    true,
	}

	repo.Create(pa)

	if err := repo.Delete("proxy-delete"); err != nil {
		t.Fatalf("Failed to delete ProxyAuth: %v", err)
	}

	_, err := repo.GetByID("proxy-delete")
	if err == nil {
		t.Error("Expected error when getting deleted ProxyAuth")
	}
}

func TestProxyAuthRepository_List(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewProxyAuthRepository(db)

	for i := 1; i <= 3; i++ {
		pa := &model.ProxyAuth{
			ID:         "proxy-list-" + string(rune('0'+i)),
			Name:       "List Proxy",
			ProxyURL:   "https://list.example.com",
			HeaderName: "X-List",
			Enabled:    true,
		}
		repo.Create(pa)
	}

	proxies, total, err := repo.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list ProxyAuth: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}

	if len(proxies) != 3 {
		t.Errorf("Expected 3 proxies, got %d", len(proxies))
	}
}

func TestEmailLogRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewEmailLogRepository(db)

	log := &model.EmailLog{
		ID:        "log-1",
		Recipient: "user@example.com",
		Subject:   "Test Subject",
		Template:  "test_template",
		Status:    "sent",
	}

	if err := repo.Create(log); err != nil {
		t.Fatalf("Failed to create EmailLog: %v", err)
	}

	retrieved, err := repo.GetByID("log-1")
	if err != nil {
		t.Fatalf("Failed to get EmailLog: %v", err)
	}

	if retrieved.Recipient != "user@example.com" {
		t.Errorf("Expected recipient 'user@example.com', got '%s'", retrieved.Recipient)
	}
}

func TestEmailLogRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewEmailLogRepository(db)

	log := &model.EmailLog{
		ID:        "log-get-by-id",
		Recipient: "test@example.com",
		Subject:   "Get By ID",
		Template:  "test",
		Status:    "delivered",
	}

	repo.Create(log)

	retrieved, err := repo.GetByID("log-get-by-id")
	if err != nil {
		t.Fatalf("Failed to get EmailLog by ID: %v", err)
	}

	if retrieved.Status != "delivered" {
		t.Errorf("Expected status 'delivered', got '%s'", retrieved.Status)
	}
}

func TestEmailLogRepository_List(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewEmailLogRepository(db)

	for i := 1; i <= 5; i++ {
		log := &model.EmailLog{
			ID:        "log-list-" + string(rune('0'+i)),
			Recipient: "user@example.com",
			Subject:   "Test",
			Template:  "test",
			Status:    "sent",
		}
		repo.Create(log)
	}

	logs, total, err := repo.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list EmailLog: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(logs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(logs))
	}
}

func TestEmailLogRepository_DeleteOlderThan(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(db)

	repo := NewEmailLogRepository(db)

	oldLog := &model.EmailLog{
		ID:        "log-old",
		Recipient: "old@example.com",
		Subject:   "Old",
		Template:  "test",
		Status:    "sent",
	}

	if err := repo.db.Exec(
		"INSERT INTO email_log (id, recipient, subject, template, status, created_at) VALUES (?, ?, ?, ?, ?, datetime('now', '-60 days'))",
		oldLog.ID, oldLog.Recipient, oldLog.Subject, oldLog.Template, oldLog.Status,
	).Error; err != nil {
		t.Fatalf("Failed to create old log: %v", err)
	}

	newLog := &model.EmailLog{
		ID:        "log-new",
		Recipient: "new@example.com",
		Subject:   "New",
		Template:  "test",
		Status:    "sent",
	}
	repo.Create(newLog)

	if err := repo.DeleteOlderThan(30); err != nil {
		t.Fatalf("Failed to delete old EmailLogs: %v", err)
	}

	_, err := repo.GetByID("log-old")
	if err == nil {
		t.Error("Expected old log to be deleted")
	}

	_, err = repo.GetByID("log-new")
	if err != nil {
		t.Error("Expected new log to still exist")
	}
}
