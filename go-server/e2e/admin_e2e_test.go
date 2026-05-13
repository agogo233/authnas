// Package e2e provides end-to-end tests for AuthNas Go server.
//
// IMPORTANT: These tests directly test the Go server (localhost:8080).
// Vite is only used as a build tool to compile frontend resources to static files.
// The Go server serves both API endpoints and frontend static resources.
//
// Tests cover:
//   - Admin user management (CRUD)
//   - Admin group management (CRUD)
//   - Admin client management (CRUD)
//   - Admin invitation management
//   - Authorization and permission checks
package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestE2E_Admin_Users(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("adminusers", "adminusers@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	_, err = server.RegisterUser("regularuser1", "regular1@example.com", "UserPass123!")
	if err != nil {
		t.Fatalf("User registration failed: %v", err)
	}

	_, err = server.RegisterUser("regularuser2", "regular2@example.com", "UserPass123!")
	if err != nil {
		t.Fatalf("User registration failed: %v", err)
	}

	t.Run("admin list all users", func(t *testing.T) {
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
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if result.Total < 3 {
			t.Errorf("Expected at least 3 users, got %d", result.Total)
		}
	})

	t.Run("regular user cannot access admin user list", func(t *testing.T) {
		regularUser, _ := server.RegisterUser("cannotlistusers", "cannotlist@example.com", "UserPass123!")

		resp, err := server.AuthenticatedRequest("GET", "/api/admin/users", nil, regularUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})

	t.Run("admin create new user", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/users", map[string]interface{}{
			"username": "newadminuser",
			"email":    "newadmin@example.com",
			"password": "NewUserPass123!",
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
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if !result.Success {
			t.Error("Expected success to be true")
		}
		if result.Data.ID == "" {
			t.Error("Expected user id to be set")
		}
	})

	t.Run("admin get specific user", func(t *testing.T) {
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
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if result.Data.ID != adminUser.ID {
			t.Errorf("Expected ID '%s', got '%s'", adminUser.ID, result.Data.ID)
		}
	})

	t.Run("admin update user", func(t *testing.T) {
		newName := "Updated Admin Name"
		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/users/"+adminUser.ID, map[string]interface{}{
			"name": newName,
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("admin delete user", func(t *testing.T) {
		newUser, _ := server.RegisterUser("todelete", "todelete@example.com", "UserPass123!")

		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/users/"+newUser.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("admin cannot delete themselves", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/users/"+adminUser.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("admin get non-existent user", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/users/nonexistent", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})

	t.Run("admin update non-existent user", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/users/nonexistent", map[string]interface{}{
			"name": "Some Name",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})

	t.Run("admin delete non-existent user", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/users/nonexistent", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})

	t.Run("admin reset user password", func(t *testing.T) {
		targetUser, _ := server.RegisterUser("passwordreset", "passwordreset@example.com", "OldPass123!")

		resp, err := server.AuthenticatedRequest("POST", "/api/admin/users/"+targetUser.ID+"/reset-password", map[string]interface{}{
			"newPassword": "ResetPass123!",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		loginResp, err := server.LoginUser("passwordreset", "ResetPass123!")
		if err != nil {
			t.Fatalf("Login with reset password failed: %v", err)
		}
		if loginResp.AccessToken == "" {
			t.Error("Expected access token after password reset")
		}
	})
}

func TestE2E_Admin_Groups(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("admingroups", "admingroups@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("admin create group", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/groups", map[string]interface{}{
			"name":        "Test Group",
			"description": "A test group for E2E testing",
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
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if !result.Success {
			t.Error("Expected success to be true")
		}
		if result.Data.ID == "" {
			t.Error("Expected group_id to be set")
		}
	})

	t.Run("admin list groups", func(t *testing.T) {
		_, _ = server.AuthenticatedRequest("POST", "/api/admin/groups", map[string]interface{}{
			"name": "List Test Group",
		}, adminUser.AccessToken)

		resp, err := server.AuthenticatedRequest("GET", "/api/admin/groups", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Groups []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"groups"`
			Total int64 `json:"total"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if result.Total < 1 {
			t.Error("Expected at least 1 group")
		}
	})

	t.Run("admin update group", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/groups", map[string]interface{}{
			"name": "Group To Update",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/groups/"+createResult.Data.ID, map[string]interface{}{
			"name":        "Updated Group Name",
			"description": "Updated description",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("admin delete group", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/groups", map[string]interface{}{
			"name": "Group To Delete",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/groups/"+createResult.Data.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("admin get non-existent group", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/groups/nonexistent", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})
}

func TestE2E_Admin_Clients(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("adminclients", "adminclients@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("admin create client", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId": "test-client",
			"name":     "Test OAuth Client",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success bool `json:"success"`
			Data    struct {
				ID       string `json:"id"`
				ClientID string `json:"clientId"`
				Name     string `json:"name"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if !result.Success {
			t.Error("Expected success to be true")
		}
	})

	t.Run("admin list clients", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/clients", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Clients []struct {
				ID       string `json:"id"`
				ClientID string `json:"clientId"`
				Name     string `json:"name"`
			} `json:"clients"`
			Total int64 `json:"total"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if result.Total < 1 {
			t.Error("Expected at least 1 client")
		}
	})

	t.Run("admin update client", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId": "client-to-update",
			"name":     "Original Name",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		newName := "Updated Client Name"
		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/clients/"+createResult.Data.ID, map[string]interface{}{
			"name": newName,
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("admin delete client", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId": "client-to-delete",
			"name":     "Client To Delete",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/clients/"+createResult.Data.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("admin update non-existent client", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/clients/nonexistent", map[string]interface{}{
			"name": "New Name",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})
}

func TestE2E_Admin_Invitations(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("admininvite", "admininvite@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("admin create invitation", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email":    "invited@example.com",
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
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if !result.Success {
			t.Error("Expected success to be true")
		}
		if result.Data.Code == "" {
			t.Error("Expected invitation code to be set")
		}
	})

	t.Run("admin list invitations", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/invitations", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Invitations []struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"invitations"`
			Total int64 `json:"total"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if result.Total < 1 {
			t.Error("Expected at least 1 invitation")
		}
	})

	t.Run("admin delete invitation", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email": "todelete@example.com",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/invitations/"+createResult.Data.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})
}

func TestE2E_Admin_ProxyAuth(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("adminproxy", "adminproxy@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("admin create proxy auth", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/proxyauth", map[string]interface{}{
			"name":       "Test Proxy Auth",
			"proxyUrl":   "https://proxy.example.com",
			"headerName": "X-User-ID",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			Success     bool   `json:"success"`
			ProxyAuthID string `json:"proxyauthId"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if !result.Success {
			t.Error("Expected success to be true")
		}
	})

	t.Run("admin list proxy auth", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/admin/proxyauth", nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var result struct {
			ProxyAuths []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"proxyauths"`
			Total int64 `json:"total"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if result.Total < 1 {
			t.Error("Expected at least 1 proxy auth")
		}
	})

	t.Run("admin update proxy auth", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/proxyauth", map[string]interface{}{
			"name":       "Proxy To Update",
			"proxyUrl":   "https://old.example.com",
			"headerName": "X-User-ID",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		newName := "Updated Proxy Name"
		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/proxyauth/"+createResult.Data.ID, map[string]interface{}{
			"name": newName,
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("admin delete proxy auth", func(t *testing.T) {
		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/proxyauth", map[string]interface{}{
			"name":       "Proxy To Delete",
			"proxyUrl":   "https://delete.example.com",
			"headerName": "X-User-ID",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		resp, err := server.AuthenticatedRequest("DELETE", "/api/admin/proxyauth/"+createResult.Data.ID, nil, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("admin update non-existent proxy auth", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("PUT", "/api/admin/proxyauth/nonexistent", map[string]interface{}{
			"name": "New Name",
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})
}

func TestE2E_Admin_ApproveUser(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	adminUser, err := server.CreateAdminUser("adminapprove", "adminapprove@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Admin creation failed: %v", err)
	}

	t.Run("admin approve unapproved user", func(t *testing.T) {
		unapprovedUser, _ := server.RegisterUser("unapproveduser", "unapproved@example.com", "UserPass123!")

		resp, err := server.AuthenticatedRequest("POST", "/api/admin/users/"+unapprovedUser.ID+"/approve", map[string]interface{}{
			"approved": true,
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("admin approve non-existent user", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("POST", "/api/admin/users/nonexistent/approve", map[string]interface{}{
			"approved": true,
		}, adminUser.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})
}

func TestE2E_Admin_UnauthorizedAccess(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	regularUser, _ := server.RegisterUser("regularforadmin", "regularforadmin@example.com", "UserPass123!")

	t.Run("unauthenticated access to admin endpoints", func(t *testing.T) {
		endpoints := []string{
			"GET", "/api/admin/users",
			"POST", "/api/admin/users",
			"GET", "/api/admin/groups",
			"POST", "/api/admin/groups",
			"GET", "/api/admin/clients",
			"POST", "/api/admin/clients",
			"GET", "/api/admin/invitations",
			"POST", "/api/admin/invitations",
			"GET", "/api/admin/proxyauth",
			"POST", "/api/admin/proxyauth",
		}

		for i := 0; i < len(endpoints); i += 2 {
			method := endpoints[i]
			path := endpoints[i+1]

			resp, err := server.DoRequest(method, path, nil, nil)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			assertResponseStatus(t, resp, http.StatusUnauthorized)
		}
	})

	t.Run("regular user cannot access admin endpoints", func(t *testing.T) {
		endpoints := []struct {
			method string
			path   string
		}{
			{"GET", "/api/admin/users"},
			{"POST", "/api/admin/users"},
			{"GET", "/api/admin/groups"},
			{"POST", "/api/admin/groups"},
			{"GET", "/api/admin/clients"},
			{"POST", "/api/admin/clients"},
		}

		for _, ep := range endpoints {
			resp, err := server.AuthenticatedRequest(ep.method, ep.path, nil, regularUser.AccessToken)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			assertResponseStatus(t, resp, http.StatusForbidden)
		}
	})
}

func TestE2E_Admin_TOTP(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	user, err := server.RegisterUser("totpadmin", "totpadmin@example.com", "UserPass123!")
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
	})

	t.Run("list passkeys", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("GET", "/api/passkey", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("delete non-existent passkey", func(t *testing.T) {
		resp, err := server.AuthenticatedRequest("DELETE", "/api/passkey/nonexistent-id", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusNotFound)
	})
}

func TestE2E_Admin_LoginFlow(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("admin login via API endpoint and access admin resources", func(t *testing.T) {
		adminUsername := "testadmin"
		adminEmail := "testadmin@example.com"
		adminPassword := "TestAdmin123!"

		_, err := server.CreateAdminUser(adminUsername, adminEmail, adminPassword)
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    adminUsername,
			"password": adminPassword,
		}, nil)
		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var loginResult struct {
			Success bool `json:"success"`
			Data    struct {
				AccessToken  string `json:"accessToken"`
				RefreshToken string `json:"refreshToken"`
			} `json:"data"`
			MFARequired bool `json:"mfa_required"`
		}
		if err := json.Unmarshal(resp.Body, &loginResult); err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		if !loginResult.Success {
			t.Error("Expected login success to be true")
		}
		if loginResult.Data.AccessToken == "" {
			t.Error("Expected access token to be set")
		}
		if loginResult.MFARequired {
			t.Error("Expected MFA not to be required for admin user")
		}

		adminResp, err := server.AuthenticatedRequest("GET", "/api/admin/users", nil, loginResult.Data.AccessToken)
		if err != nil {
			t.Fatalf("Admin API request failed: %v", err)
		}
		assertResponseStatus(t, adminResp, http.StatusOK)
	})

	t.Run("admin login with email input", func(t *testing.T) {
		adminUsername := "testadmin2"
		adminEmail := "testadmin2@example.com"
		adminPassword := "TestAdmin123!"

		_, err := server.CreateAdminUser(adminUsername, adminEmail, adminPassword)
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    adminEmail,
			"password": adminPassword,
		}, nil)
		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusOK)

		var loginResult struct {
			Success bool `json:"success"`
			Data    struct {
				AccessToken string `json:"accessToken"`
			} `json:"data"`
			MFARequired bool `json:"mfa_required"`
		}
		if err := json.Unmarshal(resp.Body, &loginResult); err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		if !loginResult.Success {
			t.Error("Expected login success to be true")
		}
		if loginResult.Data.AccessToken == "" {
			t.Error("Expected access token to be set")
		}
	})

	t.Run("admin login with wrong password", func(t *testing.T) {
		adminUsername := "testadmin3"
		adminEmail := "testadmin3@example.com"
		adminPassword := "TestAdmin123!"

		_, err := server.CreateAdminUser(adminUsername, adminEmail, adminPassword)
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    adminUsername,
			"password": "WrongPassword!",
		}, nil)
		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("regular user cannot access admin endpoints even with valid token", func(t *testing.T) {
		regularUser, err := server.RegisterUser("regularadmin", "regularadmin@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("User registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("GET", "/api/admin/users", nil, regularUser.AccessToken)
		if err != nil {
			t.Fatalf("Admin API request failed: %v", err)
		}

		assertResponseStatus(t, resp, http.StatusForbidden)
	})
}
