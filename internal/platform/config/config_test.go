package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DEPLOYMENT_ENV", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")

	cfg := Load()

	if cfg.DeploymentEnv != defaultDeploymentEnv {
		t.Fatalf("expected default deployment env %s, got %s", defaultDeploymentEnv, cfg.DeploymentEnv)
	}

	if cfg.Port != defaultPort {
		t.Fatalf("expected default port %s, got %s", defaultPort, cfg.Port)
	}

	if cfg.LogLevel != defaultLogLevel {
		t.Fatalf("expected default log level %s, got %s", defaultLogLevel, cfg.LogLevel)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("DEPLOYMENT_ENV", "staging")
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("AUTH_LOCAL_LOGIN_ENABLED", "false")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://kinetic.mlesports.gg")

	cfg := Load()

	if cfg.DeploymentEnv != "staging" {
		t.Fatalf("expected deployment env staging, got %s", cfg.DeploymentEnv)
	}

	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.Port)
	}

	if cfg.LogLevel != "debug" {
		t.Fatalf("expected log level debug, got %s", cfg.LogLevel)
	}

	if cfg.AuthLocalLogin {
		t.Fatal("expected local login override to be false")
	}

	if !cfg.AuthLocalLoginSet {
		t.Fatal("expected local login to be marked as configured by Load")
	}

	if cfg.CORSAllowedOrigins != "https://kinetic.mlesports.gg" {
		t.Fatalf("expected CORS override, got %s", cfg.CORSAllowedOrigins)
	}
}

func TestVariablesDocumentsConfigSurface(t *testing.T) {
	vars := Variables()
	expected := map[string]bool{
		"DEPLOYMENT_ENV":           false,
		"PORT":                     false,
		"LOG_LEVEL":                false,
		"DATABASE_URL":             false,
		"MIGRATIONS_DIR":           false,
		"RUN_MIGRATIONS_ON_START":  false,
		"REQUIRE_DATABASE":         false,
		"AUTH_SESSION_SECRET":      false,
		"AUTH_SESSION_COOKIE":      false,
		"AUTH_SESSION_TTL":         false,
		"AUTH_LOCAL_LOGIN_ENABLED": false,
		"WEB_BASE_URL":             false,
		"CORS_ALLOWED_ORIGINS":     false,
		"DISCORD_CLIENT_ID":        false,
		"DISCORD_CLIENT_SECRET":    false,
		"DISCORD_REDIRECT_URL":     false,
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

func TestValidateRuntimeSafetyAllowsLocalDefaults(t *testing.T) {
	cfg := Config{
		DeploymentEnv:     "local",
		AuthLocalLogin:    true,
		AuthSessionSecret: defaultAuthSessionSecret,
		WebBaseURL:        defaultWebBaseURL,
	}

	if err := ValidateRuntimeSafety(cfg); err != nil {
		t.Fatalf("expected local defaults to pass safety validation, got %v", err)
	}
}

func TestValidateRuntimeSafetyRejectsProductionLocalAuth(t *testing.T) {
	cfg := Config{
		DeploymentEnv:     "production",
		AuthLocalLogin:    true,
		AuthSessionSecret: "real-secret",
		WebBaseURL:        "https://kinetic.mlesports.gg",
	}

	if err := ValidateRuntimeSafety(cfg); err == nil {
		t.Fatal("expected production local auth to fail safety validation")
	}
}

func TestValidateRuntimeSafetyRejectsProductionDefaultSecret(t *testing.T) {
	cfg := Config{
		DeploymentEnv:     "staging",
		AuthLocalLogin:    false,
		AuthSessionSecret: defaultAuthSessionSecret,
		WebBaseURL:        "https://staging.kinetic.mlesports.gg",
	}

	if err := ValidateRuntimeSafety(cfg); err == nil {
		t.Fatal("expected production-like default secret to fail safety validation")
	}
}

func TestValidateRuntimeSafetyRejectsProductionWildcardCORS(t *testing.T) {
	cfg := Config{
		DeploymentEnv:      "prod",
		AuthLocalLogin:     false,
		AuthSessionSecret:  "real-secret",
		WebBaseURL:         "https://kinetic.mlesports.gg",
		CORSAllowedOrigins: "https://kinetic.mlesports.gg,*",
	}

	if err := ValidateRuntimeSafety(cfg); err == nil {
		t.Fatal("expected production wildcard CORS to fail safety validation")
	}
}
