package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/authz"
	"github.com/sprocketbot/sprocket-v3/internal/platform/auth"
	"github.com/sprocketbot/sprocket-v3/internal/platform/config"
)

type Server struct {
	logger *slog.Logger
	http   *http.Server
}

type Dependencies struct {
	TokenValidator auth.TokenValidator
	Evaluator      authz.Evaluator
}

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

func New(cfg config.Config, logger *slog.Logger, deps Dependencies) *Server {
	mux := http.NewServeMux()
	tokenValidator := deps.TokenValidator
	if tokenValidator == nil {
		defaultValidator := auth.NewLocalTokenValidator("local")
		tokenValidator = defaultValidator
	}
	evaluator := deps.Evaluator
	if evaluator == nil {
		defaultEvaluator := authz.NewStaticEvaluator(authz.DefaultPermissions())
		evaluator = defaultEvaluator
	}

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			Status:    "ok",
			Service:   "sprocket-v3-api",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			Status:    "ready",
			Service:   "sprocket-v3-api",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	})

	adminPingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal := auth.PrincipalFromContext(r.Context())
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"subject": principal.Subject,
		})
	})
	mux.Handle("/v1/admin/ping", auth.Authentication(
		tokenValidator,
		auth.RequirePermission(evaluator, "admin.ping", "read", adminPingHandler),
	))

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
