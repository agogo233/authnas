package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
app:
  url: "http://localhost:8080"
  name: "TestAuthNas"
  environment: "test"

database:
  path: "./test.db"

security:
  storage_key: "test-storage-key-at-least-32-chars"
  password_strength: 3
  mfa_required: false
  signup_requires_approval: false
  email_verification: false

email:
  enabled: false
  smtp_host: "localhost"
  smtp_port: 587

rate_limit:
  enabled: true
  requests_per_minute: 60

jwt:
  access_token_expiry: "15m"
  refresh_token_expiry: "168h"

cors:
  allowed_origins:
    - "http://localhost:3000"
  allow_credentials: true
  allowed_methods:
    - "GET"
    - "POST"
  allowed_headers:
    - "Content-Type"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.App.URL != "http://localhost:8080" {
		t.Errorf("Expected App.URL to be 'http://localhost:8080', got '%s'", cfg.App.URL)
	}

	if cfg.App.Name != "TestAuthNas" {
		t.Errorf("Expected App.Name to be 'TestAuthNas', got '%s'", cfg.App.Name)
	}

	if cfg.Database.Path != "./test.db" {
		t.Errorf("Expected Database.Path to be './test.db', got '%s'", cfg.Database.Path)
	}

	if cfg.Security.StorageKey != "test-storage-key-at-least-32-chars" {
		t.Errorf("Expected StorageKey to be set correctly, got '%s'", cfg.Security.StorageKey)
	}

	if cfg.Security.MFARequired != false {
		t.Errorf("Expected MFARequired to be false, got %v", cfg.Security.MFARequired)
	}

	if cfg.RateLimit.Enabled != true {
		t.Errorf("Expected RateLimit.Enabled to be true, got %v", cfg.RateLimit.Enabled)
	}

	if cfg.RateLimit.RequestsPerMinute != 60 {
		t.Errorf("Expected RateLimit.RequestsPerMinute to be 60, got %d", cfg.RateLimit.RequestsPerMinute)
	}

	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("Expected CORS.AllowedOrigins to contain 'http://localhost:3000', got %v", cfg.CORS.AllowedOrigins)
	}
}

func TestConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
app:
  url: "http://localhost:8080"
  name: "TestAuthNas"
  environment: "test"

database:
  path: "./test.db"

security:
  storage_key: "default-test-key-at-least-32-characters-long"
  password_strength: 3
  mfa_required: false
  signup_requires_approval: false
  email_verification: false

email:
  enabled: false
  smtp_host: "localhost"
  smtp_port: 587

rate_limit:
  enabled: true
  requests_per_minute: 60

jwt:
  access_token_expiry: "15m"
  refresh_token_expiry: "168h"

cors:
  allowed_origins:
    - "http://localhost:3000"
  allow_credentials: true
  allowed_methods:
    - "GET"
    - "POST"
  allowed_headers:
    - "Content-Type"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.App.URL != "http://localhost:8080" {
		t.Errorf("Expected App.URL to be 'http://localhost:8080', got '%s'", cfg.App.URL)
	}

	if cfg.App.Name != "TestAuthNas" {
		t.Errorf("Expected App.Name to be 'TestAuthNas', got '%s'", cfg.App.Name)
	}

	if cfg.Database.Path != "./test.db" {
		t.Errorf("Expected Database.Path to be './test.db', got '%s'", cfg.Database.Path)
	}

	if cfg.Security.PasswordStrength != 3 {
		t.Errorf("Expected PasswordStrength to be 3, got %d", cfg.Security.PasswordStrength)
	}

	if cfg.RateLimit.Enabled != true {
		t.Errorf("Expected RateLimit.Enabled to be true, got %v", cfg.RateLimit.Enabled)
	}

	if cfg.RateLimit.RequestsPerMinute != 60 {
		t.Errorf("Expected RateLimit.RequestsPerMinute to be 60, got %d", cfg.RateLimit.RequestsPerMinute)
	}
}
