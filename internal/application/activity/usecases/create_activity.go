package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
)

// CreateActivityUseCase handles activity creation
type CreateActivityUseCase struct {
	repo repository.ActivityRepository
}

// NewCreateActivityUseCase creates a new instance
func NewCreateActivityUseCase(repo repository.ActivityRepository) *CreateActivityUseCase {
	return &CreateActivityUseCase{repo: repo}
}

// Execute creates a new activity
func (uc *CreateActivityUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract input
	req, ok := input["request"].(*models.CreateActivityRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	// Create activity entity
	activity := &models.Activity{
		UserID:          userID,
		ActivityType:    req.ActivityType,
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		DistanceKm:      req.DistanceKm,
		CaloriesBurned:  req.CaloriesBurned,
		Notes:           req.Notes,
		ActivityDate:    req.ActivityDate,
	}

	// Create in repository
	if err := uc.repo.Create(ctx, tx, activity); err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	// Return result
	return map[string]interface{}{
		"activity":    activity,
		"activity_id": activity.ID,
	}, nil
}
