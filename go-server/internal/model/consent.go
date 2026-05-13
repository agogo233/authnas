package model

import (
	"time"
)

type Consent struct {
	ID        string     `gorm:"primaryKey;type:text" json:"id"`
	UserID    string     `gorm:"index;type:text" json:"user_id"`
	ClientID  string     `gorm:"index;type:text" json:"client_id"`
	Scopes    string     `gorm:"type:text" json:"scopes"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Consent) TableName() string {
	return "consent"
}
