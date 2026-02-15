package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// DeleteActivityInput defines the typed input for DeleteActivityUseCase
type DeleteActivityInput struct {
	UserID     int
	ActivityID int
}

// DeleteActivityOutput defines the typed output for DeleteActivityUseCase
type DeleteActivityOutput struct {
	Deleted    bool
	ActivityID int
}

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

// Execute deletes an activity (typed version)
// Decision: Use service for business logic checks, repo is available if needed
func (uc *DeleteActivityUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input DeleteActivityInput,
) (DeleteActivityOutput, error) {
	// DECISION: Use service for delete operations because we need:
	// - Ownership verification
	// - Business policy checks (e.g., preventing deletion of old activities)
	// - Cascade deletion handling
	// Alternative: Could use repo directly for simple hard deletes without checks
	err := uc.service.DeleteActivity(ctx, tx, input.UserID, input.ActivityID)
	if err != nil {
		return DeleteActivityOutput{}, fmt.Errorf("failed to delete activity: %w", err)
	}

	return DeleteActivityOutput{
		Deleted:    true,
		ActivityID: input.ActivityID,
	}, nil
}
