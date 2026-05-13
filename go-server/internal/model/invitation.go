package model

import (
	"time"
)

type Invitation struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	Email     string    `gorm:"index;type:text" json:"email"`
	Username  *string   `gorm:"type:text" json:"username,omitempty"`
	Code      string    `gorm:"uniqueIndex;type:text" json:"code"`
	Scopes    *string   `gorm:"type:text" json:"scopes,omitempty"`
	GroupID   *string   `gorm:"type:text" json:"group_id,omitempty"`
	MaxUses   *int      `json:"max_uses,omitempty"`
	UsedCount int       `gorm:"default:0" json:"used_count"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy *string   `gorm:"type:text" json:"created_by,omitempty"`
	Group     *Group    `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Creator   *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (Invitation) TableName() string {
	return "invitation"
}
