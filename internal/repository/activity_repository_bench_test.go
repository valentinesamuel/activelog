package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/valentinesamuel/activelog/internal/database"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository/testhelpers"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// setupBenchDB creates a test database for benchmarking
// Uses testcontainers to create a fresh database for each benchmark
func setupBenchDB(b *testing.B) (*database.LoggingDB, func()) {
	b.Helper()

	// Pass b directly - SetupTestDB accepts testing.TB interface
	db, cleanup := testhelpers.SetupTestDB(b)
	return db, cleanup
}

// createBenchUser creates a user for benchmark tests
func createBenchUser(b *testing.B, db *database.LoggingDB) int {
	b.Helper()

	var userID int
	err := db.QueryRow(
		"INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING id",
		fmt.Sprintf("bench%d@example.com", time.Now().UnixNano()),
		fmt.Sprintf("benchuser%d", time.Now().UnixNano()),
		"hashedpassword",
	).Scan(&userID)

	if err != nil {
		b.Fatalf("Failed to create bench user: %v", err)
	}

	return userID
}

// createBenchActivity creates an activity for benchmark tests
func createBenchActivity(b *testing.B, repo *ActivityRepository, userID int) *models.Activity {
	b.Helper()
	ctx := context.Background()

	activity := &models.Activity{
		UserID:          userID,
		ActivityType:    "running",
		Title:           "Bench Run",
		Description:     "Benchmark test activity",
		DurationMinutes: 30,
		DistanceKm:      5.0,
		CaloriesBurned:  250,
		Notes:           "Benchmark notes",
		ActivityDate:    time.Now(),
	}

	err := repo.Create(ctx, nil, activity)
	if err != nil {
		b.Fatalf("Failed to create bench activity: %v", err)
	}

	return activity
}

// ==================== BENCHMARKS ====================

// BenchmarkActivityRepository_Create benchmarks activity creation
func BenchmarkActivityRepository_Create(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	userID := createBenchUser(b, db)
	ctx := context.Background()

	// Reset timer to exclude setup time
	b.ResetTimer()
	b.ReportAllocs() // Track memory allocations

	for i := 0; i < b.N; i++ {
		activity := &models.Activity{
			UserID:          userID,
			ActivityType:    "running",
			Title:           fmt.Sprintf("Run %d", i),
			DurationMinutes: 30,
			DistanceKm:      5.0,
			ActivityDate:    time.Now(),
		}

		err := repo.Create(ctx, nil, activity)
		if err != nil {
			b.Fatalf("Create failed: %v", err)
		}
	}
}

// BenchmarkActivityRepository_GetByID benchmarks fetching activity by ID
func BenchmarkActivityRepository_GetByID(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	userID := createBenchUser(b, db)
	ctx := context.Background()

	// Create an activity to fetch
	activity := createBenchActivity(b, repo, userID)

	// Reset timer to exclude setup time
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := repo.GetByID(ctx, activity.ID)
		if err != nil {
			b.Fatalf("GetByID failed: %v", err)
		}
	}
}

// BenchmarkActivityRepository_ListByUser benchmarks listing activities
func BenchmarkActivityRepository_ListByUser(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	userID := createBenchUser(b, db)
	ctx := context.Background()

	// Create 10 activities for this user
	for i := 0; i < 10; i++ {
		_ = createBenchActivity(b, repo, userID)
	}

	// Reset timer to exclude setup time
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := repo.ListByUser(ctx, userID)
		if err != nil {
			b.Fatalf("ListByUser failed: %v", err)
		}
	}
}

// BenchmarkActivityRepository_CreateWithTags_1Tag benchmarks creating activity with 1 tag
func BenchmarkActivityRepository_CreateWithTags_1Tag(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	userID := createBenchUser(b, db)
	ctx := context.Background()

	// Reset timer to exclude setup time
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		activity := &models.Activity{
			UserID:       userID,
			ActivityType: "running",
			Title:        fmt.Sprintf("Run %d", i),
			ActivityDate: time.Now(),
		}

		tags := []*models.Tag{
			{Name: fmt.Sprintf("tag%d", i)},
		}

		err := repo.CreateWithTags(ctx, activity, tags)
		if err != nil {
			b.Fatalf("CreateWithTags failed: %v", err)
		}
	}
}

// BenchmarkActivityRepository_CreateWithTags_5Tags benchmarks creating activity with 5 tags
func BenchmarkActivityRepository_CreateWithTags_5Tags(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	userID := createBenchUser(b, db)
	ctx := context.Background()

	// Reset timer to exclude setup time
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		activity := &models.Activity{
			UserID:       userID,
			ActivityType: "running",
			Title:        fmt.Sprintf("Run %d", i),
			ActivityDate: time.Now(),
		}

		tags := []*models.Tag{
			{Name: fmt.Sprintf("tag1_%d", i)},
			{Name: fmt.Sprintf("tag2_%d", i)},
			{Name: fmt.Sprintf("tag3_%d", i)},
			{Name: fmt.Sprintf("tag4_%d", i)},
			{Name: fmt.Sprintf("tag5_%d", i)},
		}

		err := repo.CreateWithTags(ctx, activity, tags)
		if err != nil {
			b.Fatalf("CreateWithTags failed: %v", err)
		}
	}
}

// BenchmarkActivityRepository_CreateWithTags_10Tags benchmarks creating activity with 10 tags
func BenchmarkActivityRepository_CreateWithTags_10Tags(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	userID := createBenchUser(b, db)
	ctx := context.Background()

	// Reset timer to exclude setup time
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		activity := &models.Activity{
			UserID:       userID,
			ActivityType: "running",
			Title:        fmt.Sprintf("Run %d", i),
			ActivityDate: time.Now(),
		}

		tags := []*models.Tag{
			{Name: fmt.Sprintf("tag1_%d", i)},
			{Name: fmt.Sprintf("tag2_%d", i)},
			{Name: fmt.Sprintf("tag3_%d", i)},
			{Name: fmt.Sprintf("tag4_%d", i)},
			{Name: fmt.Sprintf("tag5_%d", i)},
			{Name: fmt.Sprintf("tag6_%d", i)},
			{Name: fmt.Sprintf("tag7_%d", i)},
			{Name: fmt.Sprintf("tag8_%d", i)},
			{Name: fmt.Sprintf("tag9_%d", i)},
			{Name: fmt.Sprintf("tag10_%d", i)},
		}

		err := repo.CreateWithTags(ctx, activity, tags)
		if err != nil {
			b.Fatalf("CreateWithTags failed: %v", err)
		}
	}
}

// BenchmarkActivityRepository_GetActivitiesWithTags benchmarks the JOIN query
func BenchmarkActivityRepository_GetActivitiesWithTags(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	userID := createBenchUser(b, db)
	ctx := context.Background()

	// Create 20 activities with varying tag counts
	for i := 0; i < 20; i++ {
		activity := &models.Activity{
			UserID:       userID,
			ActivityType: "running",
			Title:        fmt.Sprintf("Run %d", i),
			ActivityDate: time.Now(),
		}

		// Varying number of tags (0-3)
		numTags := i % 4
		tags := make([]*models.Tag, numTags)
		for j := 0; j < numTags; j++ {
			tags[j] = &models.Tag{Name: fmt.Sprintf("tag%d_%d", i, j)}
		}

		if len(tags) > 0 {
			err := repo.CreateWithTags(ctx, activity, tags)
			if err != nil {
				b.Fatalf("CreateWithTags failed: %v", err)
			}
		} else {
			err := repo.Create(ctx, nil, activity)
			if err != nil {
				b.Fatalf("Create failed: %v", err)
			}
		}
	}

	// Reset timer to exclude setup time
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := repo.GetActivitiesWithTags(ctx, userID, models.ActivityFilters{})
		if err != nil {
			b.Fatalf("GetActivitiesWithTags failed: %v", err)
		}
	}
}

// BenchmarkActivityRepository_GetActivitiesWithTags_N1Problem demonstrates the N+1 problem
// This benchmark shows what happens if you DON'T use JOINs (for comparison)
func BenchmarkActivityRepository_GetActivitiesWithTags_N1Problem(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	userID := createBenchUser(b, db)
	ctx := context.Background()

	// Create 20 activities with tags
	for i := 0; i < 20; i++ {
		activity := &models.Activity{
			UserID:       userID,
			ActivityType: "running",
			Title:        fmt.Sprintf("Run %d", i),
			ActivityDate: time.Now(),
		}

		tags := []*models.Tag{
			{Name: fmt.Sprintf("tag1_%d", i)},
			{Name: fmt.Sprintf("tag2_%d", i)},
		}

		err := repo.CreateWithTags(ctx, activity, tags)
		if err != nil {
			b.Fatalf("CreateWithTags failed: %v", err)
		}
	}

	// Reset timer to exclude setup time
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate N+1 problem: 1 query for activities + N queries for tags
		activities, err := repo.ListByUser(ctx, userID)
		if err != nil {
			b.Fatalf("ListByUser failed: %v", err)
		}

		// For each activity, fetch tags separately (N+1 problem!)
		for _, activity := range activities {
			_, err := tagRepo.GetTagsForActivity(ctx, int(activity.ID))
			if err != nil {
				b.Fatalf("GetTagsForActivity failed: %v", err)
			}
		}
	}
}

// BenchmarkComparison runs both approaches to compare
func BenchmarkComparison(b *testing.B) {
	b.Run("WithJOIN", BenchmarkActivityRepository_GetActivitiesWithTags)
	b.Run("N+1Problem", BenchmarkActivityRepository_GetActivitiesWithTags_N1Problem)
}
// ==================== DYNAMIC FILTERING BENCHMARKS ====================
// These benchmarks compare the old ActivityFilters approach vs the new
// QueryOptions dynamic filtering approach

// BenchmarkOldApproach_ListByUserWithFilters benchmarks the legacy filtering method
func BenchmarkOldApproach_ListByUserWithFilters(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)

	userID := createBenchUser(b, db)

	// Create 100 activities with varying types
	for i := 0; i < 100; i++ {
		createBenchActivity(b, repo, userID)
	}

	filters := models.ActivityFilters{
		ActivityType: "running",
		Limit:        20,
		Offset:       0,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := repo.ListByUserWithFilters(userID, filters)
		if err != nil {
			b.Fatalf("ListByUserWithFilters failed: %v", err)
		}
	}
}

// BenchmarkNewApproach_ListActivitiesWithQuery benchmarks the new dynamic filtering
func BenchmarkNewApproach_ListActivitiesWithQuery(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	ctx := context.Background()

	userID := createBenchUser(b, db)

	// Create 100 activities with varying types
	for i := 0; i < 100; i++ {
		createBenchActivity(b, repo, userID)
	}

	opts := &query.QueryOptions{
		Page:  1,
		Limit: 20,
		Filter: map[string]interface{}{
			"user_id":       userID,
			"activity_type": "running",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := repo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			b.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}
	}
}

// BenchmarkDynamicFiltering_WithPagination benchmarks pagination performance
func BenchmarkDynamicFiltering_WithPagination(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	ctx := context.Background()

	userID := createBenchUser(b, db)

	// Create 1000 activities
	for i := 0; i < 1000; i++ {
		createBenchActivity(b, repo, userID)
	}

	opts := &query.QueryOptions{
		Page:  10, // Middle page
		Limit: 50,
		Filter: map[string]interface{}{
			"user_id": userID,
		},
		Order: map[string]string{
			"created_at": "DESC",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := repo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			b.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}
	}
}

// BenchmarkDynamicFiltering_WithSearch benchmarks search performance
func BenchmarkDynamicFiltering_WithSearch(b *testing.B) {
	db, cleanup := setupBenchDB(b)
	defer cleanup()

	tagRepo := NewTagRepository(db)
	repo := NewActivityRepository(db, tagRepo)
	ctx := context.Background()

	userID := createBenchUser(b, db)

	// Create 500 activities
	for i := 0; i < 500; i++ {
		createBenchActivity(b, repo, userID)
	}

	opts := &query.QueryOptions{
		Page:  1,
		Limit: 20,
		Filter: map[string]interface{}{
			"user_id": userID,
		},
		Search: map[string]interface{}{
			"title": "Run", // Case-insensitive search
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := repo.ListActivitiesWithQuery(ctx, opts)
		if err != nil {
			b.Fatalf("ListActivitiesWithQuery failed: %v", err)
		}
	}
}

// BenchmarkComparison_OldVsNew runs both approaches for direct comparison
func BenchmarkComparison_OldVsNew(b *testing.B) {
	b.Run("OldApproach_ActivityFilters", BenchmarkOldApproach_ListByUserWithFilters)
	b.Run("NewApproach_QueryOptions", BenchmarkNewApproach_ListActivitiesWithQuery)
}
