package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// CreateActivityInput defines the typed input for CreateActivityUseCase
type CreateActivityInput struct {
	UserID  int
	Request *models.CreateActivityRequest
}

// CreateActivityOutput defines the typed output for CreateActivityUseCase
type CreateActivityOutput struct {
	Activity   *models.Activity
	ActivityID int64
}

// CreateActivityUseCase handles activity creation
// Has access to both service (for business logic) and repository (for simple operations)
// The use case decides which one to use based on the operation's needs
type CreateActivityUseCase struct {
	service service.ActivityServiceInterface       // For operations requiring business logic
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
func (uc *CreateActivityUseCase) RequiresTransaction() bool {
	return true
}

// Execute creates a new activity (typed version)
// Decision: Use service for business logic validation, repo is available if needed
func (uc *CreateActivityUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input CreateActivityInput,
) (CreateActivityOutput, error) {
	if input.Request == nil {
		return CreateActivityOutput{}, fmt.Errorf("request is required")
	}

	// DECISION: Use service to create operations because we need business logic validation
	// - Validates date not in future
	// - Validates duration is reasonable
	// - Validates distance is positive
	// Alternative: Could use repo directly if we wanted to skip validation (not recommended for writes)
	activity, err := uc.service.CreateActivity(ctx, tx, input.UserID, input.Request)
	if err != nil {
		return CreateActivityOutput{}, fmt.Errorf("failed to create activity: %w", err)
	}

	return CreateActivityOutput{
		Activity:   activity,
		ActivityID: activity.ID,
	}, nil
}
