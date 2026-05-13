package database

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(cfg *config.Config) (*gorm.DB, error) {
	dbPath := cfg.Database.Path

	if err := ensureDir(dbPath); err != nil {
		return nil, fmt.Errorf("failed to ensure database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.Exec("PRAGMA journal_mode=DELETE")
	sqlDB.Exec("PRAGMA foreign_keys=ON")
	sqlDB.Exec("PRAGMA synchronous=NORMAL")
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func ensureDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

type Migration struct {
	Version   string
	Name      string
	AppliedAt string
}

func RunMigrations(db *gorm.DB) error {
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	for _, file := range migrationFiles {
		version := extractVersion(file)
		if version == "" {
			continue
		}

		applied, err := isMigrationApplied(db, version)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if applied {
			continue
		}

		sql, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if err := db.Exec(string(sql)).Error; err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		if err := recordMigration(db, version, filepath.Base(file)); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		fmt.Printf("Applied migration: %s\n", version)
	}

	return nil
}

func createMigrationsTable(db *gorm.DB) error {
	sql := `
	CREATE TABLE IF NOT EXISTS migrations (
		version TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	return db.Exec(sql).Error
}

func getMigrationFiles() ([]string, error) {
	var files []string
	entries, err := os.ReadDir("migrations")
	if err != nil {
		if os.IsNotExist(err) {
			return files, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, filepath.Join("migrations", name))
		}
	}

	sort.Strings(files)
	return files, nil
}

func extractVersion(filename string) string {
	base := filepath.Base(filename)
	parts := strings.Split(base, "_")
	if len(parts) > 0 {
		return strings.TrimSuffix(parts[0], ".sql")
	}
	return ""
}

func isMigrationApplied(db *gorm.DB, version string) (bool, error) {
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM migrations WHERE version = ?", version).Scan(&count).Error
	return count > 0, err
}

func recordMigration(db *gorm.DB, version, name string) error {
	return db.Exec("INSERT INTO migrations (version, name) VALUES (?, ?)", version, name).Error
}

var _ embed.FS
