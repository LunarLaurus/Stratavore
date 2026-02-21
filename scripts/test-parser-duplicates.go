package main

import (
	"fmt"
	"path/filepath"

	"github.com/meridian-lex/stratavore/internal/migrate/parsers"
)

func main() {
	v2Dir := "/home/meridian/meridian-home/lex-internal/state"
	rankPath := filepath.Join(v2Dir, "..", "directives", "rank-status.jsonl")

	fmt.Println("Parsing V2 rank-status.jsonl...")
	rankStatus, err := parsers.ParseRankStatusFile(rankPath)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	events := rankStatus.GetRankEvents()
	fmt.Printf("Total events extracted: %d\n\n", len(events))

	// Check for duplicates
	seen := make(map[string]int)
	for _, event := range events {
		key := fmt.Sprintf("%s|%s|%s", event.Type, event.Date.Format("2006-01-02"), event.Description)
		seen[key]++
	}

	duplicates := 0
	for key, count := range seen {
		if count > 1 {
			duplicates++
			fmt.Printf("DUPLICATE (%dx): %s\n", count, key)
		}
	}

	if duplicates == 0 {
		fmt.Println("✓ No duplicates in parsed events")
	} else {
		fmt.Printf("\n✗ Found %d duplicate event keys\n", duplicates)
	}

	// Sample first 5 events
	fmt.Println("\nFirst 5 events:")
	for i := 0; i < 5 && i < len(events); i++ {
		e := events[i]
		fmt.Printf("  %d. %s | %s | %s\n", i+1, e.Type, e.Date.Format("2006-01-02"), e.Description[:min(60, len(e.Description))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
