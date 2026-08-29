package auth

import "testing"

func TestLocalTokenValidatorAnonymous(t *testing.T) {
	t.Parallel()

	validator := NewLocalTokenValidator("local")
	principal, err := validator.Validate("")
	if err != nil {
		t.Fatalf("expected no error for empty auth header, got %v", err)
	}
	if !principal.Anonymous {
		t.Fatal("expected anonymous principal for empty auth header")
	}
}

func TestLocalTokenValidatorValidToken(t *testing.T) {
	t.Parallel()

	validator := NewLocalTokenValidator("local")
	principal, err := validator.Validate("Bearer local:alice:admin,operator")
	if err != nil {
		t.Fatalf("expected no error for valid token, got %v", err)
	}
	if principal.Anonymous {
		t.Fatal("expected non-anonymous principal")
	}
	if principal.Subject != "alice" {
		t.Fatalf("expected subject alice, got %s", principal.Subject)
	}
	if len(principal.Roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(principal.Roles))
	}
}

func TestLocalTokenValidatorPreservesColonInSubject(t *testing.T) {
	t.Parallel()

	validator := NewLocalTokenValidator("local")
	principal, err := validator.Validate("Bearer local:discord:player-42:player")
	if err != nil {
		t.Fatalf("Validate returned %v", err)
	}
	if principal.Subject != "discord:player-42" {
		t.Fatalf("subject = %q", principal.Subject)
	}
	if len(principal.Roles) != 1 || principal.Roles[0] != "player" {
		t.Fatalf("roles = %#v", principal.Roles)
	}
}

func TestLocalTokenValidatorRejectsMalformedHeader(t *testing.T) {
	t.Parallel()

	validator := NewLocalTokenValidator("local")
	_, err := validator.Validate("Token local:alice:admin")
	if err == nil {
		t.Fatal("expected malformed header to be rejected")
	}
}
