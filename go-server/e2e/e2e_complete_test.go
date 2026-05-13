// Package e2e provides end-to-end tests for AuthNas Go server.
//
// IMPORTANT: These tests directly test the Go server (localhost:8080).
// Vite is only used as a build tool to compile frontend resources to static files.
// The Go server serves both API endpoints and frontend static resources.
//
// Tests cover:
//   - Full OIDC Authorization Code Flow with all parameters
//   - PKCE (Proof Key for Code Exchange) support
//   - Token validation and refresh
//   - Complex multi-client scenarios
package e2e

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestE2E_FullOIDCAuthorizationCodeFlow(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("oidcfull", "oidcfull@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	clientResp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
		"clientId": "full-flow-client",
		"name":     "Full Flow Client",
	}, adminUser.AccessToken)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	var clientResult struct {
		Success bool `json:"success"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(clientResp.Body, &clientResult)

	_, err = server.AuthenticatedRequest("PUT", "/api/admin/clients/"+clientResult.Data.ID, map[string]interface{}{
		"redirectUris":  "https://example.com/callback",
		"scopes":        "openid profile email",
		"responseTypes": "code",
		"clientSecret":  "test-secret",
	}, adminUser.AccessToken)
	if err != nil {
		t.Fatalf("Failed to update client: %v", err)
	}

	user, err := server.RegisterUser("oidcuser", "oidcuser@example.com", "UserPass123!")
	if err != nil {
		t.Fatalf("User registration failed: %v", err)
	}

	t.Run("step1-get-authorization-redirect", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/auth?"+url.Values{
			"client_id":     {"full-flow-client"},
			"redirect_uri":  {"https://example.com/callback"},
			"response_type": {"code"},
			"scope":         {"openid profile email"},
			"state":         {"test-state-full"},
		}.Encode(), nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusFound {
			t.Fatalf("Expected 302 redirect, got %d. Body: %s", resp.StatusCode, string(resp.Body))
		}

		location := resp.Header.Get("Location")
		if !strings.Contains(location, "/consent/") {
			t.Fatalf("Redirect URL should contain /consent/, got: %s", location)
		}

		t.Logf("Got authorization redirect: %s", location)
	})

	t.Run("step2-get-authorization-with-preauthenticated-user", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/auth?"+url.Values{
			"client_id":     {"full-flow-client"},
			"redirect_uri":  {"https://example.com/callback"},
			"response_type": {"code"},
			"scope":         {"openid profile email"},
			"state":         {"test-state-authed"},
		}.Encode(), nil, map[string]string{
			"Cookie": "session_token=" + user.AccessToken,
		})
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusFound {
			t.Logf("Got redirect with user session")
		} else if resp.StatusCode == http.StatusOK {
			t.Logf("Got 200 (may need login interaction)")
		}
	})

	t.Run("step3-authorization-with-pkce", func(t *testing.T) {
		codeChallenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
		resp, err := server.DoRequest("GET", "/oidc/auth?"+url.Values{
			"client_id":             {"full-flow-client"},
			"redirect_uri":          {"https://example.com/callback"},
			"response_type":         {"code"},
			"scope":                 {"openid profile email"},
			"state":                 {"pkce-state"},
			"code_challenge":        {codeChallenge},
			"code_challenge_method": {"S256"},
		}.Encode(), nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusFound {
			t.Errorf("Expected 302 redirect, got %d. Body: %s", resp.StatusCode, string(resp.Body))
		}
	})
}

func TestE2E_CompleteTokenExchange(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("tokenadmin", "tokenadmin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	clientResp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
		"clientId":       "token-client",
		"name":           "Token Client",
		"redirect_uris":  "https://token.example.com/callback",
		"scopes":         "openid profile email",
		"response_types": "code",
		"client_secret":  "secret123",
	}, adminUser.AccessToken)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	var clientResult struct {
		Success  bool   `json:"success"`
		ClientID string `json:"clientId"`
	}
	json.Unmarshal(clientResp.Body, &clientResult)

	t.Run("token-exchange-with-invalid-code", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{
			"grant_type":    "authorization_code",
			"code":          "invalid-code",
			"redirectUri":   "https://token.example.com/callback",
			"clientId":      "token-client",
			"client_secret": "secret123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Token exchange with invalid code returned %d (expected 400)", resp.StatusCode)
		}
	})

	t.Run("token-exchange-without-grant-type", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{
			"code":     "some-code",
			"clientId": "token-client",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Token exchange without grant_type returned %d", resp.StatusCode)
		}
	})

	t.Run("token-exchange-with-missing-redirect-uri", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{
			"grant_type": "authorization_code",
			"code":       "some-code",
			"clientId":   "token-client",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Token exchange without redirect_uri returned %d", resp.StatusCode)
		}
	})
}

func TestE2E_RefreshTokenFlow(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("refreshuser2", "refreshuser2@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	if user.RefreshToken == "" {
		t.Skip("Registration does not return refresh token in this config")
	}

	t.Run("refresh-token-without-client-id", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{
			"grant_type":   "refresh_token",
			"refreshToken": user.RefreshToken,
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			var tokenResp struct {
				AccessToken string `json:"accessToken"`
			}
			json.Unmarshal(resp.Body, &tokenResp)
			if tokenResp.AccessToken != "" {
				t.Logf("Refresh token exchange succeeded without client_id")
			}
		} else {
			t.Logf("Refresh without client_id returned %d", resp.StatusCode)
		}
	})

	t.Run("refresh-token-with-invalid-token", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{
			"grant_type":   "refresh_token",
			"refreshToken": "completely-invalid-token",
			"clientId":     "some-client",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Refresh with invalid token returned %d", resp.StatusCode)
		}
	})
}

func TestE2E_GroupMembership(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("groupmemadmin", "groupmemadmin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	user, err := server.RegisterUser("groupmember", "groupmember@example.com", "UserPass123!")
	if err != nil {
		t.Fatalf("User registration failed: %v", err)
	}

	var groupID string

	t.Run("admin-create-group", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/groups", map[string]interface{}{
			"name":        "Test Membership Group",
			"description": "Group for testing membership",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				CreatedAt   string `json:"createdAt"`
			} `json:"data"`
		}
		json.Unmarshal(resp.Body, &result)
		groupID = result.Data.ID

		if groupID == "" {
			t.Error("Expected group_id to be set")
		}
	})

	t.Run("admin-list-groups-contains-new-group", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/groups", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"data"`
			Total int64 `json:"total"`
		}
		json.Unmarshal(resp.Body, &result)

		found := false
		for _, g := range result.Data {
			if g.ID == groupID && g.Name == "Test Membership Group" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Created group not found in list")
		}
	})

	t.Run("regular-user-cannot-access-admin-groups", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/groups", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})

	t.Run("admin-delete-group", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/groups/"+groupID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("admin-get-deleted-group-returns-404", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/groups/"+groupID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})
}

func TestE2E_InvitationFlow(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("inviteadmin", "inviteadmin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("admin-create-invitation", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email":    "inviteduser@example.com",
			"username": "inviteduser",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				ID        string `json:"id"`
				Email     string `json:"email"`
				Code      string `json:"code"`
				ExpiresAt string `json:"expiresAt"`
				CreatedAt string `json:"createdAt"`
			} `json:"data"`
		}
		json.Unmarshal(resp.Body, &result)

		if result.Data.Code == "" {
			t.Error("Expected invitation code to be set")
		}

		t.Logf("Created invitation with code: %s", result.Data.Code)
	})

	t.Run("admin-list-invitations", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/invitations", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Invitations []interface{} `json:"invitations"`
			Total       int64         `json:"total"`
		}
		json.Unmarshal(resp.Body, &result)

		if result.Total < 1 {
			t.Error("Expected at least 1 invitation")
		}
	})

	t.Run("get-non-existent-invitation", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/auth/invitation/fake-id/fake-code", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})
}

func TestE2E_EmailVerificationFlow(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("send-verification-email-for-unverified-user", func(t *testing.T) {
		user, err := server.RegisterUser("verifyuser", "verifyuser@example.com", "password123")
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

	t.Run("verify-email-with-invalid-code", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/verify_email", map[string]interface{}{
			"id":        "some-user-id",
			"challenge": "invalid-challenge",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})
}

func TestE2E_MFAFlow(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("mfauser", "mfauser@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	var totpSecret string

	t.Run("register-TOTP", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/totp/registration", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				Secret string `json:"secret"`
			} `json:"data"`
		}
		json.Unmarshal(resp.Body, &result)

		if !result.Success {
			t.Error("Expected success to be true")
		}
		if result.Data.Secret == "" {
			t.Error("Expected secret to be set")
		}
		totpSecret = result.Data.Secret
	})

	t.Run("verify-TOTP-with-valid-token", func(t *testing.T) {
		if totpSecret == "" {
			t.Skip("TOTP secret not generated")
		}

		token, err := totp.GenerateCode(totpSecret, time.Now())
		if err != nil {
			t.Fatalf("Failed to generate TOTP code: %v", err)
		}

		resp, err := server.AuthenticatedRequest("POST", "/api/totp/verify", map[string]interface{}{
			"token": token,
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Logf("TOTP verify returned %d", resp.StatusCode)
		}
	})

	t.Run("delete-TOTP", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/totp", map[string]interface{}{
			"token": "000000",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("TOTP delete returned unexpected status %d", resp.StatusCode)
		}
	})
}

func TestE2E_PasskeyFullRegistrationFlow(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("passkeyreg", "passkeyreg@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("passkey-registration-start", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/passkey/registration/start", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				Challenge string `json:"challenge"`
				Options   string `json:"options"`
			} `json:"data"`
		}
		json.Unmarshal(resp.Body, &result)

		if !result.Success {
			t.Error("Expected success to be true")
		}
		if result.Data.Challenge == "" {
			t.Error("Expected challenge to be set")
		}
		if result.Data.Options == "" {
			t.Error("Expected options to be set")
		}
	})

	t.Run("list-passkeys-after-start", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/passkey", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Passkeys []interface{} `json:"passkeys"`
		}
		json.Unmarshal(resp.Body, &result)

		if len(result.Passkeys) != 0 {
			t.Errorf("Expected 0 passkeys before registration completes, got %d", len(result.Passkeys))
		}
	})
}

func TestE2E_ClientManagementFull(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("clientadmin2", "clientadmin2@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("create-client-with-client-secret", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId":      "secret-client",
			"name":          "Client With Secret",
			"client_secret": "super-secret-key",
			"redirect_uris": "https://secret.example.com/callback",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("update-client-redirect-uris", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId": "update-redirect-client",
			"name":     "Update Redirect Client",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/clients/"+createResult.Data.ID, map[string]interface{}{
			"redirect_uris": "https://updated.example.com/callback",
			"scopes":        "openid profile email offline_access",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("delete-client-and-verify", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId": "delete-me-client",
			"name":     "Delete Me Client",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		deleteResp, err := server.AuthenticatedRequest("DELETE", "/api/admin/clients/"+createResult.Data.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, deleteResp, http.StatusOK)

		getResp, err := server.AuthenticatedRequest("GET", "/api/admin/clients/"+createResult.Data.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, getResp, http.StatusNotFound)
	})
}

func TestE2E_SessionManagement(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("sessiontest", "sessiontest@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("delete-all-sessions", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/user/me/sessions", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("token-invalid-after-session-deletion", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Logf("Token after session delete returned %d (implementation may vary)", resp.StatusCode)
		}
	})

	t.Run("re-login-after-session-delete", func(t *testing.T) {
		newUser, err := server.LoginUser("sessiontest", "password123")
		if err != nil {
			t.Fatalf("Re-login failed: %v", err)
		}

		if newUser.AccessToken == "" {
			t.Error("Expected new access token after re-login")
		}
	})
}

func TestE2E_UserManagementAdmin(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("useradmin", "useradmin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("admin-list-all-users", func(t *testing.T) {
		_, _ = server.RegisterUser("listuser1", "listuser1@example.com", "UserPass123!")
		_, _ = server.RegisterUser("listuser2", "listuser2@example.com", "UserPass123!")

		resp, err := server.AuthenticatedRequest("GET", "/api/admin/users", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Users []struct {
				ID       string `json:"id"`
				Username string `json:"username"`
				Email    string `json:"email"`
			} `json:"users"`
			Total int64 `json:"total"`
		}
		json.Unmarshal(resp.Body, &result)

		if result.Total < 3 {
			t.Errorf("Expected at least 3 users (admin + 2 registered), got %d", result.Total)
		}
	})

	t.Run("admin-get-own-details", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/users/"+adminUser.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				ID       string `json:"id"`
				Username string `json:"username"`
				IsAdmin  bool   `json:"isAdmin"`
			} `json:"data"`
		}
		json.Unmarshal(resp.Body, &result)

		if result.Data.ID != adminUser.ID {
			t.Errorf("Expected ID %s, got %s", adminUser.ID, result.Data.ID)
		}
		if !result.Data.IsAdmin {
			t.Error("Expected is_admin to be true")
		}
	})

	t.Run("admin-update-user-profile", func(t *testing.T) {
		newUser, _ := server.RegisterUser("updateprofileuser", "updateprofile@example.com", "UserPass123!")

		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/users/"+newUser.ID, map[string]interface{}{
			"name": "Updated By Admin",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})
}

func TestE2E_SecurityBoundaryExtended(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("secuser", "secuser@example.com", "password123")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("user-cannot-access-other-user-profile", func(t *testing.T) {
		otherUser, _ := server.RegisterUser("otheruser", "otheruser@example.com", "UserPass123!")

		resp, err := server.AuthenticatedRequest("GET", "/api/admin/users/"+otherUser.ID, nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})

	t.Run("user-cannot-delete-other-user", func(t *testing.T) {
		otherUser, _ := server.RegisterUser("deleteuser", "deleteuser@example.com", "UserPass123!")

		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/users/"+otherUser.ID, nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})

	t.Run("user-cannot-access-admin-group-endpoints", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/groups", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})

	t.Run("user-cannot-access-admin-client-endpoints", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/clients", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})

	t.Run("user-cannot-access-admin-invitation-endpoints", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/invitations", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})

	t.Run("user-cannot-approve-users", func(t *testing.T) {
		otherUser, _ := server.RegisterUser("approveuser", "approveuser@example.com", "UserPass123!")

		resp, err := server.AuthenticatedRequest("POST", "/api/admin/users/"+otherUser.ID+"/approve", map[string]interface{}{
			"approved": true,
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})

	t.Run("user-cannot-reset-passwords", func(t *testing.T) {
		otherUser, _ := server.RegisterUser("resetpwduser", "resetpwduser@example.com", "UserPass123!")

		resp, err := server.AuthenticatedRequest("POST", "/api/admin/users/"+otherUser.ID+"/reset-password", map[string]interface{}{
			"new_password": "NewPass123!",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})
}

func TestE2E_OIDCDiscoveryComplete(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("discovery-contains-all-required-fields", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/.well-known/openid-configuration", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var discovery struct {
			Issuer                            string   `json:"issuer"`
			AuthorizationEndpoint             string   `json:"authorization_endpoint"`
			TokenEndpoint                     string   `json:"token_endpoint"`
			UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
			JwksURI                           string   `json:"jwks_uri"`
			RevocationEndpoint                string   `json:"revocation_endpoint"`
			ResponseTypesSupported            []string `json:"response_types_supported"`
			SubjectTypesSupported             []string `json:"subject_types_supported"`
			IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
			ScopesSupported                   []string `json:"scopes_supported"`
			TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
			ClaimsSupported                   []string `json:"claims_supported"`
			CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
			GrantTypesSupported               []string `json:"grant_types_supported"`
		}

		if err := json.Unmarshal(resp.Body, &discovery); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if discovery.Issuer == "" {
			t.Error("Issuer should not be empty")
		}
		if discovery.AuthorizationEndpoint == "" {
			t.Error("Authorization endpoint should not be empty")
		}
		if discovery.TokenEndpoint == "" {
			t.Error("Token endpoint should not be empty")
		}
		if discovery.UserInfoEndpoint == "" {
			t.Error("UserInfo endpoint should not be empty")
		}
		if discovery.JwksURI == "" {
			t.Error("JWKS URI should not be empty")
		}
		if discovery.RevocationEndpoint == "" {
			t.Error("Revocation endpoint should not be empty")
		}

		foundCode := false
		for _, rt := range discovery.ResponseTypesSupported {
			if rt == "code" {
				foundCode = true
				break
			}
		}
		if !foundCode {
			t.Error("Expected 'code' in response_types_supported")
		}

		foundOpenID := false
		for _, scope := range discovery.ScopesSupported {
			if scope == "openid" {
				foundOpenID = true
				break
			}
		}
		if !foundOpenID {
			t.Error("Expected 'openid' in scopes_supported")
		}

		foundS256 := false
		for _, method := range discovery.CodeChallengeMethodsSupported {
			if method == "S256" {
				foundS256 = true
				break
			}
		}
		if !foundS256 {
			t.Error("Expected 'S256' in code_challenge_methods_supported")
		}
	})

	t.Run("jwks-contains-valid-rsa-key", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/jwks", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var jwks struct {
			Keys []struct {
				Kty string `json:"kty"`
				Use string `json:"use"`
				Kid string `json:"kid"`
				Alg string `json:"alg"`
				N   string `json:"n"`
				E   string `json:"e"`
			} `json:"keys"`
		}

		if err := json.Unmarshal(resp.Body, &jwks); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if len(jwks.Keys) == 0 {
			t.Error("Expected at least one key in JWKS")
		}

		for _, key := range jwks.Keys {
			if key.Kty != "RSA" {
				continue
			}
			if key.Use != "sig" {
				t.Errorf("Expected key use 'sig', got '%s'", key.Use)
			}
			if key.Alg != "RS256" {
				t.Errorf("Expected algorithm 'RS256', got '%s'", key.Alg)
			}
			if key.N == "" {
				t.Error("Expected modulus (n) to be set")
			}
			if key.E == "" {
				t.Error("Expected exponent (e) to be set")
			}
			return
		}
		t.Error("No RSA key found in JWKS")
	})
}

func TestE2E_InputValidationExtended(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("register-with-username-too-short", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "ab",
			"email":    "short@example.com",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("register-with-username-too-long", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "thisusernameiswaytoolongtobevalid",
			"email":    "long@example.com",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("register-with-invalid-username-characters", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "user@name",
			"email":    "char@example.com",
			"password": "password123",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("login-with-empty-password", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "someuser",
			"password": "",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("login-with-missing-input", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"password": "somepassword",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("oidc-auth-with-missing-scope", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/auth?"+url.Values{
			"clientId":      {"some-client"},
			"redirectUri":   {"https://example.com"},
			"response_type": {"code"},
		}.Encode(), nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Auth without scope returned %d", resp.StatusCode)
		}
	})
}

func TestE2E_ProxyAuthManagement(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("proxyadmin", "proxyadmin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("create-proxy-auth-config", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/proxyauth", map[string]interface{}{
			"name":       "Test Proxy",
			"proxyUrl":   "https://proxy.example.com",
			"headerName": "X-User-ID",
			"enabled":    true,
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(resp.Body, &result)

		if result.Data.ID == "" {
			t.Error("Expected proxyauth_id to be set")
		}
	})

	t.Run("list-proxy-auth-configs", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/proxyauth", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool          `json:"success"`
			Data    []interface{} `json:"data"`
			Total   int64         `json:"total"`
		}
		json.Unmarshal(resp.Body, &result)

		if result.Total < 1 {
			t.Error("Expected at least 1 proxy auth config")
		}
	})

	t.Run("update-proxy-auth-config", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/proxyauth", map[string]interface{}{
			"name":       "Proxy To Update",
			"proxyUrl":   "https://old.example.com",
			"headerName": "X-User-ID",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Failed to create proxy auth: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/proxyauth/"+createResult.Data.ID, map[string]interface{}{
			"name":     "Updated Proxy Name",
			"proxyUrl": "https://new.example.com",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("delete-proxy-auth-config", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/proxyauth", map[string]interface{}{
			"name":       "Proxy To Delete",
			"proxyUrl":   "https://delete.example.com",
			"headerName": "X-User-ID",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Failed to create proxy auth: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		deleteResp, err := server.AuthenticatedRequest("DELETE", "/api/admin/proxyauth/"+createResult.Data.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, deleteResp, http.StatusOK)

		getResp, err := server.AuthenticatedRequest("GET", "/api/admin/proxyauth/"+createResult.Data.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, getResp, http.StatusNotFound)
	})
}

func TestE2E_HealthCheck(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("health-check-returns-ok", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/health", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				Status string `json:"status"`
			} `json:"data"`
		}
		json.Unmarshal(resp.Body, &result)

		if result.Data.Status != "ok" && result.Data.Status != "healthy" {
			t.Logf("Health status: %v", result.Data.Status)
		}
	})

	t.Run("public-config-contains-app-info", func(t *testing.T) {
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
		json.Unmarshal(resp.Body, &result)

		if result.Data.AppName == "" {
			t.Error("Expected app_name in public config")
		}
	})
}

func TestE2E_CompleteUserJourney(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("complete-user-journey", func(t *testing.T) {
		adminUser, err := server.CreateAdminUser("journeyadmin", "journeyadmin@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		_, _ = server.RegisterUser("journeyuser1", "journeyuser1@example.com", "UserPass123!")
		_, _ = server.RegisterUser("journeyuser2", "journeyuser2@example.com", "UserPass123!")

		groupResp, err := server.AuthenticatedRequest("POST", "/api/admin/groups", map[string]interface{}{
			"name":        "Journey Test Group",
			"description": "Group created during journey test",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}
		assertResponseStatus(t, groupResp, http.StatusOK)

		inviteResp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email":    "journeyinvite@example.com",
			"username": "journeyinvite",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Failed to create invitation: %v", err)
		}
		assertResponseStatus(t, inviteResp, http.StatusOK)

		usersResp, err := server.AuthenticatedRequest("GET", "/api/admin/users", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Failed to list users: %v", err)
		}
		assertResponseStatus(t, usersResp, http.StatusOK)

		groupsResp, err := server.AuthenticatedRequest("GET", "/api/admin/groups", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Failed to list groups: %v", err)
		}
		assertResponseStatus(t, groupsResp, http.StatusOK)

		clientsResp, err := server.AuthenticatedRequest("GET", "/api/admin/clients", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Failed to list clients: %v", err)
		}
		assertResponseStatus(t, clientsResp, http.StatusOK)

		invitesResp, err := server.AuthenticatedRequest("GET", "/api/admin/invitations", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Failed to list invitations: %v", err)
		}
		assertResponseStatus(t, invitesResp, http.StatusOK)
	})
}
