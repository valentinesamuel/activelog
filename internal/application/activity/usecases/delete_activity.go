package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// DeleteActivityUseCase handles activity deletion
type DeleteActivityUseCase struct {
	repo repository.ActivityRepository
}

// NewDeleteActivityUseCase creates a new instance
func NewDeleteActivityUseCase(repo repository.ActivityRepository) *DeleteActivityUseCase {
	return &DeleteActivityUseCase{repo: repo}
}

// Execute deletes an activity
func (uc *DeleteActivityUseCase) Execute(
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

	// Delete activity
	if err := uc.repo.Delete(ctx, tx, activityID, userID); err != nil {
		return nil, fmt.Errorf("failed to delete activity: %w", err)
	}

	// Return result
	return map[string]interface{}{
		"deleted":     true,
		"activity_id": activityID,
	}, nil
}
