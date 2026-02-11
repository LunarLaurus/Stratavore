package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/meridian/stratavore/pkg/types"
	"go.uber.org/zap"
)

var (
	runnerID    string
	projectName string
	projectPath string
	claudeFlags []string
)

func main() {
	// Parse flags
	flag.StringVar(&runnerID, "runner-id", "", "Runner ID")
	flag.StringVar(&projectName, "project-name", "", "Project name")
	flag.StringVar(&projectPath, "project-path", "", "Project path")
	flag.Parse()
	
	if runnerID == "" || projectName == "" || projectPath == "" {
		fmt.Fprintf(os.Stderr, "Missing required flags\n")
		os.Exit(1)
	}
	
	// Setup logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	
	logger.Info("stratavore-agent starting",
		zap.String("runner_id", runnerID),
		zap.String("project_name", projectName),
		zap.String("project_path", projectPath))
	
	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Start heartbeat goroutine
	go sendHeartbeats(ctx, runnerID, logger)
	
	// Build Claude Code command
	args := []string{"--project", projectPath}
	
	// Add custom flags
	for _, f := range claudeFlags {
		args = append(args, f)
	}
	
	// Start Claude Code
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = projectPath
	
	logger.Info("starting claude code", zap.Strings("args", args))
	
	if err := cmd.Start(); err != nil {
		logger.Error("failed to start claude code", zap.Error(err))
		os.Exit(1)
	}
	
	pid := cmd.Process.Pid
	logger.Info("claude code started", zap.Int("pid", pid))
	
	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// Wait for process or signal
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
	}()
	
	select {
	case err := <-errCh:
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = -1
			}
		}
		logger.Info("claude code exited",
			zap.Int("exit_code", exitCode))
		os.Exit(exitCode)
		
	case sig := <-sigCh:
		logger.Info("received signal, terminating",
			zap.String("signal", sig.String()))
		
		// Forward signal to Claude Code
		cmd.Process.Signal(sig)
		
		// Wait with timeout
		select {
		case <-errCh:
			logger.Info("claude code terminated gracefully")
		case <-time.After(10 * time.Second):
			logger.Warn("claude code did not exit, killing")
			cmd.Process.Kill()
		}
	}
}

func sendHeartbeats(ctx context.Context, runnerID string, logger *zap.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Send heartbeat to daemon
			hb := &types.Heartbeat{
				RunnerID:     runnerID,
				Status:       types.StatusRunning,
				Timestamp:    time.Now(),
				AgentVersion: "0.1.0",
			}
			
			// TODO: Send via gRPC to daemon
			logger.Debug("heartbeat", zap.String("runner_id", runnerID))
			_ = hb
			
		case <-ctx.Done():
			return
		}
	}
}
