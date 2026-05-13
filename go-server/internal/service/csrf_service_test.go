package service

import (
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/pkg/utils"
)

func TestCSRFService_GenerateToken(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	token, err := csrfService.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == nil {
		t.Fatal("Token should not be nil")
	}

	if token.Token == "" {
		t.Error("Token string should not be empty")
	}

	if token.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", token.UserID)
	}

	if token.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if token.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}

	if token.ExpiresAt.Before(token.CreatedAt) {
		t.Error("ExpiresAt should be after CreatedAt")
	}
}

func TestCSRFService_GenerateToken_DefaultExpiry(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 0)

	if csrfService.GetExpiry() != 60*time.Minute {
		t.Errorf("Expected default expiry of 60 minutes, got %v", csrfService.GetExpiry())
	}
}

func TestCSRFService_GenerateToken_CustomExpiry(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 30)

	if csrfService.GetExpiry() != 30*time.Minute {
		t.Errorf("Expected expiry of 30 minutes, got %v", csrfService.GetExpiry())
	}
}

func TestCSRFService_ValidateToken(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	token, err := csrfService.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	valid := csrfService.ValidateToken(token.Token, "user123")
	if !valid {
		t.Error("Token should be valid for correct user")
	}
}

func TestCSRFService_ValidateToken_WrongUser(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	token, err := csrfService.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	valid := csrfService.ValidateToken(token.Token, "wronguser")
	if valid {
		t.Error("Token should not be valid for wrong user")
	}
}

func TestCSRFService_ValidateToken_EmptyUser(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	token, err := csrfService.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	valid := csrfService.ValidateToken(token.Token, "")
	if !valid {
		t.Error("Token should be valid when userID is empty (for backwards compatibility)")
	}
}

func TestCSRFService_ValidateToken_AlreadyUsed(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	token, err := csrfService.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	valid1 := csrfService.ValidateToken(token.Token, "user123")
	if !valid1 {
		t.Error("First validation should succeed")
	}

	valid2 := csrfService.ValidateToken(token.Token, "user123")
	if valid2 {
		t.Error("Second validation should fail (token consumed)")
	}
}

func TestCSRFService_ValidateToken_NonExistent(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	valid := csrfService.ValidateToken("nonexistent-token", "user123")
	if valid {
		t.Error("Non-existent token should not be valid")
	}
}

func TestCSRFService_ValidateTokenAndRefresh(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	token, err := csrfService.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	valid, newToken := csrfService.ValidateTokenAndRefresh(token.Token, "user123")
	if !valid {
		t.Error("Token should be valid")
	}

	if newToken == "" {
		t.Error("New token should be provided")
	}

	if newToken == token.Token {
		t.Error("New token should be different from old token")
	}
}

func TestCSRFService_ValidateTokenAndRefresh_AlreadyUsed(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	token, err := csrfService.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	_, newToken1 := csrfService.ValidateTokenAndRefresh(token.Token, "user123")
	if newToken1 == "" {
		t.Error("First call should return new token")
	}

	_, newToken2 := csrfService.ValidateTokenAndRefresh(token.Token, "user123")
	if newToken2 != "" {
		t.Error("Second call should not return new token (original consumed)")
	}
}

func TestCSRFService_InvalidateToken(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	token, err := csrfService.GenerateToken("user123")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	csrfService.InvalidateToken(token.Token)

	valid := csrfService.ValidateToken(token.Token, "user123")
	if valid {
		t.Error("Token should not be valid after invalidation")
	}
}

func TestCSRFService_CleanupExpired(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 1)

	csrfService.GenerateToken("user1")
	csrfService.GenerateToken("user2")

	time.Sleep(1100 * time.Millisecond)

	csrfService.CleanupExpired()

	csrfService.CleanupExpired()
}

func TestCSRFService_MultipleTokensSameUser(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	token1, _ := csrfService.GenerateToken("user123")
	token2, _ := csrfService.GenerateToken("user123")

	if token1.Token == token2.Token {
		t.Error("Each token should be unique")
	}

	valid1 := csrfService.ValidateToken(token1.Token, "user123")
	valid2 := csrfService.ValidateToken(token2.Token, "user123")

	if !valid1 || !valid2 {
		t.Error("Both tokens should be valid")
	}
}

func TestCSRFService_MultipleUsers(t *testing.T) {
	random := utils.NewRandom()
	csrfService := NewCSRFService(random, 60)

	token1, _ := csrfService.GenerateToken("user1")
	token2, _ := csrfService.GenerateToken("user2")

	valid1ForUser1 := csrfService.ValidateToken(token1.Token, "user1")
	valid1ForUser2 := csrfService.ValidateToken(token1.Token, "user2")

	if !valid1ForUser1 {
		t.Error("Token should be valid for original user")
	}
	if valid1ForUser2 {
		t.Error("Token should not be valid for different user")
	}

	valid2ForUser2 := csrfService.ValidateToken(token2.Token, "user2")
	if !valid2ForUser2 {
		t.Error("Token should be valid for original user")
	}
}
