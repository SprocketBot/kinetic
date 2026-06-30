package http

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/authz"
	"github.com/sprocketbot/sprocket-v3/internal/platform/auth"
	"github.com/sprocketbot/sprocket-v3/internal/platform/config"
)

type routeRegistrar struct {
	logger              *slog.Logger
	deps                Dependencies
	tokenValidator      auth.TokenValidator
	evaluator           authz.Evaluator
	sessionTTL          time.Duration
	sessionCookieName   string
	sessionSecret       string
	webBaseURL          string
	discordClientID     string
	discordClientSecret string
	discordRedirectURL  string
}

func newRouteRegistrar(cfg config.Config, logger *slog.Logger, deps Dependencies) routeRegistrar {
	deps = deps.normalized()
	tokenValidator := deps.TokenValidator
	if tokenValidator == nil {
		tokenValidator = auth.NewLocalTokenValidator("local")
	}

	sessionCookieName := strings.TrimSpace(cfg.AuthSessionCookie)
	if sessionCookieName == "" {
		sessionCookieName = "sprocket_session"
	}

	sessionSecret := strings.TrimSpace(cfg.AuthSessionSecret)
	if sessionSecret == "" {
		sessionSecret = "dev-insecure-session-secret"
	}

	webBaseURL := strings.TrimRight(strings.TrimSpace(cfg.WebBaseURL), "/")
	if webBaseURL == "" {
		webBaseURL = "http://localhost:5173"
	}

	evaluator := deps.Evaluator
	if evaluator == nil {
		evaluator = authz.NewStaticEvaluator(authz.DefaultPermissions())
	}

	return routeRegistrar{
		logger:              logger,
		deps:                deps,
		tokenValidator:      tokenValidator,
		evaluator:           evaluator,
		sessionTTL:          parseSessionTTL(cfg.AuthSessionTTL),
		sessionCookieName:   sessionCookieName,
		sessionSecret:       sessionSecret,
		webBaseURL:          webBaseURL,
		discordClientID:     strings.TrimSpace(cfg.DiscordClientID),
		discordClientSecret: strings.TrimSpace(cfg.DiscordClientSecret),
		discordRedirectURL:  strings.TrimSpace(cfg.DiscordRedirectURL),
	}
}

func (d Dependencies) normalized() Dependencies {
	if d.HierarchyStore == nil {
		return d
	}
	if d.LeagueStore == nil {
		d.LeagueStore = d.HierarchyStore
	}
	if d.PlayerStore == nil {
		d.PlayerStore = d.HierarchyStore
	}
	if d.RosterStore == nil {
		d.RosterStore = d.HierarchyStore
	}
	if d.RoleStore == nil {
		d.RoleStore = d.HierarchyStore
	}
	if d.QueueStore == nil {
		d.QueueStore = d.HierarchyStore
	}
	if d.PlatformStore == nil {
		d.PlatformStore = d.HierarchyStore
	}
	if d.EligibilityStore == nil {
		d.EligibilityStore = d.HierarchyStore
	}
	if d.ScrimStore == nil {
		d.ScrimStore = d.HierarchyStore
	}
	if d.RatingStore == nil {
		d.RatingStore = d.HierarchyStore
	}
	if d.ResultStore == nil {
		d.ResultStore = d.HierarchyStore
	}
	if d.ReplayStore == nil {
		d.ReplayStore = d.HierarchyStore
	}
	if d.ExceptionStore == nil {
		d.ExceptionStore = d.HierarchyStore
	}
	if d.SchedulingStore == nil {
		d.SchedulingStore = d.HierarchyStore
	}
	return d
}

func (r routeRegistrar) register(mux *http.ServeMux) {
	r.registerAuthRoutes(mux)
	r.registerHierarchyRoutes(mux)
	r.registerQueueAndScrimRoutes(mux)
	r.registerResultsAndReplayRoutes(mux)
	r.registerExceptionRoutes(mux)
	r.registerSchedulingRoutes(mux)
	r.registerAdminMutationRoutes(mux)
	r.registerSelfServiceAndTokenRoutes(mux)
}
