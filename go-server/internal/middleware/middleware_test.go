package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/authnas/authnas/go-server/internal/config"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins:   []string{"http://localhost:3000", "http://example.com"},
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
		},
	}

	tests := []struct {
		name           string
		origin         string
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "Allowed origin",
			origin:         "http://localhost:3000",
			expectedStatus: http.StatusOK,
			expectedHeader: "http://localhost:3000",
		},
		{
			name:           "Not allowed origin",
			origin:         "http://malicious.com",
			expectedStatus: http.StatusForbidden,
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORS(cfg))
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedHeader != "" {
				actualHeader := w.Header().Get("Access-Control-Allow-Origin")
				if actualHeader != tt.expectedHeader {
					t.Errorf("Expected Access-Control-Allow-Origin '%s', got '%s'", tt.expectedHeader, actualHeader)
				}
			}
		})
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins:   []string{"http://localhost:3000"},
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type"},
		},
	}

	router := gin.New()
	router.Use(CORS(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d for preflight, got %d", http.StatusNoContent, w.Code)
	}
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name             string
		origin           string
		allowed          []string
		allowCredentials bool
		expected         bool
	}{
		{
			name:             "Exact match",
			origin:           "http://localhost:3000",
			allowed:          []string{"http://localhost:3000"},
			allowCredentials: false,
			expected:         true,
		},
		{
			name:             "Wildcard",
			origin:           "*",
			allowed:          []string{"*"},
			allowCredentials: false,
			expected:         true,
		},
		{
			name:             "Wildcard with credentials",
			origin:           "*",
			allowed:          []string{"*"},
			allowCredentials: true,
			expected:         false,
		},
		{
			name:             "No match",
			origin:           "http://notallowed.com",
			allowed:          []string{"http://allowed.com"},
			allowCredentials: false,
			expected:         false,
		},
		{
			name:             "Wildcard subdomain",
			origin:           "http://sub.example.com",
			allowed:          []string{"*.example.com"},
			allowCredentials: false,
			expected:         true,
		},
		{
			name:             "Wildcard subdomain no match",
			origin:           "http://sub.notexample.com",
			allowed:          []string{"*.example.com"},
			allowCredentials: false,
			expected:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowed, tt.allowCredentials)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 60,
		},
	}

	router := gin.New()
	router.Use(RateLimit(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}
}

func TestRateLimit_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled:           false,
			RequestsPerMinute: 1,
		},
	}

	router := gin.New()
	router.Use(RateLimit(cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}
}

func TestRateLimitByUser(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 60,
		},
	}

	user1 := "user-1"
	user2 := "user-2"

	if !RateLimitByUser(cfg, user1) {
		t.Error("First request for user1 should be allowed")
	}

	if !RateLimitByUser(cfg, user2) {
		t.Error("First request for user2 should be allowed")
	}
}
