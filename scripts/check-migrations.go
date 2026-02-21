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

	// Check if schema_migrations table exists
	var exists bool
	err = conn.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'schema_migrations'
		)`).Scan(&exists)

	if err != nil || !exists {
		fmt.Println("No schema_migrations table found")
		fmt.Println("Listing all tables to understand migration state:")

		rows, _ := conn.Query(ctx,
			`SELECT table_name FROM information_schema.tables
			 WHERE table_schema = 'public' ORDER BY table_name`)
		defer rows.Close()

		for rows.Next() {
			var name string
			rows.Scan(&name)
			fmt.Printf("- %s\n", name)
		}
		return
	}

	// Query applied migrations
	rows, err := conn.Query(ctx,
		`SELECT version, dirty FROM schema_migrations ORDER BY version`)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("Applied migrations:")
	for rows.Next() {
		var version int
		var dirty bool
		if err := rows.Scan(&version, &dirty); err != nil {
			continue
		}
		status := "✓"
		if dirty {
			status = "✗ DIRTY"
		}
		fmt.Printf("%s Migration %d\n", status, version)
	}
}
