package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidSessionToken = errors.New("invalid session token")
	ErrSessionExpired      = errors.New("session expired")
)

type SessionPrincipal struct {
	Subject     string    `json:"subject"`
	DisplayName string    `json:"displayName"`
	Roles       []string  `json:"roles"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type sessionClaims struct {
	Subject     string   `json:"sub"`
	DisplayName string   `json:"name"`
	Roles       []string `json:"roles"`
	ExpiresAt   int64    `json:"exp"`
}

func SignSessionToken(principal SessionPrincipal, secret string) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("%w: empty signing secret", ErrInvalidSessionToken)
	}
	if strings.TrimSpace(principal.Subject) == "" {
		return "", fmt.Errorf("%w: subject is required", ErrInvalidSessionToken)
	}
	if principal.ExpiresAt.IsZero() {
		return "", fmt.Errorf("%w: expiration is required", ErrInvalidSessionToken)
	}

	claims := sessionClaims{
		Subject:     principal.Subject,
		DisplayName: principal.DisplayName,
		Roles:       principal.Roles,
		ExpiresAt:   principal.ExpiresAt.UTC().Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := signPayload(encodedPayload, secret)
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	return encodedPayload + "." + encodedSignature, nil
}

func VerifySessionToken(token string, secret string, now time.Time) (SessionPrincipal, error) {
	if strings.TrimSpace(secret) == "" {
		return SessionPrincipal{}, fmt.Errorf("%w: empty signing secret", ErrInvalidSessionToken)
	}
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 {
		return SessionPrincipal{}, ErrInvalidSessionToken
	}

	encodedPayload := parts[0]
	receivedSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return SessionPrincipal{}, ErrInvalidSessionToken
	}

	expectedSignature := signPayload(encodedPayload, secret)
	if !hmac.Equal(receivedSignature, expectedSignature) {
		return SessionPrincipal{}, ErrInvalidSessionToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return SessionPrincipal{}, ErrInvalidSessionToken
	}

	var claims sessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return SessionPrincipal{}, ErrInvalidSessionToken
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return SessionPrincipal{}, ErrInvalidSessionToken
	}

	expiresAt := time.Unix(claims.ExpiresAt, 0).UTC()
	if now.UTC().After(expiresAt) {
		return SessionPrincipal{}, ErrSessionExpired
	}

	displayName := strings.TrimSpace(claims.DisplayName)
	if displayName == "" {
		displayName = claims.Subject
	}

	return SessionPrincipal{
		Subject:     claims.Subject,
		DisplayName: displayName,
		Roles:       claims.Roles,
		ExpiresAt:   expiresAt,
	}, nil
}

func signPayload(encodedPayload, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(encodedPayload))
	return mac.Sum(nil)
}
