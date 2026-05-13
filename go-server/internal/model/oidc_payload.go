package model

import (
	"time"
)

type OIDCPayload struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	UID       string    `gorm:"uniqueIndex;type:text" json:"uid"`
	Payload   string    `gorm:"type:text" json:"payload"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (OIDCPayload) TableName() string {
	return "oidc_payload"
}
