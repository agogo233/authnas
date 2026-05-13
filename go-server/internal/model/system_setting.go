package model

import "time"

type SystemSetting struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Key       string    `gorm:"uniqueIndex;type:text" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SystemSetting) TableName() string {
	return "system_setting"
}
