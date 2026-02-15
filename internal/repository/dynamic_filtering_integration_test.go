package repository

import (
	"context"
	"testing"
	"time"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository/testhelpers"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// ==================== INTEGRATION TESTS FOR DYNAMIC FILTERING ====================
// These tests verify the new TypeORM-style dynamic filtering implementation
// using real PostgreSQL testcontainers

// TestIntegration_FindAndPaginate_Activities tests the generic FindAndPaginate
// function with the Activities entity
func TestIntegration_FindAndPaginate_Activities(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	// Create test user
	userID := createIntegrationTestUser(t, db, "filter@example.com", "filteruser")

	// Create test activities with different types
	activities := []struct {
		activityType string
		title        string
		duration     int
		distance     float64
	}{
		{"running", "Morning Run", 30, 5.0},
		{"running", "Evening Run", 45, 7.5},
		{"cycling", "Bike Ride", 60, 15.0},
		{"swimming", "Pool Session", 40, 2.0},
		{"yoga", "Yoga Class", 50, 0.0},
	}

	for _, act := range activities {
		activity := &models.Activity{
			UserID:          userID,
			ActivityType:    act.activityType,
			Title:           act.title,
			DurationMinutes: act.duration,
			DistanceKm:      act.distance,
			ActivityDate:    time.Now(),
		}
		if err := activityRepo.Create(ctx, nil, activity); err != nil {
			t.Fatalf("Failed to create test activity: %v", err)
		}
	}

	t.Run("FilterByActivityType", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"user_id":       userID,
				"activity_type": "running",
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		// Verify: Got 2 running activities
		activities := result.Data.([]*models.Activity)
		if len(activities) != 2 {
			t.Errorf("Expected 2 running activities, got %d", len(activities))
		}

		// Verify: Both are running type
		for _, act := range activities {
			if act.ActivityType != "running" {
				t.Errorf("Expected activity_type 'running', got '%s'", act.ActivityType)
			}
		}

		// Verify: Pagination metadata
		if result.Meta.TotalRecords != 2 {
			t.Errorf("Expected total_records=2, got %d", result.Meta.TotalRecords)
		}
		if result.Meta.PageCount != 1 {
			t.Errorf("Expected page_count=1, got %d", result.Meta.PageCount)
		}
	})

	t.Run("FilterWithMultipleValues_IN_Clause", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"user_id":       userID,
				"activity_type": []string{"running", "cycling"},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		// Verify: Got 3 activities (2 running + 1 cycling)
		activities := result.Data.([]*models.Activity)
		if len(activities) != 3 {
			t.Errorf("Expected 3 activities (running/cycling), got %d", len(activities))
		}
	})

	t.Run("SearchByTitle", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"user_id": userID,
			},
			Search: map[string]interface{}{
				"title": "Run", // Should match "Morning Run" and "Evening Run"
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		// Verify: Got 2 activities with "Run" in title
		activities := result.Data.([]*models.Activity)
		if len(activities) != 2 {
			t.Errorf("Expected 2 activities with 'Run' in title, got %d", len(activities))
		}
	})

	t.Run("OrderByDuration_DESC", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"user_id": userID,
			},
			Order: map[string]string{
				"duration_minutes": "DESC",
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 5 {
			t.Errorf("Expected 5 activities, got %d", len(activities))
		}

		// Verify: Activities are ordered by duration DESC
		// Expected order: 60 (cycling), 50 (yoga), 45 (evening run), 40 (swimming), 30 (morning run)
		expectedDurations := []int{60, 50, 45, 40, 30}
		for i, act := range activities {
			if act.DurationMinutes != expectedDurations[i] {
				t.Errorf("Activity[%d]: expected duration %d, got %d", i, expectedDurations[i], act.DurationMinutes)
			}
		}
	})

	t.Run("Pagination_Page2", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  2,
			Limit: 2, // 2 items per page
			Filter: map[string]interface{}{
				"user_id": userID,
			},
			Order: map[string]string{
				"created_at": "ASC",
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 2 {
			t.Errorf("Expected 2 activities on page 2, got %d", len(activities))
		}

		// Verify: Pagination metadata
		if result.Meta.Page != 2 {
			t.Errorf("Expected page=2, got %d", result.Meta.Page)
		}
		if result.Meta.TotalRecords != 5 {
			t.Errorf("Expected total_records=5, got %d", result.Meta.TotalRecords)
		}
		if result.Meta.PageCount != 3 {
			t.Errorf("Expected page_count=3 (5 total / 2 per page), got %d", result.Meta.PageCount)
		}
		if result.Meta.PreviousPage != 1 {
			t.Errorf("Expected previous_page=1, got %v", result.Meta.PreviousPage)
		}
		if result.Meta.NextPage != 3 {
			t.Errorf("Expected next_page=3, got %v", result.Meta.NextPage)
		}
	})

	t.Run("EmptyResults", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"user_id":       userID,
				"activity_type": "nonexistent",
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 0 {
			t.Errorf("Expected 0 activities, got %d", len(activities))
		}

		// Verify: Metadata shows no results
		if result.Meta.TotalRecords != 0 {
			t.Errorf("Expected total_records=0, got %d", result.Meta.TotalRecords)
		}
		if result.Meta.PageCount != 0 {
			t.Errorf("Expected page_count=0, got %d", result.Meta.PageCount)
		}
		if result.Meta.PreviousPage != false {
			t.Errorf("Expected previous_page=false, got %v", result.Meta.PreviousPage)
		}
		if result.Meta.NextPage != false {
			t.Errorf("Expected next_page=false, got %v", result.Meta.NextPage)
		}
	})

	t.Run("CombineFilters_AND_Logic", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"user_id":       userID,
				"activity_type": "running",
				// Only the 45-minute evening run should match
			},
			Search: map[string]interface{}{
				"title": "Evening",
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 1 {
			t.Errorf("Expected 1 activity (Evening Run), got %d", len(activities))
		}

		if len(activities) > 0 && activities[0].Title != "Evening Run" {
			t.Errorf("Expected 'Evening Run', got '%s'", activities[0].Title)
		}
	})
}

// TestIntegration_FindAndPaginate_Tags tests the generic FindAndPaginate
// function with the Tags entity (simpler case without joins)
func TestIntegration_FindAndPaginate_Tags(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)

	// Create test tags
	tagNames := []string{"cardio", "strength", "flexibility", "outdoor", "indoor"}
	for _, name := range tagNames {
		_, err := db.Exec("INSERT INTO tags (name) VALUES ($1)", name)
		if err != nil {
			t.Fatalf("Failed to create test tag: %v", err)
		}
	}

	t.Run("ListAllTags", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
		}

		result, err := tagRepo.ListTagsWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListTagsWithQuery failed: %v", err)
		}

		tags := result.Data.([]*models.Tag)
		if len(tags) != 5 {
			t.Errorf("Expected 5 tags, got %d", len(tags))
		}
	})

	t.Run("SearchTagsByName", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Search: map[string]interface{}{
				"name": "door", // Should match "outdoor" and "indoor"
			},
		}

		result, err := tagRepo.ListTagsWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListTagsWithQuery failed: %v", err)
		}

		tags := result.Data.([]*models.Tag)
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags with 'door' in name, got %d", len(tags))
		}
	})

	t.Run("OrderByName_ASC", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Order: map[string]string{
				"name": "ASC",
			},
		}

		result, err := tagRepo.ListTagsWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListTagsWithQuery failed: %v", err)
		}

		tags := result.Data.([]*models.Tag)
		if len(tags) != 5 {
			t.Errorf("Expected 5 tags, got %d", len(tags))
		}

		// Verify: Tags are ordered alphabetically
		expectedOrder := []string{"cardio", "flexibility", "indoor", "outdoor", "strength"}
		for i, tag := range tags {
			if tag.Name != expectedOrder[i] {
				t.Errorf("Tag[%d]: expected '%s', got '%s'", i, expectedOrder[i], tag.Name)
			}
		}
	})

	t.Run("Pagination_Tags", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  2,
			Limit: 2,
			Order: map[string]string{
				"name": "ASC",
			},
		}

		result, err := tagRepo.ListTagsWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListTagsWithQuery failed: %v", err)
		}

		tags := result.Data.([]*models.Tag)
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags on page 2, got %d", len(tags))
		}

		// Verify: Got "indoor" and "outdoor" (3rd and 4th alphabetically)
		if tags[0].Name != "indoor" || tags[1].Name != "outdoor" {
			t.Errorf("Expected tags 'indoor' and 'outdoor', got '%s' and '%s'", tags[0].Name, tags[1].Name)
		}

		// Verify: Pagination metadata
		if result.Meta.PageCount != 3 {
			t.Errorf("Expected page_count=3 (5 total / 2 per page), got %d", result.Meta.PageCount)
		}
	})
}

// TestIntegration_TagFiltering_WithJoins tests automatic JOIN detection
// when filtering activities by tag names
func TestIntegration_TagFiltering_WithJoins(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	// Create test user
	userID := createIntegrationTestUser(t, db, "joins@example.com", "joinsuser")

	// Create activities with tags
	runActivity := &models.Activity{
		UserID:       userID,
		ActivityType: "running",
		Title:        "Morning Run",
		ActivityDate: time.Now(),
	}
	runTags := []*models.Tag{{Name: "cardio"}, {Name: "outdoor"}}
	if err := activityRepo.CreateWithTags(ctx, runActivity, runTags); err != nil {
		t.Fatalf("Failed to create run activity: %v", err)
	}

	yogaActivity := &models.Activity{
		UserID:       userID,
		ActivityType: "yoga",
		Title:        "Evening Yoga",
		ActivityDate: time.Now(),
	}
	yogaTags := []*models.Tag{{Name: "flexibility"}, {Name: "indoor"}}
	if err := activityRepo.CreateWithTags(ctx, yogaActivity, yogaTags); err != nil {
		t.Fatalf("Failed to create yoga activity: %v", err)
	}

	swimActivity := &models.Activity{
		UserID:       userID,
		ActivityType: "swimming",
		Title:        "Pool Session",
		ActivityDate: time.Now(),
	}
	swimTags := []*models.Tag{{Name: "cardio"}, {Name: "indoor"}}
	if err := activityRepo.CreateWithTags(ctx, swimActivity, swimTags); err != nil {
		t.Fatalf("Failed to create swim activity: %v", err)
	}

	t.Run("FilterBySingleTag", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"user_id":   userID,
				"tags.name": "cardio", // Natural column name - auto-JOINs!
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery with tag filter failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 2 {
			t.Errorf("Expected 2 activities with 'cardio' tag, got %d", len(activities))
		}

		// Verify: Got running and swimming (both have cardio tag)
		activityTypes := make(map[string]bool)
		for _, act := range activities {
			activityTypes[act.ActivityType] = true
		}

		if !activityTypes["running"] || !activityTypes["swimming"] {
			t.Error("Expected running and swimming activities")
		}
	})

	t.Run("FilterByMultipleTags_OR", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"user_id":   userID,
				"tags.name": []string{"outdoor", "flexibility"}, // Natural column name - auto-JOINs!
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery with multiple tags failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 2 {
			t.Errorf("Expected 2 activities (running OR yoga), got %d", len(activities))
		}
	})

	t.Run("CombineTagFilterWithActivityType", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"user_id":       userID,
				"activity_type": "running",
				"tags.name":     "cardio", // Natural column name - auto-JOINs!
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery with combined filters failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 1 {
			t.Errorf("Expected 1 activity (running + cardio), got %d", len(activities))
		}

		if len(activities) > 0 && activities[0].Title != "Morning Run" {
			t.Errorf("Expected 'Morning Run', got '%s'", activities[0].Title)
		}
	})
}

// TestIntegration_PaginationMetadata_EdgeCases tests pagination metadata
// calculation in various edge cases
func TestIntegration_PaginationMetadata_EdgeCases(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	userID := createIntegrationTestUser(t, db, "pagination@example.com", "paginationuser")

	// Create exactly 7 activities for testing various page sizes
	for i := 1; i <= 7; i++ {
		activity := &models.Activity{
			UserID:       userID,
			ActivityType: "test",
			Title:        "Test Activity",
			ActivityDate: time.Now(),
		}
		if err := activityRepo.Create(ctx, nil, activity); err != nil {
			t.Fatalf("Failed to create test activity: %v", err)
		}
	}

	testCases := []struct {
		name              string
		page              int
		limit             int
		expectedCount     int // Records on this page
		expectedTotal     int
		expectedPageCount int
		expectedPrevious  interface{} // int or false
		expectedNext      interface{} // int or false
	}{
		{
			name:              "FirstPage_WithNext",
			page:              1,
			limit:             3,
			expectedCount:     3,
			expectedTotal:     7,
			expectedPageCount: 3, // 7 / 3 = 3 pages
			expectedPrevious:  false,
			expectedNext:      2,
		},
		{
			name:              "MiddlePage_WithBoth",
			page:              2,
			limit:             3,
			expectedCount:     3,
			expectedTotal:     7,
			expectedPageCount: 3,
			expectedPrevious:  1,
			expectedNext:      3,
		},
		{
			name:              "LastPage_Partial",
			page:              3,
			limit:             3,
			expectedCount:     1, // Only 1 record on last page (7 % 3 = 1)
			expectedTotal:     7,
			expectedPageCount: 3,
			expectedPrevious:  2,
			expectedNext:      false,
		},
		{
			name:              "SinglePage_AllRecords",
			page:              1,
			limit:             10,
			expectedCount:     7,
			expectedTotal:     7,
			expectedPageCount: 1,
			expectedPrevious:  false,
			expectedNext:      false,
		},
		{
			name:              "ExactFit_NoPartial",
			page:              1,
			limit:             7,
			expectedCount:     7,
			expectedTotal:     7,
			expectedPageCount: 1,
			expectedPrevious:  false,
			expectedNext:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &query.QueryOptions{
				Page:  tc.page,
				Limit: tc.limit,
				Filter: map[string]interface{}{
					"user_id": userID,
				},
			}

			result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
			if err != nil {
				t.Fatalf("ListActivitiesWithQuery failed: %v", err)
			}

			activities := result.Data.([]*models.Activity)
			if len(activities) != tc.expectedCount {
				t.Errorf("Expected %d records on page, got %d", tc.expectedCount, len(activities))
			}

			if result.Meta.TotalRecords != tc.expectedTotal {
				t.Errorf("Expected total_records=%d, got %d", tc.expectedTotal, result.Meta.TotalRecords)
			}

			if result.Meta.PageCount != tc.expectedPageCount {
				t.Errorf("Expected page_count=%d, got %d", tc.expectedPageCount, result.Meta.PageCount)
			}

			if result.Meta.Page != tc.page {
				t.Errorf("Expected page=%d, got %d", tc.page, result.Meta.Page)
			}

			if result.Meta.Limit != tc.limit {
				t.Errorf("Expected limit=%d, got %d", tc.limit, result.Meta.Limit)
			}

			if result.Meta.Count != tc.expectedCount {
				t.Errorf("Expected count=%d, got %d", tc.expectedCount, result.Meta.Count)
			}

			if result.Meta.PreviousPage != tc.expectedPrevious {
				t.Errorf("Expected previous_page=%v, got %v", tc.expectedPrevious, result.Meta.PreviousPage)
			}

			if result.Meta.NextPage != tc.expectedNext {
				t.Errorf("Expected next_page=%v, got %v", tc.expectedNext, result.Meta.NextPage)
			}
		})
	}
}

// TestIntegration_MultiTenancy_UserIsolation verifies that user_id filtering
// properly isolates data between users
func TestIntegration_MultiTenancy_UserIsolation(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	// Create two users
	user1ID := createIntegrationTestUser(t, db, "user1@example.com", "user1")
	user2ID := createIntegrationTestUser(t, db, "user2@example.com", "user2")

	// Create activities for user 1
	for i := 0; i < 3; i++ {
		activity := &models.Activity{
			UserID:       user1ID,
			ActivityType: "running",
			Title:        "User 1 Activity",
			ActivityDate: time.Now(),
		}
		if err := activityRepo.Create(ctx, nil, activity); err != nil {
			t.Fatalf("Failed to create user1 activity: %v", err)
		}
	}

	// Create activities for user 2
	for i := 0; i < 5; i++ {
		activity := &models.Activity{
			UserID:       user2ID,
			ActivityType: "cycling",
			Title:        "User 2 Activity",
			ActivityDate: time.Now(),
		}
		if err := activityRepo.Create(ctx, nil, activity); err != nil {
			t.Fatalf("Failed to create user2 activity: %v", err)
		}
	}

	// Query for user 1 - should only get 3 activities
	opts1 := &query.QueryOptions{
		Page:  1,
		Limit: 10,
		Filter: map[string]interface{}{
			"user_id": user1ID,
		},
	}

	result1, err := activityRepo.ListActivitiesWithQuery(ctx, opts1)
	if err != nil {
		t.Fatalf("ListActivitiesWithQuery for user1 failed: %v", err)
	}

	activities1 := result1.Data.([]*models.Activity)
	if len(activities1) != 3 {
		t.Errorf("User 1 should have 3 activities, got %d", len(activities1))
	}

	// Verify: All activities belong to user 1
	for _, act := range activities1 {
		if act.UserID != user1ID {
			t.Errorf("Found activity from different user: user_id=%d, expected %d", act.UserID, user1ID)
		}
		if act.ActivityType != "running" {
			t.Errorf("Expected running activity, got %s", act.ActivityType)
		}
	}

	// Query for user 2 - should only get 5 activities
	opts2 := &query.QueryOptions{
		Page:  1,
		Limit: 10,
		Filter: map[string]interface{}{
			"user_id": user2ID,
		},
	}

	result2, err := activityRepo.ListActivitiesWithQuery(ctx, opts2)
	if err != nil {
		t.Fatalf("ListActivitiesWithQuery for user2 failed: %v", err)
	}

	activities2 := result2.Data.([]*models.Activity)
	if len(activities2) != 5 {
		t.Errorf("User 2 should have 5 activities, got %d", len(activities2))
	}

	// Verify: All activities belong to user 2
	for _, act := range activities2 {
		if act.UserID != user2ID {
			t.Errorf("Found activity from different user: user_id=%d, expected %d", act.UserID, user2ID)
		}
		if act.ActivityType != "cycling" {
			t.Errorf("Expected cycling activity, got %s", act.ActivityType)
		}
	}

	t.Log("✅ Multi-tenancy test passed: Users can only see their own activities")
}

// ==================== OPERATOR-BASED FILTERING INTEGRATION TESTS (v1.1.0) ====================
// These tests verify the new operator-based filtering functionality with real database queries

// TestIntegration_OperatorFiltering_DateRanges tests date comparison operators
func TestIntegration_OperatorFiltering_DateRanges(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	userID := createIntegrationTestUser(t, db, "daterange@example.com", "dateuser")

	// Create activities with specific dates
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, -10), // 10 days ago
		now.AddDate(0, 0, -5),  // 5 days ago
		now.AddDate(0, 0, -2),  // 2 days ago
		now,                    // Today
		now.AddDate(0, 0, 2),   // 2 days from now
	}

	for i, date := range dates {
		activity := &models.Activity{
			UserID:       userID,
			ActivityType: "running",
			Title:        "Test Activity",
			DistanceKm:   float64(i + 1),
			ActivityDate: date,
		}
		if err := activityRepo.Create(ctx, nil, activity); err != nil {
			t.Fatalf("Failed to create test activity: %v", err)
		}
	}

	t.Run("GTE_Operator_DateFilter", func(t *testing.T) {
		// Filter activities from 3 days ago onwards
		cutoffDate := now.AddDate(0, 0, -3)
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "activity_date", Operator: "gte", Value: cutoffDate},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get: 2 days ago, today, 2 days from now = 3 activities
		if len(activities) != 3 {
			t.Errorf("Expected 3 activities with date >= 3 days ago, got %d", len(activities))
		}
	})

	t.Run("LTE_Operator_DateFilter", func(t *testing.T) {
		// Filter activities up to 3 days ago
		cutoffDate := now.AddDate(0, 0, -3)
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "activity_date", Operator: "lte", Value: cutoffDate},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get: 10 days ago, 5 days ago = 2 activities
		if len(activities) != 2 {
			t.Errorf("Expected 2 activities with date <= 3 days ago, got %d", len(activities))
		}
	})

	t.Run("DateRange_GTE_AND_LTE", func(t *testing.T) {
		// Filter activities between 7 days ago and 1 day ago
		startDate := now.AddDate(0, 0, -7)
		endDate := now.AddDate(0, 0, -1)
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "activity_date", Operator: "gte", Value: startDate},
				{Column: "activity_date", Operator: "lte", Value: endDate},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get: 5 days ago, 2 days ago = 2 activities
		if len(activities) != 2 {
			t.Errorf("Expected 2 activities in date range, got %d", len(activities))
		}
	})

	t.Run("GT_Operator_Exclusive", func(t *testing.T) {
		// Filter activities strictly after 2 days ago (exclusive)
		cutoffDate := now.AddDate(0, 0, -2)
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "activity_date", Operator: "gt", Value: cutoffDate},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get: today, 2 days from now = 2 activities (excludes the activity AT 2 days ago)
		if len(activities) != 2 {
			t.Errorf("Expected 2 activities with date > 2 days ago, got %d", len(activities))
		}
	})

	t.Run("LT_Operator_Exclusive", func(t *testing.T) {
		// Filter activities strictly before today (exclusive)
		cutoffDate := now
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "activity_date", Operator: "lt", Value: cutoffDate},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get: 10 days ago, 5 days ago, 2 days ago = 3 activities (excludes today)
		if len(activities) != 3 {
			t.Errorf("Expected 3 activities with date < today, got %d", len(activities))
		}
	})
}

// TestIntegration_OperatorFiltering_NumericRanges tests numeric comparison operators
func TestIntegration_OperatorFiltering_NumericRanges(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	userID := createIntegrationTestUser(t, db, "numeric@example.com", "numericuser")

	// Create activities with different distances and durations
	testData := []struct {
		distance float64
		duration int
	}{
		{2.5, 20},
		{5.0, 30},
		{7.5, 45},
		{10.0, 60},
		{12.5, 75},
	}

	for _, data := range testData {
		activity := &models.Activity{
			UserID:          userID,
			ActivityType:    "running",
			Title:           "Test Run",
			DistanceKm:      data.distance,
			DurationMinutes: data.duration,
			ActivityDate:    time.Now(),
		}
		if err := activityRepo.Create(ctx, nil, activity); err != nil {
			t.Fatalf("Failed to create test activity: %v", err)
		}
	}

	t.Run("GTE_Numeric_Distance", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "distance_km", Operator: "gte", Value: 7.5},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get: 7.5, 10.0, 12.5 = 3 activities
		if len(activities) != 3 {
			t.Errorf("Expected 3 activities with distance >= 7.5, got %d", len(activities))
		}

		// Verify all distances are >= 7.5
		for _, act := range activities {
			if act.DistanceKm < 7.5 {
				t.Errorf("Found activity with distance %.2f < 7.5", act.DistanceKm)
			}
		}
	})

	t.Run("LT_Numeric_Duration", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "duration_minutes", Operator: "lt", Value: 50},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get: 20, 30, 45 = 3 activities
		if len(activities) != 3 {
			t.Errorf("Expected 3 activities with duration < 50, got %d", len(activities))
		}

		// Verify all durations are < 50
		for _, act := range activities {
			if act.DurationMinutes >= 50 {
				t.Errorf("Found activity with duration %d >= 50", act.DurationMinutes)
			}
		}
	})

	t.Run("NumericRange_Distance_Between_5_AND_10", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "distance_km", Operator: "gte", Value: 5.0},
				{Column: "distance_km", Operator: "lte", Value: 10.0},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get: 5.0, 7.5, 10.0 = 3 activities
		if len(activities) != 3 {
			t.Errorf("Expected 3 activities with distance between 5 and 10, got %d", len(activities))
		}
	})

	t.Run("GT_And_LT_ExclusiveRange", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "distance_km", Operator: "gt", Value: 5.0},
				{Column: "distance_km", Operator: "lt", Value: 12.5},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get: 7.5, 10.0 = 2 activities (excludes 5.0 and 12.5)
		if len(activities) != 2 {
			t.Errorf("Expected 2 activities with 5.0 < distance < 12.5, got %d", len(activities))
		}
	})

	t.Run("NE_Operator_NotEqual", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "distance_km", Operator: "ne", Value: 7.5},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get: 2.5, 5.0, 10.0, 12.5 = 4 activities (excludes 7.5)
		if len(activities) != 4 {
			t.Errorf("Expected 4 activities with distance != 7.5, got %d", len(activities))
		}

		// Verify none have distance 7.5
		for _, act := range activities {
			if act.DistanceKm == 7.5 {
				t.Error("Found activity with distance 7.5, should be excluded")
			}
		}
	})
}

// TestIntegration_OperatorFiltering_BackwardCompatibility verifies that old and new
// syntax can be used together without conflicts
func TestIntegration_OperatorFiltering_BackwardCompatibility(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	userID := createIntegrationTestUser(t, db, "compat@example.com", "compatuser")

	// Create test activities
	now := time.Now()
	for i := 0; i < 10; i++ {
		activity := &models.Activity{
			UserID:       userID,
			ActivityType: "running",
			Title:        "Test Run",
			DistanceKm:   float64(i + 1),
			ActivityDate: now.AddDate(0, 0, -i),
		}
		if err := activityRepo.Create(ctx, nil, activity); err != nil {
			t.Fatalf("Failed to create test activity: %v", err)
		}
	}

	t.Run("LegacySyntax_Still_Works", func(t *testing.T) {
		// Old 2-level syntax should still work
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"user_id":       userID,
				"activity_type": "running",
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery with legacy syntax failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 10 {
			t.Errorf("Expected 10 activities with legacy syntax, got %d", len(activities))
		}
	})

	t.Run("MixedSyntax_Legacy_And_Operator", func(t *testing.T) {
		// Mix old equality filter with new operator filter
		cutoffDate := now.AddDate(0, 0, -5)
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			Filter: map[string]interface{}{
				"activity_type": "running", // Legacy syntax
			},
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "activity_date", Operator: "gte", Value: cutoffDate}, // Operator syntax
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery with mixed syntax failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get activities from last 5 days
		if len(activities) != 6 { // 0, 1, 2, 3, 4, 5 days ago
			t.Errorf("Expected 6 activities with mixed syntax, got %d", len(activities))
		}
	})

	t.Run("ExplicitEQ_Operator", func(t *testing.T) {
		// Explicit eq operator should work the same as legacy
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "activity_type", Operator: "eq", Value: "running"},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery with explicit eq failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 10 {
			t.Errorf("Expected 10 activities with explicit eq, got %d", len(activities))
		}
	})
}

// TestIntegration_OperatorFiltering_ComplexScenarios tests real-world complex filtering
func TestIntegration_OperatorFiltering_ComplexScenarios(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	userID := createIntegrationTestUser(t, db, "complex@example.com", "complexuser")

	// Create diverse activities
	now := time.Now()
	testActivities := []struct {
		activityType string
		distance     float64
		duration     int
		daysAgo      int
	}{
		{"running", 5.0, 30, 1},
		{"running", 10.0, 60, 3},
		{"running", 3.0, 20, 7},
		{"cycling", 20.0, 90, 2},
		{"cycling", 15.0, 60, 5},
		{"swimming", 2.0, 45, 4},
		{"swimming", 1.5, 30, 10},
	}

	for _, data := range testActivities {
		activity := &models.Activity{
			UserID:          userID,
			ActivityType:    data.activityType,
			Title:           data.activityType + " session",
			DistanceKm:      data.distance,
			DurationMinutes: data.duration,
			ActivityDate:    now.AddDate(0, 0, -data.daysAgo),
		}
		if err := activityRepo.Create(ctx, nil, activity); err != nil {
			t.Fatalf("Failed to create test activity: %v", err)
		}
	}

	t.Run("ComplexQuery_MultipleOperators", func(t *testing.T) {
		// Find running activities from last week with distance > 4km and duration < 50 minutes
		cutoffDate := now.AddDate(0, 0, -7)
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "activity_type", Operator: "eq", Value: "running"},
				{Column: "activity_date", Operator: "gte", Value: cutoffDate},
				{Column: "distance_km", Operator: "gt", Value: 4.0},
				{Column: "duration_minutes", Operator: "lt", Value: 50},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should match: 5km/30min from 1 day ago
		if len(activities) != 1 {
			t.Errorf("Expected 1 activity matching complex criteria, got %d", len(activities))
		}

		if len(activities) > 0 {
			act := activities[0]
			if act.ActivityType != "running" {
				t.Errorf("Expected running, got %s", act.ActivityType)
			}
			if act.DistanceKm <= 4.0 {
				t.Errorf("Expected distance > 4.0, got %.2f", act.DistanceKm)
			}
			if act.DurationMinutes >= 50 {
				t.Errorf("Expected duration < 50, got %d", act.DurationMinutes)
			}
		}
	})

	t.Run("FindActivities_LastWeek_LongDuration", func(t *testing.T) {
		// Activities from last week with duration >= 60 minutes
		cutoffDate := now.AddDate(0, 0, -7)
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "activity_date", Operator: "gte", Value: cutoffDate},
				{Column: "duration_minutes", Operator: "gte", Value: 60},
			},
			Order: map[string]string{
				"duration_minutes": "DESC",
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should match: cycling 90min from 2 days ago, running 60min from 3 days ago, cycling 60min from 5 days ago
		if len(activities) != 3 {
			t.Errorf("Expected 3 activities with duration >= 60 from last week, got %d", len(activities))
		}

		// Verify ordering by duration DESC
		if len(activities) == 3 {
			if activities[0].DurationMinutes != 90 {
				t.Errorf("Expected first activity duration 90, got %d", activities[0].DurationMinutes)
			}
		}
	})

	t.Run("DistanceRange_WithPagination", func(t *testing.T) {
		// Activities with distance between 2km and 15km, page 1 with limit 2
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 2,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "distance_km", Operator: "gte", Value: 2.0},
				{Column: "distance_km", Operator: "lte", Value: 15.0},
			},
			Order: map[string]string{
				"distance_km": "DESC",
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		// Should get first 2 of 5 matching activities (15km cycling, 10km running)
		if len(activities) != 2 {
			t.Errorf("Expected 2 activities on page 1, got %d", len(activities))
		}

		// Verify pagination metadata
		if result.Meta.TotalRecords != 5 {
			t.Errorf("Expected total 5 matching activities, got %d", result.Meta.TotalRecords)
		}
		if result.Meta.PageCount != 3 {
			t.Errorf("Expected 3 pages (5 records / 2 per page), got %d", result.Meta.PageCount)
		}
		if result.Meta.NextPage != 2 {
			t.Errorf("Expected next_page=2, got %v", result.Meta.NextPage)
		}
	})
}

// TestIntegration_OperatorFiltering_EdgeCases tests edge cases and boundary conditions
func TestIntegration_OperatorFiltering_EdgeCases(t *testing.T) {
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	activityRepo := NewActivityRepository(db, tagRepo)

	userID := createIntegrationTestUser(t, db, "edge@example.com", "edgeuser")

	// Create test activity
	activity := &models.Activity{
		UserID:          userID,
		ActivityType:    "test",
		Title:           "Edge Case Test",
		DistanceKm:      5.0,
		DurationMinutes: 30,
		ActivityDate:    time.Now(),
	}
	if err := activityRepo.Create(ctx, nil, activity); err != nil {
		t.Fatalf("Failed to create test activity: %v", err)
	}

	t.Run("EQ_ShouldMatch_ExactValue", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "distance_km", Operator: "eq", Value: 5.0},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 1 {
			t.Errorf("Expected 1 activity with exact distance match, got %d", len(activities))
		}
	})

	t.Run("GTE_IncludesBoundary", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "distance_km", Operator: "gte", Value: 5.0}, // Boundary value
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 1 {
			t.Errorf("GTE should include boundary value, got %d activities", len(activities))
		}
	})

	t.Run("GT_ExcludesBoundary", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "distance_km", Operator: "gt", Value: 5.0}, // Boundary value
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 0 {
			t.Errorf("GT should exclude boundary value, got %d activities", len(activities))
		}
	})

	t.Run("NE_ExcludesValue", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "distance_km", Operator: "ne", Value: 5.0},
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 0 {
			t.Errorf("NE should exclude the value, got %d activities", len(activities))
		}
	})

	t.Run("NoResults_OutOfRange", func(t *testing.T) {
		opts := &query.QueryOptions{
			Page:  1,
			Limit: 10,
			FilterConditions: []query.FilterCondition{
				{Column: "user_id", Operator: "eq", Value: userID},
				{Column: "distance_km", Operator: "gt", Value: 100.0}, // Way out of range
			},
		}

		result, err := activityRepo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			t.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}

		activities := result.Data.([]*models.Activity)
		if len(activities) != 0 {
			t.Errorf("Expected no results for out-of-range query, got %d", len(activities))
		}

		// Verify metadata
		if result.Meta.TotalRecords != 0 {
			t.Errorf("Expected total_records=0, got %d", result.Meta.TotalRecords)
		}
	})
}
