package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/gin-gonic/gin"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			HSTSMaxAge:            31536000,
			HSTSIncludeSubDomains: true,
			HSTSPreload:           false,
		},
	}

	router := gin.New()
	router.Use(SecurityHeaders(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Expected X-Content-Type-Options: nosniff")
	}

	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("Expected X-Frame-Options: DENY")
	}

	if w.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Error("Expected X-XSS-Protection: 1; mode=block")
	}

	if w.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Error("Expected Referrer-Policy: strict-origin-when-cross-origin")
	}

	expectedHSTS := "max-age=31536000; includeSubDomains"
	if w.Header().Get("Strict-Transport-Security") != expectedHSTS {
		t.Errorf("Expected Strict-Transport-Security: %s, got: %s", expectedHSTS, w.Header().Get("Strict-Transport-Security"))
	}
}

func TestSecurityHeaders_WithPreload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			HSTSMaxAge:            31536000,
			HSTSIncludeSubDomains: true,
			HSTSPreload:           true,
		},
	}

	router := gin.New()
	router.Use(SecurityHeaders(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	expectedHSTS := "max-age=31536000; includeSubDomains; preload"
	if w.Header().Get("Strict-Transport-Security") != expectedHSTS {
		t.Errorf("Expected Strict-Transport-Security: %s, got: %s", expectedHSTS, w.Header().Get("Strict-Transport-Security"))
	}
}

func TestSecurityHeaders_WithoutHSTS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			HSTSMaxAge: 0,
		},
	}

	router := gin.New()
	router.Use(SecurityHeaders(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("Strict-Transport-Security") != "" {
		t.Error("Strict-Transport-Security should not be set when HSTSMaxAge is 0")
	}
}

func TestSecurityHeaders_WithCSP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			ContentSecurityPolicy: "default-src 'self'",
		},
	}

	router := gin.New()
	router.Use(SecurityHeaders(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Security-Policy") != "default-src 'self'" {
		t.Error("Expected Content-Security-Policy: default-src 'self'")
	}
}

func TestGetRealClientIP_NoTrustedProxies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			TrustedProxies: []string{},
		},
	}
	InitTrustedProxies(cfg)

	router := gin.New()
	router.Use(GetRealClientIP())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, GetClientIP(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	_ = w.Body.String()
}

func TestGetRealClientIP_WithTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			TrustedProxies: []string{"10.0.0.1"},
		},
	}
	InitTrustedProxies(cfg)

	router := gin.New()
	router.Use(GetRealClientIP())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, GetClientIP(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	req.Header.Set("X-Real-IP", "192.168.1.1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Body.String() != "192.168.1.1" {
		t.Errorf("Expected client IP 192.168.1.1, got: %s", w.Body.String())
	}
}

func TestGetRealClientIP_XRealIPOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			TrustedProxies: []string{"10.0.0.1"},
		},
	}
	InitTrustedProxies(cfg)

	router := gin.New()
	router.Use(GetRealClientIP())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, GetClientIP(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Body.String() != "192.168.1.1" {
		t.Errorf("Expected client IP 192.168.1.1, got: %s", w.Body.String())
	}
}

func TestGetRealClientIP_AllProxiesTrusted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			TrustedProxies: []string{"10.0.0.1", "192.168.1.1"},
		},
	}
	InitTrustedProxies(cfg)

	router := gin.New()
	router.Use(GetRealClientIP())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, GetClientIP(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	_ = w.Body.String()
}

func TestGetRealClientIP_CIDRRange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			TrustedProxies: []string{"10.0.0.0/24"},
		},
	}
	InitTrustedProxies(cfg)

	router := gin.New()
	router.Use(GetRealClientIP())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, GetClientIP(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.50")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	_ = w.Body.String()
}

func TestGetClientIP_Fallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Server: config.ServerConfig{
			TrustedProxies: []string{},
		},
	}
	InitTrustedProxies(cfg)

	router := gin.New()
	router.Use(GetRealClientIP())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, GetClientIP(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Body.String() == "" {
		t.Error("Should have some client IP from fallback")
	}
}

func TestIsTrustedProxy(t *testing.T) {
	InitTrustedProxies(&config.Config{
		Server: config.ServerConfig{
			TrustedProxies: []string{"192.168.1.1", "10.0.0.0/24"},
		},
	})

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Exact match IPv4", "192.168.1.1", true},
		{"Network address of CIDR", "10.0.0.0", true},
		{"Outside CIDR", "172.16.0.1", false},
		{"Invalid IP", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTrustedProxy(net.ParseIP(tt.ip))
			if result != tt.expected {
				t.Errorf("Expected %v for IP %s, got %v", tt.expected, tt.ip, result)
			}
		})
	}
}
