package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/authnas/authnas/go-server/internal/database"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func setupAuthMiddlewareTestDB(t *testing.T) (*gorm.DB, *model.User, func()) {
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

	email := "test@example.com"
	user := &model.User{
		ID:            "test-user-id",
		Username:      "testuser",
		Email:         &email,
		PasswordHash:  stringPtr("hash"),
		EmailVerified: true,
		Approved:      true,
		IsAdmin:       false,
		TokenVersion:  0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return db, user, cleanup
}

func stringPtr(s string) *string {
	return &s
}

func generateTestToken(user *model.User, secret string, expired bool) string {
	expiry := time.Now().Add(time.Hour)
	if expired {
		expiry = time.Now().Add(-time.Hour)
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	claims := &Claims{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        email,
		TokenVersion: user.TokenVersion,
		IsAdmin:      user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthMiddleware_Authenticate_NoHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, _, cleanup := setupAuthMiddlewareTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	am := NewAuthMiddleware(cfg, userRepo, nil, nil)

	router := gin.New()
	router.Use(am.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_Authenticate_InvalidHeaderFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, _, cleanup := setupAuthMiddlewareTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	am := NewAuthMiddleware(cfg, userRepo, nil, nil)

	router := gin.New()
	router.Use(am.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_Authenticate_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, _, cleanup := setupAuthMiddlewareTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	am := NewAuthMiddleware(cfg, userRepo, nil, nil)

	router := gin.New()
	router.Use(am.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_Authenticate_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, user, cleanup := setupAuthMiddlewareTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	am := NewAuthMiddleware(cfg, userRepo, nil, nil)

	router := gin.New()
	router.Use(am.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	token := generateTestToken(user, cfg.Security.StorageKey, true)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_Authenticate_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, _, cleanup := setupAuthMiddlewareTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	am := NewAuthMiddleware(cfg, userRepo, nil, nil)

	router := gin.New()
	router.Use(am.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	user := &model.User{
		ID:           "non-existent-user",
		Username:     "nonexistent",
		TokenVersion: 0,
	}

	token := generateTestToken(user, cfg.Security.StorageKey, false)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_Authenticate_SessionRevoked(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, user, cleanup := setupAuthMiddlewareTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	am := NewAuthMiddleware(cfg, userRepo, nil, nil)

	router := gin.New()
	router.Use(am.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	user.TokenVersion = 1
	token := generateTestToken(user, cfg.Security.StorageKey, false)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for revoked session, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_Authenticate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, user, cleanup := setupAuthMiddlewareTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	am := NewAuthMiddleware(cfg, userRepo, nil, nil)

	router := gin.New()
	router.Use(am.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		currentUser := GetCurrentUser(c)
		if currentUser == nil {
			c.String(http.StatusInternalServerError, "No user")
			return
		}
		c.String(http.StatusOK, currentUser.Username)
	})

	token := generateTestToken(user, cfg.Security.StorageKey, false)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "testuser" {
		t.Errorf("Expected body 'testuser', got '%s'", w.Body.String())
	}
}

func TestGetCurrentUser_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	user := GetCurrentUser(c)
	if user != nil {
		t.Error("Expected nil user when none set in context")
	}
}

func TestGetCurrentUser_WithUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	expectedUser := &model.User{
		ID:       "test-id",
		Username: "testuser",
	}
	c.Set("user", expectedUser)

	user := GetCurrentUser(c)
	if user == nil {
		t.Fatal("Expected user, got nil")
	}

	if user.ID != expectedUser.ID {
		t.Errorf("Expected user ID '%s', got '%s'", expectedUser.ID, user.ID)
	}
}

func TestRequireAdmin_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequireAdmin())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireAdmin_NonAdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, user, cleanup := setupAuthMiddlewareTestDB(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(db)
	am := NewAuthMiddleware(cfg, userRepo, nil, nil)

	router := gin.New()
	router.Use(am.Authenticate())
	router.Use(RequireAdmin())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	token := generateTestToken(user, cfg.Security.StorageKey, false)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestRequireAdmin_AdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			StorageKey: "test-storage-key-at-least-32-chars",
		},
	}

	db, user, cleanup := setupAuthMiddlewareTestDB(t)
	defer cleanup()

	user.IsAdmin = true
	db.Save(user)

	userRepo := repository.NewUserRepository(db)
	am := NewAuthMiddleware(cfg, userRepo, nil, nil)

	router := gin.New()
	router.Use(am.Authenticate())
	router.Use(RequireAdmin())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	token := generateTestToken(user, cfg.Security.StorageKey, false)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestClaims_Structure(t *testing.T) {
	claims := &Claims{
		UserID:       "user-123",
		Username:     "testuser",
		Email:        "test@example.com",
		TokenVersion: 5,
		IsAdmin:      true,
	}

	if claims.UserID != "user-123" {
		t.Errorf("Expected UserID 'user-123', got '%s'", claims.UserID)
	}

	if claims.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", claims.Username)
	}

	if claims.TokenVersion != 5 {
		t.Errorf("Expected TokenVersion 5, got %d", claims.TokenVersion)
	}

	if !claims.IsAdmin {
		t.Error("Expected IsAdmin to be true")
	}
}
