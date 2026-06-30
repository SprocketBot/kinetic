package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")

	cfg := Load()

	if cfg.Port != defaultPort {
		t.Fatalf("expected default port %s, got %s", defaultPort, cfg.Port)
	}

	if cfg.LogLevel != defaultLogLevel {
		t.Fatalf("expected default log level %s, got %s", defaultLogLevel, cfg.LogLevel)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.Port)
	}

	if cfg.LogLevel != "debug" {
		t.Fatalf("expected log level debug, got %s", cfg.LogLevel)
	}
}

func TestVariablesDocumentsConfigSurface(t *testing.T) {
	vars := Variables()
	expected := map[string]bool{
		"PORT":                    false,
		"LOG_LEVEL":               false,
		"DATABASE_URL":            false,
		"MIGRATIONS_DIR":          false,
		"RUN_MIGRATIONS_ON_START": false,
		"REQUIRE_DATABASE":        false,
		"AUTH_SESSION_SECRET":     false,
		"AUTH_SESSION_COOKIE":     false,
		"AUTH_SESSION_TTL":        false,
		"WEB_BASE_URL":            false,
		"DISCORD_CLIENT_ID":       false,
		"DISCORD_CLIENT_SECRET":   false,
		"DISCORD_REDIRECT_URL":    false,
	}

	for _, variable := range vars {
		if _, ok := expected[variable.Name]; !ok {
			t.Fatalf("unexpected documented variable %q", variable.Name)
		}
		if variable.Description == "" {
			t.Fatalf("variable %q is missing a description", variable.Name)
		}
		expected[variable.Name] = true
	}

	for name, seen := range expected {
		if !seen {
			t.Fatalf("config variable %q is not documented", name)
		}
	}
}
