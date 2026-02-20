package config

import "os"

const (
	defaultPort                 = "8080"
	defaultLogLevel             = "info"
	defaultDatabaseURL          = "postgres://postgres:postgres@localhost:5432/sprocket?sslmode=disable"
	defaultMigrationsDir        = "migrations"
	defaultRunMigrationsOnStart = "false"
	defaultRequireDatabase      = "false"
	defaultAuthSessionSecret    = "dev-insecure-session-secret"
	defaultAuthSessionCookie    = "sprocket_session"
	defaultAuthSessionTTL       = "12h"
	defaultWebBaseURL           = "http://localhost:5173"
)

type Config struct {
	Port                 string
	LogLevel             string
	DatabaseURL          string
	MigrationsDir        string
	RunMigrationsOnStart bool
	RequireDatabase      bool
	AuthSessionSecret    string
	AuthSessionCookie    string
	AuthSessionTTL       string
	WebBaseURL           string
}

func Load() Config {
	return Config{
		Port:                 getOrDefault("PORT", defaultPort),
		LogLevel:             getOrDefault("LOG_LEVEL", defaultLogLevel),
		DatabaseURL:          getOrDefault("DATABASE_URL", defaultDatabaseURL),
		MigrationsDir:        getOrDefault("MIGRATIONS_DIR", defaultMigrationsDir),
		RunMigrationsOnStart: getOrDefault("RUN_MIGRATIONS_ON_START", defaultRunMigrationsOnStart) == "true",
		RequireDatabase:      getOrDefault("REQUIRE_DATABASE", defaultRequireDatabase) == "true",
		AuthSessionSecret:    getOrDefault("AUTH_SESSION_SECRET", defaultAuthSessionSecret),
		AuthSessionCookie:    getOrDefault("AUTH_SESSION_COOKIE", defaultAuthSessionCookie),
		AuthSessionTTL:       getOrDefault("AUTH_SESSION_TTL", defaultAuthSessionTTL),
		WebBaseURL:           getOrDefault("WEB_BASE_URL", defaultWebBaseURL),
	}
}

func getOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
