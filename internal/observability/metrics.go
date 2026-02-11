package observability

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/meridian/stratavore/pkg/types"
	"go.uber.org/zap"
)

// MetricsServer exposes Prometheus metrics
type MetricsServer struct {
	port   int
	logger *zap.Logger
	server *http.Server

	// Metrics state (would use prometheus client_golang in production)
	mu              sync.RWMutex
	runnersByStatus map[types.RunnerStatus]int
	runnersByProject map[string]int
	totalSessions   int
	tokensUsed      int64
	heartbeatLatencies []float64
	daemonUptime    float64
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int, logger *zap.Logger) *MetricsServer {
	return &MetricsServer{
		port:             port,
		logger:           logger,
		runnersByStatus:  make(map[types.RunnerStatus]int),
		runnersByProject: make(map[string]int),
		heartbeatLatencies: []float64{},
	}
}

// Start begins serving metrics
func (m *MetricsServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", m.handleMetrics)
	mux.HandleFunc("/health", m.handleHealth)

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.port),
		Handler: mux,
	}

	m.logger.Info("metrics server starting", zap.Int("port", m.port))

	if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("metrics server error: %w", err)
	}

	return nil
}

// Stop gracefully stops the server
func (m *MetricsServer) Stop() error {
	if m.server != nil {
		return m.server.Close()
	}
	return nil
}

// handleMetrics serves Prometheus metrics in text format
func (m *MetricsServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Write metrics in Prometheus format
	// In production, use prometheus/client_golang

	// Runner metrics by status
	for status, count := range m.runnersByStatus {
		fmt.Fprintf(w, "stratavore_runners_total{status=\"%s\"} %d\n", status, count)
	}

	// Runner metrics by project
	for project, count := range m.runnersByProject {
		fmt.Fprintf(w, "stratavore_runners_by_project{project=\"%s\"} %d\n", project, count)
	}

	// Session metrics
	fmt.Fprintf(w, "stratavore_sessions_total %d\n", m.totalSessions)

	// Token metrics
	fmt.Fprintf(w, "stratavore_tokens_used_total{scope=\"global\"} %d\n", m.tokensUsed)

	// Daemon uptime
	fmt.Fprintf(w, "stratavore_daemon_uptime_seconds %f\n", m.daemonUptime)

	// Heartbeat latency histogram (simplified)
	if len(m.heartbeatLatencies) > 0 {
		sum := 0.0
		for _, lat := range m.heartbeatLatencies {
			sum += lat
		}
		avg := sum / float64(len(m.heartbeatLatencies))
		fmt.Fprintf(w, "stratavore_heartbeat_latency_seconds_sum %f\n", sum)
		fmt.Fprintf(w, "stratavore_heartbeat_latency_seconds_count %d\n", len(m.heartbeatLatencies))
		fmt.Fprintf(w, "stratavore_heartbeat_latency_seconds_avg %f\n", avg)
	}
}

// handleHealth serves health check endpoint
func (m *MetricsServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// UpdateRunnerMetrics updates runner counts
func (m *MetricsServer) UpdateRunnerMetrics(runners []*types.Runner) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset counters
	m.runnersByStatus = make(map[types.RunnerStatus]int)
	m.runnersByProject = make(map[string]int)

	// Count runners
	for _, r := range runners {
		m.runnersByStatus[r.Status]++
		m.runnersByProject[r.ProjectName]++
	}
}

// RecordTokenUsage records token usage
func (m *MetricsServer) RecordTokenUsage(tokens int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokensUsed += tokens
}

// RecordHeartbeatLatency records heartbeat processing time
func (m *MetricsServer) RecordHeartbeatLatency(latencySeconds float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.heartbeatLatencies = append(m.heartbeatLatencies, latencySeconds)

	// Keep only last 1000 measurements
	if len(m.heartbeatLatencies) > 1000 {
		m.heartbeatLatencies = m.heartbeatLatencies[len(m.heartbeatLatencies)-1000:]
	}
}

// UpdateDaemonUptime updates daemon uptime metric
func (m *MetricsServer) UpdateDaemonUptime(seconds float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.daemonUptime = seconds
}

// IncrementSessions increments total sessions counter
func (m *MetricsServer) IncrementSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalSessions++
}
