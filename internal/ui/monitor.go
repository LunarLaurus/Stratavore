package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/meridian-lex/stratavore/internal/storage"
	"github.com/meridian-lex/stratavore/pkg/types"
)

// LiveMonitor displays live runner status in terminal
type LiveMonitor struct {
	db       *storage.PostgresClient
	interval time.Duration
}

// NewLiveMonitor creates a new live monitor
func NewLiveMonitor(db *storage.PostgresClient, interval time.Duration) *LiveMonitor {
	return &LiveMonitor{
		db:       db,
		interval: interval,
	}
}

// Display shows live runner status with refresh
func (m *LiveMonitor) Display(ctx context.Context) error {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Clear screen and display initial
	fmt.Print("\033[2J\033[H")
	m.renderStatus(ctx)

	for {
		select {
		case <-ticker.C:
			// Move cursor to top and redraw
			fmt.Print("\033[H")
			m.renderStatus(ctx)
		case <-ctx.Done():
			return nil
		}
	}
}

func (m *LiveMonitor) renderStatus(ctx context.Context) {
	// Get all projects
	projects, err := m.db.ListProjects(ctx, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Header
	fmt.Println("═══════════════════════════════════════════════════════════════════════")
	fmt.Printf("  STRATAVORE LIVE MONITOR - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("═══════════════════════════════════════════════════════════════════════")
	fmt.Println()

	if len(projects) == 0 {
		fmt.Println("  No projects found.")
		fmt.Println()
		return
	}

	// Stats
	totalActiveRunners := 0
	totalSessions := 0
	var totalTokens int64

	for _, p := range projects {
		totalActiveRunners += p.ActiveRunners
		totalSessions += p.TotalSessions
		totalTokens += p.TotalTokens
	}

	fmt.Printf("  📊 Summary: %d Projects | %d Active Runners | %d Sessions | %s Tokens\n",
		len(projects), totalActiveRunners, totalSessions, formatNumber(totalTokens))
	fmt.Println()

	// Projects table
	fmt.Println("  PROJECT              STATUS    RUNNERS  SESSIONS  TOKENS")
	fmt.Println("  ─────────────────────────────────────────────────────────────────────")

	for _, p := range projects {
		statusIcon := getStatusIcon(p.Status)
		name := truncate(p.Name, 20)

		fmt.Printf("  %-20s %s %-7s  %2d       %4d      %s\n",
			name,
			statusIcon,
			p.Status,
			p.ActiveRunners,
			p.TotalSessions,
			formatNumber(p.TotalTokens))
	}

	fmt.Println()
	fmt.Println("  Press Ctrl+C to exit")
	fmt.Print("  ")
}

func getStatusIcon(status types.ProjectStatus) string {
	switch status {
	case types.ProjectActive:
		return "🟢"
	case types.ProjectIdle:
		return "⚪"
	case types.ProjectArchived:
		return "📦"
	default:
		return "⚫"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	} else if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	} else {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
}

// DisplayRunners shows detailed runner information
func (m *LiveMonitor) DisplayRunners(ctx context.Context, projectName string) error {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	fmt.Print("\033[2J\033[H")
	m.renderRunners(ctx, projectName)

	for {
		select {
		case <-ticker.C:
			fmt.Print("\033[H")
			m.renderRunners(ctx, projectName)
		case <-ctx.Done():
			return nil
		}
	}
}

func (m *LiveMonitor) renderRunners(ctx context.Context, projectName string) {
	var runners []*types.Runner
	var err error

	if projectName != "" {
		runners, err = m.db.GetActiveRunners(ctx, projectName)
	} else {
		// Get all runners (would need new query)
		runners = []*types.Runner{}
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Header
	fmt.Println("═══════════════════════════════════════════════════════════════════════")
	fmt.Printf("  ACTIVE RUNNERS - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("═══════════════════════════════════════════════════════════════════════")
	fmt.Println()

	if len(runners) == 0 {
		fmt.Println("  No active runners.")
		fmt.Println()
		return
	}

	fmt.Printf("  Total: %d active runners\n", len(runners))
	fmt.Println()

	// Runners table
	fmt.Println("  RUNNER    PROJECT          STATUS    UPTIME    CPU%   MEM(MB)  TOKENS")
	fmt.Println("  ─────────────────────────────────────────────────────────────────────")

	for _, r := range runners {
		id := truncate(r.ID, 8)
		project := truncate(r.ProjectName, 15)
		uptime := formatDuration(time.Since(r.StartedAt))

		fmt.Printf("  %-8s  %-15s  %-8s  %-8s  %5.1f  %7d  %s\n",
			id,
			project,
			r.Status,
			uptime,
			r.CPUPercent,
			r.MemoryMB,
			formatNumber(r.TokensUsed))
	}

	fmt.Println()
	fmt.Println("  Press Ctrl+C to exit")
	fmt.Print("  ")
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	} else if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
