package auth

import (
	"context"
	"testing"
)

func TestPrincipalContext(t *testing.T) {
	if got := PrincipalFromContext(context.Background()); !got.Anonymous || len(got.Roles) != 0 {
		t.Fatalf("missing principal = %#v", got)
	}

	want := NewPrincipal("subject", []string{"admin"})
	got := PrincipalFromContext(WithPrincipal(context.Background(), want))
	if got.Subject != want.Subject || got.Anonymous || len(got.Roles) != 1 || got.Roles[0] != "admin" {
		t.Fatalf("principal = %#v", got)
	}
}
