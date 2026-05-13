package model

import (
	"time"
)

type TOTP struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	UserID    string    `gorm:"uniqueIndex;type:text" json:"user_id"`
	Secret    string    `gorm:"type:text" json:"secret"`
	Issuer    string    `gorm:"type:text" json:"issuer"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (TOTP) TableName() string {
	return "totp"
}
