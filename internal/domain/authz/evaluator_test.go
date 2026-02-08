package authz

import "testing"

func TestStaticEvaluatorAllowed(t *testing.T) {
	t.Parallel()

	evaluator := NewStaticEvaluator([]Permission{
		{Role: "admin", Resource: "admin.ping", Action: "read"},
		{Role: "observer", Resource: "health", Action: "read"},
	})

	if !evaluator.Allowed([]string{"admin"}, "admin.ping", "read") {
		t.Fatal("expected admin to be allowed for admin.ping read")
	}

	if evaluator.Allowed([]string{"observer"}, "admin.ping", "read") {
		t.Fatal("expected observer to be denied for admin.ping read")
	}
}

func TestStaticEvaluatorWildcard(t *testing.T) {
	t.Parallel()

	evaluator := NewStaticEvaluator([]Permission{
		{Role: "superadmin", Resource: "*", Action: "*"},
	})

	if !evaluator.Allowed([]string{"superadmin"}, "anything", "execute") {
		t.Fatal("expected wildcard permission to allow request")
	}
}
