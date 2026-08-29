package orgconfig

import "testing"

func TestDefaultsCoverEveryKnownKey(t *testing.T) {
	keys := AllKnownKeys()
	if len(keys) == 0 {
		t.Fatal("AllKnownKeys returned no keys")
	}
	seen := make(map[string]bool, len(keys))
	for _, key := range keys {
		if seen[key] {
			t.Fatalf("duplicate key %q", key)
		}
		seen[key] = true
		if Default(key) == "" {
			t.Fatalf("key %q has no default", key)
		}
	}
	if got := Default("unknown"); got != "" {
		t.Fatalf("unknown default = %q", got)
	}
}
