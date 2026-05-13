package model

import (
	"time"
)

type Key struct {
	ID                     string    `gorm:"primaryKey;type:text" json:"id"`
	UserID                 string    `gorm:"index;type:text" json:"user_id"`
	ClientID               string    `gorm:"index;type:text" json:"client_id"`
	TokenVersion           int       `json:"token_version"`
	RefreshTokenHash       string    `gorm:"uniqueIndex;type:text" json:"refresh_token_hash"`
	RefreshTokenLookupHash string    `gorm:"index;type:text" json:"refresh_token_lookup_hash"`
	ExpiresAt              time.Time `gorm:"index" json:"expires_at"`
	CreatedAt              time.Time `json:"created_at"`
	UserAgent              string    `gorm:"type:text" json:"user_agent,omitempty"`
	GroupNames             string    `gorm:"type:text" json:"group_names,omitempty"`
	Nonce                  string    `gorm:"type:text" json:"nonce,omitempty"`
	User                   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Key) TableName() string {
	return "key"
}
