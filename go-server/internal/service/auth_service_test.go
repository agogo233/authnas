package service

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/pkg/utils"
)

func TestAuthService_HashPassword(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	randomUtil := utils.NewRandom()
	authService := &AuthService{
		cfg:    cfg,
		random: randomUtil,
	}

	password := "testpassword123"
	hash, err := authService.HashPassword(password)

	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == password {
		t.Error("Hash should not equal original password")
	}

	if len(hash) < 16 {
		t.Error("Hash should be longer than 16 bytes (includes salt)")
	}
}

func TestAuthService_VerifyPassword(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	randomUtil := utils.NewRandom()
	authService := &AuthService{
		cfg:    cfg,
		random: randomUtil,
	}

	password := "testpassword123"
	hash, _ := authService.HashPassword(password)

	if !authService.VerifyPassword(hash, password) {
		t.Error("Password verification should succeed for correct password")
	}

	if authService.VerifyPassword(hash, "wrongpassword") {
		t.Error("Password verification should fail for wrong password")
	}
}

func TestAuthService_VerifyPassword_ShortHash(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	randomUtil := utils.NewRandom()
	authService := &AuthService{
		cfg:    cfg,
		random: randomUtil,
	}

	if authService.VerifyPassword("short", "password") {
		t.Error("Should return false for hash shorter than 16 bytes")
	}
}

func TestAuthService_IsLoginLockedOut_NoAttempts(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}
	ResetLoginAttempts()
	authService := &AuthService{
		cfg: cfg,
	}

	locked := authService.IsLoginLockedOut("user1", "192.168.1.1")
	if locked {
		t.Error("User with no login attempts should not be locked out")
	}
}

func TestAuthService_RecordFailedLogin(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}
	ResetLoginAttempts()
	authService := &AuthService{
		cfg: cfg,
	}

	authService.RecordFailedLogin("user1", "192.168.1.1")
	locked := authService.IsLoginLockedOut("user1", "192.168.1.1")
	if locked {
		t.Error("User with 1 failed attempt should not be locked out")
	}

	authService.RecordFailedLogin("user1", "192.168.1.1")
	authService.RecordFailedLogin("user1", "192.168.1.1")
	authService.RecordFailedLogin("user1", "192.168.1.1")
	locked = authService.IsLoginLockedOut("user1", "192.168.1.1")
	if locked {
		t.Error("User with 4 failed attempts should not be locked out")
	}

	authService.RecordFailedLogin("user1", "192.168.1.1")
	locked = authService.IsLoginLockedOut("user1", "192.168.1.1")
	if !locked {
		t.Error("User with 5 failed attempts should be locked out")
	}
}

func TestAuthService_RecordSuccessfulLogin(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}
	ResetLoginAttempts()
	authService := &AuthService{
		cfg: cfg,
	}

	authService.RecordFailedLogin("user1", "192.168.1.1")
	authService.RecordFailedLogin("user1", "192.168.1.1")
	authService.RecordFailedLogin("user1", "192.168.1.1")

	authService.RecordSuccessfulLogin("user1", "192.168.1.1")

	locked := authService.IsLoginLockedOut("user1", "192.168.1.1")
	if locked {
		t.Error("User after successful login should not be locked out")
	}
}

func TestAuthService_GetLoginLockoutRemainingTime_NoLockout(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}
	ResetLoginAttempts()
	authService := &AuthService{
		cfg: cfg,
	}

	remaining := authService.GetLoginLockoutRemainingTime("user1", "192.168.1.1")
	if remaining != 0 {
		t.Errorf("Expected 0 remaining time, got %v", remaining)
	}
}

func TestAuthService_ResetLoginAttemptsMap(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}
	ResetLoginAttempts()
	authService := &AuthService{
		cfg: cfg,
	}

	authService.RecordFailedLogin("user1", "192.168.1.1")
	authService.RecordFailedLogin("user1", "192.168.1.1")
	authService.RecordFailedLogin("user1", "192.168.1.1")
	authService.RecordFailedLogin("user1", "192.168.1.1")
	authService.RecordFailedLogin("user1", "192.168.1.1")

	locked := authService.IsLoginLockedOut("user1", "192.168.1.1")
	if !locked {
		t.Error("User should be locked out after 5 failed attempts")
	}

	authService.ResetLoginAttemptsMap()

	locked = authService.IsLoginLockedOut("user1", "192.168.1.1")
	if locked {
		t.Error("User should not be locked out after reset")
	}
}

func TestAuthService_GetPublicKey_NoPrivateKey(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
		JWT: config.JWTConfig{
			PrivateKey: "",
		},
	}

	authService := &AuthService{
		cfg: cfg,
	}

	pubKey := authService.GetPublicKey()
	if pubKey != nil {
		t.Error("Public key should be nil when no private key is configured")
	}
}

func TestAuthService_Constants(t *testing.T) {
	if argon2Memory != 64*1024 {
		t.Errorf("Expected argon2Memory to be 65536, got %d", argon2Memory)
	}
	if argon2Iterations != 3 {
		t.Errorf("Expected argon2Iterations to be 3, got %d", argon2Iterations)
	}
	if argon2Parallelism != 4 {
		t.Errorf("Expected argon2Parallelism to be 4, got %d", argon2Parallelism)
	}
	if argon2SaltLength != 16 {
		t.Errorf("Expected argon2SaltLength to be 16, got %d", argon2SaltLength)
	}
	if argon2KeyLength != 32 {
		t.Errorf("Expected argon2KeyLength to be 32, got %d", argon2KeyLength)
	}
}
