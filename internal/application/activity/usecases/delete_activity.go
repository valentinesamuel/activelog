package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// DeleteActivityUseCase handles activity deletion
// Has access to both service (for business logic) and repository (for simple operations)
// The use case decides which one to use based on the operation's needs
type DeleteActivityUseCase struct {
	service service.ActivityServiceInterface       // For operations requiring business logic
	repo    repository.ActivityRepositoryInterface // For simple operations or when service not needed
}

// NewDeleteActivityUseCase creates a new instance with both service and repository
// The use case will decide which one to use based on what it needs
func NewDeleteActivityUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityRepositoryInterface,
) *DeleteActivityUseCase {
	return &DeleteActivityUseCase{
		service: svc,
		repo:    repo,
	}
}

// RequiresTransaction indicates this use case needs a transaction
// Write operations (DELETE) must run within a transaction for data integrity
func (uc *DeleteActivityUseCase) RequiresTransaction() bool {
	return true
}

// Execute deletes an activity
// Decision: Use service for business logic checks, repo is available if needed
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

	// DECISION: Use service for delete operations because we need:
	// - Ownership verification
	// - Business policy checks (e.g., preventing deletion of old activities)
	// - Cascade deletion handling
	// Alternative: Could use repo directly for simple hard deletes without checks
	err := uc.service.DeleteActivity(ctx, tx, userID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete activity: %w", err)
	}

	// Example of using both: Could use repo to verify deletion or clean up related data
	// remainingCount, _ := uc.repo.Count(userID)

	// Return result
	return map[string]interface{}{
		"deleted":     true,
		"activity_id": activityID,
	}, nil
}
