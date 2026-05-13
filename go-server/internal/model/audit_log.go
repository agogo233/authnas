package model

import (
	"time"
)

type AuditLog struct {
	ID           string    `gorm:"primaryKey;type:text" json:"id"`
	Timestamp    time.Time `gorm:"index" json:"timestamp"`
	EventType    string    `gorm:"index;type:text" json:"event_type"`
	UserID       string    `gorm:"index;type:text" json:"user_id,omitempty"`
	Username     string    `gorm:"type:text" json:"username,omitempty"`
	ClientID     string    `gorm:"type:text" json:"client_id,omitempty"`
	IPAddress    string    `gorm:"type:text" json:"ip_address,omitempty"`
	UserAgent    string    `gorm:"type:text" json:"user_agent,omitempty"`
	Success      bool      `json:"success"`
	ErrorMessage string    `gorm:"type:text" json:"error_message,omitempty"`
	Metadata     string    `gorm:"type:text" json:"metadata,omitempty"`
}

func (AuditLog) TableName() string {
	return "audit_log"
}
