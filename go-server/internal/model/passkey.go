package model

import (
	"time"
)

type Passkey struct {
	ID              string     `gorm:"primaryKey;type:text" json:"id"`
	UserID          string     `gorm:"index;type:text" json:"user_id"`
	Name            *string    `gorm:"type:text" json:"name"`
	CredentialID    string     `gorm:"uniqueIndex;type:text" json:"credential_id"`
	PublicKey       string     `gorm:"type:text" json:"public_key"`
	AttestationType *string    `gorm:"type:text" json:"attestation_type,omitempty"`
	Transports      *string    `gorm:"type:text" json:"transports,omitempty"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	User            *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Passkey) TableName() string {
	return "passkey"
}

type PasskeyAuthOptions struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	UserID    *string   `gorm:"index;type:text" json:"user_id,omitempty"`
	Challenge string    `gorm:"type:text" json:"challenge"`
	Options   string    `gorm:"type:text" json:"options"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (PasskeyAuthOptions) TableName() string {
	return "passkey_auth_options"
}
