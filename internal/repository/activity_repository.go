package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
)

type ActivityRepository struct {
	db *sql.DB
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
		return nil, fmt.Errorf("❌ Activity not found")
	}

	if err != nil {
		return nil, fmt.Errorf("❌ Error fetching activity: %w", err)
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
