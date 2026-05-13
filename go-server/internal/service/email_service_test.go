package service

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"gorm.io/gorm"
)

func setupEmailServiceTestDB(t *testing.T) (*gorm.DB, func()) {
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

	if err := db.AutoMigrate(&model.User{}, &model.EmailLog{}); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func newTestEmailService(db *gorm.DB) *EmailService {
	cfg := &config.Config{
		App: config.AppConfig{
			Name: "TestApp",
			URL:  "https://testapp.example.com",
		},
	}
	emailVerificationRepo := repository.NewEmailVerificationRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	emailLogRepo := repository.NewEmailLogRepository(db)
	return NewEmailService(cfg, emailVerificationRepo, passwordResetRepo, emailLogRepo, nil)
}

func TestEmailService_SendVerificationEmail_NilUserEmail(t *testing.T) {
	db, cleanup := setupEmailServiceTestDB(t)
	defer cleanup()

	svc := newTestEmailService(db)

	user := &model.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    nil,
	}

	err := svc.SendVerificationEmail(user, "123456")
	if err == nil {
		t.Error("Expected error for nil user email")
	}

	if err.Error() != "user email is nil" {
		t.Errorf("Expected 'user email is nil', got '%s'", err.Error())
	}
}

func TestEmailService_SendPasswordResetEmail_NilUserEmail(t *testing.T) {
	db, cleanup := setupEmailServiceTestDB(t)
	defer cleanup()

	svc := newTestEmailService(db)

	user := &model.User{
		ID:       "user-1",
		Username: "testuser",
		Email:    nil,
	}

	err := svc.SendPasswordResetEmail(user, "123456")
	if err == nil {
		t.Error("Expected error for nil user email")
	}

	if err.Error() != "user email is nil" {
		t.Errorf("Expected 'user email is nil', got '%s'", err.Error())
	}
}

func TestEmailService_renderTemplate(t *testing.T) {
	db, cleanup := setupEmailServiceTestDB(t)
	defer cleanup()

	svc := newTestEmailService(db)

	data := EmailTemplateData{
		AppName:  "TestApp",
		UserName: "testuser",
		Link:     "https://testapp.example.com/verify?code=123",
		Code:     "123456",
	}

	html, err := svc.renderTemplate("verification", data)
	if err != nil {
		t.Fatalf("renderTemplate failed: %v", err)
	}

	if html == "" {
		t.Error("HTML should not be empty")
	}
}

func TestEmailService_renderTemplate_PasswordReset(t *testing.T) {
	db, cleanup := setupEmailServiceTestDB(t)
	defer cleanup()

	svc := newTestEmailService(db)

	data := EmailTemplateData{
		AppName:  "TestApp",
		UserName: "testuser",
		Link:     "https://testapp.example.com/reset?code=123",
		Code:     "123456",
	}

	html, err := svc.renderTemplate("password_reset", data)
	if err != nil {
		t.Fatalf("renderTemplate failed: %v", err)
	}

	if html == "" {
		t.Error("HTML should not be empty")
	}
}

func TestEmailService_renderTemplate_UnknownTemplate(t *testing.T) {
	db, cleanup := setupEmailServiceTestDB(t)
	defer cleanup()

	svc := newTestEmailService(db)

	data := EmailTemplateData{
		AppName:  "TestApp",
		UserName: "testuser",
	}

	html, err := svc.renderTemplate("unknown_template", data)
	if err != nil {
		t.Fatalf("renderTemplate should not fail for unknown template: %v", err)
	}

	if html == "" {
		t.Error("HTML should not be empty even for unknown template")
	}
}

func TestEmailService_logEmail_NilRepo(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name: "TestApp",
			URL:  "https://testapp.example.com",
		},
	}

	svc := &EmailService{
		cfg:                   cfg,
		emailVerificationRepo: nil,
		passwordResetRepo:     nil,
		emailLogRepo:          nil,
		sender:                nil,
	}

	svc.logEmail("test@example.com", "Test Subject", "test_template", "sent", "")
}

func TestEmailService_logEmail_WithError(t *testing.T) {
	db, cleanup := setupEmailServiceTestDB(t)
	defer cleanup()

	svc := newTestEmailService(db)

	svc.logEmail("test@example.com", "Test Subject", "test_template", "failed", "connection timeout")

	emailLogRepo := repository.NewEmailLogRepository(db)
	logs, _, err := emailLogRepo.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list email logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 email log, got %d", len(logs))
	}

	if logs[0].Recipient != "test@example.com" {
		t.Errorf("Expected recipient 'test@example.com', got '%s'", logs[0].Recipient)
	}

	if logs[0].Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", logs[0].Status)
	}

	if logs[0].Error == nil || *logs[0].Error != "connection timeout" {
		t.Error("Error message mismatch")
	}
}

func TestEmailService_getEmailTemplate(t *testing.T) {
	db, cleanup := setupEmailServiceTestDB(t)
	defer cleanup()

	svc := newTestEmailService(db)

	tmpl := svc.getEmailTemplate("verification")
	if tmpl == nil {
		t.Error("Template should not be nil")
	}

	tmpl = svc.getEmailTemplate("password_reset")
	if tmpl == nil {
		t.Error("Template should not be nil")
	}

	tmpl = svc.getEmailTemplate("unknown")
	if tmpl == nil {
		t.Error("Template should not be nil even for unknown name")
	}
}

func TestEmailTemplateData_Structure(t *testing.T) {
	data := EmailTemplateData{
		AppName:  "MyApp",
		UserName: "john",
		Link:     "https://example.com/verify",
		Code:     "ABC123",
	}

	if data.AppName != "MyApp" {
		t.Errorf("Expected AppName 'MyApp', got '%s'", data.AppName)
	}
	if data.UserName != "john" {
		t.Errorf("Expected UserName 'john', got '%s'", data.UserName)
	}
	if data.Link != "https://example.com/verify" {
		t.Errorf("Expected Link 'https://example.com/verify', got '%s'", data.Link)
	}
	if data.Code != "ABC123" {
		t.Errorf("Expected Code 'ABC123', got '%s'", data.Code)
	}
}
