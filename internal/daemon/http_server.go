package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/meridian-lex/stratavore/internal/auth"
	"github.com/meridian-lex/stratavore/pkg/api"
	"github.com/meridian-lex/stratavore/pkg/config"
	"go.uber.org/zap"
)

// HTTPServer provides REST API for CLI communication
type HTTPServer struct {
	server  *http.Server
	handler *GRPCServer // Reuse gRPC handler logic
	logger  *zap.Logger
}

// NewHTTPServer creates HTTP API server.
// It wires JWT auth and per-client rate limiting when the corresponding
// config values are set; both default to disabled/permissive.
func NewHTTPServer(port int, handler *GRPCServer, logger *zap.Logger, cfg *config.SecurityConfig) *HTTPServer {
	mux := http.NewServeMux()

	httpServer := &HTTPServer{
		handler: handler,
		logger:  logger,
	}

	// Register routes
	mux.HandleFunc("/api/v1/runners/launch", httpServer.handleLaunchRunner)
	mux.HandleFunc("/api/v1/runners/stop", httpServer.handleStopRunner)
	mux.HandleFunc("/api/v1/runners/list", httpServer.handleListRunners)
	mux.HandleFunc("/api/v1/runners/get", httpServer.handleGetRunner)
	mux.HandleFunc("/api/v1/projects/create", httpServer.handleCreateProject)
	mux.HandleFunc("/api/v1/projects/list", httpServer.handleListProjects)
	mux.HandleFunc("/api/v1/heartbeat", httpServer.handleHeartbeat)
	mux.HandleFunc("/api/v1/status", httpServer.handleStatus)
	mux.HandleFunc("/api/v1/reconcile", httpServer.handleReconcile)
	mux.HandleFunc("/api/v1/health", httpServer.handleHealth)

	// Build middleware chain: rate-limit → JWT auth → mux
	var handler_ http.Handler = mux

	// JWT auth (disabled when auth_secret is empty)
	if cfg != nil {
		validator := auth.NewValidator(cfg.AuthSecret)
		if validator.Enabled() {
			logger.Info("HTTP API auth enabled")
		} else {
			logger.Info("HTTP API auth disabled (no auth_secret configured)")
		}
		handler_ = auth.Middleware(validator)(handler_)

		// Rate limiting (always active; defaults to 300 req/min, burst 50)
		ratePerMin := cfg.RateLimit.RequestsPerMinute
		if ratePerMin <= 0 {
			ratePerMin = 300
		}
		burst := cfg.RateLimit.Burst
		if burst <= 0 {
			burst = 50
		}
		rl := auth.NewRateLimiter(ratePerMin, time.Minute, burst)
		handler_ = auth.RateLimitMiddleware(rl)(handler_)

		logger.Info("HTTP API rate limiting enabled",
			zap.Int("requests_per_minute", ratePerMin),
			zap.Int("burst", burst))
	}

	httpServer.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler_,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return httpServer
}

// Start begins serving HTTP requests
func (s *HTTPServer) Start() error {
	s.logger.Info("HTTP API server starting", zap.String("addr", s.server.Addr))

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %w", err)
	}
	return nil
}

// Stop gracefully stops the server
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP API server")
	return s.server.Shutdown(ctx)
}

func (s *HTTPServer) handleLaunchRunner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.LaunchRunnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.handler.LaunchRunner(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, resp)
}

func (s *HTTPServer) handleStopRunner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.StopRunnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.handler.StopRunner(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, resp)
}

func (s *HTTPServer) handleListRunners(w http.ResponseWriter, r *http.Request) {
	projectName := r.URL.Query().Get("project")

	req := &api.ListRunnersRequest{
		ProjectName: projectName,
	}

	resp, err := s.handler.ListRunners(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, resp)
}

func (s *HTTPServer) handleGetRunner(w http.ResponseWriter, r *http.Request) {
	runnerID := r.URL.Query().Get("id")
	if runnerID == "" {
		http.Error(w, "runner_id required", http.StatusBadRequest)
		return
	}

	req := &api.GetRunnerRequest{RunnerID: runnerID}
	resp, err := s.handler.GetRunner(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, resp)
}

func (s *HTTPServer) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.handler.CreateProject(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, resp)
}

func (s *HTTPServer) handleListProjects(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	req := &api.ListProjectsRequest{Status: status}
	resp, err := s.handler.ListProjects(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, resp)
}

func (s *HTTPServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.handler.SendHeartbeat(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, resp)
}

func (s *HTTPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	req := &api.GetStatusRequest{}
	resp, err := s.handler.GetStatus(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, resp)
}

func (s *HTTPServer) handleReconcile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := &api.TriggerReconciliationRequest{}
	resp, err := s.handler.TriggerReconciliation(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, resp)
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *HTTPServer) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
