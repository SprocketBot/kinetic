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
		{Name: "PORT", Default: defaultPort, Description: "HTTP listen port inside the process."},
		{Name: "LOG_LEVEL", Default: defaultLogLevel, Description: "Structured log level: debug, info, warn, or error."},
		{Name: "DATABASE_URL", Default: defaultDatabaseURL, Secret: true, Description: "PostgreSQL connection URL used by the API and migrator."},
		{Name: "MIGRATIONS_DIR", Default: defaultMigrationsDir, Description: "Directory containing *.up.sql migration files."},
		{Name: "RUN_MIGRATIONS_ON_START", Default: defaultRunMigrationsOnStart, Description: "When true, the API applies pending migrations during startup."},
		{Name: "REQUIRE_DATABASE", Default: defaultRequireDatabase, Description: "When true, startup fails unless the database is reachable and stores can be initialized."},
		{Name: "AUTH_SESSION_SECRET", Default: defaultAuthSessionSecret, Secret: true, Description: "HMAC secret used to sign browser session cookies."},
		{Name: "AUTH_SESSION_COOKIE", Default: defaultAuthSessionCookie, Description: "Cookie name for browser sessions."},
		{Name: "AUTH_SESSION_TTL", Default: defaultAuthSessionTTL, Description: "Session lifetime parsed with Go duration syntax."},
		{Name: "WEB_BASE_URL", Default: defaultWebBaseURL, Description: "Allowed web origin and default post-login redirect base URL."},
		{Name: "DISCORD_CLIENT_ID", Description: "Discord OAuth client ID. Discord login is disabled when unset."},
		{Name: "DISCORD_CLIENT_SECRET", Secret: true, Description: "Discord OAuth client secret. Discord login is disabled when unset."},
		{Name: "DISCORD_REDIRECT_URL", Description: "Optional explicit Discord OAuth callback URL."},
	}
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
