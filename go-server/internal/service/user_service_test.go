package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/pkg/utils"
	"gorm.io/gorm"
)

func setupUserServiceTestDB(t *testing.T) (*gorm.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Group{},
		&model.UserGroup{},
		&model.Client{},
		&model.Consent{},
		&model.Passkey{},
		&model.PasskeyAuthOptions{},
		&model.TOTP{},
		&model.Key{},
		&model.Invitation{},
		&model.OIDCPayload{},
		&model.EmailVerification{},
		&model.PasswordReset{},
		&model.ProxyAuth{},
		&model.EmailLog{},
	); err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func newTestUserService(db *gorm.DB) *UserService {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey:        "test-storage-key-at-least-32-chars",
			PasswordStrength:  0,
			PasswordMinLength: 0,
		},
	}

	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	keyRepo := repository.NewKeyRepository(db)
	totpRepo := repository.NewTOTPRepository(db)
	passkeyRepo := repository.NewPasskeyRepository(db)
	emailVerificationRepo := repository.NewEmailVerificationRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	consentRepo := repository.NewConsentRepository(db)

	randomUtil := utils.NewRandom()
	timeUtil := utils.NewTime()

	return NewUserService(
		cfg,
		userRepo,
		groupRepo,
		keyRepo,
		totpRepo,
		passkeyRepo,
		emailVerificationRepo,
		passwordResetRepo,
		consentRepo,
		nil,
		nil,
		randomUtil,
		timeUtil,
		db,
	)
}

func TestUserService_Create(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	if user.Email == nil || *user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%v'", user.Email)
	}

	if user.PasswordHash == nil || *user.PasswordHash == "password123" {
		t.Error("Password should be hashed")
	}
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	_, err := svc.Create("test@example.com", "user1", "password123")
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	_, err = svc.Create("test@example.com", "user2", "password123")
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

func TestUserService_Create_DuplicateUsername(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	_, err := svc.Create("user1@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	_, err = svc.Create("user2@example.com", "testuser", "password123")
	if err == nil {
		t.Error("Expected error for duplicate username, got nil")
	}
}

func TestUserService_Create_EmptyUsername(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	_, err := svc.Create("test@example.com", "", "password123")
	if err == nil {
		t.Error("Expected error for empty username, got nil")
	}
}

func TestUserService_Create_AutoGeneratePassword(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		t.Error("Password should be auto-generated")
	}
}

func TestUserService_GetByInput(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	_, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user, err := svc.GetByInput("test@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	user, err = svc.GetByInput("testuser")
	if err != nil {
		t.Fatalf("Failed to get user by username: %v", err)
	}
	if user.Email == nil || *user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%v'", user.Email)
	}
}

func TestUserService_GetByUsername(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	_, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user, err := svc.GetByUsername("testuser")
	if err != nil {
		t.Fatalf("Failed to get user by username: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
}

func TestUserService_GetByID(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	created, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if user.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, user.ID)
	}
}

func TestUserService_List(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	for i := 0; i < 5; i++ {
		_, err := svc.Create(fmt.Sprintf("test%d@example.com", i), fmt.Sprintf("testuser%d", i), "password123")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	users, total, err := svc.List(0, 10)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(users) != 5 {
		t.Errorf("Expected 5 users, got %d", len(users))
	}
}

func TestUserService_List_Pagination(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	for i := 0; i < 5; i++ {
		_, err := svc.Create(fmt.Sprintf("test%d@example.com", i), fmt.Sprintf("testuser%d", i), "password123")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	users, total, err := svc.List(0, 2)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	users, _, err = svc.List(2, 2)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users for second page, got %d", len(users))
	}
}

func TestUserService_Count(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	for i := 0; i < 3; i++ {
		_, err := svc.Create(fmt.Sprintf("test%d@example.com", i), fmt.Sprintf("testuser%d", i), "password123")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	count, err := svc.Count()
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestUserService_Update(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	newName := "Test User"
	user.Name = &newName

	err = svc.Update(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	updated, err := svc.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updated.Name == nil || *updated.Name != newName {
		t.Errorf("Expected name '%s', got '%v'", newName, updated.Name)
	}
}

func TestUserService_Delete(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = svc.Delete(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	_, err = svc.GetByID(user.ID)
	if err == nil {
		t.Error("Expected error when getting deleted user")
	}
}

func TestUserService_CheckEmailAvailable(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	_, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	available, err := svc.CheckEmailAvailable("test@example.com", "")
	if err != nil {
		t.Fatalf("Failed to check email: %v", err)
	}
	if available {
		t.Error("Email should not be available")
	}

	available, err = svc.CheckEmailAvailable("other@example.com", "")
	if err != nil {
		t.Fatalf("Failed to check email: %v", err)
	}
	if !available {
		t.Error("Email should be available")
	}

	available, err = svc.CheckEmailAvailable("test@example.com", "exclude-user-id")
	if err != nil {
		t.Fatalf("Failed to check email: %v", err)
	}
	if available {
		t.Error("Email should not be available when excluding non-matching user")
	}
}

func TestUserService_ResetPassword(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "oldpassword")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	originalHash := user.PasswordHash

	err = svc.ResetPassword(user.ID, "newpassword123")
	if err != nil {
		t.Fatalf("Failed to reset password: %v", err)
	}

	updated, err := svc.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if updated.PasswordHash == originalHash {
		t.Error("Password hash should have changed")
	}
}

func TestUserService_ResetPassword_UserNotFound(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	err := svc.ResetPassword("non-existent-id", "newpassword123")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestUserService_EnsureInitialAdmin(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	err := svc.EnsureInitialAdmin("admin", "admin@example.com", "adminpassword")
	if err != nil {
		t.Fatalf("Failed to ensure initial admin: %v", err)
	}

	user, err := svc.GetByUsername("admin")
	if err != nil {
		t.Fatalf("Failed to get admin user: %v", err)
	}

	if !user.IsAdmin {
		t.Error("User should be admin")
	}

	err = svc.EnsureInitialAdmin("admin", "admin@example.com", "adminpassword")
	if err != nil {
		t.Fatalf("Second call should not fail: %v", err)
	}
}

func TestUserService_EnsureInitialAdmin_EmptyUsername(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	err := svc.EnsureInitialAdmin("", "", "password")
	if err != nil {
		t.Fatalf("Empty username should return nil: %v", err)
	}

	_, err = svc.GetByUsername("")
	if err == nil {
		t.Error("Admin with empty username should not be created")
	}
}

func TestUserService_HashPassword(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	password := "testpassword123"
	hash, err := svc.hashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == password {
		t.Error("Hash should not equal original password")
	}

	if len(hash) < 50 {
		t.Error("Hash should be long (includes salt and parameters)")
	}
}

func TestUserService_VerifyPassword(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	password := "testpassword123"
	hash, _ := svc.hashPassword(password)

	if !svc.verifyPassword(hash, password) {
		t.Error("Password verification should succeed for correct password")
	}

	if svc.verifyPassword(hash, "wrongpassword") {
		t.Error("Password verification should fail for wrong password")
	}
}

func TestUserService_VerifyPassword_InvalidHash(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	if svc.verifyPassword("invalid-hash", "password") {
		t.Error("Should return false for invalid hash format")
	}

	if svc.verifyPassword("", "password") {
		t.Error("Should return false for empty hash")
	}
}

func TestUserService_Search(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	svc.Create("alice@example.com", "alice", "password123")
	svc.Create("bob@example.com", "bob", "password123")
	svc.Create("charlie@example.com", "charlie", "password123")

	users, total, err := svc.Search("alice", 0, 10)
	if err != nil {
		t.Fatalf("Failed to search users: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected total 1, got %d", total)
	}
	if len(users) != 1 || users[0].Username != "alice" {
		t.Errorf("Expected to find alice, got %v", users)
	}

	users, total, err = svc.Search("example.com", 0, 10)
	if err != nil {
		t.Fatalf("Failed to search users: %v", err)
	}
	if total != 3 {
		t.Errorf("Expected total 3 for email domain search, got %d", total)
	}

	users, total, err = svc.Search("nonexistent", 0, 10)
	if err != nil {
		t.Fatalf("Failed to search users: %v", err)
	}
	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}
}

func TestUserService_CountAdmins(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	svc.Create("user1@example.com", "user1", "password123")
	svc.Create("user2@example.com", "user2", "password123")

	count, err := svc.CountAdmins()
	if err != nil {
		t.Fatalf("Failed to count admins: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 admins, got %d", count)
	}

	svc.EnsureInitialAdmin("admin", "admin@example.com", "adminpass")

	count, err = svc.CountAdmins()
	if err != nil {
		t.Fatalf("Failed to count admins: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 admin, got %d", count)
	}
}

func TestUserService_UpdatePassword(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "oldpassword")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	oldHash := user.PasswordHash

	err = svc.UpdatePassword(user.ID, "oldpassword", "newpassword123")
	if err != nil {
		t.Fatalf("Failed to update password: %v", err)
	}

	updated, err := svc.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if updated.PasswordHash == oldHash {
		t.Error("Password hash should have changed")
	}

	err = svc.UpdatePassword(user.ID, "oldpassword", "anotherpassword")
	if err == nil {
		t.Error("Should fail with wrong old password")
	}
}

func TestUserService_UpdatePassword_WrongOldPassword(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = svc.UpdatePassword(user.ID, "wrongpassword", "newpassword123")
	if err == nil {
		t.Error("Should fail with wrong old password")
	}
}

func TestUserService_UpdatePassword_UserNotFound(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	err := svc.UpdatePassword("non-existent", "old", "new")
	if err == nil {
		t.Error("Should fail for non-existent user")
	}
}

func TestUserService_UpdatePassword_NoOldPassword(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = svc.UpdatePassword(user.ID, "", "newpassword123")
	if err == nil {
		t.Error("Should fail when old password is empty")
	}
}

func TestUserService_RevokeAllSessions(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	originalTokenVersion := user.TokenVersion

	err = svc.RevokeAllSessions(user.ID)
	if err != nil {
		t.Fatalf("Failed to revoke sessions: %v", err)
	}

	updated, err := svc.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if updated.TokenVersion <= originalTokenVersion {
		t.Error("Token version should have incremented")
	}

	err = svc.RevokeAllSessions("non-existent")
	if err != nil {
		t.Fatalf("Should not fail for non-existent user: %v", err)
	}
}

func TestUserService_GetUserSessions(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	sessions, err := svc.GetUserSessions(user.ID)
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
}

func TestUserService_RevokeSession(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = svc.RevokeSession(user.ID, "non-existent-session")
	if err == nil {
		t.Error("Should fail for non-existent session")
	}
}

func TestUserService_ResetPasswordByCode(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "oldpassword")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	code := "test-reset-code-12345"
	pr := &model.PasswordReset{
		ID:        generateID(),
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: svc.time.Now().Add(1 * time.Hour),
		CreatedAt: svc.time.Now(),
	}

	prRepo := repository.NewPasswordResetRepository(db)
	err = prRepo.Create(pr)
	if err != nil {
		t.Fatalf("Failed to create password reset: %v", err)
	}

	err = svc.ResetPasswordByCode(code, "newpassword123")
	if err != nil {
		t.Fatalf("Failed to reset password by code: %v", err)
	}

	updated, err := svc.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if updated.TokenVersion != 1 {
		t.Errorf("Expected token version 1, got %d", updated.TokenVersion)
	}

	_, err = prRepo.GetByCode(code)
	if err == nil {
		t.Error("Reset code should be deleted after use")
	}
}

func TestUserService_ResetPasswordByCode_InvalidCode(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	err := svc.ResetPasswordByCode("invalid-code", "newpassword")
	if err == nil {
		t.Error("Should fail with invalid code")
	}
}

func TestUserService_ResetPasswordByCode_ExpiredCode(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	user, err := svc.Create("test@example.com", "testuser", "oldpassword")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	code := "expired-reset-code"
	pr := &model.PasswordReset{
		ID:        generateID(),
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: svc.time.Now().Add(-1 * time.Hour),
		CreatedAt: svc.time.Now().Add(-2 * time.Hour),
	}

	prRepo := repository.NewPasswordResetRepository(db)
	err = prRepo.Create(pr)
	if err != nil {
		t.Fatalf("Failed to create password reset: %v", err)
	}

	err = svc.ResetPasswordByCode(code, "newpassword")
	if err == nil {
		t.Error("Should fail with expired code")
	}
}

func TestUserService_GetConfig(t *testing.T) {
	db, cleanup := setupUserServiceTestDB(t)
	defer cleanup()

	svc := newTestUserService(db)

	cfg := svc.GetConfig()
	if cfg == nil {
		t.Error("Config should not be nil")
	}
	if cfg.Security.StorageKey != "test-storage-key-at-least-32-chars" {
		t.Errorf("Unexpected storage key: %s", cfg.Security.StorageKey)
	}
}
