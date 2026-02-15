package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetDefaults verifies that LoadConfig returns sensible defaults when no
// config file or environment overrides are present.
func TestSetDefaults(t *testing.T) {
	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 50051, cfg.Daemon.Port_GRPC, "default grpc_port should be 50051")
	assert.Equal(t, 8080, cfg.Daemon.Port_HTTP, "default http_port should be 8080")
	assert.NotEmpty(t, cfg.Database.PostgreSQL.Host, "daemon id / host should be non-empty")
}

// TestEnvOverride verifies that STRATAVORE_DAEMON_HTTP_PORT overrides the
// default http_port value.
func TestEnvOverride(t *testing.T) {
	t.Setenv("STRATAVORE_DAEMON_HTTP_PORT", "9999")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 9999, cfg.Daemon.Port_HTTP, "http_port should be overridden to 9999 via env")
}

// TestConnectionString verifies that GetConnectionString produces a URL
// containing the configured host, database name, and user.
func TestConnectionString(t *testing.T) {
	pgCfg := &PostgreSQLConfig{
		Host:     "db.example.com",
		Port:     5432,
		Database: "mydb",
		User:     "myuser",
		Password: "secret",
		SSLMode:  "disable",
	}

	connStr := pgCfg.GetConnectionString()

	assert.Contains(t, connStr, "db.example.com", "connection string should contain host")
	assert.Contains(t, connStr, "mydb", "connection string should contain database name")
	assert.Contains(t, connStr, "myuser", "connection string should contain user")

	// Verify overall format
	assert.Contains(t, connStr, "postgres://")
}

// TestEnvOverrideCleanup verifies cleanup via t.Setenv (automatic) leaves no
// residual state for subsequent tests.
func TestEnvOverrideCleanup(t *testing.T) {
	// Confirm the env var is not set in this test (relies on t.Setenv cleanup).
	_, set := os.LookupEnv("STRATAVORE_DAEMON_HTTP_PORT")
	assert.False(t, set, "env var should not be set outside TestEnvOverride")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Daemon.Port_HTTP, "http_port should revert to default 8080")
}
