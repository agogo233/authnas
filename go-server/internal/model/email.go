package model

import (
	"time"
)

type EmailVerification struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	UserID    string    `gorm:"index;type:text" json:"user_id"`
	Email     string    `gorm:"type:text" json:"email"`
	Code      string    `gorm:"uniqueIndex;type:text" json:"code"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (EmailVerification) TableName() string {
	return "email_verification"
}

type PasswordReset struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	UserID    string    `gorm:"index;type:text" json:"user_id"`
	Code      string    `gorm:"uniqueIndex;type:text" json:"code"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (PasswordReset) TableName() string {
	return "password_reset"
}
