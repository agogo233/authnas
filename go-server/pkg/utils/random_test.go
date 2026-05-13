package utils

import (
	"strings"
	"testing"
)

func TestRandomUtil_NewRandom(t *testing.T) {
	util := NewRandom()
	if util == nil {
		t.Fatal("NewRandom() returned nil")
	}
}

func TestRandomUtil_GenerateToken(t *testing.T) {
	util := NewRandom()

	testCases := []int{16, 32, 64}

	for _, length := range testCases {
		token, err := util.GenerateToken(length)
		if err != nil {
			t.Errorf("GenerateToken(%d) returned error: %v", length, err)
			continue
		}

		if len(token) == 0 {
			t.Errorf("GenerateToken(%d) returned empty token", length)
		}

		if !strings.ContainsAny(token, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=") {
			t.Errorf("GenerateToken(%d) returned non-base64 string: %q", length, token)
		}
	}
}

func TestRandomUtil_GenerateToken_ZeroLength(t *testing.T) {
	util := NewRandom()
	token, err := util.GenerateToken(0)
	if err != nil {
		t.Errorf("GenerateToken(0) returned error: %v", err)
	}
	if token != "" {
		t.Errorf("GenerateToken(0) should return empty string, got %q", token)
	}
}

func TestRandomUtil_GenerateToken_Uniqueness(t *testing.T) {
	util := NewRandom()
	tokens := make(map[string]bool)

	for i := 0; i < 100; i++ {
		token, err := util.GenerateToken(32)
		if err != nil {
			t.Fatalf("GenerateToken(32) returned error: %v", err)
		}

		if tokens[token] {
			t.Errorf("GenerateToken produced duplicate token after %d iterations", i)
		}
		tokens[token] = true
	}
}

func TestRandomUtil_GenerateRandomBytes(t *testing.T) {
	util := NewRandom()

	testCases := []int{16, 32, 64}

	for _, length := range testCases {
		bytes, err := util.GenerateRandomBytes(length)
		if err != nil {
			t.Errorf("GenerateRandomBytes(%d) returned error: %v", length, err)
			continue
		}

		if len(bytes) != length {
			t.Errorf("GenerateRandomBytes(%d) returned %d bytes", length, len(bytes))
		}
	}
}

func TestRandomUtil_GenerateRandomBytes_ZeroLength(t *testing.T) {
	util := NewRandom()
	bytes, err := util.GenerateRandomBytes(0)
	if err != nil {
		t.Errorf("GenerateRandomBytes(0) returned error: %v", err)
	}
	if len(bytes) != 0 {
		t.Errorf("GenerateRandomBytes(0) should return empty slice, got %d bytes", len(bytes))
	}
}

func TestRandomUtil_GenerateRandomBytes_Uniqueness(t *testing.T) {
	util := NewRandom()
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		bytes, err := util.GenerateRandomBytes(32)
		if err != nil {
			t.Fatalf("GenerateRandomBytes(32) returned error: %v", err)
		}

		key := string(bytes)
		if seen[key] {
			t.Errorf("GenerateRandomBytes produced duplicate bytes after %d iterations", i)
		}
		seen[key] = true
	}
}
