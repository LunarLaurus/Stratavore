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

	// Query indexes on rank_tracking
	rows, err := conn.Query(ctx,
		`SELECT indexname, indexdef
		 FROM pg_indexes
		 WHERE tablename = 'rank_tracking' AND schemaname = 'public'`)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("Indexes on rank_tracking:")
	for rows.Next() {
		var name, def string
		if err := rows.Scan(&name, &def); err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}
		fmt.Printf("- %s\n  %s\n", name, def)
	}
}
