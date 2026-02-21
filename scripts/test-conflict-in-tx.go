package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dbURL := os.Getenv("STRATAVORE_DB_URL")
	if dbURL == "" {
		fmt.Println("STRATAVORE_DB_URL not set")
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Printf("Connect error: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	fmt.Println("Testing ON CONFLICT within transaction...")
	fmt.Println()

	// Begin transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		fmt.Printf("Begin transaction error: %v\n", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	// Get first existing event from database
	var eventType, description, evidence, currentRank, progress string
	var strikes, commendations int
	var eventDate, metadata string

	err = tx.QueryRow(ctx, `
		SELECT event_type, event_date::text, description, evidence, current_rank, progress, strikes, commendations, metadata::text
		FROM rank_tracking
		ORDER BY id
		LIMIT 1
	`).Scan(&eventType, &eventDate, &description, &evidence, &currentRank, &progress, &strikes, &commendations, &metadata)

	if err != nil {
		fmt.Printf("Error fetching test data: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Test event:\n")
	fmt.Printf("  Type: %s\n", eventType)
	fmt.Printf("  Date: %s\n", eventDate)
	fmt.Printf("  Description: %s\n", description[:min(60, len(description))])
	fmt.Println()

	// Try to insert this same event WITH ON CONFLICT (inside transaction)
	query := `
		INSERT INTO rank_tracking (
			current_rank, progress, strikes, commendations,
			event_type, event_date, description, evidence, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (event_type, event_date, COALESCE(description, ''))
		DO NOTHING
	`

	fmt.Println("Attempting to re-insert WITH ON CONFLICT (inside transaction)...")
	result, err := tx.Exec(ctx, query,
		currentRank, progress, strikes, commendations,
		eventType, eventDate, description, evidence, metadata)

	if err != nil {
		fmt.Printf("✗ Insert failed: %v\n", err)
		fmt.Println("\nThis is the bug! ON CONFLICT doesn't work inside transaction.")
	} else {
		rows := result.RowsAffected()
		if rows == 0 {
			fmt.Printf("✓ Insert succeeded with 0 rows affected (conflict handled correctly)\n")
			fmt.Println("ON CONFLICT works fine inside transaction!")
		} else {
			fmt.Printf("✗ Insert succeeded but %d rows were inserted (should be 0)\n", rows)
		}
	}

	// Rollback (don't commit)
	tx.Rollback(ctx)
	fmt.Println("Transaction rolled back")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
