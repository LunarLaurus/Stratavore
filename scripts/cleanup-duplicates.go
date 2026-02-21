//go:build ignore

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

	fmt.Println("Cleaning up duplicate rank_tracking entries...")
	fmt.Println("Strategy: Keep oldest entry (min created_at) for each unique event")
	fmt.Println()

	// Use CTE to identify duplicates and delete all but the oldest
	deleteQuery := `
		WITH duplicates AS (
			SELECT
				id,
				ROW_NUMBER() OVER (
					PARTITION BY event_type, event_date, COALESCE(description, '')
					ORDER BY created_at ASC
				) as rn
			FROM rank_tracking
		)
		DELETE FROM rank_tracking
		WHERE id IN (
			SELECT id FROM duplicates WHERE rn > 1
		)
	`

	// Execute in transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		fmt.Printf("Begin transaction error: %v\n", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	// Get count before cleanup
	var beforeCount int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM rank_tracking`).Scan(&beforeCount)
	if err != nil {
		fmt.Printf("Count error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Before cleanup: %d total rank_tracking rows\n", beforeCount)

	// Delete duplicates
	result, err := tx.Exec(ctx, deleteQuery)
	if err != nil {
		fmt.Printf("Delete error: %v\n", err)
		os.Exit(1)
	}

	deleted := result.RowsAffected()
	fmt.Printf("Deleted: %d duplicate rows\n", deleted)

	// Get count after cleanup
	var afterCount int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM rank_tracking`).Scan(&afterCount)
	if err != nil {
		fmt.Printf("Count error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("After cleanup: %d rows remaining\n", afterCount)
	fmt.Printf("Reduction: %.1f%%\n", float64(deleted)/float64(beforeCount)*100)

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		fmt.Printf("Commit error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Cleanup complete")
	fmt.Println("Next step: Apply unique index with scripts/apply-fix.go")
}
