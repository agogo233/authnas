package config

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Security  SecurityConfig  `mapstructure:"security"`
	Email     EmailConfig     `mapstructure:"email"`
	OIDC      OIDCConfig      `mapstructure:"oidc"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Server    ServerConfig    `mapstructure:"server"`
}

type AppConfig struct {
	URL         string `mapstructure:"url"`
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type SecurityConfig struct {
	StorageKey             string `mapstructure:"storage_key"`
	PasswordStrength       int    `mapstructure:"password_strength"`
	PasswordMinLength      int    `mapstructure:"password_min_length"`
	MFARequired            bool   `mapstructure:"mfa_required"`
	SignupRequiresApproval bool   `mapstructure:"signup_requires_approval"`
	EmailVerification      bool   `mapstructure:"email_verification"`
	InitialAdminUsername   string `mapstructure:"initial_admin_username"`
	InitialAdminEmail      string `mapstructure:"initial_admin_email"`
	InitialAdminPassword   string `mapstructure:"initial_admin_password"`
}

type EmailConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromAddress  string `mapstructure:"from_address"`
	FromName     string `mapstructure:"from_name"`
}

type OIDCConfig struct {
	Issuer      string `mapstructure:"issuer"`
	PrivateKey  string `mapstructure:"private_key"`
	Certificate string `mapstructure:"certificate"`
}

type JWTConfig struct {
	AccessTokenExpiry  string `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry string `mapstructure:"refresh_token_expiry"`
	PrivateKey         string `mapstructure:"private_key"`
	PublicKey          string `mapstructure:"public_key"`
}

func (c *JWTConfig) GetAccessTokenExpiry() time.Duration {
	d, err := time.ParseDuration(c.AccessTokenExpiry)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

func (c *JWTConfig) GetRefreshTokenExpiry() time.Duration {
	d, err := time.ParseDuration(c.RefreshTokenExpiry)
	if err != nil {
		return 168 * time.Hour
	}
	return d
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
}

type ServerConfig struct {
	HSTSMaxAge            int      `mapstructure:"hsts_max_age"`
	HSTSIncludeSubDomains bool     `mapstructure:"hsts_include_sub_domains"`
	HSTSPreload           bool     `mapstructure:"hsts_preload"`
	ContentSecurityPolicy string   `mapstructure:"content_security_policy"`
	TrustedProxies        []string `mapstructure:"trusted_proxies"`
}

var (
	viperMu sync.Mutex
)

func resetViper() {
	viperMu.Lock()
	defer viperMu.Unlock()
	viper.Reset()
}

func Load() (*Config, error) {
	resetViper()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	viper.SetDefault("app.url", "http://localhost:8080")
	viper.SetDefault("app.name", "AuthNas")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("database.path", "./data/authnas.db")
	viper.SetDefault("security.password_strength", 3)
	viper.SetDefault("security.password_min_length", 8)
	viper.SetDefault("security.mfa_required", false)
	viper.SetDefault("security.signup_requires_approval", false)
	viper.SetDefault("security.email_verification", false)
	viper.SetDefault("email.enabled", false)
	viper.SetDefault("oidc.issuer", "http://localhost:8080")
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_minute", 60)
	viper.SetDefault("jwt.access_token_expiry", "15m")
	viper.SetDefault("jwt.refresh_token_expiry", "168h")
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:8080"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Accept", "Origin", "Cache-Control", "X-Requested-With"})
	viper.SetDefault("server.hsts_max_age", 31536000)
	viper.SetDefault("server.hsts_include_sub_domains", true)
	viper.SetDefault("server.hsts_preload", false)
	viper.SetDefault("server.content_security_policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		} else {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "mapstructure"
	}); err != nil {
		return nil, err
	}

	if cfg.Security.StorageKey == "" {
		return nil, errors.New("security.storage_key is required and cannot be empty")
	}

	if len(cfg.Security.StorageKey) < 32 {
		return nil, errors.New("security.storage_key must be at least 32 characters for security")
	}

	if !isRandomEnough(cfg.Security.StorageKey) {
		log.Println("[WARNING] security.storage_key appears to be a weak key. For production, use a cryptographically random key with at least 20 unique characters.")
	}

	if envPassword := viper.GetString("security.initial_admin_password"); envPassword != "" {
		cfg.Security.InitialAdminPassword = envPassword
	} else if cfg.Security.InitialAdminPassword != "" {
		log.Println("[WARNING] Initial admin password is set in config file. For security, use INITIAL_ADMIN_PASSWORD environment variable instead.")
	}

	if cfg.OIDC.Issuer == "" || cfg.OIDC.Issuer == "http://localhost:8080" {
		cfg.OIDC.Issuer = cfg.App.URL
	}

	if cfg.App.Environment == "production" && strings.HasPrefix(cfg.OIDC.Issuer, "http://") {
		log.Println("[WARNING] OIDC Issuer is using HTTP in production. For security, use HTTPS. Setting issuer to app URL.")
		cfg.OIDC.Issuer = cfg.App.URL
	}

	return &cfg, nil
}

func isRandomEnough(s string) bool {
	if len(s) < 32 {
		return false
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	var uniqueChars int
	seen := make(map[rune]bool)

	for _, c := range s {
		seen[c] = true
		switch {
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsDigit(c):
			hasDigit = true
		case !unicode.IsLetter(c) && !unicode.IsDigit(c):
			hasSpecial = true
		}
	}

	uniqueChars = len(seen)

	categories := 0
	if hasLower {
		categories++
	}
	if hasUpper {
		categories++
	}
	if hasDigit {
		categories++
	}
	if hasSpecial {
		categories++
	}

	if uniqueChars < 10 || categories < 2 {
		return false
	}

	return true
}
