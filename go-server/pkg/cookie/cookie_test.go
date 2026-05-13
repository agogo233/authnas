package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupTestContext(method, url, host string, headers map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, url, nil)
	if host != "" {
		c.Request.Host = host
	}
	for k, v := range headers {
		c.Request.Header.Set(k, v)
	}
	return c, w
}

func TestDefaultCookieConfig(t *testing.T) {
	t.Run("secure mode", func(t *testing.T) {
		cfg := DefaultCookieConfig(true, 24*time.Hour)
		if !cfg.Secure {
			t.Error("Expected Secure to be true")
		}
		if !cfg.HTTPOnly {
			t.Error("Expected HTTPOnly to be true")
		}
		if cfg.Path != "/" {
			t.Errorf("Expected Path '/', got %s", cfg.Path)
		}
		if cfg.SameSite != http.SameSiteLaxMode {
			t.Errorf("Expected SameSiteLaxMode, got %v", cfg.SameSite)
		}
	})

	t.Run("non-secure mode", func(t *testing.T) {
		cfg := DefaultCookieConfig(false, 1*time.Hour)
		if cfg.Secure {
			t.Error("Expected Secure to be false")
		}
		if cfg.MaxAge != 1*time.Hour {
			t.Errorf("Expected MaxAge 1h, got %v", cfg.MaxAge)
		}
	})
}

func TestSetCookie(t *testing.T) {
	t.Run("basic cookie", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "localhost", nil)
		cfg := CookieConfig{
			Domain:   "",
			Path:     "/",
			Secure:   false,
			HTTPOnly: true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   3600 * time.Second,
		}
		SetCookie(c, "test_cookie", "test_value", cfg)

		cookie := w.Header().Get("Set-Cookie")
		if cookie == "" {
			t.Fatal("Expected Set-Cookie header")
		}
		if !containsCookieAttr(cookie, "test_cookie=test_value") {
			t.Errorf("Cookie should contain test_cookie=test_value, got: %s", cookie)
		}
		if !containsCookieAttr(cookie, "HttpOnly") {
			t.Error("Cookie should be HttpOnly")
		}
		if !containsCookieAttr(cookie, "Path=/") {
			t.Error("Cookie should have Path=/")
		}
		if !containsCookieAttr(cookie, "SameSite=Strict") {
			t.Errorf("Cookie should have SameSite=Strict, got: %s", cookie)
		}
	})

	t.Run("secure cookie", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "localhost", nil)
		cfg := CookieConfig{
			Path:     "/",
			Secure:   true,
			HTTPOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   7200 * time.Second,
		}
		SetCookie(c, "secure_cookie", "secure_value", cfg)

		cookie := w.Header().Get("Set-Cookie")
		if !containsCookieAttr(cookie, "Secure") {
			t.Error("Cookie should be Secure")
		}
	})
}

func TestSetRefreshTokenCookie(t *testing.T) {
	t.Run("http host sets non-secure", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "localhost:8080", nil)
		SetRefreshTokenCookie(c, "refresh_token_value", 24*time.Hour)

		cookie := w.Header().Get("Set-Cookie")
		if !containsCookieAttr(cookie, "auth_refresh_token=refresh_token_value") {
			t.Errorf("Wrong cookie value: %s", cookie)
		}
		if containsCookieAttr(cookie, "Secure") {
			t.Error("HTTP host should not set Secure flag")
		}
		if !containsCookieAttr(cookie, "HttpOnly") {
			t.Error("Refresh token should be HttpOnly")
		}
	})

	t.Run("https host sets secure", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "example.com", map[string]string{
			"X-Forwarded-Proto": "https",
		})
		SetRefreshTokenCookie(c, "refresh_token_value", 24*time.Hour)

		cookie := w.Header().Get("Set-Cookie")
		if !containsCookieAttr(cookie, "Secure") {
			t.Error("HTTPS host should set Secure flag")
		}
		if !containsCookieAttr(cookie, "HttpOnly") {
			t.Error("Refresh token should be HttpOnly")
		}
	})

	t.Run("https prefix in host sets secure", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "https://example.com", nil)
		SetRefreshTokenCookie(c, "refresh_token_value", 24*time.Hour)

		cookie := w.Header().Get("Set-Cookie")
		if !containsCookieAttr(cookie, "Secure") {
			t.Error("https:// prefix should set Secure flag")
		}
	})
}

func TestGetRefreshTokenFromCookie(t *testing.T) {
	t.Run("existing token", func(t *testing.T) {
		c, _ := setupTestContext("GET", "/test", "localhost", nil)
		c.Request.Header.Set("Cookie", "auth_refresh_token=my_refresh_token")

		token := GetRefreshTokenFromCookie(c)
		if token != "my_refresh_token" {
			t.Errorf("Expected 'my_refresh_token', got '%s'", token)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		c, _ := setupTestContext("GET", "/test", "localhost", nil)

		token := GetRefreshTokenFromCookie(c)
		if token != "" {
			t.Errorf("Expected empty token, got '%s'", token)
		}
	})
}

func TestClearRefreshTokenCookie(t *testing.T) {
	t.Run("clears token on http", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "localhost", nil)
		ClearRefreshTokenCookie(c)

		cookie := w.Header().Get("Set-Cookie")
		if !containsCookieAttr(cookie, "auth_refresh_token=") {
			t.Errorf("Should clear cookie: %s", cookie)
		}
		if !containsCookieAttr(cookie, "HttpOnly") {
			t.Error("Clear should set HttpOnly flag")
		}
	})

	t.Run("clears token on https", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "example.com", map[string]string{
			"X-Forwarded-Proto": "https",
		})
		ClearRefreshTokenCookie(c)

		cookie := w.Header().Get("Set-Cookie")
		if !containsCookieAttr(cookie, "Secure") {
			t.Error("Clear on HTTPS should set Secure flag")
		}
		if !containsCookieAttr(cookie, "HttpOnly") {
			t.Error("Clear should set HttpOnly flag")
		}
	})
}

func TestSetSessionCookie(t *testing.T) {
	t.Run("session cookie is not httponly", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "localhost", nil)
		SetSessionCookie(c, "session_abc123", 24*time.Hour)

		cookie := w.Header().Get("Set-Cookie")
		if !containsCookieAttr(cookie, "auth_session=session_abc123") {
			t.Errorf("Wrong cookie value: %s", cookie)
		}
		if containsCookieAttr(cookie, "HttpOnly") {
			t.Error("Session cookie should NOT be HttpOnly")
		}
	})

	t.Run("https sets secure flag", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "example.com", map[string]string{
			"X-Forwarded-Proto": "https",
		})
		SetSessionCookie(c, "session_abc123", 24*time.Hour)

		cookie := w.Header().Get("Set-Cookie")
		if !containsCookieAttr(cookie, "Secure") {
			t.Error("HTTPS should set Secure flag")
		}
	})
}

func TestGetSessionFromCookie(t *testing.T) {
	t.Run("existing session", func(t *testing.T) {
		c, _ := setupTestContext("GET", "/test", "localhost", nil)
		c.Request.Header.Set("Cookie", "auth_session=session_xyz")

		session := GetSessionFromCookie(c)
		if session != "session_xyz" {
			t.Errorf("Expected 'session_xyz', got '%s'", session)
		}
	})

	t.Run("missing session", func(t *testing.T) {
		c, _ := setupTestContext("GET", "/test", "localhost", nil)

		session := GetSessionFromCookie(c)
		if session != "" {
			t.Errorf("Expected empty session, got '%s'", session)
		}
	})
}

func TestClearSessionCookie(t *testing.T) {
	t.Run("clears session cookie", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "localhost", nil)
		ClearSessionCookie(c)

		cookie := w.Header().Get("Set-Cookie")
		if !containsCookieAttr(cookie, "auth_session=") {
			t.Errorf("Should clear cookie: %s", cookie)
		}
	})

	t.Run("https sets secure on clear", func(t *testing.T) {
		c, w := setupTestContext("GET", "/test", "example.com", map[string]string{
			"X-Forwarded-Proto": "https",
		})
		ClearSessionCookie(c)

		cookie := w.Header().Get("Set-Cookie")
		if !containsCookieAttr(cookie, "Secure") {
			t.Error("Clear on HTTPS should set Secure flag")
		}
		if containsCookieAttr(cookie, "HttpOnly") {
			t.Error("Session cookie should NOT be HttpOnly")
		}
	})
}

func containsCookieAttr(cookie, attr string) bool {
	return len(cookie) > 0 && (cookie == attr || len(cookie) > len(attr) && (cookie[:len(attr)+1] == attr+";" || containsSubstring(cookie, attr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
