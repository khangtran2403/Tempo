package domain

import "time"

type Integration struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	Provider       string                 `json:"provider"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	AccessToken    string                 `json:"-"`
	RefreshToken   string                 `json:"-"`
	TokenExpiresAt *time.Time             `json:"token_expires_at"`
	Metadata       map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	IsActive       bool                   `json:"is_active"`
	IsRevoked      bool                   `json:"is_revoked"`
	LastUsedAt     *time.Time             `json:"last_used_at"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ProviderConfig định nghĩa config cho mỗi provider
type ProviderConfig struct {
	Provider     string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	DataURL      string // API endpoint
}

// IntegrationAction định nghĩa available actions
type IntegrationAction struct {
	ID          string
	Provider    string
	Action      string
	Description string
	Parameters  []ActionParameter
}

type ActionParameter struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Options     []string
}
