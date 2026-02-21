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

	fmt.Println("Finding duplicate rank_tracking entries...")
	fmt.Println()

	// Find duplicates
	rows, err := conn.Query(ctx,
		`SELECT
			event_type,
			event_date::text,
			COALESCE(description, '') as description,
			COUNT(*) as count,
			array_agg(id ORDER BY created_at) as ids,
			array_agg(created_at::text ORDER BY created_at) as created_ats
		FROM rank_tracking
		GROUP BY event_type, event_date, COALESCE(description, '')
		HAVING COUNT(*) > 1
		ORDER BY event_date DESC`)

	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	totalDuplicates := 0
	for rows.Next() {
		var eventType, description, eventDate string
		var count int
		var ids []int
		var createdAts []string

		if err := rows.Scan(&eventType, &eventDate, &description, &count, &ids, &createdAts); err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}

		totalDuplicates++
		fmt.Printf("Duplicate set %d:\n", totalDuplicates)
		fmt.Printf("  Event: %s | %s\n", eventType, eventDate)
		fmt.Printf("  Description: %s\n", description)
		fmt.Printf("  Count: %d duplicates\n", count)
		fmt.Printf("  IDs: %v\n", ids)
		fmt.Printf("  Created: %v\n", createdAts)
		fmt.Println()
	}

	if totalDuplicates == 0 {
		fmt.Println("✓ No duplicates found")
	} else {
		fmt.Printf("✗ Found %d duplicate event sets\n", totalDuplicates)
		fmt.Println("\nSuggestion: Keep first occurrence (oldest created_at), delete others")
	}
}
