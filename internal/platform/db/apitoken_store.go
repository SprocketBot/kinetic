package db

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kineticbot/kinetic-v3/internal/domain/apitoken"
)

// APITokenStore implements apitoken.Store backed by PostgreSQL.
type APITokenStore struct {
	db *sql.DB
}

// NewAPITokenStore constructs an APITokenStore using the given connection.
func NewAPITokenStore(db *sql.DB) *APITokenStore {
	return &APITokenStore{db: db}
}

// generateRawToken creates a plaintext token of the form spr_<32 random hex chars>.
func generateRawToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token bytes: %w", err)
	}
	return "spr_" + hex.EncodeToString(buf), nil
}

// hashToken returns the hex-encoded SHA-256 hash of the given plaintext token.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func mapAPITokenSQLError(err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.Code {
		case "23505":
			return fmt.Errorf("%w: %s", apitoken.ErrConflict, pgError.ConstraintName)
		case "23514":
			return fmt.Errorf("%w: %s", apitoken.ErrInvalidInput, pgError.ConstraintName)
		}
	}
	return err
}

func scanAPIToken(scanner interface {
	Scan(dest ...any) error
}) (apitoken.APIToken, error) {
	var t apitoken.APIToken
	var createdBy sql.NullInt64
	var expiresAt sql.NullTime
	var revokedAt sql.NullTime
	var lastUsedAt sql.NullTime

	// scopes column is TEXT[] — pgx/v5 driver decodes it as []string when
	// pointed at a *[]string, but the standard database/sql path requires a
	// PostgreSQL array literal. We read it as a string and parse manually.
	// Alternatively accept []byte and decode; simplest is to read into a
	// pq.Array-compatible value. Since we only use pgx here, use the pgx
	// array scanning approach via sql.Scanner on a custom type.
	var scopesRaw pgTextArray

	if err := scanner.Scan(
		&t.ID,
		&t.Name,
		&t.TokenPrefix,
		&t.Subject,
		&scopesRaw,
		&createdBy,
		&expiresAt,
		&revokedAt,
		&lastUsedAt,
		&t.CreatedAt,
	); err != nil {
		return apitoken.APIToken{}, err
	}

	t.Scopes = []string(scopesRaw)
	if createdBy.Valid {
		v := createdBy.Int64
		t.CreatedBy = &v
	}
	if expiresAt.Valid {
		v := expiresAt.Time
		t.ExpiresAt = &v
	}
	if revokedAt.Valid {
		v := revokedAt.Time
		t.RevokedAt = &v
	}
	if lastUsedAt.Valid {
		v := lastUsedAt.Time
		t.LastUsedAt = &v
	}

	return t, nil
}

// pgTextArray is a sql.Scanner that decodes a PostgreSQL text[] literal into
// a Go []string without requiring the lib/pq package.
type pgTextArray []string

func (a *pgTextArray) Scan(src any) error {
	var raw string
	switch v := src.(type) {
	case []byte:
		raw = string(v)
	case string:
		raw = v
	case nil:
		*a = []string{}
		return nil
	default:
		return fmt.Errorf("pgTextArray: unsupported type %T", src)
	}

	// PostgreSQL array literal: {} or {"a","b"} or {a,b}
	// Strip outer braces.
	if len(raw) < 2 || raw[0] != '{' || raw[len(raw)-1] != '}' {
		return fmt.Errorf("pgTextArray: invalid array literal: %q", raw)
	}
	inner := raw[1 : len(raw)-1]
	if inner == "" {
		*a = []string{}
		return nil
	}

	// Simple CSV-like split respecting double-quoted elements.
	result := make([]string, 0)
	i := 0
	for i < len(inner) {
		if inner[i] == '"' {
			// quoted element
			i++ // skip opening quote
			elem := []byte{}
			for i < len(inner) {
				if inner[i] == '\\' && i+1 < len(inner) {
					i++
					elem = append(elem, inner[i])
				} else if inner[i] == '"' {
					i++ // skip closing quote
					break
				} else {
					elem = append(elem, inner[i])
				}
				i++
			}
			result = append(result, string(elem))
			if i < len(inner) && inner[i] == ',' {
				i++ // skip comma
			}
		} else {
			// unquoted element
			start := i
			for i < len(inner) && inner[i] != ',' {
				i++
			}
			result = append(result, inner[start:i])
			if i < len(inner) && inner[i] == ',' {
				i++ // skip comma
			}
		}
	}

	*a = result
	return nil
}

// CreateAPIToken generates a new token, stores its hash, and returns both the
// record and the plaintext token (only time the plaintext is available).
func (s *APITokenStore) CreateAPIToken(ctx context.Context, input apitoken.CreateAPITokenInput) (apitoken.APIToken, string, error) {
	if input.Name == "" {
		return apitoken.APIToken{}, "", fmt.Errorf("%w: name is required", apitoken.ErrInvalidInput)
	}
	if input.Subject == "" {
		return apitoken.APIToken{}, "", fmt.Errorf("%w: subject is required", apitoken.ErrInvalidInput)
	}

	rawToken, err := generateRawToken()
	if err != nil {
		return apitoken.APIToken{}, "", err
	}

	tokenHash := hashToken(rawToken)
	// The first 8 chars of the plaintext serve as the display prefix (e.g. "spr_abcd").
	prefix := rawToken
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}

	scopes := input.Scopes
	if scopes == nil {
		scopes = []string{}
	}

	const stmt = `
INSERT INTO api_tokens(name, token_hash, subject, scopes, created_by, expires_at)
VALUES ($1, $2, $3, $4::text[], $5, $6)
RETURNING id, name, LEFT(token_hash, 8) AS token_prefix, subject, scopes, created_by, expires_at, revoked_at, last_used_at, created_at;`

	// Build a PostgreSQL array literal for scopes.
	scopesLiteral := buildPGTextArray(scopes)

	row := s.db.QueryRowContext(ctx, stmt,
		input.Name,
		tokenHash,
		input.Subject,
		scopesLiteral,
		nullInt64(input.CreatedBy),
		nullTime(input.ExpiresAt),
	)

	// Override the returned prefix with the plaintext prefix for display.
	token, err := scanAPIToken(row)
	if err != nil {
		return apitoken.APIToken{}, "", mapAPITokenSQLError(err)
	}
	token.TokenPrefix = prefix
	return token, rawToken, nil
}

// ListAPITokens returns all tokens for the given subject that have not been revoked.
func (s *APITokenStore) ListAPITokens(ctx context.Context, subject string) ([]apitoken.APIToken, error) {
	const stmt = `
SELECT id, name, LEFT(token_hash, 8) AS token_prefix, subject, scopes, created_by, expires_at, revoked_at, last_used_at, created_at
FROM api_tokens
WHERE subject = $1
  AND revoked_at IS NULL
ORDER BY created_at DESC;`

	rows, err := s.db.QueryContext(ctx, stmt, subject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]apitoken.APIToken, 0)
	for rows.Next() {
		t, err := scanAPIToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

// RevokeAPIToken marks the identified token as revoked, but only if it belongs
// to the provided subject.
func (s *APITokenStore) RevokeAPIToken(ctx context.Context, input apitoken.RevokeAPITokenInput) (apitoken.APIToken, error) {
	const stmt = `
UPDATE api_tokens
SET revoked_at = NOW()
WHERE id = $1
  AND subject = $2
  AND revoked_at IS NULL
RETURNING id, name, LEFT(token_hash, 8) AS token_prefix, subject, scopes, created_by, expires_at, revoked_at, last_used_at, created_at;`

	row := s.db.QueryRowContext(ctx, stmt, input.TokenID, input.Subject)
	t, err := scanAPIToken(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apitoken.APIToken{}, fmt.Errorf("%w: token not found or does not belong to subject", apitoken.ErrNotFound)
		}
		return apitoken.APIToken{}, mapAPITokenSQLError(err)
	}
	return t, nil
}

// ValidateAPIToken hashes rawToken, looks it up in the database, verifies it
// is active and not expired, updates last_used_at, and returns the subject and
// scopes.
func (s *APITokenStore) ValidateAPIToken(ctx context.Context, rawToken string) (apitoken.ValidateAPITokenResult, error) {
	tokenHash := hashToken(rawToken)

	const stmt = `
UPDATE api_tokens
SET last_used_at = NOW()
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > NOW())
RETURNING subject, scopes;`

	var subject string
	var scopesRaw pgTextArray

	row := s.db.QueryRowContext(ctx, stmt, tokenHash)
	if err := row.Scan(&subject, &scopesRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apitoken.ValidateAPITokenResult{}, fmt.Errorf("%w: token not found or inactive", apitoken.ErrNotFound)
		}
		return apitoken.ValidateAPITokenResult{}, err
	}

	return apitoken.ValidateAPITokenResult{
		Subject: subject,
		Scopes:  []string(scopesRaw),
	}, nil
}

// buildPGTextArray converts a Go string slice to a PostgreSQL array literal
// such as {"a","b"} suitable for use as a query parameter with ::text[].
func buildPGTextArray(values []string) string {
	if len(values) == 0 {
		return "{}"
	}
	result := "{"
	for i, v := range values {
		if i > 0 {
			result += ","
		}
		result += `"` + escapePGString(v) + `"`
	}
	result += "}"
	return result
}

func escapePGString(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '"' || s[i] == '\\' {
			out = append(out, '\\')
		}
		out = append(out, s[i])
	}
	return string(out)
}

func nullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func nullTime(v *time.Time) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *v, Valid: true}
}
