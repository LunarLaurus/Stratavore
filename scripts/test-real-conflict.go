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

	fmt.Println("Testing ON CONFLICT with real rank_tracking data...")
	fmt.Println()

	// Get first existing event from database
	var eventType, description, evidence, currentRank, progress string
	var strikes, commendations int
	var eventDate, metadata string

	err = conn.QueryRow(ctx, `
		SELECT event_type, event_date::text, description, evidence, current_rank, progress, strikes, commendations, metadata::text
		FROM rank_tracking
		ORDER BY id
		LIMIT 1
	`).Scan(&eventType, &eventDate, &description, &evidence, &currentRank, &progress, &strikes, &commendations, &metadata)

	if err != nil {
		fmt.Printf("Error fetching test data: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Test event from database:\n")
	fmt.Printf("  Type: %s\n", eventType)
	fmt.Printf("  Date: %s\n", eventDate)
	fmt.Printf("  Description: %s\n", description)
	fmt.Println()

	// Try to insert this exact same event WITH ON CONFLICT
	query := `
		INSERT INTO rank_tracking (
			current_rank, progress, strikes, commendations,
			event_type, event_date, description, evidence, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (event_type, event_date, COALESCE(description, ''))
		DO NOTHING
	`

	fmt.Println("Attempting to re-insert this event WITH ON CONFLICT...")
	result, err := conn.Exec(ctx, query,
		currentRank, progress, strikes, commendations,
		eventType, eventDate, description, evidence, metadata)

	if err != nil {
		fmt.Printf("✗ Insert failed: %v\n", err)
		fmt.Println("\nThis explains why sync is failing!")
		fmt.Println("The ON CONFLICT clause is not working as expected.")
	} else {
		rows := result.RowsAffected()
		if rows == 0 {
			fmt.Printf("✓ Insert succeeded with 0 rows affected (conflict detected, DO NOTHING worked)\n")
			fmt.Println("ON CONFLICT is working correctly!")
		} else {
			fmt.Printf("✗ Insert succeeded but %d rows were inserted (should be 0)\n", rows)
		}
	}
}
