package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDB creates a PostgreSQL testcontainer and runs migrations
// Returns: (*sql.DB, cleanup func())
// The cleanup function must be called with defer to stop the container
func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	// 1. Start postgres container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:latest",
		postgres.WithDatabase("activelog"),
		postgres.WithUsername("activelog_user"),
		postgres.WithPassword("activelog"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	// 2. Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// 3. Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("❌ Failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("❌ Failed to ping test database: %v", err)
	}

	// 4. Run migrations
	if err := runMigrations(t, db); err != nil {
		db.Close()
		postgresContainer.Terminate(ctx)
		t.Fatalf("❌ Failed to run migrations: %v", err)
	}

	t.Log("✅ Test database ready with migrations applied")

	// 5. Return cleanup function
	cleanup := func() {
		db.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("⚠️ Warning: Failed to terminate container: %v", err)
		}
	}

	return db, cleanup
}

// runMigrations reads and executes all .up.sql migration files in order
func runMigrations(t *testing.T, db *sql.DB) error {
	t.Helper()

	// Get the project root (assumes testhelpers is in internal/repository/testhelpers)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Navigate to migrations directory from anywhere in the project
	migrationsDir := filepath.Join(cwd, "..", "..", "..", "migrations")

	// If that doesn't exist, try alternative path (when running from project root)
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = filepath.Join(cwd, "migrations")
	}

	// Read all migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationsDir)
	}

	// Sort files to ensure correct order (000001, 000002, etc.)
	sort.Strings(files)

	// Execute each migration file
	for _, file := range files {
		t.Logf("Running migration: %s", filepath.Base(file))

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Execute the migration SQL
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	t.Logf("✅ Applied %d migrations successfully", len(files))
	return nil
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	// Get all table names from the database
	rows, err := db.Query(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
	`)
	if err != nil {
		t.Logf("⚠️ Warning: Failed to get table names: %v", err)
		db.Close()
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			t.Logf("⚠️ Warning: Failed to scan table name: %v", err)
			continue
		}
		tables = append(tables, tableName)
	}

	// Truncate all tables with CASCADE to handle foreign key constraints
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			t.Logf("⚠️ Warning: Failed to truncate table %s: %v", table, err)
		}
	}

	db.Close()
}