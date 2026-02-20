package auth

import (
	"errors"
	"testing"
	"time"
)

func TestSessionTokenRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	token, err := SignSessionToken(SessionPrincipal{
		Subject:     "alice",
		DisplayName: "Alice",
		Roles:       []string{"league_admin"},
		ExpiresAt:   now.Add(time.Hour),
	}, "test-secret")
	if err != nil {
		t.Fatalf("expected token signing success, got error: %v", err)
	}

	principal, err := VerifySessionToken(token, "test-secret", now)
	if err != nil {
		t.Fatalf("expected token verify success, got error: %v", err)
	}
	if principal.Subject != "alice" {
		t.Fatalf("expected subject alice, got %s", principal.Subject)
	}
	if principal.DisplayName != "Alice" {
		t.Fatalf("expected displayName Alice, got %s", principal.DisplayName)
	}
	if len(principal.Roles) != 1 || principal.Roles[0] != "league_admin" {
		t.Fatalf("expected league_admin role, got %#v", principal.Roles)
	}
}

func TestSessionTokenRejectsTamperingAndExpiry(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	token, err := SignSessionToken(SessionPrincipal{
		Subject:   "bob",
		Roles:     []string{"player"},
		ExpiresAt: now.Add(time.Minute),
	}, "test-secret")
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if _, err := VerifySessionToken(token+"x", "test-secret", now); !errors.Is(err, ErrInvalidSessionToken) {
		t.Fatalf("expected invalid token error on tamper, got: %v", err)
	}

	expiredToken, err := SignSessionToken(SessionPrincipal{
		Subject:   "bob",
		Roles:     []string{"player"},
		ExpiresAt: now.Add(-time.Minute),
	}, "test-secret")
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	if _, err := VerifySessionToken(expiredToken, "test-secret", now); !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("expected expired token error, got: %v", err)
	}
}
