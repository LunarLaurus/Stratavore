package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
)

var (
	auditDBURL string
)

func init() {
	schemaAuditCmd.Flags().StringVar(&auditDBURL, "db-url", "", "PostgreSQL connection URL (default: from STRATAVORE_DB_URL env)")
}

var schemaAuditCmd = &cobra.Command{
	Use:   "schema-audit",
	Short: "Audit V3 database schema completeness",
	Long:  "Verify all expected tables, indexes, constraints, and types exist in the V3 database",
	RunE:  runSchemaAudit,
}

type TableInfo struct {
	Name    string
	Columns int
	Indexes int
	FKs     int
}

type ExtensionInfo struct {
	Name      string
	Installed bool
}

type TypeInfo struct {
	Name   string
	Kind   string
	Exists bool
}

func runSchemaAudit(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get database URL
	if auditDBURL == "" {
		auditDBURL = os.Getenv("STRATAVORE_DB_URL")
	}
	if auditDBURL == "" {
		return fmt.Errorf("database URL required: use --db-url flag or STRATAVORE_DB_URL env var")
	}

	// Connect to database
	conn, err := pgx.Connect(ctx, auditDBURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer conn.Close(ctx)

	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Println("│ Stratavore V3 Database Schema Completeness Audit    │")
	fmt.Println("└─────────────────────────────────────────────────────┘")
	fmt.Println()

	// 1. Check extensions
	fmt.Println("═══ PostgreSQL Extensions ═══")
	extensions := []string{"pgcrypto", "vector"}
	extResults := checkExtensions(ctx, conn, extensions)
	printExtensions(extResults)

	// 2. Check custom types
	fmt.Println("\n═══ Custom Types (ENUMs) ═══")
	types := []string{
		"runner_status",
		"project_status",
		"conversation_mode",
		"runtime_type",
		"event_severity",
	}
	typeResults := checkTypes(ctx, conn, types)
	printTypes(typeResults)

	// 3. Check tables
	fmt.Println("\n═══ Tables ═══")
	expectedTables := []string{
		// Migration 0001 - Core tables
		"projects",
		"runners",
		"project_capabilities",
		"sessions",
		"session_blobs",
		"outbox",
		"events",
		"token_budgets",
		"resource_quotas",
		"daemon_state",
		"agent_tokens",

		// Migration 0002 - Sprint system
		"model_registry",
		"sprints",
		"sprint_tasks",
		"sprint_executions",

		// Migration 0003 - V2 migration support
		"rank_tracking",
		"directives",
		"v2_sync_state",
		"v2_import_log",
	}

	tableResults, err := checkTables(ctx, conn, expectedTables)
	if err != nil {
		return fmt.Errorf("check tables: %w", err)
	}
	printTables(tableResults)

	// 4. Check foreign keys
	fmt.Println("\n═══ Foreign Key Constraints ═══")
	fkCount, err := checkForeignKeys(ctx, conn)
	if err != nil {
		return fmt.Errorf("check foreign keys: %w", err)
	}
	fmt.Printf("✓ Total foreign key constraints: %d\n", fkCount)

	// 5. Check for orphaned data
	fmt.Println("\n═══ Data Integrity Checks ═══")
	orphans, err := checkOrphanedData(ctx, conn)
	if err != nil {
		return fmt.Errorf("check orphaned data: %w", err)
	}
	if orphans == 0 {
		fmt.Println("✓ No orphaned foreign key references")
	} else {
		fmt.Printf("✗ Found %d orphaned references (FKs pointing to non-existent rows)\n", orphans)
	}

	// 6. Check unique constraints for idempotency
	fmt.Println("\n═══ Idempotency Constraints ═══")
	idempotentTables := map[string]string{
		"rank_tracking": "idx_rank_tracking_unique_event",
		"v2_sync_state": "v2_sync_state_pkey", // PRIMARY KEY on file_path
	}
	for table, constraint := range idempotentTables {
		exists, err := checkConstraintExists(ctx, conn, table, constraint)
		if err != nil {
			return fmt.Errorf("check constraint %s.%s: %w", table, constraint, err)
		}
		if exists {
			fmt.Printf("✓ %s: %s\n", table, constraint)
		} else {
			fmt.Printf("✗ %s: %s MISSING\n", table, constraint)
		}
	}

	// 7. Summary
	fmt.Println("\n═══ Schema Completeness Summary ═══")
	missingTables := 0
	for _, tbl := range tableResults {
		if tbl.Columns == 0 {
			missingTables++
		}
	}

	missingTypes := 0
	for _, typ := range typeResults {
		if !typ.Exists {
			missingTypes++
		}
	}

	missingExt := 0
	for _, ext := range extResults {
		if !ext.Installed {
			missingExt++
		}
	}

	totalExpected := len(expectedTables)
	totalFound := totalExpected - missingTables

	fmt.Printf("Extensions: %d/%d installed\n", len(extensions)-missingExt, len(extensions))
	fmt.Printf("Types:      %d/%d present\n", len(types)-missingTypes, len(types))
	fmt.Printf("Tables:     %d/%d present\n", totalFound, totalExpected)
	fmt.Printf("Foreign Keys: %d constraints\n", fkCount)

	if missingTables == 0 && missingTypes == 0 && missingExt == 0 && orphans == 0 {
		fmt.Println("\n✓ Schema complete and valid")
		return nil
	} else {
		fmt.Println("\n✗ Schema validation FAILED - missing or invalid components")
		os.Exit(1)
	}

	return nil
}

func checkExtensions(ctx context.Context, conn *pgx.Conn, extensions []string) []ExtensionInfo {
	results := make([]ExtensionInfo, len(extensions))

	for i, ext := range extensions {
		var installed bool
		err := conn.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = $1)`,
			ext,
		).Scan(&installed)

		if err != nil {
			installed = false
		}

		results[i] = ExtensionInfo{
			Name:      ext,
			Installed: installed,
		}
	}

	return results
}

func checkTypes(ctx context.Context, conn *pgx.Conn, types []string) []TypeInfo {
	results := make([]TypeInfo, len(types))

	for i, typ := range types {
		var kind string
		var exists bool

		err := conn.QueryRow(ctx,
			`SELECT typtype::text FROM pg_type
			 JOIN pg_namespace ON pg_type.typnamespace = pg_namespace.oid
			 WHERE typname = $1 AND nspname = 'public'`,
			typ,
		).Scan(&kind)

		if err == nil {
			exists = true
		}

		results[i] = TypeInfo{
			Name:   typ,
			Kind:   kind,
			Exists: exists,
		}
	}

	return results
}

func checkTables(ctx context.Context, conn *pgx.Conn, tables []string) ([]TableInfo, error) {
	results := make([]TableInfo, len(tables))

	for i, table := range tables {
		// Count columns
		var colCount int
		err := conn.QueryRow(ctx,
			`SELECT COUNT(*) FROM information_schema.columns
			 WHERE table_name = $1 AND table_schema = 'public'`,
			table,
		).Scan(&colCount)

		if err != nil {
			return nil, fmt.Errorf("count columns for %s: %w", table, err)
		}

		// Count indexes
		var idxCount int
		err = conn.QueryRow(ctx,
			`SELECT COUNT(*) FROM pg_indexes
			 WHERE tablename = $1 AND schemaname = 'public'`,
			table,
		).Scan(&idxCount)

		if err != nil {
			return nil, fmt.Errorf("count indexes for %s: %w", table, err)
		}

		// Count foreign keys
		var fkCount int
		err = conn.QueryRow(ctx,
			`SELECT COUNT(*) FROM information_schema.table_constraints
			 WHERE constraint_type = 'FOREIGN KEY'
			 AND table_name = $1 AND table_schema = 'public'`,
			table,
		).Scan(&fkCount)

		if err != nil {
			return nil, fmt.Errorf("count FKs for %s: %w", table, err)
		}

		results[i] = TableInfo{
			Name:    table,
			Columns: colCount,
			Indexes: idxCount,
			FKs:     fkCount,
		}
	}

	return results, nil
}

func checkForeignKeys(ctx context.Context, conn *pgx.Conn) (int, error) {
	var count int
	err := conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM information_schema.table_constraints
		 WHERE constraint_type = 'FOREIGN KEY' AND table_schema = 'public'`,
	).Scan(&count)

	return count, err
}

func checkOrphanedData(ctx context.Context, conn *pgx.Conn) (int, error) {
	// Check for common FK violations (sample checks)
	orphans := 0

	// runners.project_name → projects.name
	var orphanedRunners int
	err := conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM runners r
		 WHERE NOT EXISTS (SELECT 1 FROM projects p WHERE p.name = r.project_name)`,
	).Scan(&orphanedRunners)
	if err == nil {
		orphans += orphanedRunners
	}

	// sessions.project_name → projects.name
	var orphanedSessions int
	err = conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM sessions s
		 WHERE NOT EXISTS (SELECT 1 FROM projects p WHERE p.name = s.project_name)`,
	).Scan(&orphanedSessions)
	if err == nil {
		orphans += orphanedSessions
	}

	// sessions.runner_id → runners.id
	var orphanedSessionRunners int
	err = conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM sessions s
		 WHERE NOT EXISTS (SELECT 1 FROM runners r WHERE r.id = s.runner_id)`,
	).Scan(&orphanedSessionRunners)
	if err == nil {
		orphans += orphanedSessionRunners
	}

	return orphans, nil
}

func checkConstraintExists(ctx context.Context, conn *pgx.Conn, table, constraint string) (bool, error) {
	var exists bool

	// Check both constraint name and index name
	err := conn.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM information_schema.table_constraints
			WHERE constraint_name = $1 AND table_name = $2 AND table_schema = 'public'
		) OR EXISTS(
			SELECT 1 FROM pg_indexes
			WHERE indexname = $1 AND tablename = $2 AND schemaname = 'public'
		)`,
		constraint,
		table,
	).Scan(&exists)

	return exists, err
}

func printExtensions(extensions []ExtensionInfo) {
	for _, ext := range extensions {
		status := "✓"
		if !ext.Installed {
			status = "✗"
		}
		fmt.Printf("%s %s\n", status, ext.Name)
	}
}

func printTypes(types []TypeInfo) {
	for _, typ := range types {
		status := "✓"
		kind := "enum"
		if !typ.Exists {
			status = "✗"
			kind = "MISSING"
		}
		fmt.Printf("%s %-25s (%s)\n", status, typ.Name, kind)
	}
}

func printTables(tables []TableInfo) {
	// Group by migration
	migration0001 := tables[0:11]
	migration0002 := tables[11:15]
	migration0003 := tables[15:19]

	fmt.Println("\nMigration 0001 - Core Tables:")
	printTableGroup(migration0001)

	fmt.Println("\nMigration 0002 - Sprint System:")
	printTableGroup(migration0002)

	fmt.Println("\nMigration 0003 - V2 Migration:")
	printTableGroup(migration0003)
}

func printTableGroup(tables []TableInfo) {
	for _, tbl := range tables {
		status := "✓"
		if tbl.Columns == 0 {
			status = "✗"
		}

		name := tbl.Name
		if len(name) < 22 {
			name = name + strings.Repeat(" ", 22-len(name))
		}

		fmt.Printf("%s %-22s | %2d cols | %2d idx | %d FKs\n",
			status, name, tbl.Columns, tbl.Indexes, tbl.FKs)
	}
}
