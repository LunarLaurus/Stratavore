package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	dbURL := os.Getenv("STRATAVORE_DB_URL")
	if dbURL == "" {
		fmt.Println("STRATAVORE_DB_URL not set")
		os.Exit(1)
	}

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		fmt.Printf("Connect error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	fmt.Println("Applying fix: rank_tracking unique index...")

	// Create unique index
	_, err = conn.Exec(ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_rank_tracking_unique_event
		 ON rank_tracking(event_type, event_date, COALESCE(description, ''))`)

	if err != nil {
		fmt.Printf("✗ Failed to create index: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Index created successfully")

	// Verify
	var exists bool
	err = conn.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM pg_indexes
			WHERE tablename = 'rank_tracking'
			  AND indexname = 'idx_rank_tracking_unique_event'
			  AND schemaname = 'public'
		)`).Scan(&exists)

	if err != nil || !exists {
		fmt.Println("✗ Index verification failed")
		os.Exit(1)
	}

	fmt.Println("✓ Index verified present")
	fmt.Println("\nRank tracking idempotency constraint now active.")
}
