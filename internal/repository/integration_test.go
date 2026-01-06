package repository

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository/testhelpers"
)

// ==================== HELPER FUNCTIONS ====================
// Note: These helpers are prefixed with "Integration" to avoid conflicts
// with existing test helpers in activity_repository_test.go

// createIntegrationTestUser creates a test user and returns the user ID
func createIntegrationTestUser(t *testing.T, db *sql.DB, email, username string) int {
	t.Helper()

	var userID int
	query := `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := db.QueryRow(query, email, username, "hashedpassword123").Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

// createIntegrationTestActivity creates a test activity and returns the activity
func createIntegrationTestActivity(t *testing.T, repo *ActivityRepository, userID int, activityType, title string) *models.Activity {
	t.Helper()
	ctx := context.Background()

	activity := &models.Activity{
		UserID:          userID,
		ActivityType:    activityType,
		Title:           title,
		Description:     "Test description",
		DurationMinutes: 30,
		DistanceKm:      5.0,
		CaloriesBurned:  250,
		Notes:           "Test notes",
		ActivityDate:    time.Now(),
	}

	err := repo.Create(ctx, nil, activity)
	if err != nil {
		t.Fatalf("Failed to create test activity: %v", err)
	}

	return activity
}

// verifyActivityExists checks if an activity exists in the database
func verifyActivityExists(t *testing.T, db *sql.DB, activityID int64) bool {
	t.Helper()

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM activities WHERE id = $1)", activityID).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check activity existence: %v", err)
	}

	return exists
}

// verifyTagExists checks if a tag exists in the database
func verifyTagExists(t *testing.T, db *sql.DB, tagName string) (bool, int) {
	t.Helper()

	var tagID int
	err := db.QueryRow("SELECT id FROM tags WHERE name = $1", tagName).Scan(&tagID)
	if err == sql.ErrNoRows {
		return false, 0
	}
	if err != nil {
		t.Fatalf("Failed to check tag existence: %v", err)
	}

	return true, tagID
}

// verifyActivityTagLink checks if an activity-tag link exists
func verifyActivityTagLink(t *testing.T, db *sql.DB, activityID int64, tagID int) bool {
	t.Helper()

	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM activity_tags WHERE activity_id = $1 AND tag_id = $2)",
		activityID, tagID,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check activity-tag link: %v", err)
	}

	return exists
}

// countActivities counts total activities for a user
func countActivities(t *testing.T, db *sql.DB, userID int) int {
	t.Helper()

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM activities WHERE user_id = $1", userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count activities: %v", err)
	}

	return count
}

// ==================== INTEGRATION TESTS ====================

// TestIntegration_CreateActivityWithTags tests the full transaction flow
// of creating an activity with associated tags
func TestIntegration_CreateActivityWithTags(t *testing.T) {
	// Setup: Create test database with migrations
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	// Create a test user
	userID := createIntegrationTestUser(t, db, "test@example.com", "testuser")

	// Test data
	activity := &models.Activity{
		UserID:          userID,
		ActivityType:    "running",
		Title:           "Morning Run",
		Description:     "5K around the park",
		DurationMinutes: 35,
		DistanceKm:      5.2,
		CaloriesBurned:  300,
		Notes:           "Felt great!",
		ActivityDate:    time.Now(),
	}

	tags := []*models.Tag{
		{Name: "outdoor"},
		{Name: "cardio"},
		{Name: "morning"},
	}

	// Execute: Create activity with tags in a transaction
	err := activityRepo.CreateWithTags(ctx, activity, tags)
	if err != nil {
		t.Fatalf("CreateWithTags failed: %v", err)
	}

	// Verify: Activity was created
	if activity.ID == 0 {
		t.Fatal("Activity ID should be set after creation")
	}

	exists := verifyActivityExists(t, db, activity.ID)
	if !exists {
		t.Fatal("Activity was not created in database")
	}

	// Verify: Tags were created
	for _, tag := range tags {
		exists, tagID := verifyTagExists(t, db, tag.Name)
		if !exists {
			t.Fatalf("Tag '%s' was not created", tag.Name)
		}

		// Verify: Activity-tag link exists
		linked := verifyActivityTagLink(t, db, activity.ID, tagID)
		if !linked {
			t.Fatalf("Activity-tag link missing for tag '%s'", tag.Name)
		}
	}

	// Verify: Can retrieve activity with tags
	retrieved, err := activityRepo.GetByID(ctx, activity.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve activity: %v", err)
	}

	if retrieved.Title != "Morning Run" {
		t.Errorf("Expected title 'Morning Run', got '%s'", retrieved.Title)
	}

	t.Log("✅ Transaction test passed: Activity created with tags successfully")
}

// TestIntegration_CreateActivityWithTags_RollbackOnError tests that
// the transaction rolls back if tag creation fails
func TestIntegration_CreateActivityWithTags_RollbackOnError(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	userID := createIntegrationTestUser(t, db, "rollback@example.com", "rollbackuser")

	activity := &models.Activity{
		UserID:          userID,
		ActivityType:    "running",
		Title:           "Test Rollback",
		DurationMinutes: 30,
		ActivityDate:    time.Now(),
	}

	// Create a tag with an extremely long name that will violate VARCHAR(50) constraint
	tags := []*models.Tag{
		{Name: "validtag"},
		{Name: "this_tag_name_is_way_too_long_and_exceeds_fifty_characters_limit_by_a_lot"},
	}

	// Execute: This should fail and rollback
	err := activityRepo.CreateWithTags(ctx, activity, tags)
	if err == nil {
		t.Fatal("Expected CreateWithTags to fail with long tag name, but it succeeded")
	}

	// Verify: Activity was NOT created (transaction rolled back)
	count := countActivities(t, db, userID)
	if count != 0 {
		t.Errorf("Expected 0 activities after rollback, got %d", count)
	}

	// Verify: First tag was NOT created either (even though it was valid)
	exists, _ := verifyTagExists(t, db, "validtag")
	if exists {
		t.Error("Tag 'validtag' should not exist - transaction should have rolled back")
	}

	t.Log("✅ Rollback test passed: Transaction properly rolled back on error")
}

// TestIntegration_ConcurrentInsertions tests that multiple goroutines
// can safely insert activities concurrently
func TestIntegration_ConcurrentInsertions(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	userID := createIntegrationTestUser(t, db, "concurrent@example.com", "concurrentuser")

	// Number of concurrent insertions
	numGoroutines := 10

	// WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Channel to collect errors
	errChan := make(chan error, numGoroutines)

	// Channel to collect created activity IDs
	idChan := make(chan int64, numGoroutines)

	// Launch multiple goroutines to create activities concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			activity := &models.Activity{
				UserID:          userID,
				ActivityType:    "running",
				Title:           "Concurrent Activity",
				DurationMinutes: 30 + index,
				DistanceKm:      float64(5 + index),
				ActivityDate:    time.Now(),
			}

			err := activityRepo.Create(ctx, nil, activity)
			if err != nil {
				errChan <- err
				return
			}

			idChan <- activity.ID
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)
	close(idChan)

	// Check for errors
	for err := range errChan {
		t.Errorf("Concurrent insertion failed: %v", err)
	}

	// Collect all IDs
	var ids []int64
	for id := range idChan {
		ids = append(ids, id)
	}

	// Verify: All activities were created
	if len(ids) != numGoroutines {
		t.Fatalf("Expected %d activities, got %d", numGoroutines, len(ids))
	}

	// Verify: All IDs are unique
	idMap := make(map[int64]bool)
	for _, id := range ids {
		if idMap[id] {
			t.Fatalf("Duplicate activity ID found: %d", id)
		}
		idMap[id] = true
	}

	// Verify: Database count matches
	count := countActivities(t, db, userID)
	if count != numGoroutines {
		t.Errorf("Expected %d activities in database, got %d", numGoroutines, count)
	}

	t.Logf("✅ Concurrency test passed: %d activities created concurrently without conflicts", numGoroutines)
}

// TestIntegration_ForeignKeyConstraint tests that foreign key constraints
// are properly enforced
func TestIntegration_ForeignKeyConstraint(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	// Test 1: Cannot create activity for non-existent user
	t.Run("RejectActivityForNonExistentUser", func(t *testing.T) {
		activity := &models.Activity{
			UserID:          99999, // Non-existent user ID
			ActivityType:    "running",
			Title:           "Should Fail",
			DurationMinutes: 30,
			ActivityDate:    time.Now(),
		}

		err := activityRepo.Create(ctx, nil, activity)
		if err == nil {
			t.Fatal("Expected foreign key constraint violation, but insert succeeded")
		}

		// Verify the error is related to foreign key constraint
		if !contains(err.Error(), "foreign key") && !contains(err.Error(), "violates") {
			t.Logf("Error message: %v (may or may not mention foreign key explicitly)", err)
		}

		t.Log("✅ Foreign key constraint properly rejected invalid user_id")
	})

	// Test 2: Deleting a user cascades to delete activities
	t.Run("CascadeDeleteActivities", func(t *testing.T) {
		userID := createIntegrationTestUser(t, db, "cascade@example.com", "cascadeuser")

		// Create activities for this user
		_ = createIntegrationTestActivity(t, activityRepo, userID, "running", "Activity 1")
		_ = createIntegrationTestActivity(t, activityRepo, userID, "cycling", "Activity 2")

		// Verify activities exist
		initialCount := countActivities(t, db, userID)
		if initialCount != 2 {
			t.Fatalf("Expected 2 activities, got %d", initialCount)
		}

		// Delete the user
		_, err := db.Exec("DELETE FROM users WHERE id = $1", userID)
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		// Verify activities were cascade deleted
		finalCount := countActivities(t, db, userID)
		if finalCount != 0 {
			t.Errorf("Expected 0 activities after cascade delete, got %d", finalCount)
		}

		t.Log("✅ CASCADE DELETE properly removed activities when user was deleted")
	})
}

// TestIntegration_UniqueConstraintViolations tests unique constraint enforcement
func TestIntegration_UniqueConstraintViolations(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	// Test 1: Duplicate email
	t.Run("RejectDuplicateEmail", func(t *testing.T) {
		email := "duplicate@example.com"

		// Create first user
		_ = createIntegrationTestUser(t, db, email, "user1")

		// Try to create second user with same email (should fail)
		var userID int
		err := db.QueryRow(
			"INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING id",
			email, "user2", "password",
		).Scan(&userID)

		if err == nil {
			t.Fatal("Expected unique constraint violation for duplicate email, but insert succeeded")
		}

		if !contains(err.Error(), "unique") && !contains(err.Error(), "duplicate") {
			t.Logf("Error message: %v (may or may not mention unique constraint explicitly)", err)
		}

		t.Log("✅ Unique constraint properly rejected duplicate email")
	})

	// Test 2: Duplicate username
	t.Run("RejectDuplicateUsername", func(t *testing.T) {
		username := "duplicateuser"

		// Create first user
		_ = createIntegrationTestUser(t, db, "user1@example.com", username)

		// Try to create second user with same username (should fail)
		var userID int
		err := db.QueryRow(
			"INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING id",
			"user2@example.com", username, "password",
		).Scan(&userID)

		if err == nil {
			t.Fatal("Expected unique constraint violation for duplicate username, but insert succeeded")
		}

		t.Log("✅ Unique constraint properly rejected duplicate username")
	})

	// Test 3: Tag names are unique (ON CONFLICT DO UPDATE test)
	t.Run("TagUniqueConstraintWithConflictHandling", func(t *testing.T) {
		ctx := context.Background()
		tagRepo := NewTagRepository(db)
		activityRepo := NewActivityRepository(db, tagRepo)

		userID := createIntegrationTestUser(t, db, "tagtest@example.com", "taguser")

		// Create first activity with "outdoor" tag
		activity1 := &models.Activity{
			UserID:       userID,
			ActivityType: "running",
			Title:        "Run 1",
			ActivityDate: time.Now(),
		}
		tags1 := []*models.Tag{{Name: "outdoor"}}
		err := activityRepo.CreateWithTags(ctx, activity1, tags1)
		if err != nil {
			t.Fatalf("Failed to create first activity: %v", err)
		}

		// Create second activity with same "outdoor" tag
		// Should reuse existing tag due to ON CONFLICT DO UPDATE
		activity2 := &models.Activity{
			UserID:       userID,
			ActivityType: "cycling",
			Title:        "Ride 1",
			ActivityDate: time.Now(),
		}
		tags2 := []*models.Tag{{Name: "outdoor"}}
		err = activityRepo.CreateWithTags(ctx, activity2, tags2)
		if err != nil {
			t.Fatalf("Failed to create second activity: %v", err)
		}

		// Verify: Only ONE "outdoor" tag exists in database
		var tagCount int
		err = db.QueryRow("SELECT COUNT(*) FROM tags WHERE name = 'outdoor'").Scan(&tagCount)
		if err != nil {
			t.Fatalf("Failed to count tags: %v", err)
		}

		if tagCount != 1 {
			t.Errorf("Expected 1 'outdoor' tag, got %d (ON CONFLICT should reuse existing tag)", tagCount)
		}

		// Verify: Both activities are linked to the same tag
		exists, tagID := verifyTagExists(t, db, "outdoor")
		if !exists {
			t.Fatal("outdoor tag should exist")
		}

		linked1 := verifyActivityTagLink(t, db, activity1.ID, tagID)
		linked2 := verifyActivityTagLink(t, db, activity2.ID, tagID)

		if !linked1 || !linked2 {
			t.Error("Both activities should be linked to the same outdoor tag")
		}

		t.Log("✅ ON CONFLICT DO UPDATE properly reused existing tag")
	})
}

// TestIntegration_ComplexQueryWithJoins tests querying activities with tags
func TestIntegration_ComplexQueryWithJoins(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	userID := createIntegrationTestUser(t, db, "joins@example.com", "joinsuser")

	// Create activities with varying tag counts
	activity1 := &models.Activity{
		UserID:       userID,
		ActivityType: "running",
		Title:        "Morning Run",
		ActivityDate: time.Now(),
	}
	tags1 := []*models.Tag{{Name: "outdoor"}, {Name: "cardio"}}
	_ = activityRepo.CreateWithTags(ctx, activity1, tags1)

	activity2 := &models.Activity{
		UserID:       userID,
		ActivityType: "yoga",
		Title:        "Evening Yoga",
		ActivityDate: time.Now(),
	}
	tags2 := []*models.Tag{{Name: "indoor"}}
	_ = activityRepo.CreateWithTags(ctx, activity2, tags2)

	activity3 := &models.Activity{
		UserID:       userID,
		ActivityType: "swimming",
		Title:        "Pool Session",
		ActivityDate: time.Now(),
	}
	// No tags
	_ = activityRepo.Create(ctx, nil, activity3)

	// Query activities with tags
	activities, err := activityRepo.GetActivitiesWithTags(ctx, userID, models.ActivityFilters{})
	if err != nil {
		t.Fatalf("GetActivitiesWithTags failed: %v", err)
	}

	// Verify: Got all 3 activities
	if len(activities) != 3 {
		t.Fatalf("Expected 3 activities, got %d", len(activities))
	}

	// Verify: Activity with multiple tags has correct tag count
	for _, act := range activities {
		if act.Title == "Morning Run" {
			if len(act.Tags) != 2 {
				t.Errorf("Expected 2 tags for Morning Run, got %d", len(act.Tags))
			}
		}
		if act.Title == "Evening Yoga" {
			if len(act.Tags) != 1 {
				t.Errorf("Expected 1 tag for Evening Yoga, got %d", len(act.Tags))
			}
		}
		if act.Title == "Pool Session" {
			if len(act.Tags) != 0 {
				t.Errorf("Expected 0 tags for Pool Session, got %d", len(act.Tags))
			}
		}
	}

	t.Log("✅ Complex JOIN query properly retrieved activities with their tags")
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
