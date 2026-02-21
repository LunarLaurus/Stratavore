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

	var nullCount int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM rank_tracking WHERE description IS NULL`).Scan(&nullCount)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		os.Exit(1)
	}

	var notNullCount int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM rank_tracking WHERE description IS NOT NULL`).Scan(&notNullCount)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Rank tracking descriptions:\n")
	fmt.Printf("  NULL:     %d\n", nullCount)
	fmt.Printf("  NOT NULL: %d\n", notNullCount)
	fmt.Printf("  TOTAL:    %d\n", nullCount+notNullCount)

	if nullCount > 0 {
		fmt.Println("\nSample NULL description events:")
		rows, _ := conn.Query(ctx, `
			SELECT event_type, event_date, current_rank
			FROM rank_tracking
			WHERE description IS NULL
			LIMIT 5
		`)
		defer rows.Close()

		for rows.Next() {
			var eventType, eventDate, currentRank string
			rows.Scan(&eventType, &eventDate, &currentRank)
			fmt.Printf("  %s | %s | %s\n", eventType, eventDate, currentRank)
		}
	}
}
