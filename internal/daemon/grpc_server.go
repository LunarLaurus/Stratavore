package daemon

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/meridian/stratavore/internal/storage"
	"github.com/meridian/stratavore/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer implements the Stratavore gRPC API
type GRPCServer struct {
	// UnimplementedStratavoreServiceServer
	server        *grpc.Server
	runnerManager *RunnerManager
	storage       *storage.PostgresClient
	logger        *zap.Logger
	port          int
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(
	port int,
	runnerManager *RunnerManager,
	storage *storage.PostgresClient,
	logger *zap.Logger,
) *GRPCServer {
	return &GRPCServer{
		server:        grpc.NewServer(),
		runnerManager: runnerManager,
		storage:       storage,
		logger:        logger,
		port:          port,
	}
}

// Start begins serving gRPC requests
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Register service
	// pb.RegisterStratavoreServiceServer(s.server, s)

	s.logger.Info("gRPC server starting", zap.Int("port", s.port))

	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the server
func (s *GRPCServer) Stop() {
	s.logger.Info("stopping gRPC server")
	s.server.GracefulStop()
}

// LaunchRunner handles runner launch requests
func (s *GRPCServer) LaunchRunner(ctx context.Context, req *LaunchRunnerRequest) (*LaunchRunnerResponse, error) {
	s.logger.Info("launch runner request",
		zap.String("project", req.ProjectName),
		zap.String("runtime", req.RuntimeType))

	// Convert to internal request
	launchReq := &types.LaunchRequest{
		ProjectName:      req.ProjectName,
		ProjectPath:      req.ProjectPath,
		Flags:            req.Flags,
		Capabilities:     req.Capabilities,
		Environment:      req.Environment,
		ConversationMode: types.ConversationMode(req.ConversationMode),
		SessionID:        req.SessionId,
		RuntimeType:      types.RuntimeType(req.RuntimeType),
	}

	// Launch runner
	runner, err := s.runnerManager.Launch(ctx, launchReq)
	if err != nil {
		s.logger.Error("failed to launch runner", zap.Error(err))
		return &LaunchRunnerResponse{
			Error: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	return &LaunchRunnerResponse{
		Runner: convertRunnerToProto(runner),
	}, nil
}

// StopRunner handles runner stop requests
func (s *GRPCServer) StopRunner(ctx context.Context, req *StopRunnerRequest) (*StopRunnerResponse, error) {
	s.logger.Info("stop runner request",
		zap.String("runner_id", req.RunnerId),
		zap.Bool("force", req.Force))

	err := s.runnerManager.StopRunner(ctx, req.RunnerId)
	if err != nil {
		s.logger.Error("failed to stop runner", zap.Error(err))
		return &StopRunnerResponse{
			Success: false,
			Error:   err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	return &StopRunnerResponse{
		Success: true,
	}, nil
}

// GetRunner retrieves runner details
func (s *GRPCServer) GetRunner(ctx context.Context, req *GetRunnerRequest) (*GetRunnerResponse, error) {
	runner, err := s.storage.GetRunner(ctx, req.RunnerId)
	if err != nil {
		return &GetRunnerResponse{
			Error: err.Error(),
		}, status.Error(codes.NotFound, err.Error())
	}

	return &GetRunnerResponse{
		Runner: convertRunnerToProto(runner),
	}, nil
}

// ListRunners lists active runners
func (s *GRPCServer) ListRunners(ctx context.Context, req *ListRunnersRequest) (*ListRunnersResponse, error) {
	var runners []*types.Runner
	var err error

	if req.ProjectName != "" {
		runners, err = s.storage.GetActiveRunners(ctx, req.ProjectName)
	} else {
		// Get all active runners (would need new query)
		runners = s.runnerManager.GetActiveRunners()
	}

	if err != nil {
		return &ListRunnersResponse{
			Error: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	protoRunners := make([]*Runner, len(runners))
	for i, r := range runners {
		protoRunners[i] = convertRunnerToProto(r)
	}

	return &ListRunnersResponse{
		Runners: protoRunners,
		Total:   int32(len(runners)),
	}, nil
}

// CreateProject creates a new project
func (s *GRPCServer) CreateProject(ctx context.Context, req *CreateProjectRequest) (*CreateProjectResponse, error) {
	project := &types.Project{
		Name:        req.Name,
		Path:        req.Path,
		Description: req.Description,
		Tags:        req.Tags,
		Status:      types.ProjectIdle,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.storage.CreateProject(ctx, project)
	if err != nil {
		return &CreateProjectResponse{
			Error: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	return &CreateProjectResponse{
		Project: convertProjectToProto(project),
	}, nil
}

// GetProject retrieves project details
func (s *GRPCServer) GetProject(ctx context.Context, req *GetProjectRequest) (*GetProjectResponse, error) {
	project, err := s.storage.GetProject(ctx, req.Name)
	if err != nil {
		return &GetProjectResponse{
			Error: err.Error(),
		}, status.Error(codes.NotFound, err.Error())
	}

	return &GetProjectResponse{
		Project: convertProjectToProto(project),
	}, nil
}

// ListProjects lists all projects
func (s *GRPCServer) ListProjects(ctx context.Context, req *ListProjectsRequest) (*ListProjectsResponse, error) {
	projects, err := s.storage.ListProjects(ctx, req.Status)
	if err != nil {
		return &ListProjectsResponse{
			Error: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	protoProjects := make([]*Project, len(projects))
	for i, p := range projects {
		protoProjects[i] = convertProjectToProto(p)
	}

	return &ListProjectsResponse{
		Projects: protoProjects,
	}, nil
}

// SendHeartbeat processes heartbeat from agent
func (s *GRPCServer) SendHeartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error) {
	hb := &types.Heartbeat{
		RunnerID:     req.RunnerId,
		Status:       types.RunnerStatus(req.Status),
		Timestamp:    time.Now(),
		CPUPercent:   req.CpuPercent,
		MemoryMB:     req.MemoryMb,
		TokensUsed:   req.TokensUsed,
		SessionID:    req.SessionId,
		AgentVersion: req.AgentVersion,
		Hostname:     req.Hostname,
	}

	err := s.runnerManager.ProcessHeartbeat(ctx, hb)
	if err != nil {
		s.logger.Error("heartbeat processing failed",
			zap.String("runner_id", req.RunnerId),
			zap.Error(err))
		return &HeartbeatResponse{
			Success: false,
			Error:   err.Error(),
		}, nil // Don't fail heartbeat on error
	}

	return &HeartbeatResponse{
		Success: true,
	}, nil
}

// GetStatus returns daemon status
func (s *GRPCServer) GetStatus(ctx context.Context, req *GetStatusRequest) (*GetStatusResponse, error) {
	// Get daemon info
	// Get metrics

	activeRunners := len(s.runnerManager.GetActiveRunners())

	metrics := &GlobalMetrics{
		ActiveRunners: int32(activeRunners),
		// Would need to query database for other metrics
	}

	daemonStatus := &DaemonStatus{
		Healthy:       true,
		LastHeartbeat: time.Now().Format(time.RFC3339),
	}

	return &GetStatusResponse{
		Daemon:  daemonStatus,
		Metrics: metrics,
	}, nil
}

// TriggerReconciliation manually triggers stale runner cleanup
func (s *GRPCServer) TriggerReconciliation(ctx context.Context, req *TriggerReconciliationRequest) (*TriggerReconciliationResponse, error) {
	s.logger.Info("manual reconciliation triggered")

	err := s.runnerManager.ReconcileRunners(ctx)
	if err != nil {
		return &TriggerReconciliationResponse{
			Error: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	return &TriggerReconciliationResponse{
		ReconciledCount: 0, // Would need to return count from reconciliation
	}, nil
}

// Helper functions to convert between types

func convertRunnerToProto(r *types.Runner) *Runner {
	proto := &Runner{
		Id:                  r.ID,
		RuntimeType:         string(r.RuntimeType),
		RuntimeId:           r.RuntimeID,
		NodeId:              r.NodeID,
		ProjectName:         r.ProjectName,
		ProjectPath:         r.ProjectPath,
		Status:              string(r.Status),
		Flags:               r.Flags,
		Capabilities:        r.Capabilities,
		Environment:         r.Environment,
		SessionId:           r.SessionID,
		ConversationMode:    string(r.ConversationMode),
		TokensUsed:          r.TokensUsed,
		CpuPercent:          r.CPUPercent,
		MemoryMb:            r.MemoryMB,
		RestartAttempts:     int32(r.RestartAttempts),
		MaxRestartAttempts:  int32(r.MaxRestartAttempts),
		StartedAt:           r.StartedAt.Format(time.RFC3339),
		HeartbeatTtlSeconds: int32(r.HeartbeatTTL),
		CreatedAt:           r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           r.UpdatedAt.Format(time.RFC3339),
	}

	if r.LastHeartbeat != nil {
		proto.LastHeartbeat = r.LastHeartbeat.Format(time.RFC3339)
	}
	if r.TerminatedAt != nil {
		proto.TerminatedAt = r.TerminatedAt.Format(time.RFC3339)
	}
	if r.ExitCode != nil {
		proto.ExitCode = int32(*r.ExitCode)
	}

	return proto
}

func convertProjectToProto(p *types.Project) *Project {
	proto := &Project{
		Name:          p.Name,
		Path:          p.Path,
		Status:        string(p.Status),
		Description:   p.Description,
		Tags:          p.Tags,
		TotalRunners:  int32(p.TotalRunners),
		ActiveRunners: int32(p.ActiveRunners),
		TotalSessions: int32(p.TotalSessions),
		TotalTokens:   p.TotalTokens,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}

	if p.LastAccessedAt != nil {
		proto.LastAccessedAt = p.LastAccessedAt.Format(time.RFC3339)
	}
	if p.ArchivedAt != nil {
		proto.ArchivedAt = p.ArchivedAt.Format(time.RFC3339)
	}

	return proto
}

// Placeholder types (would be generated from proto)
type (
	LaunchRunnerRequest           struct{ ProjectName, ProjectPath, RuntimeType, ConversationMode, SessionId string; Flags, Capabilities []string; Environment map[string]string }
	LaunchRunnerResponse          struct{ Runner *Runner; Error string }
	StopRunnerRequest             struct{ RunnerId string; Force bool; TimeoutSeconds int32 }
	StopRunnerResponse            struct{ Success bool; Error string }
	GetRunnerRequest              struct{ RunnerId string }
	GetRunnerResponse             struct{ Runner *Runner; Error string }
	ListRunnersRequest            struct{ ProjectName, Status string; Limit, Offset int32 }
	ListRunnersResponse           struct{ Runners []*Runner; Total int32; Error string }
	CreateProjectRequest          struct{ Name, Path, Description string; Tags []string }
	CreateProjectResponse         struct{ Project *Project; Error string }
	GetProjectRequest             struct{ Name string }
	GetProjectResponse            struct{ Project *Project; Error string }
	ListProjectsRequest           struct{ Status string }
	ListProjectsResponse          struct{ Projects []*Project; Error string }
	HeartbeatRequest              struct{ RunnerId, Status, SessionId, AgentVersion, Hostname string; CpuPercent float64; MemoryMb, TokensUsed int64 }
	HeartbeatResponse             struct{ Success bool; Command, Error string }
	GetStatusRequest              struct{}
	GetStatusResponse             struct{ Daemon *DaemonStatus; Metrics *GlobalMetrics; Error string }
	TriggerReconciliationRequest  struct{}
	TriggerReconciliationResponse struct{ ReconciledCount int32; FailedRunnerIds []string; Error string }

	Runner        struct{ Id, RuntimeType, RuntimeId, NodeId, ProjectName, ProjectPath, Status, SessionId, ConversationMode, StartedAt, LastHeartbeat, TerminatedAt, CreatedAt, UpdatedAt string; Flags, Capabilities []string; Environment map[string]string; TokensUsed, MemoryMb int64; CpuPercent float64; RestartAttempts, MaxRestartAttempts, HeartbeatTtlSeconds, ExitCode int32 }
	Project       struct{ Name, Path, Status, Description, CreatedAt, LastAccessedAt, ArchivedAt, UpdatedAt string; Tags []string; TotalRunners, ActiveRunners, TotalSessions int32; TotalTokens int64 }
	DaemonStatus  struct{ DaemonId, Hostname, Version, StartedAt, LastHeartbeat string; Healthy bool }
	GlobalMetrics struct{ ActiveRunners, ActiveProjects, TotalSessions int32; TokensUsed, TokenLimit int64 }
)
