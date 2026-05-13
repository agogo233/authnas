package model

import (
	"time"
)

type Group struct {
	ID          string    `gorm:"primaryKey;type:text" json:"id"`
	Name        string    `gorm:"uniqueIndex;type:text" json:"name"`
	Description *string   `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Users       []User    `gorm:"many2many:user_group" json:"users,omitempty"`
}

func (Group) TableName() string {
	return "groups"
}
