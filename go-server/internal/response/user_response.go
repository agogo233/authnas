package response

type UserResponse struct {
	ID            string  `json:"id"`
	Email         *string `json:"email,omitempty"`
	Username      string  `json:"username"`
	Name          *string `json:"name,omitempty"`
	IsAdmin       bool    `json:"isAdmin"`
	EmailVerified bool    `json:"emailVerified"`
	Approved      bool    `json:"approved"`
	MFARequired   bool    `json:"mfaRequired"`
	HasTotp       bool    `json:"hasTotp"`
	HasPasskeys   bool    `json:"hasPasskeys"`
	HasPassword   bool    `json:"hasPassword"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     *string `json:"updatedAt,omitempty"`
}

type LoginUserResponse struct {
	ID            string  `json:"id"`
	Email         *string `json:"email,omitempty"`
	Username      string  `json:"username"`
	Name          *string `json:"name,omitempty"`
	EmailVerified bool    `json:"emailVerified"`
	Approved      bool    `json:"approved"`
	IsAdmin       bool    `json:"isAdmin"`
	MFARequired   bool    `json:"mfaRequired"`
	HasTotp       bool    `json:"hasTotp,omitempty"`
	HasPasskeys   bool    `json:"hasPasskeys,omitempty"`
	HasPassword   bool    `json:"hasPassword,omitempty"`
	TokenVersion  int     `json:"tokenVersion,omitempty"`
	ExpiresAt     *string `json:"expiresAt,omitempty"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     *string `json:"updatedAt,omitempty"`
}
