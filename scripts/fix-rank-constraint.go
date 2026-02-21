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

	fmt.Println("Fixing rank_tracking idempotency constraint...")
	fmt.Println()

	// Drop the index (can't use ON CONFLICT with functional index)
	fmt.Println("1. Dropping functional index idx_rank_tracking_unique_event...")
	_, err = conn.Exec(ctx, `DROP INDEX IF EXISTS idx_rank_tracking_unique_event`)
	if err != nil {
		fmt.Printf("✗ Failed to drop index: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Index dropped")

	// Create unique constraint (can use ON CONFLICT with this)
	// Note: We need a generated column or change the approach
	// PostgreSQL doesn't allow UNIQUE constraints with expressions directly
	// So we'll create an exclusion constraint or modify the import code

	// Actually, let's use a different approach: add a computed column
	fmt.Println("2. Adding computed column for idempotency...")
	_, err = conn.Exec(ctx, `
		ALTER TABLE rank_tracking
		ADD COLUMN IF NOT EXISTS event_key TEXT
		GENERATED ALWAYS AS (event_type || '::' || event_date::text || '::' || COALESCE(description, '')) STORED
	`)
	if err != nil {
		fmt.Printf("✗ Failed to add computed column: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Computed column added")

	// Create unique constraint on the computed column
	fmt.Println("3. Creating unique constraint...")
	_, err = conn.Exec(ctx, `
		ALTER TABLE rank_tracking
		ADD CONSTRAINT rank_tracking_event_key_unique
		UNIQUE (event_key)
	`)
	if err != nil {
		fmt.Printf("✗ Failed to create constraint: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Unique constraint created")

	// Verify
	var exists bool
	err = conn.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM information_schema.table_constraints
			WHERE table_name = 'rank_tracking'
			  AND constraint_name = 'rank_tracking_event_key_unique'
			  AND constraint_type = 'UNIQUE'
		)`).Scan(&exists)

	if err != nil || !exists {
		fmt.Println("✗ Constraint verification failed")
		os.Exit(1)
	}

	fmt.Println("✓ Constraint verified")
	fmt.Println()
	fmt.Println("Idempotency constraint fixed!")
	fmt.Println("Sync code will now need to use: ON CONFLICT (event_key) DO NOTHING")
}
