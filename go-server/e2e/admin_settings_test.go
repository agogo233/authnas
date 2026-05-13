package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestE2E_AdminSettings(t *testing.T) {
	if os.Getenv("TEST_E2E") != "1" {
		t.Skip("Skipping e2e test")
	}

	server := NewE2ETestServer(t)
	defer server.Close()
	defer server.Cleanup()

	admin, err := server.CreateAdminUser("adminsettings", "adminsettings@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	t.Run("get and update general settings", func(t *testing.T) {
		resp, err := server.AuthenticatedCSRFRequest("GET", "/api/admin/settings/general", nil, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)
		assert.True(t, result["success"].(bool))

		payload := map[string]interface{}{
			"app_name": "TestApp",
			"app_url":  "http://test.example.com",
		}
		resp, err = server.AuthenticatedCSRFRequest("POST", "/api/admin/settings/general", payload, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = server.AuthenticatedCSRFRequest("GET", "/api/admin/settings/general", nil, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		json.Unmarshal(resp.Body, &result)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, "TestApp", data["app_name"])
		assert.Equal(t, "http://test.example.com", data["app_url"])
	})

	t.Run("get and update security settings", func(t *testing.T) {
		resp, err := server.AuthenticatedCSRFRequest("GET", "/api/admin/settings/security", nil, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		payload := map[string]interface{}{
			"email_verification_required": true,
			"signup_requires_approval":    true,
			"mfa_required":                false,
			"password_min_length":         10,
			"password_strength":           4,
		}
		resp, err = server.AuthenticatedCSRFRequest("POST", "/api/admin/settings/security", payload, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = server.AuthenticatedCSRFRequest("GET", "/api/admin/settings/security", nil, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(10), data["password_min_length"])
		assert.Equal(t, float64(4), data["password_strength"])
	})

	t.Run("get and update session settings", func(t *testing.T) {
		resp, err := server.AuthenticatedCSRFRequest("GET", "/api/admin/settings/session", nil, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		payload := map[string]interface{}{
			"access_token_expiry":   30,
			"refresh_token_expiry":  14,
			"max_sessions_per_user": 10,
		}
		resp, err = server.AuthenticatedCSRFRequest("POST", "/api/admin/settings/session", payload, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = server.AuthenticatedCSRFRequest("GET", "/api/admin/settings/session", nil, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(30), data["access_token_expiry"])
		assert.Equal(t, float64(14), data["refresh_token_expiry"])
	})

	t.Run("get and update rate limit settings", func(t *testing.T) {
		resp, err := server.AuthenticatedCSRFRequest("GET", "/api/admin/settings/ratelimit", nil, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		payload := map[string]interface{}{
			"enabled":        true,
			"login_limit":    10,
			"register_limit": 5,
			"api_limit":      100,
		}
		resp, err = server.AuthenticatedCSRFRequest("POST", "/api/admin/settings/ratelimit", payload, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = server.AuthenticatedCSRFRequest("GET", "/api/admin/settings/ratelimit", nil, admin.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		var result map[string]interface{}
		json.Unmarshal(resp.Body, &result)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(10), data["login_limit"])
		assert.Equal(t, float64(100), data["api_limit"])
	})

	t.Run("admin only access", func(t *testing.T) {
		user, err := server.RegisterUser("settingsuser", "settingsuser@example.com", "ValidPass123!")
		if !assert.NoError(t, err) {
			return
		}

		resp, err := server.AuthenticatedCSRFRequest("GET", "/api/admin/settings/general", nil, user.AccessToken)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
