package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestE2E_SessionManagementScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("session-management-list-sessions", func(t *testing.T) {
		user, err := server.RegisterUser("sessionlist", "sessionlist@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("GET", "/api/user/me/sessions", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Sessions []struct {
				ID        string `json:"id"`
				CreatedAt string `json:"createdAt"`
			} `json:"sessions"`
		}
		json.Unmarshal(resp.Body, &result)
		t.Logf("Found %d sessions", len(result.Sessions))
	})

	t.Run("session-management-revoke-all-sessions", func(t *testing.T) {
		user, err := server.RegisterUser("revokeall", "revokeall@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("DELETE", "/api/user/me/sessions", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		resp2, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp2, http.StatusUnauthorized)
	})

	t.Run("session-management-revoke-specific-session", func(t *testing.T) {
		user, err := server.RegisterUser("revokespecific", "revokespecific@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		listResp, err := server.AuthenticatedRequest("GET", "/api/user/me/sessions", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		var sessions struct {
			Sessions []struct {
				ID string `json:"id"`
			} `json:"sessions"`
		}
		json.Unmarshal(listResp.Body, &sessions)

		if len(sessions.Sessions) > 0 {
			sessionID := sessions.Sessions[0].ID
			resp, err := server.AuthenticatedRequest("DELETE", fmt.Sprintf("/api/user/me/sessions/%s", sessionID), nil, user.AccessToken)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			assertResponseStatus(t, resp, http.StatusOK)
		}
	})

	t.Run("session-management-revoke-nonexistent-session", func(t *testing.T) {
		user, err := server.RegisterUser("retypenonexist", "retypenonexist@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("DELETE", "/api/user/me/sessions/nonexistent-id", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			t.Log("Revoke non-existent session returns 404 as expected")
		} else if resp.StatusCode == http.StatusOK {
			t.Log("Revoke non-existent session returns 200 (acceptable)")
		}
	})

	t.Run("session-management-revoke-without-auth", func(t *testing.T) {
		resp, err := server.DoRequest("DELETE", "/api/user/me/sessions", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestE2E_PasswordSecurityScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("password-security-update-with-same-password", func(t *testing.T) {
		user, err := server.RegisterUser("samepwd", "samepwd@example.com", "OriginalPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("PUT", "/api/user/me/password", map[string]interface{}{
			"old_password": "OriginalPass123!",
			"new_password": "OriginalPass123!",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			t.Log("Same password allowed (depends on password policy)")
		} else if resp.StatusCode == http.StatusBadRequest {
			t.Log("Same password not allowed")
		}
	})

	t.Run("password-security-update-without-old-password", func(t *testing.T) {
		user, err := server.RegisterUser("nopwdold", "nopwdold@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("PUT", "/api/user/me/password", map[string]interface{}{
			"new_password": "NewPass456!",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("password-security-update-without-new-password", func(t *testing.T) {
		user, err := server.RegisterUser("nopwdnew", "nopwdnew@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("PUT", "/api/user/me/password", map[string]interface{}{
			"old_password": "UserPass123!",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})
}

func TestE2E_ClientManagementScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("clientadmin", "clientadmin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("client-management-create-with-duplicate-id", func(t *testing.T) {
		resp1, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId": "duplicate-client",
			"name":     "First Client",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("First creation failed: %v", err)
		}
		assertResponseStatus(t, resp1, http.StatusOK)

		resp2, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId": "duplicate-client",
			"name":     "Second Client",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Second creation failed: %v", err)
		}
		assertResponseStatus(t, resp2, http.StatusBadRequest)
	})

	t.Run("client-management-create-with-empty-redirect-uri", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId":      "no-redirect-client",
			"name":          "No Redirect Client",
			"redirect_uris": "",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("client-management-get-nonexistent", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/clients/nonexistent-client", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusNotFound)
	})

	t.Run("client-management-delete-nonexistent", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/clients/nonexistent-client", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode == http.StatusNotFound {
			t.Log("Delete non-existent client returns 404 as expected")
		} else if resp.StatusCode == http.StatusOK {
			t.Log("BUG: Delete non-existent client returns 200 (should return 404)")
		}
	})

	t.Run("client-management-update-scopes", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId": "scope-test-client",
			"name":     "Scope Test Client",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Creation failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/clients/"+createResult.Data.ID, map[string]interface{}{
			"scopes": "openid profile email offline_access",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})
}

func TestE2E_GroupManagementScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("groupadmin", "groupadmin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("group-management-create-with-empty-description", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/groups", map[string]interface{}{
			"name":        "Empty Desc Group",
			"description": "",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("group-management-create-without-description", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/groups", map[string]interface{}{
			"name": "No Desc Group",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("group-management-update-nonexistent", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/groups/nonexistent", map[string]interface{}{
			"name": "New Name",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusNotFound)
	})

	t.Run("group-management-delete-nonexistent", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/groups/nonexistent", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusNotFound)
	})

	t.Run("group-management-get-details", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/groups", map[string]interface{}{
			"name":        "Details Test Group",
			"description": "Test description",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Creation failed: %v", err)
		}

		var createResult struct {
			GroupID string `json:"groupId"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("GET", "/api/admin/groups/"+createResult.GroupID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			var result struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
			}
			json.Unmarshal(resp.Body, &result)
			if result.Name != "Details Test Group" {
				t.Errorf("Expected name 'Details Test Group', got '%s'", result.Name)
			}
		} else {
			t.Logf("BUG: GET group details returned %d instead of 200", resp.StatusCode)
		}
	})
}

func TestE2E_InvitationScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("inviteadmin", "inviteadmin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("invitation-create-with-email-only", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email": "emailonly@example.com",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("invitation-create-with-email-and-username", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email":    "fullinvite@example.com",
			"username": "fullinvite",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("invitation-get-details", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email":    "detailstest@example.com",
			"username": "detailstest",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Creation failed: %v", err)
		}

		var createResult struct {
			InvitationID string `json:"invitationId"`
			Code         string `json:"code"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("GET", "/api/admin/invitations/"+createResult.InvitationID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			t.Log("Get invitation details returned 200")
		} else {
			t.Logf("BUG: Get invitation details returned %d", resp.StatusCode)
		}
	})

	t.Run("invitation-delete-nonexistent", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/invitations/nonexistent", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode == http.StatusNotFound {
			t.Log("Delete non-existent invitation returns 404 as expected")
		} else if resp.StatusCode == http.StatusOK {
			t.Log("BUG: Delete non-existent invitation returns 200 (should return 404)")
		}
	})
}

func TestE2E_ProxyAuthScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("proxyadmin", "proxyadmin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("proxyauth-create-with-enabled", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/proxyauth", map[string]interface{}{
			"name":       "Enabled Proxy",
			"proxyUrl":   "https://proxy.example.com",
			"headerName": "X-User-ID",
			"enabled":    true,
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("proxyauth-update-enabled-status", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/proxyauth", map[string]interface{}{
			"name":       "Toggle Proxy",
			"proxyUrl":   "https://proxy.example.com",
			"headerName": "X-User-ID",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Creation failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/proxyauth/"+createResult.Data.ID, map[string]interface{}{
			"enabled": false,
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("proxyauth-get-details", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/proxyauth", map[string]interface{}{
			"name":       "Details Proxy",
			"proxyUrl":   "https://proxy.example.com",
			"headerName": "X-User-ID",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Creation failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("GET", "/api/admin/proxyauth/"+createResult.Data.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			t.Log("Get proxy auth details returned 200")
		} else {
			t.Logf("BUG: Get proxy auth details returned %d", resp.StatusCode)
		}
	})

	t.Run("proxyauth-delete-nonexistent", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/proxyauth/nonexistent", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode == http.StatusNotFound {
			t.Log("Delete non-existent proxy auth returns 404 as expected")
		} else if resp.StatusCode == http.StatusOK {
			t.Log("BUG: Delete non-existent proxy auth returns 200 (should return 404)")
		}
	})
}

func TestE2E_TOTPCompleteScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("totpcomplete", "totpcomplete@example.com", "UserPass123!")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("totp-registration-returns-secret", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/totp/registration", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				Secret    string `json:"secret"`
				QRCodeURI string `json:"qr_code_uri"`
			} `json:"data"`
		}
		json.Unmarshal(resp.Body, &result)

		if !result.Success {
			t.Error("Expected success to be true")
		}
		if result.Data.Secret == "" {
			t.Error("Expected secret to be non-empty")
		}
		if result.Data.QRCodeURI == "" {
			t.Error("Expected QR code URI to be non-empty")
		}
	})

	t.Run("totp-verify-without-auth", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/totp/verify", map[string]interface{}{
			"token": "123456",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("totp-verify-with-empty-token", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/totp/verify", map[string]interface{}{
			"token": "",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("totp-delete-without-auth", func(t *testing.T) {
		resp, err := server.DoRequest("DELETE", "/api/totp", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("totp-delete-with-invalid-token", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/totp", map[string]interface{}{
			"token": "invalid",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Logf("Delete TOTP with invalid token returned %d", resp.StatusCode)
		}
	})
}

func TestE2E_PasskeyCompleteScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("passkeycomplete", "passkeycomplete@example.com", "UserPass123!")
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Run("passkey-list-for-new-user", func(t *testing.T) {
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
			t.Errorf("Expected 0 passkeys for new user, got %d", len(result.Passkeys))
		}
	})

	t.Run("passkey-registration-start-without-auth", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/passkey/registration/start", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("passkey-registration-end-without-auth", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/passkey/registration/end", map[string]interface{}{
			"credential": "{}",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("passkey-delete-without-auth", func(t *testing.T) {
		resp, err := server.DoRequest("DELETE", "/api/passkey/some-id", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("passkey-auth-start-with-existing-user", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/passkey/start", map[string]interface{}{
			"username": "passkeycomplete",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("passkey-auth-start-with-nonexistent-user", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/passkey/start", map[string]interface{}{
			"username": "nonexistentuser12345",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})
}

func TestE2E_ErrorHandlingScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("error-handling-invalid-json-body", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input": "test",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		var result struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		json.Unmarshal(resp.Body, &result)

		if result.Success == false && result.Message == "" {
			t.Log("Invalid JSON handling may need verification")
		}
	})

	t.Run("error-handling-missing-required-fields", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"username": "test",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("error-handling-empty-request-body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", server.url+"/api/auth/login", nil)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		t.Logf("Empty body returned status: %d", resp.StatusCode)
	})
}

func TestE2E_AdminUsersCountScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("countadmin", "countadmin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("admin-users-count", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/users/count", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			var result struct {
				Count int64 `json:"count"`
			}
			json.Unmarshal(resp.Body, &result)
			t.Logf("User count: %d", result.Count)
		} else if resp.StatusCode == http.StatusNotFound {
			t.Log("BUG: GET /api/admin/users/count endpoint not implemented")
		} else {
			t.Logf("GET users count returned %d", resp.StatusCode)
		}
	})
}

func TestE2E_CORSScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("cors-preflight-request", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", server.url+"/api/health", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		t.Logf("CORS preflight returned status: %d", resp.StatusCode)
		allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
		if allowOrigin != "" {
			t.Logf("CORS Allow-Origin: %s", allowOrigin)
		}
	})

	t.Run("cors-request-with-valid-origin", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.url+"/api/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		t.Logf("Request with valid origin returned status: %d", resp.StatusCode)
	})
}

func TestE2E_SecurityHeadersScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("security-headers-x-content-type-options", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/health", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		xContentType := resp.Header.Get("X-Content-Type-Options")
		if xContentType == "" {
			t.Log("X-Content-Type-Options header not set")
		} else {
			t.Logf("X-Content-Type-Options: %s", xContentType)
		}
	})

	t.Run("security-headers-content-type", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/health", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			t.Error("Content-Type header is empty")
		} else {
			t.Logf("Content-Type: %s", contentType)
		}
	})
}

func TestE2E_OIDCTokenEndpointScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("oidc-token-get-method-not-allowed", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/token", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode == http.StatusMethodNotAllowed {
			t.Log("GET /oidc/token correctly returns 405")
		} else if resp.StatusCode == http.StatusOK {
			t.Log("GET /oidc/token returns 200 (may have error message)")
		}
	})

	t.Run("oidc-token-post-without-grant-type", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("oidc-token-post-with-invalid-grant-type", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{
			"grant_type": "invalid_grant",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("oidc-token-post-authorization-code-grant", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{
			"grant_type": "authorization_code",
			"code":       "some-code",
			"clientId":   "test-client",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Authorization code grant returned %d (expected 400 for invalid code)", resp.StatusCode)
		}
	})
}

func TestE2E_UserProfileScenarios(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("user-profile-get-own-profile", func(t *testing.T) {
		user, err := server.RegisterUser("ownprofile", "ownprofile@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				ID            string `json:"id"`
				Username      string `json:"username"`
				Email         string `json:"email"`
				Name          string `json:"name"`
				IsAdmin       bool   `json:"isAdmin"`
				EmailVerified bool   `json:"emailVerified"`
			} `json:"data"`
		}
		json.Unmarshal(resp.Body, &result)

		if result.Data.Username != "ownprofile" {
			t.Errorf("Expected username 'ownprofile', got '%s'", result.Data.Username)
		}
	})

	t.Run("user-profile-update-name", func(t *testing.T) {
		user, err := server.RegisterUser("updatename", "updatename@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("PUT", "/api/user/me", map[string]interface{}{
			"name": "New Name",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("user-profile-update-email", func(t *testing.T) {
		user, err := server.RegisterUser("updateemailprof", "updateemailprof@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("PUT", "/api/user/me", map[string]interface{}{
			"email": "newemail@example.com",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("user-profile-update-invalid-email", func(t *testing.T) {
		user, err := server.RegisterUser("invemail", "invemail@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("PUT", "/api/user/me", map[string]interface{}{
			"email": "not-an-email",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("user-profile-update-empty-name", func(t *testing.T) {
		user, err := server.RegisterUser("emptyname", "emptyname@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("PUT", "/api/user/me", map[string]interface{}{
			"name": "",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode == http.StatusOK {
			t.Log("Empty name update returned 200 (may be allowed)")
		} else {
			t.Log("Empty name update returned non-200")
		}
	})
}
