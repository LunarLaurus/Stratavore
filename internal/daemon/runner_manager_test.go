package daemon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestNewRunnerManager verifies that NewRunnerManager correctly sets the
// httpAPIURL field when constructed with port 8080.
func TestNewRunnerManager(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRunnerManager(nil, nil, logger, 8080)

	assert.Equal(t, "http://localhost:8080", rm.httpAPIURL,
		"httpAPIURL should be http://localhost:8080 for port 8080")
}

// TestNewRunnerManagerCustomPort verifies that NewRunnerManager sets the
// correct httpAPIURL when constructed with a non-default port.
func TestNewRunnerManagerCustomPort(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRunnerManager(nil, nil, logger, 9090)

	assert.Equal(t, "http://localhost:9090", rm.httpAPIURL,
		"httpAPIURL should be http://localhost:9090 for port 9090")
}
