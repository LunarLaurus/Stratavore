package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/meridian-lex/stratavore/internal/migrate/parsers"
)

func main() {
	ctx := context.Background()

	// Parse V2 data
	v2Dir := "/home/meridian/meridian-home/lex-internal/state"
	rankPath := filepath.Join(v2Dir, "..", "directives", "rank-status.jsonl")

	fmt.Println("Parsing V2 rank-status.jsonl...")
	rankStatus, err := parsers.ParseRankStatusFile(rankPath)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	v2Events := rankStatus.GetRankEvents()
	fmt.Printf("V2 events: %d\n", len(v2Events))

	// Get V3 data
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

	var v3Count int
	err = conn.QueryRow(ctx, `SELECT COUNT(*) FROM rank_tracking`).Scan(&v3Count)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("V3 events: %d\n", v3Count)
	fmt.Println()

	// Compare first strike event
	v2Strike := v2Events[0]
	fmt.Println("V2 Strike 1:")
	fmt.Printf("  Type: %s\n", v2Strike.Type)
	fmt.Printf("  Date: %s\n", v2Strike.Date.Format("2006-01-02"))
	fmt.Printf("  Description: %s\n", v2Strike.Description)
	fmt.Printf("  Evidence: %s\n", v2Strike.Evidence)
	fmt.Println()

	var v3Type, v3Date, v3Desc, v3Evidence string
	err = conn.QueryRow(ctx, `
		SELECT event_type, event_date::text, description, COALESCE(evidence, '') as evidence
		FROM rank_tracking
		WHERE event_type = 'strike'
		ORDER BY event_date
		LIMIT 1
	`).Scan(&v3Type, &v3Date, &v3Desc, &v3Evidence)

	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("V3 Strike 1:")
	fmt.Printf("  Type: %s\n", v3Type)
	fmt.Printf("  Date: %s\n", v3Date)
	fmt.Printf("  Description: %s\n", v3Desc)
	fmt.Printf("  Evidence: %s\n", v3Evidence)
	fmt.Println()

	// Check if they match
	v2DateStr := v2Strike.Date.Format("2006-01-02")
	v3DateStr := v3Date[:10] // Extract just the date part

	if v2Strike.Type == v3Type && v2DateStr == v3DateStr && v2Strike.Description == v3Desc {
		fmt.Println("✓ V2 and V3 data MATCH for first strike")
	} else {
		fmt.Println("✗ V2 and V3 data MISMATCH:")
		if v2Strike.Type != v3Type {
			fmt.Printf("  Type: '%s' vs '%s'\n", v2Strike.Type, v3Type)
		}
		if v2DateStr != v3DateStr {
			fmt.Printf("  Date: '%s' vs '%s'\n", v2DateStr, v3DateStr)
		}
		if v2Strike.Description != v3Desc {
			fmt.Printf("  Description:\n    V2: '%s'\n    V3: '%s'\n", v2Strike.Description, v3Desc)
		}
	}
}
