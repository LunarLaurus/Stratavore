package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/meridian/stratavore/internal/storage"
	"github.com/meridian/stratavore/internal/ui"
	"github.com/meridian/stratavore/pkg/config"
	"github.com/meridian/stratavore/pkg/types"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

var (
	flagsVar   string
	godMode    bool
	preset     string
	configFile string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "stratavore [project]",
	Short: "AI Development Workspace Orchestrator",
	Long: `Stratavore manages multiple Claude Code sessions across projects,
providing global state visibility, session resumption, and resource management.`,
	Version: fmt.Sprintf("%s (built %s, commit %s)", Version, BuildTime, Commit),
	Run:     rootHandler,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVar(&flagsVar, "flags", "", "Claude Code flags")
	rootCmd.PersistentFlags().BoolVar(&godMode, "god", false, "God mode (full access)")
	rootCmd.PersistentFlags().StringVar(&preset, "preset", "", "Use preset configuration")
	
	// Add subcommands
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(runnersCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(killCmd)
	rootCmd.AddCommand(daemonCmd)
}

func rootHandler(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		// Interactive launcher (TUI)
		fmt.Println("Interactive launcher not yet implemented")
		fmt.Println("Usage: stratavore <project-name>")
		os.Exit(1)
	}
	
	projectName := args[0]
	
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	
	// Connect to database
	ctx := context.Background()
	db, err := storage.NewPostgresClient(
		ctx,
		cfg.Database.PostgreSQL.GetConnectionString(),
		5, 1,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	
	// Check for existing runners
	runners, err := db.GetActiveRunners(ctx, projectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking runners: %v\n", err)
		os.Exit(1)
	}
	
	if len(runners) == 0 {
		// Launch new runner
		fmt.Printf("Launching new runner for project: %s\n", projectName)
		launchNewRunner(ctx, db, projectName, cfg)
	} else if len(runners) == 1 {
		// Offer attach or new
		fmt.Printf("Found 1 active runner for %s\n", projectName)
		fmt.Printf("  Runner ID: %s (started %v)\n", runners[0].ID, runners[0].StartedAt)
		fmt.Println("\nOptions:")
		fmt.Println("  1. Attach to existing runner")
		fmt.Println("  2. Launch new runner")
		// TODO: Interactive choice
		fmt.Println("\nAttaching to existing runner...")
	} else {
		// Show picker
		fmt.Printf("Found %d active runners for %s:\n", len(runners), projectName)
		for i, r := range runners {
			fmt.Printf("  %d. %s (started %v)\n", i+1, r.ID, r.StartedAt)
		}
		// TODO: Interactive picker
	}
}

func launchNewRunner(ctx context.Context, db *storage.PostgresClient, projectName string, cfg *config.Config) {
	// This would typically communicate with the daemon via gRPC
	// For now, just show what would happen
	
	fmt.Println("Would launch runner with daemon...")
	fmt.Printf("  Project: %s\n", projectName)
	if godMode {
		fmt.Println("  Mode: GOD MODE (full access)")
	}
	if preset != "" {
		fmt.Printf("  Preset: %s\n", preset)
	}
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show global dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== Stratavore Status ===")
		fmt.Println("Active Runners: 0")
		fmt.Println("Active Projects: 0")
		fmt.Println("Total Sessions: 0")
		fmt.Println("Tokens Used: 0")
		fmt.Println("\nDaemon: Not running (TODO: check via gRPC)")
	},
}

var runnersCmd = &cobra.Command{
	Use:   "runners",
	Short: "List all active runners",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.LoadConfig()
		ctx := context.Background()
		db, err := storage.NewPostgresClient(
			ctx,
			cfg.Database.PostgreSQL.GetConnectionString(),
			5, 1,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
			return
		}
		defer db.Close()
		
		// Get all active runners (would need to add this query)
		fmt.Println("=== Active Runners ===")
		fmt.Println("(Query implementation TODO)")
	},
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all projects",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.LoadConfig()
		ctx := context.Background()
		db, err := storage.NewPostgresClient(
			ctx,
			cfg.Database.PostgreSQL.GetConnectionString(),
			5, 1,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
			return
		}
		defer db.Close()
		
		projects, err := db.ListProjects(ctx, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing projects: %v\n", err)
			return
		}
		
		fmt.Println("=== Projects ===")
		for _, p := range projects {
			fmt.Printf("%s (%s) - %d active runners\n", p.Name, p.Status, p.ActiveRunners)
		}
	},
}

var newCmd = &cobra.Command{
	Use:   "new <project-name>",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		
		cfg, _ := config.LoadConfig()
		ctx := context.Background()
		db, err := storage.NewPostgresClient(
			ctx,
			cfg.Database.PostgreSQL.GetConnectionString(),
			5, 1,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
			return
		}
		defer db.Close()
		
		// Get current directory
		pwd, _ := os.Getwd()
		
		project := &types.Project{
			Name:   projectName,
			Path:   pwd,
			Status: types.ProjectIdle,
		}
		
		if err := db.CreateProject(ctx, project); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating project: %v\n", err)
			return
		}
		
		fmt.Printf("Created project: %s at %s\n", projectName, pwd)
	},
}

var attachCmd = &cobra.Command{
	Use:   "attach <runner-id>",
	Short: "Attach to running instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runnerID := args[0]
		fmt.Printf("Attaching to runner: %s\n", runnerID)
		fmt.Println("(Attach implementation TODO - requires PTY handling)")
	},
}

var killCmd = &cobra.Command{
	Use:   "kill <runner-id>",
	Short: "Stop a runner",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runnerID := args[0]
		fmt.Printf("Stopping runner: %s\n", runnerID)
		fmt.Println("(Would communicate with daemon via gRPC)")
	},
}

var watchCmd = &cobra.Command{
	Use:   "watch [project]",
	Short: "Live monitor of runners",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.LoadConfig()
		ctx := context.Background()
		db, err := storage.NewPostgresClient(
			ctx,
			cfg.Database.PostgreSQL.GetConnectionString(),
			5, 1,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
			return
		}
		defer db.Close()

		monitor := ui.NewLiveMonitor(db, 2*time.Second)
		
		// Setup signal handler
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			<-sigCh
			cancel()
		}()

		if len(args) > 0 {
			// Watch specific project runners
			monitor.DisplayRunners(ctx, args[0])
		} else {
			// Watch all projects
			monitor.Display(ctx)
		}
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon [start|stop|status]",
	Short: "Manage daemon",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]
		switch action {
		case "start":
			fmt.Println("Starting daemon...")
			fmt.Println("(Would start stratavored)")
		case "stop":
			fmt.Println("Stopping daemon...")
		case "status":
			fmt.Println("Daemon status: Unknown")
		default:
			fmt.Printf("Unknown action: %s\n", action)
		}
	},
}
