package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/authnas/authnas/go-server/internal/config"
	"gorm.io/gorm"
)

func getSQLDB(t *testing.T, db *gorm.DB) *sql.DB {
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	return sqlDB
}

func TestNewDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer getSQLDB(t, db).Close()

	sqlDB := getSQLDB(t, db)

	var journalMode string
	if err := sqlDB.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("Failed to query journal_mode: %v", err)
	}

	if journalMode != "delete" {
		t.Errorf("Expected journal_mode to be 'delete', got '%s'", journalMode)
	}
}

func TestRunMigrations_InProjectContext(t *testing.T) {
	if os.Getenv("RUN_MIGRATION_TESTS") != "true" {
		t.Skip("Skipping migration tests that require project root (migrations directory)")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer getSQLDB(t, db).Close()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	var tableCount int
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&tableCount).Error; err != nil {
		t.Fatalf("Failed to count tables: %v", err)
	}

	if tableCount < 10 {
		t.Errorf("Expected at least 10 tables, got %d", tableCount)
	}
}

func TestMigrationsTableCreated_InProjectContext(t *testing.T) {
	if os.Getenv("RUN_MIGRATION_TESTS") != "true" {
		t.Skip("Skipping migration tests that require project root (migrations directory)")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer getSQLDB(t, db).Close()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM migrations").Scan(&count).Error; err != nil {
		t.Fatalf("Failed to query migrations table: %v", err)
	}

	if count == 0 {
		t.Error("Expected at least one migration record, got 0")
	}
}

func TestDatabaseAutoMigration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer getSQLDB(t, db).Close()

	type TestUser struct {
		ID        string `gorm:"primaryKey"`
		Username  string
		Email     string
		CreatedAt string
	}

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	var tableExists int
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_users'").Scan(&tableExists).Error; err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}

	if tableExists != 1 {
		t.Error("Expected test_users table to exist after AutoMigrate")
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "subsubdir", "test.db")

	err := ensureDir(dbPath)
	if err != nil {
		t.Errorf("ensureDir returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}
