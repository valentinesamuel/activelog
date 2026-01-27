package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// UpdateActivityUseCase handles activity updates
// Has access to both service (for business logic) and repository (for simple operations)
// The use case decides which one to use based on the operation's needs
type UpdateActivityUseCase struct {
	service service.ActivityServiceInterface       // For operations requiring business logic
	repo    repository.ActivityRepositoryInterface // For simple operations or when service not needed
}

// NewUpdateActivityUseCase creates a new instance with both service and repository
// The use case will decide which one to use based on what it needs
func NewUpdateActivityUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityRepositoryInterface,
) *UpdateActivityUseCase {
	return &UpdateActivityUseCase{
		service: svc,
		repo:    repo,
	}
}

// RequiresTransaction indicates this use case needs a transaction
// Write operations (UPDATE) must run within a transaction for data integrity
func (uc *UpdateActivityUseCase) RequiresTransaction() bool {
	return true
}

// Execute updates an existing activity
// Decision: Use service for business logic validation, repo is available if needed
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

	// DECISION: Use service for update operations because we need:
	// - Ownership verification
	// - Business rule validation (date, duration, distance)
	// - Consistent error handling
	// Alternative: Could use repo directly for simple field updates without validation
	activity, err := uc.service.UpdateActivity(ctx, tx, userID, activityID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update activity: %w", err)
	}

	// Example of using both: Could use repo for related data if needed
	// relatedActivities, _ := uc.repo.ListByUser(ctx, userID)

	// Return result
	return map[string]interface{}{
		"activity": activity,
		"updated":  true,
	}, nil
}
