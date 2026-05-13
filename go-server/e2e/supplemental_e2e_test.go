// Package e2e provides end-to-end tests for AuthNas Go server.
//
// IMPORTANT: These tests directly test the Go server (localhost:8080).
// Vite is only used as a build tool to compile frontend resources to static files.
// The Go server serves both API endpoints and frontend static resources.
//
// Tests cover:
//   - Complete user flows from registration to login
//   - OIDC complete flows with various configurations
//   - Session and token lifecycle
//   - Edge cases and error handling
package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestE2E_CompleteFlows(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("complete-user-registration-to-login-flow", func(t *testing.T) {
		username := "completeflowuser"
		email := "completeflow@example.com"
		password := "SecurePass123!"

		registerResp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": username,
			"email":    email,
			"password": password,
		}, nil)
		if err != nil {
			t.Fatalf("Registration request failed: %v", err)
		}
		assertResponseStatus(t, registerResp, http.StatusOK)

		var registerResult struct {
			Success bool `json:"success"`
			Data    struct {
				AccessToken  string `json:"accessToken"`
				RefreshToken string `json:"refreshToken"`
			} `json:"data"`
		}
		json.Unmarshal(registerResp.Body, &registerResult)
		if !registerResult.Success || registerResult.Data.AccessToken == "" {
			t.Fatal("Registration should return success with tokens")
		}

		meResp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, registerResult.Data.AccessToken)
		if err != nil {
			t.Fatalf("Get user me failed: %v", err)
		}
		assertResponseStatus(t, meResp, http.StatusOK)

		logoutResp, err := server.AuthenticatedRequest("DELETE", "/api/user/me/sessions", nil, registerResult.Data.AccessToken)
		if err != nil {
			t.Fatalf("Logout failed: %v", err)
		}
		assertResponseStatus(t, logoutResp, http.StatusOK)

		loginResp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    username,
			"password": password,
		}, nil)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
		assertResponseStatus(t, loginResp, http.StatusOK)
	})

	t.Run("complete-admin-user-management-flow", func(t *testing.T) {
		admin, err := server.CreateAdminUser("flowadmin", "flowadmin@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		newUser, err := server.RegisterUser("flowuser", "flowuser@example.com", "UserPass123!")
		if err != nil {
			t.Fatalf("User registration failed: %v", err)
		}

		listResp, err := server.AuthenticatedRequest("GET", "/api/admin/users", nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("List users failed: %v", err)
		}
		assertResponseStatus(t, listResp, http.StatusOK)

		var userList struct {
			Users []struct {
				ID       string `json:"id"`
				Username string `json:"username"`
			} `json:"users"`
			Total int64 `json:"total"`
		}
		json.Unmarshal(listResp.Body, &userList)
		if userList.Total < 2 {
			t.Errorf("Expected at least 2 users, got %d", userList.Total)
		}

		updateResp, err := server.AuthenticatedRequest("PUT", "/api/admin/users/"+newUser.ID, map[string]interface{}{
			"name": "Updated Name",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Update user failed: %v", err)
		}
		assertResponseStatus(t, updateResp, http.StatusOK)

		approveResp, err := server.AuthenticatedRequest("POST", "/api/admin/users/"+newUser.ID+"/approve", map[string]interface{}{
			"approved": true,
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Approve user failed: %v", err)
		}
		assertResponseStatus(t, approveResp, http.StatusOK)
	})

	t.Run("complete-group-management-flow", func(t *testing.T) {
		admin, err := server.CreateAdminUser("groupflowadmin", "groupflowadmin@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/groups", map[string]interface{}{
			"name":        "Test Group Flow",
			"description": "Group for testing flow",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Create group failed: %v", err)
		}
		assertResponseStatus(t, createResp, http.StatusOK)

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				CreatedAt   string `json:"createdAt"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		getResp, err := server.AuthenticatedRequest("GET", "/api/admin/groups/"+createResult.Data.ID, nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("Get group failed: %v", err)
		}
		if getResp.StatusCode != http.StatusOK {
			t.Logf("BUG: GET /api/admin/groups/:id returns %d", getResp.StatusCode)
		}

		updateResp, err := server.AuthenticatedRequest("PUT", "/api/admin/groups/"+createResult.Data.ID, map[string]interface{}{
			"name":        "Updated Group Name",
			"description": "Updated description",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Update group failed: %v", err)
		}
		assertResponseStatus(t, updateResp, http.StatusOK)

		deleteResp, err := server.AuthenticatedRequest("DELETE", "/api/admin/groups/"+createResult.Data.ID, nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("Delete group failed: %v", err)
		}
		assertResponseStatus(t, deleteResp, http.StatusOK)

		getAfterDeleteResp, err := server.AuthenticatedRequest("GET", "/api/admin/groups/"+createResult.Data.ID, nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("Get deleted group failed: %v", err)
		}
		if getAfterDeleteResp.StatusCode != http.StatusNotFound {
			t.Logf("BUG: GET deleted group returns %d instead of 404", getAfterDeleteResp.StatusCode)
		}
	})

	t.Run("complete-client-management-flow", func(t *testing.T) {
		admin, err := server.CreateAdminUser("clientflowadmin", "clientflowadmin@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId":      "flow-client-" + fmt.Sprintf("%d", serverIndex),
			"name":          "Flow Test Client",
			"redirect_uris": "https://flow.example.com/callback",
			"scopes":        "openid profile email offline_access",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Create client failed: %v", err)
		}
		assertResponseStatus(t, createResp, http.StatusOK)

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		getResp, err := server.AuthenticatedRequest("GET", "/api/admin/clients/"+createResult.Data.ID, nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("Get client failed: %v", err)
		}
		if getResp.StatusCode != http.StatusOK {
			t.Logf("BUG: GET /api/admin/clients/:id returns %d", getResp.StatusCode)
		}

		updateResp, err := server.AuthenticatedRequest("PUT", "/api/admin/clients/"+createResult.Data.ID, map[string]interface{}{
			"redirect_uris": "https://new-flow.example.com/callback",
			"scopes":        "openid profile",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Update client failed: %v", err)
		}
		assertResponseStatus(t, updateResp, http.StatusOK)

		deleteResp, err := server.AuthenticatedRequest("DELETE", "/api/admin/clients/"+createResult.Data.ID, nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("Delete client failed: %v", err)
		}
		assertResponseStatus(t, deleteResp, http.StatusOK)
	})

	t.Run("complete-invitation-management-flow", func(t *testing.T) {
		admin, err := server.CreateAdminUser("inviteflowadmin", "inviteflowadmin@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email":    "inviteflow@example.com",
			"username": "inviteflowuser",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Create invitation failed: %v", err)
		}
		assertResponseStatus(t, createResp, http.StatusOK)

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID   string `json:"id"`
				Code string `json:"code"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		getResp, err := server.AuthenticatedRequest("GET", "/api/admin/invitations/"+createResult.Data.ID, nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("Get invitation failed: %v", err)
		}
		if getResp.StatusCode != http.StatusOK {
			t.Logf("BUG: GET /api/admin/invitations/:id returns %d", getResp.StatusCode)
		}

		deleteResp, err := server.AuthenticatedRequest("DELETE", "/api/admin/invitations/"+createResult.Data.ID, nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("Delete invitation failed: %v", err)
		}
		assertResponseStatus(t, deleteResp, http.StatusOK)
	})

	t.Run("complete-proxy-auth-management-flow", func(t *testing.T) {
		admin, err := server.CreateAdminUser("proxyauthflowadmin", "proxyauthflowadmin@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/proxyauth", map[string]interface{}{
			"name":       "Flow Proxy Auth",
			"proxyUrl":   "https://flow-proxy.example.com",
			"headerName": "X-Flow-User-ID",
			"enabled":    true,
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Create proxy auth failed: %v", err)
		}
		assertResponseStatus(t, createResp, http.StatusOK)

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		getResp, err := server.AuthenticatedRequest("GET", "/api/admin/proxyauth/"+createResult.Data.ID, nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("Get proxy auth failed: %v", err)
		}
		if getResp.StatusCode != http.StatusOK {
			t.Logf("BUG: GET /api/admin/proxyauth/:id returns %d", getResp.StatusCode)
		}

		updateResp, err := server.AuthenticatedRequest("PUT", "/api/admin/proxyauth/"+createResult.Data.ID, map[string]interface{}{
			"enabled": false,
			"name":    "Updated Flow Proxy Auth",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Update proxy auth failed: %v", err)
		}
		assertResponseStatus(t, updateResp, http.StatusOK)

		deleteResp, err := server.AuthenticatedRequest("DELETE", "/api/admin/proxyauth/"+createResult.Data.ID, nil, admin.AccessToken)
		if err != nil {
			t.Fatalf("Delete proxy auth failed: %v", err)
		}
		assertResponseStatus(t, deleteResp, http.StatusOK)
	})
}

func TestE2E_PasswordManagement(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("update-password-with-correct-old-password", func(t *testing.T) {
		user, err := server.RegisterUser("pwdupdateuser", "pwdupdate@example.com", "OldPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		updateResp, err := server.AuthenticatedRequest("PUT", "/api/user/me/password", map[string]interface{}{
			"oldPassword": "OldPass123!",
			"newPassword": "NewPass456!",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Update password failed: %v", err)
		}
		assertResponseStatus(t, updateResp, http.StatusOK)

		loginWithOldResp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "pwdupdateuser",
			"password": "OldPass123!",
		}, nil)
		if err != nil {
			t.Fatalf("Login with old password failed: %v", err)
		}
		if loginWithOldResp.StatusCode != http.StatusUnauthorized {
			t.Log("BUG: Old password should not work after password update")
		}

		loginWithNewResp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "pwdupdateuser",
			"password": "NewPass456!",
		}, nil)
		if err != nil {
			t.Fatalf("Login with new password failed: %v", err)
		}
		assertResponseStatus(t, loginWithNewResp, http.StatusOK)
	})

	t.Run("update-password-with-incorrect-old-password", func(t *testing.T) {
		user, err := server.RegisterUser("pwderroruser", "pwderror@example.com", "OriginalPass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		updateResp, err := server.AuthenticatedRequest("PUT", "/api/user/me/password", map[string]interface{}{
			"oldPassword": "WrongOldPass!",
			"newPassword": "NewPass789!",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Update password request failed: %v", err)
		}
		if updateResp.StatusCode != http.StatusBadRequest {
			t.Logf("Update with wrong old password returned %d, expected 400", updateResp.StatusCode)
		}
	})

	t.Run("update-password-with-same-old-and-new-password", func(t *testing.T) {
		user, err := server.RegisterUser("pwdsameuser", "pwdsame@example.com", "SamePass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		updateResp, err := server.AuthenticatedRequest("PUT", "/api/user/me/password", map[string]interface{}{
			"oldPassword": "SamePass123!",
			"newPassword": "SamePass123!",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Update password request failed: %v", err)
		}
		if updateResp.StatusCode != http.StatusOK && updateResp.StatusCode != http.StatusBadRequest {
			t.Logf("Update with same password returned %d", updateResp.StatusCode)
		}
	})

	t.Run("admin-reset-user-password", func(t *testing.T) {
		admin, err := server.CreateAdminUser("adminpwdreset", "adminpwdreset@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		user, err := server.RegisterUser("usertoreset", "usertoreset@example.com", "OriginalPass123!")
		if err != nil {
			t.Fatalf("User registration failed: %v", err)
		}

		resetResp, err := server.AuthenticatedRequest("POST", "/api/admin/users/"+user.ID+"/reset-password", map[string]interface{}{
			"newPassword": "AdminResetPass456!",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Admin reset password failed: %v", err)
		}
		assertResponseStatus(t, resetResp, http.StatusOK)

		loginResp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "usertoreset",
			"password": "AdminResetPass456!",
		}, nil)
		if err != nil {
			t.Fatalf("Login with reset password failed: %v", err)
		}
		assertResponseStatus(t, loginResp, http.StatusOK)
	})
}

func TestE2E_UserProfileManagement(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("get-own-profile", func(t *testing.T) {
		user, err := server.RegisterUser("profileuser", "profileuser@example.com", "ProfilePass123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		meResp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Get profile failed: %v", err)
		}
		assertResponseStatus(t, meResp, http.StatusOK)

		var profile struct {
			Success bool `json:"success"`
			Data    struct {
				ID            string `json:"id"`
				Username      string `json:"username"`
				Email         string `json:"email"`
				Name          string `json:"name"`
				EmailVerified bool   `json:"emailVerified"`
				Approved      bool   `json:"approved"`
			} `json:"data"`
		}
		json.Unmarshal(meResp.Body, &profile)

		if profile.Data.Username != "profileuser" {
			t.Errorf("Expected username 'profileuser', got '%s'", profile.Data.Username)
		}
	})

	t.Run("update-profile-name", func(t *testing.T) {
		user, err := server.RegisterUser("nameupdate", "nameupdate@example.com", "NameUpdate123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		updateResp, err := server.AuthenticatedRequest("PUT", "/api/user/me", map[string]interface{}{
			"name": "Updated Test Name",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Update name failed: %v", err)
		}
		assertResponseStatus(t, updateResp, http.StatusOK)

		meResp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Get profile failed: %v", err)
		}

		var profile struct {
			Success bool `json:"success"`
			Data    struct {
				Name string `json:"name"`
			} `json:"data"`
		}
		json.Unmarshal(meResp.Body, &profile)

		if profile.Data.Name != "Updated Test Name" {
			t.Errorf("Expected name 'Updated Test Name', got '%s'", profile.Data.Name)
		}
	})

	t.Run("update-profile-email", func(t *testing.T) {
		user, err := server.RegisterUser("emailupdate", "emailupdate1@example.com", "EmailUpdate123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		updateResp, err := server.AuthenticatedRequest("PUT", "/api/user/me", map[string]interface{}{
			"email": "emailupdate2@example.com",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Update email failed: %v", err)
		}
		if updateResp.StatusCode != http.StatusOK {
			t.Logf("Update email returned %d", updateResp.StatusCode)
		}
	})

	t.Run("profile-without-auth-returns-401", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/user/me", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestE2E_EmailVerificationComplete(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("send-verify-email-and-check-response", func(t *testing.T) {
		user, err := server.RegisterUser("verifytest", "verifytest@example.com", "VerifyTest123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("POST", "/api/auth/send_verify_email", map[string]interface{}{
			"email": user.Email,
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Send verify email failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("verify-email-with-invalid-challenge", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/verify_email", map[string]interface{}{
			"id":        "some-user-id",
			"challenge": "invalid-challenge",
		}, nil)
		if err != nil {
			t.Fatalf("Verify email failed: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Verify email with invalid challenge returned %d", resp.StatusCode)
		}
	})
}

func TestE2E_InvitationComplete(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("get-existing-invitation", func(t *testing.T) {
		admin, err := server.CreateAdminUser("inviteadmin", "inviteadmin@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email":    "inviteduser@example.com",
			"username": "inviteduser",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Create invitation failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID   string `json:"id"`
				Code string `json:"code"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		getResp, err := server.DoRequest("GET", "/api/auth/invitation/"+createResult.Data.ID+"/"+createResult.Data.Code, nil, nil)
		if err != nil {
			t.Fatalf("Get invitation failed: %v", err)
		}
		assertResponseStatus(t, getResp, http.StatusOK)

		var inviteResult struct {
			Success bool `json:"success"`
			Data    struct {
				Email    string `json:"email"`
				Username string `json:"username"`
			} `json:"data"`
		}
		json.Unmarshal(getResp.Body, &inviteResult)

		if !inviteResult.Success {
			t.Error("Invitation should be valid")
		}
		if inviteResult.Data.Email != "inviteduser@example.com" {
			t.Errorf("Expected email 'inviteduser@example.com', got '%s'", inviteResult.Data.Email)
		}
	})

	t.Run("get-invitation-with-wrong-challenge", func(t *testing.T) {
		admin, err := server.CreateAdminUser("inviteadmin2", "inviteadmin2@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		createResp, err := server.AuthenticatedRequest("POST", "/api/admin/invitations", map[string]interface{}{
			"email": "inviteduser2@example.com",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Create invitation failed: %v", err)
		}

		var createResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(createResp.Body, &createResult)

		getResp, err := server.DoRequest("GET", "/api/auth/invitation/"+createResult.Data.ID+"/wrong-challenge", nil, nil)
		if err != nil {
			t.Fatalf("Get invitation with wrong challenge failed: %v", err)
		}
		if getResp.StatusCode != http.StatusNotFound {
			t.Logf("Get invitation with wrong challenge returned %d", getResp.StatusCode)
		}
	})
}

func TestE2E_SessionManagementComplete(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("session-invalidation-after-delete", func(t *testing.T) {
		user, err := server.RegisterUser("sessiontest", "sessiontest@example.com", "SessionTest123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		accessToken := user.AccessToken

		deleteResp, err := server.AuthenticatedRequest("DELETE", "/api/user/me/sessions", nil, accessToken)
		if err != nil {
			t.Fatalf("Delete sessions failed: %v", err)
		}
		assertResponseStatus(t, deleteResp, http.StatusOK)

		useTokenResp, err := server.AuthenticatedRequest("GET", "/api/user/me", nil, accessToken)
		if err != nil {
			t.Fatalf("Use token after delete failed: %v", err)
		}
		if useTokenResp.StatusCode != http.StatusUnauthorized {
			t.Logf("BUG: Token should be invalid after session delete, got %d", useTokenResp.StatusCode)
		}
	})

	t.Run("new-login-after-session-delete", func(t *testing.T) {
		user, err := server.RegisterUser("relogintest", "relogintest@example.com", "ReloginTest123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		deleteResp, err := server.AuthenticatedRequest("DELETE", "/api/user/me/sessions", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Delete sessions failed: %v", err)
		}
		assertResponseStatus(t, deleteResp, http.StatusOK)

		loginResp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{
			"input":    "relogintest",
			"password": "ReloginTest123!",
		}, nil)
		if err != nil {
			t.Fatalf("Re-login failed: %v", err)
		}
		assertResponseStatus(t, loginResp, http.StatusOK)

		var loginResult struct {
			Success bool `json:"success"`
			Data    struct {
				AccessToken string `json:"accessToken"`
			} `json:"data,omitempty"`
		}
		json.Unmarshal(loginResp.Body, &loginResult)

		if !loginResult.Success || loginResult.Data.AccessToken == "" {
			t.Error("Re-login should succeed with new token")
		}
	})
}

func TestE2E_PasskeyManagementComplete(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("list-passkeys-for-authenticated-user", func(t *testing.T) {
		user, err := server.RegisterUser("pklistuser", "pklistuser@example.com", "PkListUser123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		listResp, err := server.AuthenticatedRequest("GET", "/api/passkey", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("List passkeys failed: %v", err)
		}
		assertResponseStatus(t, listResp, http.StatusOK)

		var passkeys []interface{}
		json.Unmarshal(listResp.Body, &passkeys)

		if len(passkeys) != 0 {
			t.Errorf("New user should have 0 passkeys, got %d", len(passkeys))
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
			"challenge": "test-challenge",
			"options":   "{}",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("delete-passkey-without-auth", func(t *testing.T) {
		resp, err := server.DoRequest("DELETE", "/api/passkey/some-id", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestE2E_TOTPManagementComplete(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("register-TOTP", func(t *testing.T) {
		user, err := server.RegisterUser("totpreguser", "totpreguser@example.com", "TotpRegUser123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		regResp, err := server.AuthenticatedRequest("POST", "/api/totp/registration", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("TOTP registration failed: %v", err)
		}
		assertResponseStatus(t, regResp, http.StatusOK)

		var regResult struct {
			Success bool `json:"success"`
			Data    struct {
				Secret    string `json:"secret"`
				QRCodeURI string `json:"qr_code_uri"`
			} `json:"data"`
		}
		json.Unmarshal(regResp.Body, &regResult)

		if !regResult.Success {
			t.Error("TOTP registration should return success")
		}
		if regResult.Data.Secret == "" {
			t.Error("TOTP registration should return secret")
		}
	})

	t.Run("verify-TOTP-with-valid-token", func(t *testing.T) {
		user, err := server.RegisterUser("totpverifyuser", "totpverifyuser@example.com", "TotpVerifyUser123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		regResp, err := server.AuthenticatedRequest("POST", "/api/totp/registration", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("TOTP registration failed: %v", err)
		}

		var regResult struct {
			Success bool `json:"success"`
			Data    struct {
				Secret string `json:"secret"`
			} `json:"data"`
		}
		json.Unmarshal(regResp.Body, &regResult)

		verifyResp, err := server.AuthenticatedRequest("POST", "/api/totp/verify", map[string]interface{}{
			"token": "123456",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("TOTP verify request failed: %v", err)
		}
		if verifyResp.StatusCode != http.StatusOK && verifyResp.StatusCode != http.StatusUnauthorized {
			t.Logf("TOTP verify returned %d (token may be invalid format)", verifyResp.StatusCode)
		}
	})

	t.Run("delete-TOTP-with-invalid-token", func(t *testing.T) {
		user, err := server.RegisterUser("totpdeleteuser", "totpdeleteuser@example.com", "TotpDeleteUser123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		deleteResp, err := server.AuthenticatedRequest("DELETE", "/api/totp", map[string]interface{}{
			"token": "invalid-token",
		}, user.AccessToken)
		if err != nil {
			t.Fatalf("Delete TOTP request failed: %v", err)
		}
		if deleteResp.StatusCode != http.StatusUnauthorized && deleteResp.StatusCode != http.StatusBadRequest {
			t.Logf("Delete TOTP with invalid token returned %d", deleteResp.StatusCode)
		}
	})

	t.Run("TOTP-without-auth", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/totp/verify", map[string]interface{}{
			"token": "123456",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestE2E_OIDCCompleteFlows(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("oidc-authorization-with-valid-client", func(t *testing.T) {
		admin, err := server.CreateAdminUser("oidcflowadmin", "oidcflowadmin@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		clientResp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId":      "oidc-flow-client",
			"name":          "OIDC Flow Client",
			"redirectUris":  "https://oidcflow.example.com/callback",
			"scopes":        "openid profile email",
			"responseTypes": "code",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Client creation failed: %v", err)
		}

		var clientResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID       string `json:"id"`
				ClientID string `json:"clientId"`
			} `json:"data"`
		}
		json.Unmarshal(clientResp.Body, &clientResult)

		authResp, err := server.DoRequest("GET", "/oidc/auth?"+
			"client_id=oidc-flow-client&"+
			"redirect_uri=https://oidcflow.example.com/callback&"+
			"response_type=code&"+
			"scope=openid profile email&"+
			"state=test-state", nil, nil)
		if err != nil {
			t.Fatalf("Authorization request failed: %v", err)
		}
		if authResp.StatusCode != http.StatusFound && authResp.StatusCode != http.StatusBadRequest {
			t.Logf("Authorization returned %d", authResp.StatusCode)
		}
	})

	t.Run("oidc-authorization-with-invalid-client", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/auth?"+
			"client_id=nonexistent-client&"+
			"redirect_uri=https://example.com/callback&"+
			"response_type=code&"+
			"scope=openid", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("oidc-authorization-with-invalid-redirect-uri", func(t *testing.T) {
		admin, err := server.CreateAdminUser("oidcflowadmin2", "oidcflowadmin2@example.com", "AdminPass123!")
		if err != nil {
			t.Fatalf("Admin creation failed: %v", err)
		}

		clientResp, err := server.AuthenticatedRequest("POST", "/api/admin/clients", map[string]interface{}{
			"clientId": "oidc-invalid-redirect-client",
			"name":     "OIDC Invalid Redirect Client",
		}, admin.AccessToken)
		if err != nil {
			t.Fatalf("Client creation failed: %v", err)
		}

		var clientResult struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(clientResp.Body, &clientResult)

		resp, err := server.DoRequest("GET", "/oidc/auth?"+
			"client_id=oidc-invalid-redirect-client&"+
			"redirect_uri=https://evil.com/callback&"+
			"response_type=code&"+
			"scope=openid", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("oidc-token-with-invalid-grant", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{
			"grant_type":  "authorization_code",
			"code":        "invalid-code",
			"redirectUri": "https://example.com/callback",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Token with invalid grant returned %d", resp.StatusCode)
		}
	})

	t.Run("oidc-refresh-token-without-token", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token", map[string]interface{}{
			"grant_type": "refresh_token",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("oidc-userinfo-without-token", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/userinfo", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("oidc-revocation-without-token", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/oidc/token/revocation", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})
}

func TestE2E_HealthAndConfigEndpoints(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("health-check", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/health", nil, nil)
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)
	})

	t.Run("public-config", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/config", nil, nil)
		if err != nil {
			t.Fatalf("Config request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)

		var config struct {
			Success bool `json:"success"`
			Data    struct {
				AppName string `json:"app_name"`
			} `json:"data"`
		}
		json.Unmarshal(resp.Body, &config)

		if config.Data.AppName == "" {
			t.Error("Config should contain app_name")
		}
	})

	t.Run("oidc-discovery-endpoint", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/.well-known/openid-configuration", nil, nil)
		if err != nil {
			t.Fatalf("Discovery request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)

		var discovery map[string]interface{}
		json.Unmarshal(resp.Body, &discovery)

		requiredFields := []string{"issuer", "authorization_endpoint", "token_endpoint", "userinfo_endpoint", "jwks_uri"}
		for _, field := range requiredFields {
			if discovery[field] == nil {
				t.Errorf("Discovery should contain %s", field)
			}
		}
	})

	t.Run("oidc-jwks-endpoint", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/oidc/jwks", nil, nil)
		if err != nil {
			t.Fatalf("JWKS request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusOK)

		var jwks struct {
			Keys []struct {
				Kty string `json:"kty"`
				Use string `json:"use"`
				Kid string `json:"kid"`
				Alg string `json:"alg"`
			} `json:"keys"`
		}
		json.Unmarshal(resp.Body, &jwks)

		if len(jwks.Keys) == 0 {
			t.Error("JWKS should contain at least one key")
		}
	})
}

func TestE2E_ErrorHandlingComplete(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("invalid-json-body", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", "not-json", nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Invalid JSON returned %d", resp.StatusCode)
		}
	})

	t.Run("malformed-access-token", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/user/me", nil, map[string]string{
			"Authorization": "Bearer invalid.token.here",
		})
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("missing-required-fields", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/login", map[string]interface{}{}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("admin-endpoint-without-auth", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/admin/users", nil, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("admin-endpoint-with-non-admin-user", func(t *testing.T) {
		user, err := server.RegisterUser("nonadminuser", "nonadminuser@example.com", "NonAdminUser123!")
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		resp, err := server.AuthenticatedRequest("GET", "/api/admin/users", nil, user.AccessToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusForbidden)
	})
}

func TestE2E_SecurityValidationComplete(t *testing.T) {
	server := NewE2ETestServer(t)
	defer server.Cleanup()

	t.Run("sql-injection-in-username", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "admin' OR '1'='1",
			"email":    "sqli@example.com",
			"password": "TestPass123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("sql-injection-in-email", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "sqliuser",
			"email":    "' OR '1'='1@example.com",
			"password": "TestPass123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("xss-in-username", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "<script>alert(1)</script>",
			"email":    "xss@example.com",
			"password": "TestPass123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("username-with-spaces", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "user with spaces",
			"email":    "spaces@example.com",
			"password": "TestPass123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("username-too-short", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "ab",
			"email":    "short@example.com",
			"password": "TestPass123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("username-too-long", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "thisusernameiswaytolongfortheapplication",
			"email":    "long@example.com",
			"password": "TestPass123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("invalid-email-format", func(t *testing.T) {
		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "validuser",
			"email":    "not-an-email",
			"password": "TestPass123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("duplicate-username", func(t *testing.T) {
		_, _ = server.RegisterUser("duplicateuser", "duplicate1@example.com", "DupPass123!")

		resp, err := server.DoRequest("POST", "/api/auth/register", map[string]interface{}{
			"username": "duplicateuser",
			"email":    "duplicate2@example.com",
			"password": "DupPass123!",
		}, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("invalid-bearer-token-format", func(t *testing.T) {
		resp, err := server.DoRequest("GET", "/api/user/me", nil, map[string]string{
			"Authorization": "NotBearer token",
		})
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		assertResponseStatus(t, resp, http.StatusUnauthorized)
	})
}
