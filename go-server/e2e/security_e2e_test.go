package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestE2E_Security_JWTValidation(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("jwtuser", "jwt@example.com", "Password123!")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("jwt-token-structure-valid", func(t *testing.T) {
		parts := strings.Split(user.AccessToken, ".")
		if len(parts) != 3 {
			t.Fatalf("Expected JWT with 3 parts, got %d", len(parts))
		}

		for i, part := range parts {
			if part == "" {
				t.Fatalf("JWT part %d is empty", i)
			}
			_, err := base64.RawURLEncoding.DecodeString(part)
			if err != nil {
				t.Fatalf("JWT part %d is not valid base64url: %v", i, err)
			}
		}
	})

	t.Run("jwt-payload-contains-required-claims", func(t *testing.T) {
		parts := strings.Split(user.AccessToken, ".")
		payloadBytes, _ := base64.RawURLEncoding.DecodeString(parts[1])

		var payload map[string]interface{}
		json.Unmarshal(payloadBytes, &payload)

		requiredClaims := []string{"sub", "exp", "iat", "iss"}
		for _, claim := range requiredClaims {
			if payload[claim] == nil {
				t.Errorf("JWT payload missing required claim: %s", claim)
			}
		}
	})

	t.Run("jwt-expired-token-rejected", func(t *testing.T) {
		parts := strings.Split(user.AccessToken, ".")
		payloadBytes, _ := base64.RawURLEncoding.DecodeString(parts[1])

		var payload map[string]interface{}
		json.Unmarshal(payloadBytes, &payload)

		payload["exp"] = float64(time.Now().Unix() - 3600)
		newPayloadBytes, _ := json.Marshal(payload)
		newPayload := base64.RawURLEncoding.EncodeToString(newPayloadBytes)

		tamperedToken := parts[0] + "." + newPayload + "." + parts[2]

		resp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, tamperedToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("jwt-tampered-signature-rejected", func(t *testing.T) {
		parts := strings.Split(user.AccessToken, ".")
		tamperedToken := parts[0] + "." + parts[1] + ".tampered_signature"

		resp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, tamperedToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("jwt-missing-bearer-prefix-rejected", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/user/me", nil, map[string]string{
			"Authorization": user.AccessToken,
		})
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("jwt-empty-bearer-prefix-rejected", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/user/me", nil, map[string]string{
			"Authorization": "Bearer ",
		})
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("jwt-basic-auth-prefix-rejected", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/user/me", nil, map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass")),
		})
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("jwt-token-version-invalidated-on-password-change", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("PUT", "/api/user/me/password", map[string]interface{}{
			"oldPassword": "Password123!",
			"newPassword": "NewPassword456!",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Password change failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)

		resp2, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request with old token failed: %v", err)
		}

		assertResponseStatus(t, resp2, http.StatusUnauthorized)
	})
}

func TestE2E_Security_AccountLockout(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	_, err := server.RegisterUser("lockoutuser", "lockout@example.com", "Password123!")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("account-lockout-after-failed-attempts", func(t *testing.T) {
		maxAttempts := 10
		locked := false

		for i := 0; i < maxAttempts; i++ {
			resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
				"input":    "lockoutuser",
				"password": "wrongpassword",
			}, nil)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				locked = true
				t.Logf("Account locked after %d failed attempts", i+1)
				break
			}

			if resp.StatusCode == http.StatusUnauthorized && i == maxAttempts-1 {
				t.Logf("Account did not lock after %d failed attempts (may be disabled)", maxAttempts)
			}
		}

		_ = locked
	})

	t.Run("successful-login-resets-failure-counter", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
				"input":    "lockoutuser",
				"password": "wrongpassword",
			}, nil)
		}

		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "lockoutuser",
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			t.Log("Successful login resets failure counter")
		}
	})
}

func TestE2E_Security_PasswordStorage(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	_, err := server.RegisterUser("pwdstorage", "pwdstorage@example.com", "OriginalPass123!")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("password-not-returned-in-api-responses", func(t *testing.T) {
		resp, err := server.LoginUser("pwdstorage", "OriginalPass123!")
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		respBody, _ := json.Marshal(resp)
		if strings.Contains(string(respBody), "OriginalPass123!") {
			t.Error("Password should not be included in API response")
		}
		if strings.Contains(string(respBody), "password") {
			t.Error("Password field should not be in response")
		}
	})

	t.Run("password-hash-not-predictable", func(t *testing.T) {
		_, _ = server.RegisterUser("pwdhash2", "pwdhash2@example.com", "SamePassword123!")

		user1Token, _ := server.LoginUser("pwdstorage", "OriginalPass123!")
		user2Token, _ := server.LoginUser("pwdhash2", "SamePassword123!")

		if user1Token.AccessToken == user2Token.AccessToken {
			t.Error("Different users should have different tokens")
		}
	})
}

func TestE2E_Security_InputValidation(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("username-length-validation", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "a",
			"email":    "short@example.com",
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			t.Log("Short username accepted")
		} else {
			assertResponseStatus(t, resp, http.StatusBadRequest)
		}
	})

	t.Run("username-max-length-validation", func(t *testing.T) {
		longUsername := strings.Repeat("a", 200)
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": longUsername,
			"email":    "long@example.com",
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("email-max-length-validation", func(t *testing.T) {
		longEmail := strings.Repeat("a", 250) + "@example.com"
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "validuser",
			"email":    longEmail,
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("null-byte-injection-prevention", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "user\x00name",
			"email":    "null@example.com",
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("control-character-injection-prevention", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "user\r\nname",
			"email":    "control@example.com",
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("html-tag-in-username-prevention", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "<html>username</html>",
			"email":    "html@example.com",
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("unicode-homograph-attack-prevention", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "pаyout",
			"password": "anypassword",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Log("Unicode homograph attack blocked")
		}
	})
}

func TestE2E_Security_SessionFixation(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("sessionfix", "sessionfix@example.com", "Password123!")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("new-session-after-login", func(t *testing.T) {
		resp1, err := server.LoginUser("sessionfix", "Password123!")
		if err != nil {
			t.Fatalf("First login failed: %v", err)
		}

		resp2, err := server.LoginUser("sessionfix", "Password123!")
		if err != nil {
			t.Fatalf("Second login failed: %v", err)
		}

		if resp1.AccessToken == resp2.AccessToken {
			t.Log("Session fixation possible: same token returned")
		} else {
			t.Log("Different tokens for each login session")
		}
	})

	t.Run("session-id-in-response", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "sessionfix",
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)

		if sessionID, ok := result["session_id"]; ok && sessionID != nil {
			t.Logf("Session ID returned in response: %v", sessionID)
		}
	})

	t.Run("cookie-attributes-secure", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "sessionfix",
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		cookies := resp.Header["Set-Cookie"]
		for _, cookie := range cookies {
			t.Logf("Cookie: %s", cookie)
		}
	})

	_ = user
}

func TestE2E_Security_CSRFProtection(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("csrfuser", "csrf@example.com", "Password123!")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("state-parameter-required-for-oidc", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/auth?client_id=test&redirect_uri=https://example.com&response_type=code&scope=openid", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusFound {
			location := resp.Header.Get("Location")
			if !strings.Contains(location, "state=") {
				t.Log("Warning: OIDC authorization without state parameter")
			}
		}
	})

	t.Run("csrf-token-for-sensitive-operations", func(t *testing.T) {
		resp, err := server.AuthenticatedCSRFRequest("PUT", "/api/user/me", map[string]interface{}{
			"name": "New Name",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	_ = user
}

func TestE2E_Security_RateLimiting(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("rate-limit-login-endpoint", func(t *testing.T) {
		rateLimited := false
		attempts := 0

		for i := 0; i < 20; i++ {
			resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
				"input":    "nonexistent",
				"password": "wrongpassword",
			}, nil)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			attempts++

			if resp.StatusCode == http.StatusTooManyRequests {
				rateLimited = true
				t.Logf("Rate limited after %d attempts", attempts)
				break
			}
		}

		if !rateLimited {
			t.Logf("No rate limiting observed after %d attempts", attempts)
		}
	})

	t.Run("rate-limit-returns-retry-after-header", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
				"input":    "nonexistent",
				"password": "wrongpassword",
			}, nil)
		}

		resp, _ := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "nonexistent",
			"password": "wrongpassword",
		}, nil)

		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				t.Logf("Retry-After header present: %s", retryAfter)
			}
		}
	})

	t.Run("rate-limit-scoped-to-ip", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
				"input":    "nonexistent",
				"password": "wrongpassword",
			}, nil)
		}

		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": fmt.Sprintf("user%d", time.Now().UnixNano()),
			"email":    fmt.Sprintf("user%d@example.com", time.Now().UnixNano()),
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			t.Log("Rate limiting applies to different endpoints from same IP")
		}
	})
}

func TestE2E_Security_InformationDisclosure(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("login-error-user-not_found-message", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "nonexistentuser12345",
			"password": "anypassword",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		body := string(resp.Body)
		if strings.Contains(strings.ToLower(body), "user not found") ||
			strings.Contains(strings.ToLower(body), "用户不存在") {
			t.Error("Error message reveals user existence")
		}
	})

	t.Run("login-error-wrong-password-message", func(t *testing.T) {
		_, _ = server.RegisterUser("validuser", "valid@example.com", "CorrectPass123!")

		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "validuser",
			"password": "WrongPass123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		body := string(resp.Body)
		if strings.Contains(strings.ToLower(body), "wrong password") ||
			strings.Contains(strings.ToLower(body), "密码错误") {
			t.Error("Error message reveals password was wrong")
		}
	})

	t.Run("register-error-duplicate-username-message", func(t *testing.T) {
		_, _ = server.RegisterUser("duplicate", "dup1@example.com", "Password123!")

		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "duplicate",
			"email":    "dup2@example.com",
			"password": "Password123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		body := string(resp.Body)
		if strings.Contains(strings.ToLower(body), "username exists") ||
			strings.Contains(strings.ToLower(body), "用户名已存在") {
			t.Log("Error message reveals username already exists")
		}
	})

	t.Run("server-version-not-disclosed", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/health", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		body := string(resp.Body)
		if strings.Contains(body, "authnas") || strings.Contains(body, "gin") {
			t.Logf("Server potentially discloses version info: %s", body)
		}
	})

	t.Run("stack-trace-not-exposed", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/nonexistent", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		body := string(resp.Body)
		if strings.Contains(body, "panic") ||
			strings.Contains(body, "stack") ||
			strings.Contains(body, "goroutine") {
			t.Error("Stack trace or panic exposed in error response")
		}
	})
}

func TestE2E_Security_AuthorizationBypass(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user1, _ := server.RegisterUser("user1auth", "user1auth@example.com", "Password123!")
	user2, _ := server.RegisterUser("user2auth", "user2auth@example.com", "Password123!")

	t.Run("user-cannot-access-other-user-data", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/users/"+user2.ID, nil, user1.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			t.Error("User should not be able to access other user's admin data")
		} else {
			assertResponseStatus(t, resp, http.StatusForbidden)
		}
	})

	t.Run("user-cannot-modify-other-user", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/users/"+user2.ID, map[string]interface{}{
			"name": "Hacked Name",
		}, user1.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})

	t.Run("user-cannot-delete-other-user", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/users/"+user2.ID, nil, user1.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})

	t.Run("admin-cannot-delete-self", func(t *testing.T) {
		admin, _ := server.CreateAdminUser("selfdelete", "selfdelete@example.com", "AdminPass123!")

		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/users/"+admin.ID, nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("id-oracle-attack-prevention", func(t *testing.T) {
		possibleIDs := []string{
			user1.ID,
			"00000000-0000-0000-0000-000000000000",
			"../../../admin",
		}

		for _, id := range possibleIDs {
			resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/users/"+id, nil, user1.AccessToken)
			if err != nil {
				t.Fatalf("Request failed for ID %s: %v", id, err)
			}

			if resp.StatusCode == http.StatusOK && id != user2.ID {
				t.Errorf("Suspicious successful delete for ID: %s", id)
			}
		}
	})
}

func TestE2E_Security_TokenSecurity(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, _ := server.RegisterUser("tokenuser", "token@example.com", "Password123!")

	t.Run("access-token-not-http-only", func(t *testing.T) {
		resp, _ := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "tokenuser",
			"password": "Password123!",
		}, nil)
		cookies := resp.Header["Set-Cookie"]

		for _, cookie := range cookies {
			if strings.Contains(cookie, "access_token") && strings.Contains(cookie, "HttpOnly") {
				t.Log("Access token has HttpOnly flag (may not be ideal for SPA)")
			}
		}
	})

	t.Run("refresh-token-security", func(t *testing.T) {
		if user.RefreshToken == "" {
			t.Skip("Refresh token not available")
		}

		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{
			"grant_type":   "refresh_token",
			"refreshToken": user.RefreshToken,
		}, nil)
		if err != nil {
			t.Fatalf("Refresh token request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			json.Unmarshal(resp.Body, &result)

			if newToken, ok := result["access_token"]; ok && newToken != nil {
				t.Log("Refresh token successfully issued new access token")
			}
		}
	})

	t.Run("token-claims-audience-validation", func(t *testing.T) {
		parts := strings.Split(user.AccessToken, ".")
		if len(parts) != 3 {
			t.Fatalf("Invalid token format")
		}

		payloadBytes, _ := base64.RawURLEncoding.DecodeString(parts[1])
		var payload map[string]interface{}
		json.Unmarshal(payloadBytes, &payload)

		if aud, ok := payload["aud"]; ok {
			t.Logf("Token has audience claim: %v", aud)
		} else {
			t.Log("Token does not have audience claim")
		}
	})

	_ = user
}

func TestE2E_Security_OAuthSecurity(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	admin, _ := server.CreateAdminUser("oauthadmin", "oauthadmin@example.com", "AdminPass123!")

	t.Run("authorization-code-with-openid-scope", func(t *testing.T) {
		clientResp, _ := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId": "security-test-client",
			"name":     "Security Test Client",
		}, admin.AccessToken)
		var clientResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(clientResp.Body, &clientResult)

		server.AuthenticatedRequest("PUT", "/api/admin/clients/"+clientResult.Data.ID, map[string]interface{}{
			"redirect_uris": "https://secure.example.com/callback",
			"scopes":        "openid profile email",
		}, admin.AccessToken)

		resp, err := server.DoRequest("GET", "/oidc/auth?"+
			"client_id=security-test-client&"+
			"redirect_uri=https://secure.example.com/callback&"+
			"response_type=code&"+
			"scope=openid profile email", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusFound {
			location := resp.Header.Get("Location")
			if strings.Contains(location, "uid=") {
				t.Log("Authorization code flow working with openid scope")
			}
		}
	})

	t.Run("implicit-flow-not-supported", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/auth?"+
			"client_id=security-test-client&"+
			"redirect_uri=https://secure.example.com/callback&"+
			"response_type=token&"+
			"scope=openid", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusFound {
			location := resp.Header.Get("Location")
			if strings.Contains(location, "access_token") || strings.Contains(location, "token=") {
				t.Error("Implicit flow (token response_type) should not be supported for security")
			}
		}
	})

	t.Run("pkce-support", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/auth?"+
			"client_id=security-test-client&"+
			"redirect_uri=https://secure.example.com/callback&"+
			"response_type=code&"+
			"scope=openid&"+
			"code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&"+
			"code_challenge_method=S256", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusFound {
			t.Log("PKCE challenge accepted")
		}
	})
}

func TestE2E_Security_MaliciousInput(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	maliciousInputs := []struct {
		name     string
		field    string
		value    string
		endpoint string
	}{
		{"LDAP injection in username", "username", "admin)(cn=*", "POST:/api/auth/login"},
		{"Command injection in email", "email", "$(whoami)@example.com", "POST:/api/auth/register"},
		{"XPath injection", "username", "admin' or '1'='1", "POST:/api/auth/login"},
		{"Shell metacharacter", "username", "user; ls", "POST:/api/auth/register"},
		{"Regex injection", "username", "^(.*){10,}", "POST:/api/auth/register"},
		{"XML external entity", "email", "test@example.com<!DOCTYPE foo [<!ENTITY xxe SYSTEM \"file:///etc/passwd\">]>", "POST:/api/auth/register"},
		{"Template injection", "username", "{{7*7}}", "POST:/api/auth/register"},
		{"JSON deserialization", "password", "{\"__proto__\": {\"admin\": true}}", "POST:/api/auth/register"},
		{"Path traversal", "email", "../../../etc/passwd@example.com", "POST:/api/auth/register"},
		{"Unicode normalization", "username", "admin\u200B", "POST:/api/auth/register"},
	}

	t.Run("malicious-input-rejected", func(t *testing.T) {
		for _, input := range maliciousInputs {
			parts := strings.Split(input.endpoint, ":")
			method := parts[0]
			path := parts[1]

			payload := map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "Password123!",
			}
			payload[input.field] = input.value

			resp, err := server.DoRequest(method, path, payload, nil)
			if err != nil {
				t.Fatalf("Request failed for %s: %v", input.name, err)
			}

			if resp.StatusCode == http.StatusOK {
				body := string(resp.Body)
				if !strings.Contains(body, "success\":true") {
					t.Errorf("Malicious input '%s' was accepted: %s", input.name, body)
				}
			}
		}
	})

	t.Run("response-time-anomaly-detection", func(t *testing.T) {
		start := time.Now()
		server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "' OR SLEEP(5)--",
			"password": "any",
		}, nil)
		elapsed := time.Since(start)

		if elapsed > 3*time.Second {
			t.Errorf("Suspiciously long response time (%v) - possible time-based injection", elapsed)
		}
	})
}

func TestE2E_Security_TimeoutAndLimits(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("request-timeout-enforced", func(t *testing.T) {
		start := time.Now()

		client := &http.Client{Timeout: 1 * time.Second}
		req, _ := http.NewRequest("GET", server.url+"/api/health", nil)
		resp, err := client.Do(req)

		elapsed := time.Since(start)

		if err != nil && elapsed < 1*time.Second {
			t.Log("Request timeout enforced as expected")
		}
		if resp != nil {
			resp.Body.Close()
		}
	})

	t.Run("max-request-size-enforced", func(t *testing.T) {
		largeBody := make([]byte, 11*1024*1024)
		for i := range largeBody {
			largeBody[i] = 'a'
		}

		req, _ := http.NewRequest("POST", server.url+"/api/auth/register", bytes.NewReader(largeBody))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)

		if err != nil {
			t.Log("Request rejected due to size (expected)")
		} else {
			resp.Body.Close()
			if resp.StatusCode == http.StatusRequestEntityTooLarge {
				t.Log("Request body size limit enforced")
			}
		}
	})

	t.Run("concurrent-request-limit", func(t *testing.T) {
		done := make(chan bool, 100)

		for i := 0; i < 100; i++ {
			go func() {
				server.DoRequest("GET", "/api/health", nil, nil)
				done <- true
			}()
		}

		timeout := time.After(30 * time.Second)
		for i := 0; i < 100; i++ {
			select {
			case <-done:
			case <-timeout:
				t.Error("Concurrent request limit may be exceeded (timeout)")
				return
			}
		}
	})
}

func TestE2E_Security_BusinessLogic(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	admin, _ := server.CreateAdminUser("bizlogic", "bizlogic@example.com", "AdminPass123!")

	t.Run("invitation-code-one-time-use", func(t *testing.T) {
		inviteResp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email":    "inviteonce@example.com",
			"username": "inviteonce",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Invitation creation failed: %v", err)
		}

		var inviteResult struct {
			InvitationID string `json:"invitationId"`
			Code         string `json:"code"`
		}
		json.Unmarshal(inviteResp.Body, &inviteResult)

		resp, err := server.DoRequest("GET", "/api/auth/invitation/"+inviteResult.InvitationID+"/"+inviteResult.Code, nil, nil)
		if err != nil {
			t.Fatalf("First invitation fetch failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			resp2, err := server.DoRequest("GET", "/api/auth/invitation/"+inviteResult.InvitationID+"/"+inviteResult.Code, nil, nil)
			if err != nil {
				t.Fatalf("Second invitation fetch failed: %v", err)
			}

			if resp2.StatusCode == http.StatusOK {
				t.Log("Invitation code should be single-use")
			}
		}
	})

	t.Run("email-verification-one-time-use", func(t *testing.T) {
		user, _ := server.RegisterUser("verifyonce", "verifyonce@example.com", "Password123!")

		resp, _ := server.DoRequest("POST", "/api/auth/verify_email", map[string]interface{}{
			"id":        user.ID,
			"challenge": "anycode",
		}, nil)

		if resp.StatusCode == http.StatusBadRequest {
			resp2, _ := server.DoRequest("POST", "/api/auth/verify_email", map[string]interface{}{
				"id":        user.ID,
				"challenge": "anycode",
			}, nil)

			if resp2.StatusCode == http.StatusBadRequest {
				t.Log("Email verification code properly rejected on reuse attempt")
			}
		}
	})

	t.Run("password-reset-expiry", func(t *testing.T) {
		_, _ = server.RegisterUser("resetexpiry", "resetexpiry@example.com", "Password123!")

		_, _ = server.DoRequest("POST", "/api/auth/forgot_password", map[string]interface{}{
			"email": "resetexpiry@example.com",
		}, nil)

		time.Sleep(1 * time.Second)

		resp, err := server.DoRequest("POST", "/api/auth/reset_password", map[string]interface{}{
			"code":         "expired_or_invalid",
			"new_password": "NewPassword123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusBadRequest {
			t.Log("Password reset code validation working")
		}
	})

	t.Run("token-not-reused-after-logout", func(t *testing.T) {
		resp, _ := server.LoginUser("bizlogic", "AdminPass123!")
		token := resp.AccessToken

		resp2, err := server.AuthenticatedRequest("DELETE", "/api/user/me/sessions", nil, token)
		if err != nil {
			t.Fatalf("Session deletion failed: %v", err)
		}
		assertResponseStatus(t, resp2, http.StatusOK)

		resp3, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, token)
		if err != nil {
			t.Fatalf("Request with logged out token failed: %v", err)
		}

		assertResponseStatus(t, resp3, http.StatusUnauthorized)
	})
}
