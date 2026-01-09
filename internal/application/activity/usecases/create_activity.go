package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// CreateActivityUseCase handles activity creation
// Has access to both service (for business logic) and repository (for simple operations)
// The use case decides which one to use based on the operation's needs
type CreateActivityUseCase struct {
	service service.ActivityServiceInterface      // For operations requiring business logic
	repo    repository.ActivityRepositoryInterface // For simple operations or when service not needed
}

// NewCreateActivityUseCase creates a new instance with both service and repository
// The use case will decide which one to use based on what it needs
func NewCreateActivityUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityRepositoryInterface,
) *CreateActivityUseCase {
	return &CreateActivityUseCase{
		service: svc,
		repo:    repo,
	}
}

// RequiresTransaction indicates this use case needs a transaction
// Write operations (CREATE) must run within a transaction for data integrity
func (uc *CreateActivityUseCase) RequiresTransaction() bool {
	return true
}

// Execute creates a new activity
// Decision: Use service for business logic validation, repo is available if needed
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

	// DECISION: Use service for create operations because we need business logic validation
	// - Validates date not in future
	// - Validates duration is reasonable
	// - Validates distance is positive
	// Alternative: Could use repo directly if we wanted to skip validation (not recommended for writes)
	activity, err := uc.service.CreateActivity(ctx, tx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	// Return result
	return map[string]interface{}{
		"activity":    activity,
		"activity_id": activity.ID,
	}, nil
}
