package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/database"
	"github.com/valentinesamuel/activelog/internal/models"
)

func setupTestDB(t *testing.T) (*sql.DB, DBConn) {
	cfg := config.Load()
	db, err := sql.Open("postgres", "postgres://activelog_user:activelog@localhost:5444/activelog_test?sslmode=disable")

	if err != nil {
		t.Fatalf("❌ Failed to connect to test database %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("❌ Failed to ping test database %v", err)
	}

	var dbConn DBConn = db
	if cfg.EnableQueryLogging {
		queryLogger := log.New(os.Stdout, "[SQL] ", log.LstdFlags)
		dbConn = database.NewLoggingDB(db, queryLogger)
		log.Println("🔍 Query logging enabled")
	}

	return db, dbConn
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

func createTestUser(t *testing.T, dbConn DBConn) int {
	var userID int
	err := dbConn.QueryRow(`
	INSERT INTO users (email, username, password_hash)
	VALUES ($1, $2, $3)
	RETURNING id
	`, "test@test.com", "testuser", "password").Scan(&userID)

	if err != nil {
		t.Fatalf("❌ Failed to create test user %v", err)
	}
	return userID
}

func TestActivityRepository_Create(t *testing.T) {
	db, dbConn := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userID := createTestUser(t, dbConn)
	repo := NewActivityRepository(dbConn, nil)

	activity := &models.Activity{
		UserID:          userID,
		ActivityType:    "running",
		Title:           "Test Run",
		DurationMinutes: 30,
		DistanceKm:      5.0,
		ActivityDate:    time.Now(),
	}

	err := repo.Create(t.Context(), activity)

	if err != nil {
		t.Fatalf("❌ Failed to create activity %v", err)
	}

	if activity.ID == 0 {
		t.Errorf("❌ Activity ID should be set after creation")
	}

	if activity.CreatedAt.IsZero() {
		t.Errorf("❌ CreatedAt should be set after creation")
	}
}

func TestActivityRepository_GetById(t *testing.T) {
	db, dbConn := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userID := createTestUser(t, dbConn)
	repo := NewActivityRepository(dbConn, nil)

	activity := &models.Activity{
		UserID:       userID,
		ActivityType: "basketball",
		Title:        "Pickup Game",
		ActivityDate: time.Now(),
	}

	err := repo.Create(t.Context(), activity)
	if err != nil {
		t.Fatalf("❌ Failed to create activity: %v", err)
	}

	fetched, err := repo.GetByID(t.Context(), (activity.ID))
	if err != nil {
		t.Fatalf("❌ Failed to get activity: %v", err)
	}

	if fetched.ID != activity.ID {
		t.Errorf("Expected ID %d, got %d", activity.ID, fetched.ID)
	}

	if fetched.Title != "Pickup Game" {
		t.Errorf("❌ Expected title 'Pickup Game', got '%s'", fetched.Title)
	}
}

func TestActivityRepository_ListByUser(t *testing.T) {
	db, dbConn := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userID := createTestUser(t, dbConn)
	repo := NewActivityRepository(dbConn, nil)

	for i := range 3 {
		activity := &models.Activity{
			UserID:       userID,
			ActivityType: "Running",
			Title:        fmt.Sprintf("Run %d", i),
			ActivityDate: time.Now(),
		}
		err := repo.Create(t.Context(), activity)
		if err != nil {
			t.Fatalf("❌ Failed to create activity %v", err)
		}
	}

	activities, err := repo.ListByUser(t.Context(), userID)
	if err != nil {
		t.Fatalf("❌ Failed to list activities %v", err)
	}

	if len(activities) != 3 {
		t.Errorf("❌ Expected 3 activities, got %d", len(activities))
	}
}
