package model

import (
	"time"
)

type UserGroup struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	UserID    string    `gorm:"type:text;index" json:"user_id"`
	GroupID   string    `gorm:"type:text;index" json:"group_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (UserGroup) TableName() string {
	return "user_group"
}
