package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/apitoken"
	"github.com/sprocketbot/sprocket-v3/internal/domain/authz"
	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
	"github.com/sprocketbot/sprocket-v3/internal/domain/notifications"
	"github.com/sprocketbot/sprocket-v3/internal/domain/orgconfig"
	"github.com/sprocketbot/sprocket-v3/internal/domain/replaystats"
	"github.com/sprocketbot/sprocket-v3/internal/domain/skillgroup"
	"github.com/sprocketbot/sprocket-v3/internal/platform/auth"
	"github.com/sprocketbot/sprocket-v3/internal/platform/config"
)

type Server struct {
	logger *slog.Logger
	http   *http.Server
}

type Dependencies struct {
	TokenValidator     auth.TokenValidator
	Evaluator          authz.Evaluator
	LeagueStore        hierarchy.LeagueStore
	PlayerStore        hierarchy.PlayerStore
	RosterStore        hierarchy.RosterStore
	RoleStore          hierarchy.RoleStore
	QueueStore         hierarchy.QueueStore
	PlatformStore      hierarchy.PlatformAccountStore
	EligibilityStore   hierarchy.EligibilityStore
	ScrimStore         hierarchy.ScrimStore
	RatingStore        hierarchy.RatingStore
	ResultStore        hierarchy.ResultStore
	ReplayStore        hierarchy.ReplayStore
	ExceptionStore     hierarchy.ExceptionStore
	SchedulingStore    hierarchy.SchedulingStore
	HierarchyStore     hierarchy.Store
	SkillGroupStore    skillgroup.Store
	OrgConfigStore     orgconfig.Store
	NotificationsStore notifications.Store
	APITokenStore      apitoken.Store
	ReplayStatsStore   replaystats.Store
}

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

func New(cfg config.Config, logger *slog.Logger, deps Dependencies) *Server {
	mux := http.NewServeMux()
	routes := newRouteRegistrar(cfg, logger, deps)
	routes.register(mux)
	handler := corsMiddleware(routes.corsAllowedOrigins, mux)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return &Server{logger: logger, http: srv}
}
func (s *Server) Start() error {
	s.logger.Info("http server starting", "addr", s.http.Addr)
	return s.http.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("http server shutting down")
	return s.http.Shutdown(ctx)
}

func (s *Server) Handler() http.Handler {
	return s.http.Handler
}
