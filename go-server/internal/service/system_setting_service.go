package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/nbutton23/zxcvbn-go"
	"gorm.io/gorm"
)

type SystemSettingService struct {
	db *gorm.DB
}

func NewSystemSettingService(db *gorm.DB) *SystemSettingService {
	return &SystemSettingService{db: db}
}

type GeneralSettings struct {
	AppName string `json:"app_name"`
	AppURL  string `json:"app_url"`
}

type SecuritySettings struct {
	EmailVerificationRequired bool `json:"email_verification_required"`
	SignupRequiresApproval    bool `json:"signup_requires_approval"`
	MFARequired               bool `json:"mfa_required"`
	PasswordMinLength         int  `json:"password_min_length"`
	PasswordStrength          int  `json:"password_strength"`
}

type EmailSettings struct {
	Enabled    bool   `json:"enabled"`
	SMTPHost   string `json:"smtp_host"`
	SMTPPort   int    `json:"smtp_port"`
	SMTPUser   string `json:"smtp_user"`
	SMTPPass   string `json:"smtp_pass"`
	FromEmail  string `json:"from_email"`
	FromName   string `json:"from_name"`
}

type SessionSettings struct {
	AccessTokenExpiry  int `json:"access_token_expiry"`
	RefreshTokenExpiry int `json:"refresh_token_expiry"`
	MaxSessionsPerUser int `json:"max_sessions_per_user"`
}

type RateLimitSettings struct {
	Enabled       bool `json:"enabled"`
	LoginLimit    int  `json:"login_limit"`
	RegisterLimit int  `json:"register_limit"`
	APILimit      int  `json:"api_limit"`
}

func (s *SystemSettingService) get(key string, target interface{}) error {
	var setting model.SystemSetting
	if err := s.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return err
	}
	return json.Unmarshal([]byte(setting.Value), target)
}

func (s *SystemSettingService) set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal setting %s: %w", key, err)
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var existing model.SystemSetting
		if err := tx.Where("key = ?", key).First(&existing).Error; err == nil {
			return tx.Model(&existing).Update("value", string(data)).Error
		}
		return tx.Create(&model.SystemSetting{Key: key, Value: string(data)}).Error
	})
}

func (s *SystemSettingService) GetGeneral() (GeneralSettings, error) {
	var settings GeneralSettings
	err := s.get("general", &settings)
	return settings, err
}

func (s *SystemSettingService) SetGeneral(settings GeneralSettings) error {
	return s.set("general", settings)
}

func (s *SystemSettingService) GetSecurity() (SecuritySettings, error) {
	var settings SecuritySettings
	err := s.get("security", &settings)
	return settings, err
}

func (s *SystemSettingService) SetSecurity(settings SecuritySettings) error {
	return s.set("security", settings)
}

func (s *SystemSettingService) GetEmail() (EmailSettings, error) {
	var settings EmailSettings
	err := s.get("email", &settings)
	return settings, err
}

func (s *SystemSettingService) SetEmail(settings EmailSettings) error {
	return s.set("email", settings)
}

func (s *SystemSettingService) GetSession() (SessionSettings, error) {
	var settings SessionSettings
	err := s.get("session", &settings)
	return settings, err
}

func (s *SystemSettingService) SetSession(settings SessionSettings) error {
	return s.set("session", settings)
}

func (s *SystemSettingService) GetRateLimit() (RateLimitSettings, error) {
	var settings RateLimitSettings
	err := s.get("ratelimit", &settings)
	return settings, err
}

func (s *SystemSettingService) SetRateLimit(settings RateLimitSettings) error {
	return s.set("ratelimit", settings)
}

func (s *SystemSettingService) InitializeDefaults() error {
	defaults := map[string]interface{}{
		"general": GeneralSettings{
			AppName: "AuthNas",
			AppURL:  "http://localhost:8080",
		},
		"security": SecuritySettings{
			EmailVerificationRequired: false,
			SignupRequiresApproval:    false,
			MFARequired:               false,
			PasswordMinLength:         8,
			PasswordStrength:          3,
		},
		"email": EmailSettings{
			Enabled:   false,
			SMTPHost:  "",
			SMTPPort:  587,
			SMTPUser:  "",
			SMTPPass:  "",
			FromEmail: "",
			FromName:  "AuthNas",
		},
		"session": SessionSettings{
			AccessTokenExpiry:  15,
			RefreshTokenExpiry: 7,
			MaxSessionsPerUser: 5,
		},
		"ratelimit": RateLimitSettings{
			Enabled:       true,
			LoginLimit:    5,
			RegisterLimit: 3,
			APILimit:      60,
		},
	}

	for key, value := range defaults {
		var existing model.SystemSetting
		if err := s.db.Where("key = ?", key).First(&existing).Error; err == gorm.ErrRecordNotFound {
			data, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("marshal default %s: %w", key, err)
			}
			if err := s.db.Create(&model.SystemSetting{Key: key, Value: string(data)}).Error; nil != err {
				return fmt.Errorf("create default %s: %w", key, err)
			}
		}
	}
	return nil
}

func (s *SystemSettingService) GetSMTPConfig() (host string, port int, user string, pass string, from string, enabled bool, err error) {
	settings, err := s.GetEmail()
	if err != nil {
		return "", 0, "", "", "", false, err
	}
	return settings.SMTPHost, settings.SMTPPort, settings.SMTPUser, settings.SMTPPass, settings.FromEmail, settings.Enabled, nil
}

func (s *SystemSettingService) GetPasswordPolicy() (minLength int, strength int, err error) {
	settings, err := s.GetSecurity()
	if err != nil {
		return 8, 3, err
	}
	if settings.PasswordMinLength == 0 {
		settings.PasswordMinLength = 8
	}
	if settings.PasswordStrength == 0 {
		settings.PasswordStrength = 3
	}
	return settings.PasswordMinLength, settings.PasswordStrength, nil
}

func (s *SystemSettingService) ValidatePassword(password string) error {
	settings, err := s.GetSecurity()
	if err != nil {
		return nil
	}
	if settings.PasswordMinLength > 0 && len(password) < settings.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters", settings.PasswordMinLength)
	}
	if settings.PasswordStrength > 0 {
		result := zxcvbn.PasswordStrength(password, nil)
		if result.Score < settings.PasswordStrength {
			return fmt.Errorf("password is too weak (score: %d, required: %d)", result.Score, settings.PasswordStrength)
		}
	}
	return nil
}

func (s *SystemSettingService) IsEmailVerificationRequired() bool {
	settings, err := s.GetSecurity()
	if err != nil {
		return false
	}
	return settings.EmailVerificationRequired
}

func (s *SystemSettingService) IsSignupRequiresApproval() bool {
	settings, err := s.GetSecurity()
	if err != nil {
		return false
	}
	return settings.SignupRequiresApproval
}

func (s *SystemSettingService) IsEmailEnabled() bool {
	settings, err := s.GetEmail()
	if err != nil {
		return false
	}
	return settings.Enabled
}

func (s *SystemSettingService) GetAppURL() string {
	settings, err := s.GetGeneral()
	if err != nil {
		return "http://localhost:8080"
	}
	if settings.AppURL == "" {
		return "http://localhost:8080"
	}
	return strings.TrimRight(settings.AppURL, "/")
}

func (s *SystemSettingService) GetAppName() string {
	settings, err := s.GetGeneral()
	if err != nil {
		return "AuthNas"
	}
	if settings.AppName == "" {
		return "AuthNas"
	}
	return settings.AppName
}

func (s *SystemSettingService) GetAccessTokenExpiry() int {
	settings, err := s.GetSession()
	if err != nil {
		return 15
	}
	if settings.AccessTokenExpiry == 0 {
		return 15
	}
	return settings.AccessTokenExpiry
}

func (s *SystemSettingService) GetRefreshTokenExpiry() int {
	settings, err := s.GetSession()
	if err != nil {
		return 7
	}
	if settings.RefreshTokenExpiry == 0 {
		return 7
	}
	return settings.RefreshTokenExpiry
}
