package testhelpers

import (
	"testing"
)

func TestSetupTestDB(t *testing.T) {
	// This test verifies that SetupTestDB:
	// 1. Starts a postgres container
	// 2. Connects to the database
	// 3. Runs migrations successfully
	// 4. Returns a working database connection and cleanup function

	db, cleanup := SetupTestDB(t)
	defer cleanup() // Cleanup will stop the container

	// Verify database connection works
	if err := db.Ping(); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Verify migrations ran (check that users table exists)
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'users'
		)
	`).Scan(&exists)

	if err != nil {
		t.Fatalf("Failed to check if users table exists: %v", err)
	}

	if !exists {
		t.Fatal("Users table does not exist - migrations did not run properly")
	}

	// Verify we can insert data (schema is fully set up)
	_, err = db.Exec(`
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
	`, "test@example.com", "testuser", "hashedpassword123")

	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Verify we can read data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 1 {
		t.Fatalf("Expected 1 user, got %d", count)
	}

	t.Log("✅ SetupTestDB working correctly!")
}
