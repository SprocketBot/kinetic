package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	defaultDeploymentEnv        = "local"
	defaultPort                 = "8080"
	defaultLogLevel             = "info"
	defaultDatabaseURL          = "postgres://postgres:postgres@localhost:5432/kinetic?sslmode=disable"
	defaultMigrationsDir        = "migrations"
	defaultRunMigrationsOnStart = "false"
	defaultRequireDatabase      = "false"
	defaultAuthSessionSecret    = "dev-insecure-session-secret"
	defaultAuthSessionCookie    = "kinetic_session"
	defaultAuthSessionTTL       = "12h"
	defaultAuthLocalLogin       = "true"
	defaultWebBaseURL           = "http://localhost:5173"
	defaultCORSAllowedOrigins   = ""
)

type Config struct {
	DeploymentEnv        string
	Port                 string
	LogLevel             string
	DatabaseURL          string
	MigrationsDir        string
	RunMigrationsOnStart bool
	RequireDatabase      bool
	AuthSessionSecret    string
	AuthSessionCookie    string
	AuthSessionTTL       string
	AuthLocalLogin       bool
	AuthLocalLoginSet    bool
	WebBaseURL           string
	CORSAllowedOrigins   string
	DiscordClientID      string
	DiscordClientSecret  string
	DiscordRedirectURL   string
}

type Variable struct {
	Name        string
	Default     string
	Required    bool
	Secret      bool
	Description string
}

func Variables() []Variable {
	return []Variable{
		{Name: "DEPLOYMENT_ENV", Default: defaultDeploymentEnv, Description: "Runtime lane: local, development, staging, or production. Production-like lanes enable safety validation."},
		{Name: "PORT", Default: defaultPort, Description: "HTTP listen port inside the process."},
		{Name: "LOG_LEVEL", Default: defaultLogLevel, Description: "Structured log level: debug, info, warn, or error."},
		{Name: "DATABASE_URL", Default: defaultDatabaseURL, Secret: true, Description: "PostgreSQL connection URL used by the API and migrator."},
		{Name: "MIGRATIONS_DIR", Default: defaultMigrationsDir, Description: "Directory containing *.up.sql migration files."},
		{Name: "RUN_MIGRATIONS_ON_START", Default: defaultRunMigrationsOnStart, Description: "When true, the API applies pending migrations during startup."},
		{Name: "REQUIRE_DATABASE", Default: defaultRequireDatabase, Description: "When true, startup fails unless the database is reachable and stores can be initialized."},
		{Name: "AUTH_SESSION_SECRET", Default: defaultAuthSessionSecret, Secret: true, Description: "HMAC secret used to sign browser session cookies."},
		{Name: "AUTH_SESSION_COOKIE", Default: defaultAuthSessionCookie, Description: "Cookie name for browser sessions."},
		{Name: "AUTH_SESSION_TTL", Default: defaultAuthSessionTTL, Description: "Session lifetime parsed with Go duration syntax."},
		{Name: "AUTH_LOCAL_LOGIN_ENABLED", Default: defaultAuthLocalLogin, Description: "Enables local query-parameter login. Keep true only for local development and release-evidence rehearsal."},
		{Name: "WEB_BASE_URL", Default: defaultWebBaseURL, Description: "Allowed web origin and default post-login redirect base URL."},
		{Name: "CORS_ALLOWED_ORIGINS", Default: defaultCORSAllowedOrigins, Description: "Comma-separated credentialed CORS origins. Defaults to WEB_BASE_URL when unset."},
		{Name: "DISCORD_CLIENT_ID", Description: "Discord OAuth client ID. Discord login is disabled when unset."},
		{Name: "DISCORD_CLIENT_SECRET", Secret: true, Description: "Discord OAuth client secret. Discord login is disabled when unset."},
		{Name: "DISCORD_REDIRECT_URL", Description: "Optional explicit Discord OAuth callback URL."},
	}
}

func Load() Config {
	return Config{
		DeploymentEnv:        getOrDefault("DEPLOYMENT_ENV", defaultDeploymentEnv),
		Port:                 getOrDefault("PORT", defaultPort),
		LogLevel:             getOrDefault("LOG_LEVEL", defaultLogLevel),
		DatabaseURL:          getOrDefault("DATABASE_URL", defaultDatabaseURL),
		MigrationsDir:        getOrDefault("MIGRATIONS_DIR", defaultMigrationsDir),
		RunMigrationsOnStart: parseBool(getOrDefault("RUN_MIGRATIONS_ON_START", defaultRunMigrationsOnStart)),
		RequireDatabase:      parseBool(getOrDefault("REQUIRE_DATABASE", defaultRequireDatabase)),
		AuthSessionSecret:    getOrDefault("AUTH_SESSION_SECRET", defaultAuthSessionSecret),
		AuthSessionCookie:    getOrDefault("AUTH_SESSION_COOKIE", defaultAuthSessionCookie),
		AuthSessionTTL:       getOrDefault("AUTH_SESSION_TTL", defaultAuthSessionTTL),
		AuthLocalLogin:       parseBool(getOrDefault("AUTH_LOCAL_LOGIN_ENABLED", defaultAuthLocalLogin)),
		AuthLocalLoginSet:    true,
		WebBaseURL:           getOrDefault("WEB_BASE_URL", defaultWebBaseURL),
		CORSAllowedOrigins:   getOrDefault("CORS_ALLOWED_ORIGINS", defaultCORSAllowedOrigins),
		DiscordClientID:      os.Getenv("DISCORD_CLIENT_ID"),
		DiscordClientSecret:  os.Getenv("DISCORD_CLIENT_SECRET"),
		DiscordRedirectURL:   os.Getenv("DISCORD_REDIRECT_URL"),
	}
}

func getOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func parseBool(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "true")
}

func ValidateRuntimeSafety(cfg Config) error {
	if !isProductionLike(cfg.DeploymentEnv) {
		return nil
	}

	if cfg.AuthLocalLogin {
		return fmt.Errorf("AUTH_LOCAL_LOGIN_ENABLED must be false when DEPLOYMENT_ENV=%s", cfg.DeploymentEnv)
	}
	if strings.TrimSpace(cfg.AuthSessionSecret) == "" || cfg.AuthSessionSecret == defaultAuthSessionSecret {
		return fmt.Errorf("AUTH_SESSION_SECRET must be overridden when DEPLOYMENT_ENV=%s", cfg.DeploymentEnv)
	}
	if looksLocalURL(cfg.WebBaseURL) {
		return fmt.Errorf("WEB_BASE_URL must not be a local development URL when DEPLOYMENT_ENV=%s", cfg.DeploymentEnv)
	}
	if strings.Contains(cfg.CORSAllowedOrigins, "*") {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS must not contain wildcard origins when DEPLOYMENT_ENV=%s", cfg.DeploymentEnv)
	}
	return nil
}

func isProductionLike(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "prod", "production", "staging":
		return true
	default:
		return false
	}
}

func looksLocalURL(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return normalized == "" ||
		strings.Contains(normalized, "localhost") ||
		strings.Contains(normalized, "127.0.0.1") ||
		strings.Contains(normalized, "0.0.0.0")
}
