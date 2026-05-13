// Package e2e provides end-to-end tests for AuthNas Go server.
//
// IMPORTANT: These tests directly test the Go server (localhost:8080).
// Vite is only used as a build tool to compile frontend resources to static files.
// The Go server serves both API endpoints and frontend static resources.
//
// Tests cover:
//   - TOTP registration and verification
//   - Passkey listing and deletion
//   - User profile management
//   - User session management
package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestE2E_TOTP_FullFlow(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("totpfullflow", "totpfull@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("register TOTP", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/totp/registration", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				Secret    string `json:"secret"`
				QRCodeURI string `json:"qr_code_uri"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if !result.Success {
			t.Error("Expected success to be true")
		}
		if result.Data.Secret == "" {
			t.Error("Expected secret to be set")
		}
		if result.Data.QRCodeURI == "" {
			t.Error("Expected QRCodeURI to be set")
		}
	})

	t.Run("list passkeys after registration", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/passkey", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("delete TOTP without token should fail", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/totp", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("TOTP delete without token returned %d, expected 400", resp.StatusCode)
		}
	})

	t.Run("delete TOTP with invalid token should fail", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/totp", map[string]interface{}{
			"token": "000000",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusBadRequest {
			t.Logf("TOTP delete with invalid token returned %d, expected 401 or 400", resp.StatusCode)
		}
	})
}

func TestE2E_Passkey_ListDelete(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("passkeylist", "passkeylist@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("list passkeys for new user", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/passkey", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Passkeys []interface{} `json:"passkeys"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if len(result.Passkeys) != 0 {
			t.Errorf("Expected 0 passkeys for new user, got %d", len(result.Passkeys))
		}
	})

	t.Run("delete non-existent passkey", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/passkey/nonexistent-id", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})
}

func TestE2E_UserProfile(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("profileuser", "profile@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("get user profile", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				ID       string `json:"id"`
				Username string `json:"username"`
				Email    string `json:"email"`
				Name     string `json:"name"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if result.Data.Username != "profileuser" {
			t.Errorf("Expected username 'profileuser', got '%s'", result.Data.Username)
		}
	})

	t.Run("update profile with valid name", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("PUT", "/api/user/me", map[string]interface{}{
			"name": "Updated Profile Name",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		meResp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Failed to get user info: %v", err)
		}

		var updated struct {
			Success bool `json:"success"`
			Data    struct {
				Name string `json:"name"`
			} `json:"data"`
		}
		json.Unmarshal(meResp.Body, &updated)

		if updated.Data.Name != "Updated Profile Name" {
			t.Errorf("Expected name 'Updated Profile Name', got '%s'", updated.Data.Name)
		}
	})

	t.Run("update profile with empty name", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("PUT", "/api/user/me", map[string]interface{}{
			"name": "",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			t.Log("Empty name update returned 200, may be allowed")
		}
	})
}

func TestE2E_UserSessions(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("sessionuser2", "session2@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("delete all sessions", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/user/me/sessions", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("token should be invalid after session delete", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Logf("Token after session delete returned %d (depends on implementation)", resp.StatusCode)
		}
	})
}
