package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetStatusResponseJSON verifies that GetStatusResponse marshals to JSON
// with the expected camelCase top-level keys "daemon" and "metrics".
func TestGetStatusResponseJSON(t *testing.T) {
	resp := GetStatusResponse{
		Daemon: &DaemonStatus{
			DaemonID: "test-daemon",
			Hostname: "localhost",
			Version:  "1.0.0",
			Healthy:  true,
		},
		Metrics: &GlobalMetrics{
			ActiveRunners:  3,
			ActiveProjects: 2,
			TotalSessions:  10,
			TokensUsed:     5000,
			TokenLimit:     100000,
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Contains(t, raw, "daemon", "top-level key should be 'daemon' (camelCase)")
	assert.Contains(t, raw, "metrics", "top-level key should be 'metrics' (camelCase)")
	assert.NotContains(t, raw, "Daemon", "PascalCase key 'Daemon' must not appear")
	assert.NotContains(t, raw, "Metrics", "PascalCase key 'Metrics' must not appear")
}

// TestRunnerJSON verifies that Runner marshals to JSON with camelCase field
// names matching the struct tags.
func TestRunnerJSON(t *testing.T) {
	runner := Runner{
		ID:          "runner-abc123",
		RuntimeType: "process",
		ProjectName: "my-project",
		Status:      "running",
	}

	data, err := json.Marshal(runner)
	require.NoError(t, err)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Contains(t, raw, "runtimeType", "JSON key should be 'runtimeType'")
	assert.Contains(t, raw, "projectName", "JSON key should be 'projectName'")
	assert.NotContains(t, raw, "RuntimeType", "PascalCase 'RuntimeType' must not appear")
	assert.NotContains(t, raw, "ProjectName", "PascalCase 'ProjectName' must not appear")
}

// TestOmitEmpty verifies that an empty Error field is omitted from the JSON
// output of GetStatusResponse.
func TestOmitEmpty(t *testing.T) {
	resp := GetStatusResponse{
		Daemon: &DaemonStatus{
			DaemonID: "test-daemon",
			Healthy:  true,
		},
		Metrics: &GlobalMetrics{},
		// Error intentionally left empty
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))

	_, hasError := raw["error"]
	assert.False(t, hasError, "'error' key should be absent when Error field is empty")
}
