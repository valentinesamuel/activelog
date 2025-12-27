package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/pkg/errors"
)

type ActivityRepository struct {
	db *sql.DB
}

type ActivityStats struct {
	TotalActivities int
	TotalDuration   int
	TotalDistance   float64
	TotalCalories   int
	ActivityTypes   map[string]int
}

func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{
		db: db,
	}
}

func (ar *ActivityRepository) Create(ctx context.Context, activity *models.Activity) error {
	query := `
		INSERT INTO activities 
		(user_id, activity_type, title, description, duration_minutes, distance_km, calories_burned, notes, activity_date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := ar.db.QueryRowContext(ctx, query, activity.UserID, activity.ActivityType, activity.Title, activity.Description, activity.DurationMinutes, activity.DistanceKm, activity.CaloriesBurned, activity.Notes, activity.ActivityDate).Scan(&activity.ID, &activity.CreatedAt, &activity.UpdatedAt)

	if err != nil {
		return fmt.Errorf("❌ Error creating activity %w", err)
	}

	fmt.Println("✅ Activity created successfully!")

	return nil
}

func (ar *ActivityRepository) GetByID(ctx context.Context, id int64) (*models.Activity, error) {
	query := `
		SELECT id, user_id, activity_type, title, description, duration_minutes, distance_km, calories_burned, notes, activity_date, created_at, updated_at
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
			distance_km, calories_burned, notes, activity_date, created_at, updated_at
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

func (ar *ActivityRepository) ListByUserWithFilters(UserID int, filters models.ActivityFilters) ([]*models.Activity, error) {
	query := `
		SELECT id, user_id, activity_type, title, description, duration_minutes,
			distance_km, calories_burned, notes, activity_date, created_at, updated_at
		FROM activities
		WHERE user_id = $1
	`

	args := []interface{}{UserID}
	argsCount := 1

	// add activity filter type
	if filters.ActivityType != "" {
		argsCount++
		query += fmt.Sprintf(" AND activity_type = $%d", argsCount)
		args = append(args, filters.ActivityType)
	}

	if filters.StartDate != nil {
		argsCount++
		query += fmt.Sprintf(" AND activity_date >= $%d", argsCount)
		args = append(args, filters.StartDate)
	}

	if filters.EndDate != nil {
		argsCount++
		query += fmt.Sprintf(" AND activity_date <= $%d", argsCount)
		args = append(args, filters.EndDate)
	}

	query += " ORDER BY activity_date DESC"

	// add pagination
	if filters.Limit > 0 {
		argsCount++
		query += fmt.Sprintf(" LIMIT $%d", argsCount)
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		argsCount++
		query += fmt.Sprintf(" OFFSET $%d", argsCount)
		args = append(args, filters.Limit)
	}

	rows, err := ar.db.Query(query, args...)
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
		)

		if err != nil {
			return nil, fmt.Errorf("❌ Error scanning activity: %w", err)
		}
		activities = append(activities, activity)
	}

	fmt.Println("✅ Activities fetched successfully!")

	return activities, rows.Err()
}

func (ar *ActivityRepository) Count(userID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM activities WHERE user_id = $1"
	err := ar.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (r *ActivityRepository) Update(id int, activity *models.Activity) error {
	query := `
		UPDATE activities
		SET activity_type = $1, title = $2, description = $3,
			duration_minutes = $4, distance_km = $5, calories_burned = $6,
			notes = $7, activity_date = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9 AND user_id = $10
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
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
	).Scan(&activity.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("❌ Activity not found")
	}

	return err
}

func (r *ActivityRepository) Delete(id int, userID int) error {
	query := "DELETE FROM activities WHERE id = $1 AND user_id = $2"
	result, err := r.db.Exec(query, id, userID)
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
