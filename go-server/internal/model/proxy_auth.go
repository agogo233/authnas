package model

import (
	"time"
)

type ProxyAuth struct {
	ID         string    `gorm:"primaryKey;type:text" json:"id"`
	Name       string    `gorm:"type:text" json:"name"`
	ProxyURL   string    `gorm:"type:text" json:"proxy_url"`
	HeaderName string    `gorm:"type:text" json:"header_name"`
	Scopes     *string   `gorm:"type:text" json:"scopes,omitempty"`
	GroupID    *string   `gorm:"type:text" json:"group_id,omitempty"`
	Enabled    bool      `gorm:"default:true" json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Group      *Group    `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

func (ProxyAuth) TableName() string {
	return "proxy_auth"
}

type EmailLog struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	Recipient string    `gorm:"type:text" json:"recipient"`
	Subject   string    `gorm:"type:text" json:"subject"`
	Template  string    `gorm:"type:text" json:"template"`
	Status    string    `gorm:"type:text" json:"status"`
	Error     *string   `gorm:"type:text" json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (EmailLog) TableName() string {
	return "email_log"
}
