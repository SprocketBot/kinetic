package apitoken

import "context"

// Store defines the persistence operations for API tokens.
type Store interface {
	// CreateAPIToken generates a new token, stores its hash, and returns both
	// the APIToken record and the plaintext token. The plaintext token is
	// returned exactly once and is never stored; callers must present it to
	// users immediately.
	CreateAPIToken(ctx context.Context, input CreateAPITokenInput) (APIToken, string, error)

	// ListAPITokens returns all non-revoked tokens whose subject matches the
	// provided subject string.
	ListAPITokens(ctx context.Context, subject string) ([]APIToken, error)

	// RevokeAPIToken marks a token as revoked. Only the token owner (matched
	// by subject) may revoke their own tokens.
	RevokeAPIToken(ctx context.Context, input RevokeAPITokenInput) (APIToken, error)

	// ValidateAPIToken looks up a token by its SHA-256 hash, verifies it is
	// active and not expired, updates last_used_at, and returns the associated
	// subject and scopes. Returns ErrNotFound if the token is unknown,
	// ErrUnauthorized if it has been revoked or has expired.
	ValidateAPIToken(ctx context.Context, rawToken string) (ValidateAPITokenResult, error)
}
