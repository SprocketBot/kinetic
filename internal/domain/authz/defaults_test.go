package authz

import "testing"

func TestDefaultPermissionsAndContextEvaluation(t *testing.T) {
	permissions := DefaultPermissions()
	if len(permissions) == 0 {
		t.Fatal("DefaultPermissions returned no permissions")
	}
	evaluator := NewStaticEvaluator(permissions)

	if !evaluator.Allowed([]string{"admin"}, "anything", "delete") {
		t.Fatal("admin wildcard permission was not applied")
	}
	if evaluator.Allowed([]string{"player"}, ResourceRoleAssignment, ActionCreate) {
		t.Fatal("player unexpectedly allowed to create role assignments")
	}
	if !evaluator.AllowedInContext(nil, []string{"captain"}, ResourceScrim, ActionCreate, GlobalContext()) {
		t.Fatal("scoped captain permission was not applied")
	}
	if !evaluator.AllowedInContext([]string{"operator"}, nil, "anything", "anything", AuthzContext{}) {
		t.Fatal("global operator permission was not applied")
	}
}

func TestMatches(t *testing.T) {
	for _, tt := range []struct {
		policy, request string
		want            bool
	}{{"*", "read", true}, {"read", "read", true}, {"write", "read", false}} {
		if got := matches(tt.policy, tt.request); got != tt.want {
			t.Fatalf("matches(%q, %q)=%v", tt.policy, tt.request, got)
		}
	}
}
