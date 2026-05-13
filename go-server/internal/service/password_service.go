package service

import (
	"github.com/nbutton23/zxcvbn-go"
	"github.com/authnas/authnas/go-server/internal/config"
)

type PasswordService struct {
	cfg *config.Config
}

func NewPasswordService(cfg *config.Config) *PasswordService {
	return &PasswordService{
		cfg: cfg,
	}
}

type PasswordStrengthResult struct {
	Score            int    `json:"score"`
	Strength         string `json:"strength"`
	CrackTimeDisplay string `json:"crack_time_display"`
}

func (s *PasswordService) CheckStrength(password string) *PasswordStrengthResult {
	result := zxcvbn.PasswordStrength(password, nil)

	strength := "weak"
	switch result.Score {
	case 0:
		strength = "very_weak"
	case 1:
		strength = "weak"
	case 2:
		strength = "fair"
	case 3:
		strength = "strong"
	case 4:
		strength = "very_strong"
	}

	return &PasswordStrengthResult{
		Score:            result.Score,
		Strength:         strength,
		CrackTimeDisplay: result.CrackTimeDisplay,
	}
}

func (s *PasswordService) IsStrongEnough(password string) bool {
	result := s.CheckStrength(password)
	return result.Score >= s.cfg.Security.PasswordStrength
}
