package cookie

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RefreshTokenCookieName = "auth_refresh_token"
	SessionCookieName      = "auth_session"
)

type CookieConfig struct {
	Domain   string
	Path     string
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
	MaxAge   time.Duration
}

func DefaultCookieConfig(isSecure bool, maxAge time.Duration) CookieConfig {
	return CookieConfig{
		Domain:   "",
		Path:     "/",
		Secure:   isSecure,
		HTTPOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
}

func SetCookie(c *gin.Context, name string, value string, cfg CookieConfig) {
	c.SetSameSite(cfg.SameSite)
	c.SetCookie(
		name,
		value,
		int(cfg.MaxAge.Seconds()),
		cfg.Path,
		cfg.Domain,
		cfg.Secure,
		cfg.HTTPOnly,
	)
}

func SetRefreshTokenCookie(c *gin.Context, token string, maxAge time.Duration) {
	isSecure := strings.HasPrefix(c.Request.Host, "https") ||
		strings.HasPrefix(c.GetHeader("X-Forwarded-Proto"), "https")
	cfg := DefaultCookieConfig(isSecure, maxAge)
	SetCookie(c, RefreshTokenCookieName, token, cfg)
}

func GetRefreshTokenFromCookie(c *gin.Context) string {
	token, err := c.Cookie(RefreshTokenCookieName)
	if err != nil {
		return ""
	}
	return token
}

func ClearRefreshTokenCookie(c *gin.Context) {
	isSecure := strings.HasPrefix(c.Request.Host, "https") ||
		strings.HasPrefix(c.GetHeader("X-Forwarded-Proto"), "https")
	cfg := DefaultCookieConfig(isSecure, 0)
	cfg.MaxAge = 0
	c.SetSameSite(cfg.SameSite)
	c.SetCookie(
		RefreshTokenCookieName,
		"",
		0,
		cfg.Path,
		cfg.Domain,
		cfg.Secure,
		cfg.HTTPOnly,
	)
}

func SetSessionCookie(c *gin.Context, sessionID string, maxAge time.Duration) {
	isSecure := strings.HasPrefix(c.Request.Host, "https") ||
		strings.HasPrefix(c.GetHeader("X-Forwarded-Proto"), "https")
	cfg := CookieConfig{
		Domain:   "",
		Path:     "/",
		Secure:   isSecure,
		HTTPOnly: false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
	SetCookie(c, SessionCookieName, sessionID, cfg)
}

func GetSessionFromCookie(c *gin.Context) string {
	session, err := c.Cookie(SessionCookieName)
	if err != nil {
		return ""
	}
	return session
}

func ClearSessionCookie(c *gin.Context) {
	isSecure := strings.HasPrefix(c.Request.Host, "https") ||
		strings.HasPrefix(c.GetHeader("X-Forwarded-Proto"), "https")
	cfg := CookieConfig{
		Domain:   "",
		Path:     "/",
		Secure:   isSecure,
		HTTPOnly: false,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetSameSite(cfg.SameSite)
	c.SetCookie(
		SessionCookieName,
		"",
		0,
		cfg.Path,
		cfg.Domain,
		cfg.Secure,
		cfg.HTTPOnly,
	)
}
