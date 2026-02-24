package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/pkg/errors"
	"github.com/valentinesamuel/activelog/pkg/query"
)

type ActivityRepository struct {
	db       DBConn
	tagRepo  *TagRepository
	registry *query.RelationshipRegistry
}

type ActivityStats struct {
	TotalActivities int
	TotalDuration   int
	TotalDistance   float64
	TotalCalories   int
	ActivityTypes   map[string]int
}

func NewActivityRepository(db DBConn, tagRepo *TagRepository) *ActivityRepository {
	// Initialize RelationshipRegistry for auto-JOIN support
	registry := query.NewRelationshipRegistry("activities")

	// Register Many-to-Many relationship: activities <-> tags
	// WithConditions ensures soft-deleted tags/junctions are auto-excluded from JOINs
	registry.Register(query.ManyToManyRelationship(
		"tags",          // Relationship name (users write: tags.name)
		"tags",          // Target table
		"activity_tags", // Junction table
		"activity_id",   // FK to activities
		"tag_id",        // FK to tags
	).WithConditions(
		query.AdditionalCondition{Column: "tags.deleted_at", Operator: "eq", Value: nil},
		query.AdditionalCondition{Column: "activity_tags.deleted_at", Operator: "eq", Value: nil},
	))

	// Register Many-to-One relationship: activities -> users
	// Name matches table name so "users.col" maps correctly to SQL "users.col"
	registry.Register(query.ManyToOneRelationship(
		"users",   // Relationship name matches table name for SQL WHERE correctness
		"users",   // Target table
		"user_id", // FK in activities table
	))

	return &ActivityRepository{
		db:       db,
		tagRepo:  tagRepo,
		registry: registry,
	}
}

// GetRegistry returns the RelationshipRegistry for this repository (v3.0)
// Used by RegistryManager for cross-registry path resolution
func (ar *ActivityRepository) GetRegistry() *query.RelationshipRegistry {
	return ar.registry
}

// Create creates a new activity
// tx is optional - if nil, uses direct DB connection; if provided, uses the transaction
func (ar *ActivityRepository) Create(ctx context.Context, tx TxConn, activity *models.Activity) error {
	query := `
		INSERT INTO activities
		(user_id, activity_type, title, description, duration_minutes, distance_km, calories_burned, notes, activity_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	// Use helper - automatically chooses tx or db
	row := QueryRowInTx(ctx, tx, ar.db, query,
		activity.UserID, activity.ActivityType, activity.Title, activity.Description,
		activity.DurationMinutes, activity.DistanceKm, activity.CaloriesBurned,
		activity.Notes, activity.ActivityDate)

	err := row.Scan(&activity.ID, &activity.CreatedAt, &activity.UpdatedAt)
	if err != nil {
		return fmt.Errorf("❌ Error creating activity %w", err)
	}

	fmt.Println("✅ Activity created successfully!")
	return nil
}

func (ar *ActivityRepository) GetByID(ctx context.Context, id int64) (*models.Activity, error) {
	query := `
		SELECT id, user_id, activity_type, title, description, duration_minutes, distance_km, calories_burned, notes, activity_date, created_at, updated_at, deleted_at
		FROM activities
		WHERE id = $1
	`

	activity := &models.Activity{}

	err := ar.db.QueryRowContext(ctx, query, id).Scan(
		&activity.ID,
		&activity.UserID,
		&activity.ActivityType,
		&activity.Title,
		&activity.Description,
		&activity.DurationMinutes,
		&activity.DistanceKm,
		&activity.CaloriesBurned,
		&activity.Notes,
		&activity.ActivityDate,
		&activity.CreatedAt,
		&activity.UpdatedAt,
		&activity.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}

	if err != nil {
		return nil, &errors.DatabaseError{
			Op:    "SELECT",
			Table: "activities",
			Err:   err,
		}
	}

	fmt.Println("✅ Activity fetched successfully!")

	return activity, nil
}

func (ar *ActivityRepository) ListByUser(ctx context.Context, UserID int) ([]*models.Activity, error) {
	query := `
		SELECT id, user_id, activity_type, title, description, duration_minutes,
			distance_km, calories_burned, notes, activity_date, created_at, updated_at, deleted_at
		FROM activities
		WHERE user_id = $1
		ORDER BY activity_date DESC
	`

	rows, err := ar.db.QueryContext(ctx, query, UserID)
	if err != nil {
		return nil, fmt.Errorf("❌ Error listing activities: %w", err)
	}

	defer rows.Close()

	var activities []*models.Activity

	for rows.Next() {
		activity := &models.Activity{}
		err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.ActivityType,
			&activity.Title,
			&activity.Description,
			&activity.DurationMinutes,
			&activity.DistanceKm,
			&activity.CaloriesBurned,
			&activity.Notes,
			&activity.ActivityDate,
			&activity.CreatedAt,
			&activity.UpdatedAt,
			&activity.DeletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("❌ Error scanning activity: %w", err)
		}
		activities = append(activities, activity)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("❌ Error listing activities")
		return nil, err
	}

	fmt.Println("✅ Activities fetched successfully!")

	return activities, nil
}

func (ar *ActivityRepository) Count(userID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM activities WHERE user_id = $1"
	err := ar.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// Update updates an existing activity
// tx is optional - if nil, uses direct DB connection; if provided, uses the transaction
func (ar *ActivityRepository) Update(ctx context.Context, tx TxConn, id int, activity *models.Activity) error {
	query := `
		UPDATE activities
		SET activity_type = $1, title = $2, description = $3,
			duration_minutes = $4, distance_km = $5, calories_burned = $6,
			notes = $7, activity_date = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9 AND user_id = $10
		RETURNING updated_at
	`

	// Use helper - automatically chooses tx or db
	row := QueryRowInTx(ctx, tx, ar.db, query,
		activity.ActivityType,
		activity.Title,
		activity.Description,
		activity.DurationMinutes,
		activity.DistanceKm,
		activity.CaloriesBurned,
		activity.Notes,
		activity.ActivityDate,
		id,
		activity.UserID,
	)

	err := row.Scan(&activity.UpdatedAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("❌ Activity not found")
	}

	return err
}

// Delete deletes an activity
// tx is optional - if nil, uses direct DB connection; if provided, uses the transaction
func (ar *ActivityRepository) Delete(ctx context.Context, tx TxConn, id int, userID int) error {
	// query := "DELETE FROM activities WHERE id = $1 AND user_id = $2"
	query := "UPDATE activities set deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND user_id = $2"

	// Use helper - automatically chooses tx or db
	result, err := ExecInTx(ctx, tx, ar.db, query, id, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("❌ Activity not found")
	}

	return nil
}

func (r *ActivityRepository) GetStats(userID int, startDate, endDate *time.Time) (*ActivityStats, error) {

	query := `
	SELECT 
		COUNT(*) as total,
		COALESCE(SUM(duration_minutes), 0) as total_duration,
		COALESCE(SUM(distance_km), 0) as total_distance,
		COALESCE(SUM(calories_burned), 0) as total_calories
	FROM activities
	WHERE user_id = $1
	`

	args := []interface{}{userID}
	argsCount := 1

	if startDate != nil {
		argsCount++
		query += fmt.Sprintf(" AND activity_date >= $%d", argsCount)
		args = append(args, startDate)
	}

	if endDate != nil {
		argsCount++
		query += fmt.Sprintf(" AND activity_date <= $%d", argsCount)
		args = append(args, endDate)
	}

	stats := &ActivityStats{
		ActivityTypes: make(map[string]int),
	}

	err := r.db.QueryRow(query, args...).Scan(
		&stats.TotalActivities,
		&stats.TotalDuration,
		&stats.TotalDistance,
		&stats.TotalCalories,
	)

	if err != nil {
		return nil, err
	}

	typeQuery := `
		SELECT activity_type, COUNT(*)
		FROM activities
		WHERE user_id = $1
	`

	typeArgs := []interface{}{userID}
	typeArgCount := 1

	if startDate != nil {
		typeArgCount++
		typeQuery += fmt.Sprintf(" AND activity_date >= $%d", typeArgCount)
		typeArgs = append(typeArgs, startDate)
	}
	if endDate != nil {
		typeArgCount++
		typeQuery += fmt.Sprintf(" AND activity_date <= $%d", typeArgCount)
		typeArgs = append(typeArgs, endDate)
	}

	typeQuery += " GROUP BY activity_type"

	rows, err := r.db.Query(typeQuery, typeArgs...)
	if err != nil {
		return stats, nil
	}

	defer rows.Close()

	for rows.Next() {
		var activityType string
		var count int
		if err := rows.Scan(&activityType, &count); err == nil {
			stats.ActivityTypes[activityType] = count
		}
	}

	return stats, nil
}

// CreateWithTags creates an activity with associated tags in a transaction
// This demonstrates a multi-step operation that requires a transaction
func (ar *ActivityRepository) CreateWithTags(ctx context.Context, activity *models.Activity, tags []*models.Tag) error {
	// Use WithTransaction helper for automatic commit/rollback
	return WithTransaction(ctx, ar.db, func(tx TxConn) error {
		// 1. Insert activity
		activityQuery := `
			INSERT INTO activities
			(user_id, activity_type, title, description, duration_minutes, distance_km, calories_burned, notes, activity_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, created_at, updated_at
		`
		row := QueryRowInTx(ctx, tx, ar.db, activityQuery,
			activity.UserID, activity.ActivityType, activity.Title, activity.Description,
			activity.DurationMinutes, activity.DistanceKm, activity.CaloriesBurned,
			activity.Notes, activity.ActivityDate)

		if err := row.Scan(&activity.ID, &activity.CreatedAt, &activity.UpdatedAt); err != nil {
			return fmt.Errorf("failed to insert activity: %w", err)
		}

		// 2. Create tags and link them (within the same transaction)
		for _, tag := range tags {
			// Get or create tag
			tagQuery := `
				INSERT INTO tags (name)
				VALUES ($1)
				ON CONFLICT (name) DO UPDATE
				SET name = EXCLUDED.name
				RETURNING id
			`
			var tagID int
			row := QueryRowInTx(ctx, tx, ar.db, tagQuery, tag.Name)
			if err := row.Scan(&tagID); err != nil {
				return fmt.Errorf("failed to create tag: %w", err)
			}

			// Link activity to tag
			linkQuery := `
				INSERT INTO activity_tags (tag_id, activity_id)
				VALUES ($1, $2)
				ON CONFLICT (tag_id, activity_id) DO NOTHING
			`
			if _, err := ExecInTx(ctx, tx, ar.db, linkQuery, tagID, activity.ID); err != nil {
				return fmt.Errorf("failed to link activity to tag: %w", err)
			}
		}

		return nil // Commit happens automatically on success
	})
}

// scanActivity is a reusable function to scan a single activity row
// Used by the generic FindAndPaginate function for dynamic filtering
func (ar *ActivityRepository) scanActivity(rows *sql.Rows) (*models.Activity, error) {
	activity := &models.Activity{}
	err := rows.Scan(
		&activity.ID,
		&activity.UserID,
		&activity.ActivityType,
		&activity.Title,
		&activity.Description,
		&activity.DurationMinutes,
		&activity.DistanceKm,
		&activity.CaloriesBurned,
		&activity.Notes,
		&activity.ActivityDate,
		&activity.CreatedAt,
		&activity.UpdatedAt,
		&activity.DeletedAt,
	)
	return activity, err
}

// ListActivitiesWithQuery uses the new dynamic filtering pattern with QueryOptions
// This method leverages the generic FindAndPaginate function for flexible, type-safe queries.
//
// Supports automatic JOIN detection using natural column names:
//   - filter[tags.name]=cardio → Automatically JOINs tags table and filters by tag name
//   - filter[user.username]=john → Automatically JOINs users table and filters by username
//   - search[tags.name]=run → Automatically JOINs and searches tag names
//   - order[tags.name]=ASC → Automatically JOINs and orders by tag name
//
// Example usage in handler:
//
//	opts := &query.QueryOptions{
//	    Page: 1,
//	    Limit: 20,
//	    Filter: map[string]interface{}{
//	        "activity_type": "running",
//	        "user_id": 123,
//	        "tags.name": "cardio",  // Natural column name - auto-JOINs!
//	    },
//	    Search: map[string]interface{}{
//	        "title": "morning",
//	        "tags.name": "run",     // Auto-JOINs for search too!
//	    },
//	    Order: map[string]string{
//	        "created_at": "DESC",
//	        "tags.name": "ASC",     // Auto-JOINs for ordering!
//	    },
//	}
//	result, err := repo.ListActivitiesWithQuery(ctx, opts)
func (ar *ActivityRepository) ListActivitiesWithQuery(
	ctx context.Context,
	opts *query.QueryOptions,
) (*query.PaginatedResult, error) {
	// Auto-generate JOINs based on relationship column names
	// The registry detects columns like "tags.name" and "user.username"
	// and automatically generates the appropriate JOINs
	joins := ar.registry.GenerateJoins(opts)

	// Use the generic FindAndPaginate function with auto-generated JOINs
	return FindAndPaginate[models.Activity](
		ctx,
		ar.db,
		"activities",
		opts,
		ar.scanActivity,
		joins...,
	)
}
