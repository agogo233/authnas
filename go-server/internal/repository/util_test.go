package repository

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
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

	if err := db.AutoMigrate(&model.User{}, &model.Group{}, &model.UserGroup{}, &model.Client{}, &model.Consent{}, &model.Passkey{}, &model.PasskeyAuthOptions{}, &model.TOTP{}, &model.Key{}, &model.Invitation{}, &model.OIDCPayload{}, &model.EmailVerification{}, &model.PasswordReset{}, &model.ProxyAuth{}, &model.EmailLog{}); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	return db
}

func CleanupDB(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}
}

func StringPtr(s string) *string {
	return &s
}
