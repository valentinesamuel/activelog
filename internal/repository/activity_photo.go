package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/pkg/errors"
	"github.com/valentinesamuel/activelog/pkg/query"
)

type ActivityPhotoRepository struct {
	db           DBConn
	registry     *query.RelationshipRegistry
	activityRepo ActivityRepository
}

func NewActivityPhotoRepository(db DBConn, activityRepo *ActivityRepository) *ActivityPhotoRepository {
	registry := query.NewRelationshipRegistry("activity_photos")

	registry.Register((query.ManyToOneRelationship("photos", "activities", "activity_id")))

	return &ActivityPhotoRepository{
		registry:     registry,
		db:           db,
		activityRepo: *activityRepo,
	}
}

func (apr *ActivityPhotoRepository) GetRegistry() *query.RelationshipRegistry {
	return apr.registry
}

func (apr *ActivityPhotoRepository) Create(ctx context.Context, tx TxConn, activityPhoto *models.ActivityPhoto) error {
	query := `
		INSERT INTO activity_photos
		(activity_id, s3_key, thumbnail_key, content_type, file_size, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	// Use helper - automatically chooses tx or db
	row := QueryRowInTx(ctx, tx, apr.db, query,
		activityPhoto.ActivityID, activityPhoto.S3Key, activityPhoto.ThumbnailKey, activityPhoto.ContentType, activityPhoto.FileSize, activityPhoto.UploadedAt)

	err := row.Scan(&activityPhoto.ID, &activityPhoto.CreatedAt, &activityPhoto.UpdatedAt)
	if err != nil {
		return fmt.Errorf("❌ Error creating activity photo %w", err)
	}

	fmt.Println("✅ Activity photocreated successfully!")
	return nil
}

func (apr *ActivityPhotoRepository) GetByActivityID(ctx context.Context, id int) ([]*models.ActivityPhoto, error) {
	query := `
		SELECT *
		FROM activity_photos
		WHERE activity_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := apr.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("❌ Error listing activities: %w", err)
	}

	defer rows.Close()

	var activityPhotos []*models.ActivityPhoto

	for rows.Next() {
		activityPhoto := &models.ActivityPhoto{}
		err := rows.Scan(
			&activityPhoto.ID,
			&activityPhoto.ActivityID,
			&activityPhoto.ContentType,
			&activityPhoto.FileSize,
			&activityPhoto.S3Key,
			&activityPhoto.ThumbnailKey,
			&activityPhoto.UploadedAt,
			&activityPhoto.CreatedAt,
			&activityPhoto.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("❌ Error scanning activity: %w", err)
		}
		activityPhotos = append(activityPhotos, activityPhoto)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("❌ Error retrieving activity photos")
		return nil, err
	}

	fmt.Println("✅ Activities fetched successfully!")

	return activityPhotos, nil
}

func (apr *ActivityPhotoRepository) GetByID(ctx context.Context, id int) (*models.ActivityPhoto, error) {
	query := `
		SELECT *
		FROM activity_photos
		WHERE id = $1
	`

	activityPhoto := &models.ActivityPhoto{}

	err := apr.db.QueryRowContext(ctx, query, id).Scan(
		&activityPhoto.ID,
		&activityPhoto.ActivityID,
		&activityPhoto.ContentType,
		&activityPhoto.FileSize,
		&activityPhoto.S3Key,
		&activityPhoto.ThumbnailKey,
		&activityPhoto.UploadedAt,
		&activityPhoto.CreatedAt,
		&activityPhoto.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}

	if err != nil {
		return nil, &errors.DatabaseError{
			Op:    "SELECT",
			Table: "activity_photos",
			Err:   err,
		}
	}

	fmt.Println("✅ Activity photo fetched successfully!")

	return activityPhoto, nil
}

func (apr *ActivityPhotoRepository) Delete(ctx context.Context, tx TxConn, id int, userID int) error {
	query := "DELETE FROM activity_photos WHERE id = $1"

	// Use helper - automatically chooses tx or db
	result, err := ExecInTx(ctx, tx, apr.db, query, id, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("❌ Activity photo not found")
	}

	return nil
}
