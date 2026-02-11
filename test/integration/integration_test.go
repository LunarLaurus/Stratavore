package integration

import (
	"context"
	"testing"
	"time"

	"github.com/meridian/stratavore/internal/storage"
	"github.com/meridian/stratavore/pkg/api"
	"github.com/meridian/stratavore/pkg/client"
	"github.com/meridian/stratavore/pkg/config"
	"github.com/meridian/stratavore/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDaemonStartup tests that daemon starts and API is reachable
func TestDaemonStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	apiClient := client.NewClient("localhost", 50051)

	// Ping daemon
	err := apiClient.Ping(ctx)
	require.NoError(t, err, "daemon should be reachable")
}

// TestProjectLifecycle tests complete project workflow
func TestProjectLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	apiClient := client.NewClient("localhost", 50051)

	// Create project
	req := &api.CreateProjectRequest{
		Name:        "test-project-" + time.Now().Format("20060102150405"),
		Path:        "/tmp/test-project",
		Description: "Integration test project",
		Tags:        []string{"test", "integration"},
	}

	resp, err := apiClient.CreateProject(ctx, req)
	require.NoError(t, err)
	assert.Empty(t, resp.Error)
	assert.NotNil(t, resp.Project)
	assert.Equal(t, req.Name, resp.Project.Name)

	// List projects
	listResp, err := apiClient.ListProjects(ctx, "")
	require.NoError(t, err)
	assert.NotEmpty(t, listResp.Projects)

	// Find our project
	found := false
	for _, p := range listResp.Projects {
		if p.Name == req.Name {
			found = true
			assert.Equal(t, "idle", p.Status)
			break
		}
	}
	assert.True(t, found, "created project should be in list")
}

// TestRunnerLifecycle tests runner creation and management
func TestRunnerLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	apiClient := client.NewClient("localhost", 50051)

	// Create test project first
	projectName := "test-runner-" + time.Now().Format("20060102150405")
	projReq := &api.CreateProjectRequest{
		Name: projectName,
		Path: "/tmp/" + projectName,
	}
	_, err := apiClient.CreateProject(ctx, projReq)
	require.NoError(t, err)

	// Launch runner
	launchReq := &api.LaunchRunnerRequest{
		ProjectName:      projectName,
		ProjectPath:      "/tmp/" + projectName,
		RuntimeType:      "process",
		ConversationMode: "new",
	}

	launchResp, err := apiClient.LaunchRunner(ctx, launchReq)
	// Note: This may fail if Claude Code not installed
	if err != nil {
		t.Skipf("skipping runner test: %v", err)
		return
	}

	assert.Empty(t, launchResp.Error)
	assert.NotNil(t, launchResp.Runner)
	runnerID := launchResp.Runner.ID

	// Wait for runner to start
	time.Sleep(2 * time.Second)

	// Get runner details
	getResp, err := apiClient.GetRunner(ctx, runnerID)
	require.NoError(t, err)
	assert.NotNil(t, getResp.Runner)

	// List runners
	listResp, err := apiClient.ListRunners(ctx, projectName)
	require.NoError(t, err)
	assert.NotEmpty(t, listResp.Runners)

	// Stop runner
	stopResp, err := apiClient.StopRunner(ctx, runnerID, false)
	require.NoError(t, err)
	assert.True(t, stopResp.Success)

	// Wait for cleanup
	time.Sleep(1 * time.Second)

	// Verify runner stopped
	getResp2, err := apiClient.GetRunner(ctx, runnerID)
	require.NoError(t, err)
	assert.Contains(t, []string{"stopped", "failed", "terminated"}, getResp2.Runner.Status)
}

// TestDaemonStatus tests status endpoint
func TestDaemonStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	apiClient := client.NewClient("localhost", 50051)

	resp, err := apiClient.GetStatus(ctx)
	require.NoError(t, err)
	assert.Empty(t, resp.Error)
	assert.NotNil(t, resp.Daemon)
	assert.True(t, resp.Daemon.Healthy)
	assert.NotNil(t, resp.Metrics)
}

// TestTokenBudget tests budget enforcement
func TestTokenBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	cfg, _ := config.LoadConfig()

	// Connect to database directly
	db, err := storage.NewPostgresClient(
		ctx,
		cfg.Database.PostgreSQL.GetConnectionString(),
		5, 1,
	)
	require.NoError(t, err)
	defer db.Close()

	// Create test budget
	projectName := "budget-test-" + time.Now().Format("20060102150405")
	budget := &types.TokenBudget{
		Scope:             "project",
		ScopeID:           projectName,
		LimitTokens:       1000,
		UsedTokens:        0,
		PeriodGranularity: "daily",
		PeriodStart:       time.Now(),
		PeriodEnd:         time.Now().Add(24 * time.Hour),
	}

	err = db.CreateTokenBudget(ctx, budget)
	require.NoError(t, err)

	// Retrieve budget
	retrieved, err := db.GetTokenBudget(ctx, "project", projectName)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, int64(1000), retrieved.LimitTokens)
	assert.Equal(t, int64(0), retrieved.UsedTokens)

	// Increment usage
	err = db.IncrementTokenUsage(ctx, "project", projectName, 500)
	require.NoError(t, err)

	// Verify increment
	updated, err := db.GetTokenBudget(ctx, "project", projectName)
	require.NoError(t, err)
	assert.Equal(t, int64(500), updated.UsedTokens)
}

// TestReconciliation tests stale runner cleanup
func TestReconciliation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	apiClient := client.NewClient("localhost", 50051)

	// Trigger reconciliation
	resp, err := apiClient.TriggerReconciliation(ctx)
	require.NoError(t, err)
	assert.Empty(t, resp.Error)
}

// BenchmarkAPILatency benchmarks API call latency
func BenchmarkAPILatency(b *testing.B) {
	ctx := context.Background()
	apiClient := client.NewClient("localhost", 50051)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		apiClient.GetStatus(ctx)
	}
}

// BenchmarkDatabaseQuery benchmarks database query performance
func BenchmarkDatabaseQuery(b *testing.B) {
	ctx := context.Background()
	cfg, _ := config.LoadConfig()

	db, err := storage.NewPostgresClient(
		ctx,
		cfg.Database.PostgreSQL.GetConnectionString(),
		5, 1,
	)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.ListProjects(ctx, "")
	}
}
