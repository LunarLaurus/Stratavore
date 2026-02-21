//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

	fmt.Println("Testing ON CONFLICT with rank_tracking unique index...")
	fmt.Println()

	// Test 1: Insert a new test event
	testDate := time.Now().Format("2006-01-02")

	query1 := `
		INSERT INTO rank_tracking (
			current_rank, progress, strikes, commendations,
			event_type, event_date, description, evidence, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	fmt.Println("Test 1: Insert new test event...")
	_, err = conn.Exec(ctx, query1,
		"Ensign", "0/5", 0, 0,
		"test", testDate, "Test conflict handling", "", `{"test":true}`)

	if err != nil {
		fmt.Printf("✗ Test 1 failed: %v\n", err)
	} else {
		fmt.Println("✓ Test 1 passed: New event inserted")
	}

	// Test 2: Try to insert the same event again WITHOUT ON CONFLICT (should fail)
	fmt.Println("\nTest 2: Insert duplicate WITHOUT ON CONFLICT (should fail)...")
	_, err = conn.Exec(ctx, query1,
		"Ensign", "0/5", 0, 0,
		"test", testDate, "Test conflict handling", "", `{"test":true}`)

	if err != nil {
		fmt.Printf("✓ Test 2 passed: Duplicate rejected as expected\n   Error: %v\n", err)
	} else {
		fmt.Println("✗ Test 2 failed: Duplicate was inserted (should have been rejected)")
	}

	// Test 3: Try with ON CONFLICT DO NOTHING (should succeed silently)
	query3 := `
		INSERT INTO rank_tracking (
			current_rank, progress, strikes, commendations,
			event_type, event_date, description, evidence, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (event_type, event_date, COALESCE(description, ''))
		DO NOTHING
	`

	fmt.Println("\nTest 3: Insert duplicate WITH ON CONFLICT DO NOTHING (should succeed silently)...")
	result, err := conn.Exec(ctx, query3,
		"Ensign", "0/5", 0, 0,
		"test", testDate, "Test conflict handling", "", `{"test":true}`)

	if err != nil {
		fmt.Printf("✗ Test 3 failed: %v\n", err)
	} else {
		rows := result.RowsAffected()
		fmt.Printf("✓ Test 3 passed: Query succeeded, %d rows affected (0 expected)\n", rows)
	}

	// Clean up test data
	fmt.Println("\nCleaning up test data...")
	_, err = conn.Exec(ctx, `DELETE FROM rank_tracking WHERE event_type = 'test'`)
	if err != nil {
		fmt.Printf("Warning: cleanup failed: %v\n", err)
	} else {
		fmt.Println("✓ Test data removed")
	}

	fmt.Println("\n=== Test Summary ===")
	fmt.Println("If all tests passed, ON CONFLICT syntax is working correctly.")
	fmt.Println("The sync issue may be elsewhere (e.g., description NULL handling).")
}
