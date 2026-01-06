package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/valentinesamuel/activelog/internal/database"
)

// SetupTestDB creates a PostgreSQL testcontainer and runs migrations
// Returns: (*database.LoggingDB, cleanup func())
// The cleanup function must be called with defer to stop the container
// Accepts testing.TB interface so it works with both *testing.T and *testing.B
func SetupTestDB(t testing.TB) (*database.LoggingDB, func()) {
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
	rawDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("❌ Failed to connect to test database: %v", err)
	}

	if err := rawDB.Ping(); err != nil {
		t.Fatalf("❌ Failed to ping test database: %v", err)
	}

	// Configure connection pool for test workloads (especially benchmarks)
	rawDB.SetMaxOpenConns(25)        // Allow up to 25 concurrent connections
	rawDB.SetMaxIdleConns(25)         // Keep 25 idle connections ready
	rawDB.SetConnMaxLifetime(5 * time.Minute) // Connections live for 5 minutes max

	// 4. Run migrations
	if err := runMigrations(t, rawDB); err != nil {
		rawDB.Close()
		postgresContainer.Terminate(ctx)
		t.Fatalf("❌ Failed to run migrations: %v", err)
	}

	// 5. Wrap in LoggingDB for transaction support
	// Use a silent logger for tests/benchmarks to reduce noise
	logger := log.New(io.Discard, "", 0) // Silent logger
	db := database.NewLoggingDB(rawDB, logger)

	t.Log("✅ Test database ready with migrations applied")

	// 6. Return cleanup function
	cleanup := func() {
		rawDB.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("⚠️ Warning: Failed to terminate container: %v", err)
		}
	}

	return db, cleanup
}

// findProjectRoot walks up the directory tree to find the project root (where go.mod is)
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			return "", fmt.Errorf("could not find project root (go.mod not found)")
		}
		dir = parent
	}
}

// runMigrations reads and executes all .up.sql migration files in order
func runMigrations(t testing.TB, db *sql.DB) error {
	t.Helper()

	// Find project root by looking for go.mod
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Migrations are in the migrations/ directory at project root
	migrationsDir := filepath.Join(projectRoot, "migrations")

	// Verify migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found at %s", migrationsDir)
	}

	// Read all migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationsDir)
	}

	t.Logf("📂 Found %d migration files in %s", len(files), migrationsDir)

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