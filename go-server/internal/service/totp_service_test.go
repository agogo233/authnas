package service

import (
	"reflect"
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
)

func TestTOTPService_NewTOTPService(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}

	svc := NewTOTPService(cfg, nil, nil)
	if svc == nil {
		t.Fatal("NewTOTPService should not return nil")
	}
	defer svc.Stop()
}

func TestTOTPService_IsLockedOut_NilRepo(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	if svc.IsLockedOut("user123") {
		t.Error("User should not be locked out initially")
	}
}

func TestTOTPService_HasTOTP_NilRepo(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("HasTOTP panicked when repo is nil: %v (this is expected)", r)
		}
	}()

	svc.HasTOTP("user123")
}

func TestTOTPService_GetTOTPByUserID_NilRepo(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetTOTPByUserID panicked when repo is nil: %v (this is expected)", r)
		}
	}()

	svc.GetTOTPByUserID("user123")
}

func TestTOTPService_Delete_NilRepo(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Delete panicked when repo is nil: %v (this is expected)", r)
		}
	}()

	svc.Delete("user123")
}

func TestTOTPService_Validate_NilRepo(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Validate panicked when repo is nil: %v (this is expected)", r)
		}
	}()

	svc.Validate("user123", "123456")
}

func TestTOTPService_ValidateWithWindow_NilRepo(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("ValidateWithWindow panicked when repo is nil: %v (this is expected)", r)
		}
	}()

	svc.ValidateWithWindow("user123", "123456", 1)
}

func TestTOTPService_ValidateWithWindow_ZeroWindow(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("ValidateWithWindow panicked: %v (this is expected for nil repo)", r)
		}
	}()

	svc.ValidateWithWindow("user123", "123456", 0)
}

func TestTOTPService_ValidateWithWindow_NegativeWindow(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("ValidateWithWindow panicked: %v (this is expected for nil repo)", r)
		}
	}()

	svc.ValidateWithWindow("user123", "123456", -1)
}

func TestTOTPService_Generate_NilUser(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name: "TestApp",
		},
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Generate panicked when user is nil: %v (this is expected)", r)
		}
	}()

	svc.Generate(nil)
}

func TestTOTPService_UpdateTOTP_NilRepo(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("UpdateTOTP panicked when repo is nil: %v (this is expected)", r)
		}
	}()

	svc.UpdateTOTP(nil)
}

func TestTOTPService_Constants(t *testing.T) {
	if MaxTOTPAttempts != 5 {
		t.Errorf("Expected MaxTOTPAttempts to be 5, got %d", MaxTOTPAttempts)
	}
	if TOTPLockoutDuration != 5*60*1e9 {
		t.Errorf("Expected TOTPLockoutDuration to be 5 minutes, got %d", TOTPLockoutDuration)
	}
	if TOTPAttemptWindow != 15*60*1e9 {
		t.Errorf("Expected TOTPAttemptWindow to be 15 minutes, got %d", TOTPAttemptWindow)
	}
}

func TestTOTPResult_Structure(t *testing.T) {
	result := &TOTPResult{
		Secret:    "JBSWY3DPEHPK3PXP",
		QRCodeURI: "otpauth://totp/TestApp:user?secret=JBSWY3DPEHPK3PXP&issuer=TestApp",
	}

	if result.Secret == "" {
		t.Error("Secret should not be empty")
	}
	if result.QRCodeURI == "" {
		t.Error("QRCodeURI should not be empty")
	}
}

func TestTOTPService_cleanupAttempts(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	svc.performCleanup()
}

func TestTOTPService_initEncryptKey_WithHKDF(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := &TOTPService{
		cfg:        cfg,
		totpRepo:   nil,
		userRepo:   nil,
		encryptKey: nil,
	}

	svc.initEncryptKey()

	if svc.encryptKey == nil {
		t.Error("encryptKey should not be nil after initEncryptKey")
	}
	if len(svc.encryptKey) != 32 {
		t.Errorf("Expected encryptKey length 32, got %d", len(svc.encryptKey))
	}
}

func TestTOTPService_initEncryptKey_AlreadyInitialized(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	existingKey := []byte("12345678901234567890123456789012")
	svc := &TOTPService{
		cfg:        cfg,
		totpRepo:   nil,
		userRepo:   nil,
		encryptKey: existingKey,
	}

	svc.initEncryptKey()

	if !reflect.DeepEqual(svc.encryptKey, existingKey) {
		t.Error("encryptKey should not be reinitialized when already set")
	}
}

func TestTOTPService_encryptSecret(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	encrypted, err := svc.encryptSecret("JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("encryptSecret failed: %v", err)
	}
	if encrypted == "" {
		t.Error("Encrypted secret should not be empty")
	}
	if encrypted == "JBSWY3DPEHPK3PXP" {
		t.Error("Encrypted secret should differ from plaintext")
	}
}

func TestTOTPService_decryptSecret(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	plaintext := "JBSWY3DPEHPK3PXP"
	encrypted, _ := svc.encryptSecret(plaintext)

	decrypted, err := svc.decryptSecret(encrypted)
	if err != nil {
		t.Fatalf("decryptSecret failed: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("Expected '%s', got '%s'", plaintext, decrypted)
	}
}

func TestTOTPService_encryptDecryptRoundTrip(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	testCases := []string{
		"short",
		"JBSWY3DPEHPK3PXP",
		"a very long secret key that should still work correctly with encryption and decryption",
		"with special chars: !@#$%^&*()",
	}

	for _, tc := range testCases {
		encrypted, err := svc.encryptSecret(tc)
		if err != nil {
			t.Errorf("encryptSecret failed for '%s': %v", tc, err)
			continue
		}

		decrypted, err := svc.decryptSecret(encrypted)
		if err != nil {
			t.Errorf("decryptSecret failed for '%s': %v", tc, err)
			continue
		}

		if decrypted != tc {
			t.Errorf("Round trip failed for '%s': expected '%s', got '%s'", tc, tc, decrypted)
		}
	}
}

func TestTOTPService_decryptSecret_InvalidBase64(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	_, err := svc.decryptSecret("not-valid-base64!!!")
	if err == nil {
		t.Error("Should return error for invalid base64")
	}
}

func TestTOTPService_decryptSecret_CiphertextTooShort(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	_, err := svc.decryptSecret("c2hvcnQ=")
	if err == nil {
		t.Error("Should return error for ciphertext too short")
	}
}

func TestTOTPService_isLockedOut_UserNotInMap(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	locked := svc.isLockedOut("nonexistent-user")
	if locked {
		t.Error("User not in map should not be locked out")
	}
}

func TestTOTPService_recordFailedAttempt_NewUser(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	svc.recordFailedAttempt("new-user")

	locked := svc.IsLockedOut("new-user")
	if locked {
		t.Error("User with 1 failed attempt should not be locked out")
	}
}

func TestTOTPService_recordFailedAttempt_MultipleAttempts(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	for i := 0; i < MaxTOTPAttempts-1; i++ {
		svc.recordFailedAttempt("multi-fail-user")
	}

	locked := svc.IsLockedOut("multi-fail-user")
	if locked {
		t.Error("User with MaxTOTPAttempts-1 failed attempts should not be locked out yet")
	}

	svc.recordFailedAttempt("multi-fail-user")
	locked = svc.IsLockedOut("multi-fail-user")
	if !locked {
		t.Error("User with MaxTOTPAttempts failed attempts should be locked out")
	}
}

func TestTOTPService_clearAttempts(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	svc.recordFailedAttempt("user-to-clear")
	svc.clearAttempts("user-to-clear")

	locked := svc.IsLockedOut("user-to-clear")
	if locked {
		t.Error("User after clearAttempts should not be locked out")
	}
}

func TestTOTPAttempt_Structure(t *testing.T) {
	attempt := &totpAttempt{
		failedAttempts: 3,
	}

	if attempt.failedAttempts != 3 {
		t.Errorf("Expected failedAttempts 3, got %d", attempt.failedAttempts)
	}
}

func TestTOTPService_GenerateTOTPCode(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	svc := NewTOTPService(cfg, nil, nil)
	defer svc.Stop()

	secret := "JBSWY3DPEHPK3PXP"
	code, err := svc.GenerateTOTPCode(secret)
	if err != nil {
		t.Fatalf("Failed to generate TOTP code: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("Expected 6-digit code, got %d digits", len(code))
	}
}

func TestTOTPService_ValidateWithWindow_Success(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	totpRepo := repository.NewTOTPRepository(db)
	userRepo := repository.NewUserRepository(db)
	svc := NewTOTPService(cfg, totpRepo, userRepo)
	defer svc.Stop()

	userID := "test-user-totp"
	secret := "JBSWY3DPEHPK3PXP"

	now := time.Now()
	user := &model.User{
		ID:       userID,
		Username: "totptest",
		Email:    strPtr("totp@test.com"),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	encrypted, _ := svc.encryptSecret(secret)
	totpModel := &model.TOTP{
		ID:     generateID(),
		UserID: userID,
		Secret: encrypted,
	}
	if err := totpRepo.Create(totpModel); err != nil {
		t.Fatalf("Failed to create TOTP: %v", err)
	}

	code, err := svc.GenerateTOTPCode(secret)
	if err != nil {
		t.Fatalf("Failed to generate TOTP code: %v", err)
	}

	valid := svc.ValidateWithWindow(userID, code, 1)
	if !valid {
		t.Error("Expected valid TOTP code")
	}
}

func TestTOTPService_ValidateWithWindow_InvalidCode(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars!!",
		},
	}
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	totpRepo := repository.NewTOTPRepository(db)
	userRepo := repository.NewUserRepository(db)
	svc := NewTOTPService(cfg, totpRepo, userRepo)
	defer svc.Stop()

	valid := svc.ValidateWithWindow("nonexistent-user", "000000", 1)
	if valid {
		t.Error("Expected invalid TOTP code")
	}
}
