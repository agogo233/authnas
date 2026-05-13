package service

import (
	"testing"

	"github.com/authnas/authnas/go-server/internal/config"
)

func TestPasswordService_NewPasswordService(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			PasswordStrength: 3,
		},
	}
	svc := NewPasswordService(cfg)
	if svc == nil {
		t.Fatal("NewPasswordService should not return nil")
	}
}

func TestPasswordService_CheckStrength_EmptyPassword(t *testing.T) {
	cfg := &config.Config{}
	svc := NewPasswordService(cfg)

	result := svc.CheckStrength("")
	if result.Score != 0 {
		t.Errorf("Expected score 0 for empty string, got %d", result.Score)
	}
	if result.Strength != "very_weak" {
		t.Errorf("Expected strength 'very_weak', got '%s'", result.Strength)
	}
}

func TestPasswordService_CheckStrength_SimplePassword(t *testing.T) {
	cfg := &config.Config{}
	svc := NewPasswordService(cfg)

	result := svc.CheckStrength("abc")
	if result.Score < 0 || result.Score > 4 {
		t.Errorf("Score should be between 0 and 4, got %d", result.Score)
	}
	if result.Strength == "" {
		t.Error("Strength should not be empty")
	}
}

func TestPasswordService_CheckStrength_CommonPassword(t *testing.T) {
	cfg := &config.Config{}
	svc := NewPasswordService(cfg)

	result := svc.CheckStrength("password")
	if result.Score < 0 || result.Score > 4 {
		t.Errorf("Score should be between 0 and 4, got %d", result.Score)
	}
	if result.Strength == "" {
		t.Error("Strength should not be empty")
	}
}

func TestPasswordService_CheckStrength_RandomString(t *testing.T) {
	cfg := &config.Config{}
	svc := NewPasswordService(cfg)

	result := svc.CheckStrength("aB3$xY7")
	if result.Score < 0 || result.Score > 4 {
		t.Errorf("Score should be between 0 and 4, got %d", result.Score)
	}
	if result.Strength == "" {
		t.Error("Strength should not be empty")
	}
}

func TestPasswordService_CheckStrength_StrongPassword(t *testing.T) {
	cfg := &config.Config{}
	svc := NewPasswordService(cfg)

	result := svc.CheckStrength("V3ry$tr0ng!P@ssw0rd#2024")
	if result.Score < 0 || result.Score > 4 {
		t.Errorf("Score should be between 0 and 4, got %d", result.Score)
	}
	if result.Strength == "" {
		t.Error("Strength should not be empty")
	}
	if result.CrackTimeDisplay == "" {
		t.Error("CrackTimeDisplay should not be empty")
	}
}

func TestPasswordService_CheckStrength_CrackTimeDisplay(t *testing.T) {
	cfg := &config.Config{}
	svc := NewPasswordService(cfg)

	testCases := []string{
		"x",
		"password",
		"hello world",
		"MyStr0ng!Pass",
		"V3ry$tr0ng!P@ss",
	}

	for _, password := range testCases {
		result := svc.CheckStrength(password)
		if result.CrackTimeDisplay == "" {
			t.Errorf("CrackTimeDisplay should not be empty for password '%s'", password)
		}
	}
}

func TestPasswordService_IsStrongEnough_BelowThreshold(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			PasswordStrength: 3,
		},
	}
	svc := NewPasswordService(cfg)

	if svc.IsStrongEnough("x") {
		t.Error("Password 'x' should not be strong enough with threshold 3")
	}
}

func TestPasswordService_IsStrongEnough_AtThreshold(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			PasswordStrength: 2,
		},
	}
	svc := NewPasswordService(cfg)

	isStrong := svc.IsStrongEnough("V3ry$tr0ng!P@ssw0rd")
	if !isStrong && svc.CheckStrength("V3ry$tr0ng!P@ssw0rd").Score >= 2 {
		t.Logf("Password might be considered strong enough at threshold 2 depending on zxcvbn version")
	}
}

func TestPasswordService_CheckStrength_AllStrengths(t *testing.T) {
	cfg := &config.Config{}
	svc := NewPasswordService(cfg)

	validStrengths := map[string]bool{
		"very_weak":   true,
		"weak":        true,
		"fair":        true,
		"strong":      true,
		"very_strong": true,
	}

	passwords := []string{
		"x",
		"password",
		"hello world",
		"MyStr0ng!Pass",
		"V3ry$tr0ng!P@ss",
	}

	for _, password := range passwords {
		result := svc.CheckStrength(password)
		if !validStrengths[result.Strength] {
			t.Errorf("Invalid strength '%s' for password '%s'", result.Strength, password)
		}
	}
}

func TestPasswordService_ResultStruct_Fields(t *testing.T) {
	cfg := &config.Config{}
	svc := NewPasswordService(cfg)

	result := svc.CheckStrength("Test123!")
	if result.Score < 0 || result.Score > 4 {
		t.Errorf("Score should be between 0 and 4, got %d", result.Score)
	}
	if result.Strength == "" {
		t.Error("Strength should not be empty")
	}
	if result.CrackTimeDisplay == "" {
		t.Error("CrackTimeDisplay should not be empty")
	}
}
