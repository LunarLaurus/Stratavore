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

	var total int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM rank_tracking`).Scan(&total)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total rank_tracking events in V3: %d\n", total)

	// Count by type
	rows, err := conn.Query(ctx, `
		SELECT event_type, COUNT(*) as count
		FROM rank_tracking
		GROUP BY event_type
		ORDER BY count DESC
	`)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("\nBreakdown by event type:")
	for rows.Next() {
		var eventType string
		var count int
		rows.Scan(&eventType, &count)
		fmt.Printf("  %-15s: %d\n", eventType, count)
	}
}
