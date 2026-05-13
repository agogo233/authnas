package model

import (
	"time"
)

type Client struct {
	ID                     string     `gorm:"primaryKey;type:text" json:"id"`
	ClientID               string     `gorm:"uniqueIndex;type:text" json:"client_id"`
	ClientSecret           *string    `gorm:"type:text" json:"client_secret,omitempty"`
	PreviousClientSecret   *string    `gorm:"type:text" json:"-"`
	ClientSecretRotatedAt  *time.Time `json:"-"`
	Name                   string     `gorm:"type:text" json:"name"`
	LogoURI                *string    `gorm:"type:text" json:"logo_uri,omitempty"`
	RedirectURIs           string     `gorm:"type:text" json:"redirect_uris"`
	PostLogoutRedirectURIs *string    `gorm:"type:text" json:"post_logout_redirect_uris,omitempty"`
	GrantTypes             *string    `gorm:"type:text" json:"grant_types,omitempty"`
	ResponseTypes          *string    `gorm:"type:text" json:"response_types,omitempty"`
	Scopes                 *string    `gorm:"type:text" json:"scopes,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

func (Client) TableName() string {
	return "client"
}
