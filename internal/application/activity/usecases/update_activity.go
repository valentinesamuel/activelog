package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
)

// UpdateActivityUseCase handles activity updates
type UpdateActivityUseCase struct {
	repo repository.ActivityRepository
}

// NewUpdateActivityUseCase creates a new instance
func NewUpdateActivityUseCase(repo repository.ActivityRepository) *UpdateActivityUseCase {
	return &UpdateActivityUseCase{repo: repo}
}

// Execute updates an existing activity
func (uc *UpdateActivityUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract input
	activityID, ok := input["activity_id"].(int)
	if !ok {
		return nil, fmt.Errorf("activity_id is required")
	}

	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	req, ok := input["request"].(*models.UpdateActivityRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Fetch existing activity
	activity, err := uc.repo.GetByID(ctx, int64(activityID))
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}

	// Verify ownership
	if activity.UserID != userID {
		return nil, fmt.Errorf("activity not found or access denied")
	}

	// Apply updates
	if req.ActivityType != nil {
		activity.ActivityType = *req.ActivityType
	}
	if req.Title != nil {
		activity.Title = *req.Title
	}
	if req.Description != nil {
		activity.Description = *req.Description
	}
	if req.DurationMinutes != nil {
		activity.DurationMinutes = *req.DurationMinutes
	}
	if req.DistanceKm != nil {
		activity.DistanceKm = *req.DistanceKm
	}
	if req.CaloriesBurned != nil {
		activity.CaloriesBurned = *req.CaloriesBurned
	}
	if req.Notes != nil {
		activity.Notes = *req.Notes
	}
	if req.ActivityDate != nil {
		activity.ActivityDate = *req.ActivityDate
	}

	// Save updates
	if err := uc.repo.Update(ctx, tx, activityID, activity); err != nil {
		return nil, fmt.Errorf("failed to update activity: %w", err)
	}

	// Return result
	return map[string]interface{}{
		"activity": activity,
		"updated":  true,
	}, nil
}
