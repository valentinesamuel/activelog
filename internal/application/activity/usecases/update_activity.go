package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// UpdateActivityInput defines the typed input for UpdateActivityUseCase
type UpdateActivityInput struct {
	UserID     int
	ActivityID int
	Request    *models.UpdateActivityRequest
}

// UpdateActivityOutput defines the typed output for UpdateActivityUseCase
type UpdateActivityOutput struct {
	Activity *models.Activity
	Updated  bool
}

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

// Execute updates an existing activity (typed version)
// Decision: Use service for business logic validation, repo is available if needed
func (uc *UpdateActivityUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input UpdateActivityInput,
) (UpdateActivityOutput, error) {
	if input.Request == nil {
		return UpdateActivityOutput{}, fmt.Errorf("request is required")
	}

	// DECISION: Use service for update operations because we need:
	// - Ownership verification
	// - Business rule validation (date, duration, distance)
	// - Consistent error handling
	// Alternative: Could use repo directly for simple field updates without validation
	activity, err := uc.service.UpdateActivity(ctx, tx, input.UserID, input.ActivityID, input.Request)
	if err != nil {
		return UpdateActivityOutput{}, fmt.Errorf("failed to update activity: %w", err)
	}

	return UpdateActivityOutput{
		Activity: activity,
		Updated:  true,
	}, nil
}
