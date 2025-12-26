package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/valentinesamuel/activelog/internal/models"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", "postgres://activelog_user:activelog@localhost:5444/activelog_test?sslmode=disable")

	if err != nil {
		t.Fatalf("❌ Failed to connect to test database %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("❌ Failed to ping test database %v", err)
	}

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM activities")
	db.Exec("DELETE FROM users")
	if err != nil {
		t.Logf("⚠️ Warning: Failed to clean activities :%v", err)
	}
	db.Close()
}

func createTestUser(t *testing.T, db *sql.DB) int {
	var userID int
	err := db.QueryRow(`
	INSERT INTO users (email, username)
	VALUES ($1, $2)
	RETURNING id
	`, "test@test.com", "testuser").Scan(&userID)

	if err != nil {
		t.Fatalf("❌ Failed to create test user %v", err)
	}
	return userID
}

func TestActivityRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userID := createTestUser(t, db)
	repo := NewActivityRepository(db)

	activity := &models.Activity{
		UserID:          userID,
		ActivityType:    "running",
		Title:           "Test Run",
		DurationMinutes: 30,
		DistanceKm:      5.0,
		ActivityDate:    time.Now(),
	}

	err := repo.Create(t.Context(),
		activity)

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
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userID := createTestUser(t, db)
	repo := NewActivityRepository(db)

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
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userID := createTestUser(t, db)
	repo := NewActivityRepository(db)

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
