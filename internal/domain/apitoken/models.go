package apitoken

import "time"

// APIToken represents a stored API token record. The hashed_token is never
// returned to callers; token_prefix is shown for display purposes only.
type APIToken struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"tokenPrefix"`
	Subject     string     `json:"subject"`
	Scopes      []string   `json:"scopes"`
	CreatedBy   *int64     `json:"createdBy,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	RevokedAt   *time.Time `json:"revokedAt,omitempty"`
	LastUsedAt  *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// CreateAPITokenInput carries the parameters needed to create a new API token.
type CreateAPITokenInput struct {
	Name      string     `json:"name"`
	Subject   string     `json:"subject"`
	Scopes    []string   `json:"scopes"`
	CreatedBy *int64     `json:"createdBy,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// RevokeAPITokenInput identifies a token to revoke.
type RevokeAPITokenInput struct {
	TokenID int64  `json:"tokenId"`
	Subject string `json:"subject"`
}

// ValidateAPITokenResult is returned by ValidateAPIToken and carries
// the information needed to build a principal for an authenticated request.
type ValidateAPITokenResult struct {
	Subject string
	Scopes  []string
}
