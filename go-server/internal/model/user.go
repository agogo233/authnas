package model

import (
	"time"
)

type User struct {
	ID               string     `gorm:"primaryKey;type:text" json:"id"`
	Email            *string    `gorm:"uniqueIndex;type:text" json:"email"`
	Username         string     `gorm:"uniqueIndex;type:text" json:"username"`
	Name             *string    `gorm:"type:text" json:"name"`
	PasswordHash     *string    `gorm:"type:text" json:"-"`
	EmailVerified    bool       `gorm:"default:false" json:"email_verified"`
	Approved         bool       `gorm:"default:false" json:"approved"`
	IsAdmin          bool       `gorm:"default:false" json:"is_admin"`
	MFARequired      bool       `gorm:"default:false" json:"mfa_required"`
	TokenVersion     int        `gorm:"default:0" json:"token_version"`
	MustChangePassword bool     `gorm:"default:false" json:"must_change_password"`
	ExpiresAt        *time.Time `json:"expires_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Groups           []Group    `gorm:"many2many:user_group" json:"groups,omitempty"`
}

func (User) TableName() string {
	return "user"
}
