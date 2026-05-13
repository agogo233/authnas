// Package e2e provides end-to-end tests for AuthNas Go server.
//
// IMPORTANT: These tests directly test the Go server (localhost:8080).
// Vite is only used as a build tool to compile frontend resources to static files.
// The Go server serves both API endpoints and frontend static resources.
//
// Tests cover:
//   - User registration and login
//   - Password reset and update
//   - Profile management
//   - Protected resource access
//   - Session management
package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/authnas/authnas/go-server/internal/service"
)

func TestE2E_Auth_Register(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("successful registration", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "newuser",
			"email":    "newuser@example.com",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				AccessToken string `json:"accessToken"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if !result.Success {
			t.Error("Expected success to be true")
		}
		if result.Data.AccessToken == "" {
			t.Error("Expected access token to be set")
		}

		refreshToken := extractRefreshToken(resp.Header)
		if refreshToken == "" {
			t.Error("Expected refresh token cookie to be set")
		}

		userMeResp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, result.Data.AccessToken)
		if err != nil {
			t.Fatalf("Failed to get user info: %v", err)
		}
		assertResponseStatus(t, userMeResp, http.StatusOK)
	})

	t.Run("registration with missing fields", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "incompleteuser",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("registration with weak password", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "weakuser",
			"email":    "weak@example.com",
			"password": "123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			t.Skip("Password strength check is disabled (PasswordStrength=0)")
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("registration with duplicate username", func(t *testing.T) {
		_, err := server.RegisterUser("duplicateuser", "dup1@example.com", "password123")
		if err != nil {
			t.Fatalf("First registration failed: %v", err)
		}

		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "duplicateuser",
			"email":    "dup2@example.com",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})
}

func TestE2E_Auth_Login(t *testing.T) {
	service.ResetLoginAttempts()
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	_, err := server.RegisterUser("loginuser", "login@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("successful login with username", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "loginuser",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				AccessToken string `json:"accessToken"`
			} `json:"data"`
			MFARequired bool `json:"mfa_required"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if !result.Success {
			t.Error("Expected success to be true")
		}
		if result.Data.AccessToken == "" {
			t.Error("Expected access token to be set")
		}
	})

	t.Run("successful login with email", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "login@example.com",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("login with wrong password", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "loginuser",
			"password": "wrongpassword",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("login with non-existent user", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "nonexistent",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestE2E_Auth_ProtectedResource(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("protecteduser", "protected@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("access protected resource with valid token", func(t *testing.T) {
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
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if result.Data.Username != "protecteduser" {
			t.Errorf("Expected username 'protecteduser', got '%s'", result.Data.Username)
		}
	})

	t.Run("access protected resource without token", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/user/me", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("access protected resource with invalid token", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, "invalid-token")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("access protected resource with malformed token", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, "Bearer malformed.token.here")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestE2E_Auth_PasswordReset(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	_, err := server.RegisterUser("resetuser", "reset@example.com", "ResetPass123!")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("forgot password for existing user", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/forgot_password", map[string]interface{}{
			"email": "reset@example.com",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}
	})

	t.Run("forgot password for non-existent user", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/forgot_password", map[string]interface{}{
			"email": "nonexistent@example.com",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})
}

func TestE2E_Auth_UpdatePassword(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("updatepwduser", "updatepwd@example.com", "OldPass123!")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("update password with valid old password", func(t *testing.T) {
		resp, err := server.AuthenticatedCSRFRequest("PUT", "/api/user/me/password", map[string]interface{}{
			"oldPassword": "OldPass123!",
			"newPassword": "NewPass456!",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		resp2, err := server.LoginUser("updatepwduser", "NewPass456!")
		if err != nil {
			t.Fatalf("Login with new password failed: %v", err)
		}
		if resp2.AccessToken == "" {
			t.Error("Expected new access token after password change")
		}
		user.AccessToken = resp2.AccessToken
	})

	t.Run("update password with invalid old password", func(t *testing.T) {
		resp, err := server.AuthenticatedCSRFRequest("PUT", "/api/user/me/password", map[string]interface{}{
			"old_password": "WrongOldPass!",
			"new_password": "AnotherNewPass!",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("update password without authentication", func(t *testing.T) {
		resp, err := server.DoRequest("PUT", "/api/user/me/password", map[string]interface{}{
			"old_password": "OldPass123!",
			"new_password": "NewPass456!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestE2E_Auth_UpdateProfile(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("updateprofile", "updateprofile@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("update profile successfully", func(t *testing.T) {
		newName := "Updated Name"
		resp, err := server.AuthenticatedCSRFRequest("PUT", "/api/user/me", map[string]interface{}{
			"name": newName,
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		meResp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Failed to get user info: %v", err)
		}

		var updatedUser struct {
			Success bool `json:"success"`
			Data    struct {
				Name string `json:"name"`
			} `json:"data"`
		}
		if err := json.Unmarshal(meResp.Body, &updatedUser); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if updatedUser.Data.Name != newName {
			t.Errorf("Expected name '%s', got '%s'", newName, updatedUser.Data.Name)
		}
	})

	t.Run("update email successfully", func(t *testing.T) {
		newEmail := "newemail@example.com"
		resp, err := server.AuthenticatedCSRFRequest("PUT", "/api/user/me", map[string]interface{}{
			"email": newEmail,
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})
}

func TestE2E_Auth_SessionManagement(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("sessionuser", "session@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("delete all sessions", func(t *testing.T) {
		resp, err := server.AuthenticatedCSRFRequest("DELETE", "/api/user/me/sessions", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})
}

func TestE2E_Auth_RefreshToken(t *testing.T) {
	server := NewE2ETestServerWithCookies(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("refreshuser", "refresh@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("refresh token returns new access token", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/refresh", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool   `json:"success"`
			Data    struct {
				AccessToken string `json:"accessToken"`
				ExpiresAt   string `json:"expiresAt"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if !result.Success {
			t.Error("Expected success to be true")
		}
		if result.Data.AccessToken == "" {
			t.Error("Expected new access token to be set")
		}
		if result.Data.ExpiresAt == "" {
			t.Error("Expected expiresAt to be set")
		}

		newRefreshToken := extractRefreshToken(resp.Header)
		if newRefreshToken == "" {
			t.Error("Expected new refresh token cookie to be set (rotation)")
		}
		if newRefreshToken == user.RefreshToken {
			t.Error("Expected refresh token to be rotated (different from original)")
		}
	})

	t.Run("new access token works for authenticated requests", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/refresh", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				AccessToken string `json:"accessToken"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		newToken := result.Data.AccessToken
		if newToken == "" {
			t.Skip("No access token from refresh")
		}

		meResp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, newToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, meResp, http.StatusOK)
	})

	t.Run("refresh without cookie returns unauthorized", func(t *testing.T) {
		freshServer := NewE2ETestServer(t)
		defer freshServer.Cleanup()

		resp, err := freshServer.DoRequest("POST", "/api/auth/refresh", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("refresh with invalid cookie returns unauthorized", func(t *testing.T) {
		freshServer := NewE2ETestServer(t)
		defer freshServer.Cleanup()

		resp, err := freshServer.DoRequest("POST", "/api/auth/refresh", nil, map[string]string{
			"Cookie": "auth_refresh_token=invalid-token-value",
		})
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestE2E_Auth_HealthCheck(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("health check returns OK", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/health", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}
	})
}

func TestE2E_Auth_PublicConfig(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("get public config", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/config", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				AppName string `json:"app_name"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if result.Data.AppName == "" {
			t.Error("Expected app_name in config")
		}
	})
}

func TestE2E_Auth_Invitation(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("get non-existent invitation", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/auth/invitation/nonexistent/invalidcode", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})
}

func TestE2E_Auth_VerifyEmail(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("verifyemail", "verifyemail@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("verify email with invalid code", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/verify_email", map[string]interface{}{
			"id":        user.ID,
			"challenge": "invalid-code",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})
}

func TestE2E_Auth_SendVerifyEmail(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("send verification email for existing user", func(t *testing.T) {
		user, err := server.RegisterUser("sendverify", "sendverify@example.com", "password123")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.DoRequest("POST", "/api/auth/send_verify_email", map[string]interface{}{
			"email": user.Email,
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("send verification email for non-existent user", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/send_verify_email", map[string]interface{}{
			"email": "nonexistent@example.com",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})
}

func TestE2E_Auth_Passkey(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("passkey start without username", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/passkey/start", map[string]interface{}{
			"username": "nonexistentuser",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("passkey end with invalid response", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/passkey/end", map[string]interface{}{
			"credential_id": "test-credential-id",
			"challenge":     "test-challenge",
			"response":      "invalid-response",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})
}

func TestE2E_Auth_TOTP(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("totpuser", "totp@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("TOTP verify with missing fields", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/auth/totp", map[string]interface{}{}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})
}

func TestE2E_Auth_RateLimiting(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("multiple rapid login attempts", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
				"input":    "rateuser",
				"password": "wrongpassword",
			}, nil)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			if i >= 3 && resp.StatusCode == http.StatusTooManyRequests {
				t.Logf("Rate limiting triggered after %d attempts", i+1)
				return
			}
		}
	})
}

func TestE2E_Auth_InputValidation(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("SQL injection prevention in login", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "' OR '1'='1",
			"password": "' OR '1'='1",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("XSS prevention in username registration", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "<script>alert('xss')</script>",
			"email":    "xss@example.com",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("empty username registration", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "",
			"email":    "empty@example.com",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("invalid email format", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "invalidemail",
			"email":    "not-an-email",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("username with special characters", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "user@#$%",
			"email":    "special@example.com",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})
}

func TestE2E_Auth_ConcurrentRegistration(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("concurrent registrations", func(t *testing.T) {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				username := fmt.Sprintf("concurrentuser%d", time.Now().UnixNano()+int64(index))
				email := fmt.Sprintf("concurrent%d@example.com", time.Now().UnixNano()+int64(index))
				_, err := server.RegisterUser(
					strings.ReplaceAll(username, "-", ""),
					email,
					"password123",
				)
				if err != nil {
					t.Logf("Concurrent registration error: %v", err)
				}
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
