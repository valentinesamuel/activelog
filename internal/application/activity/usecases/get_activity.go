package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// GetActivityInput defines the typed input for GetActivityUseCase
type GetActivityInput struct {
	ActivityID int64
}

// GetActivityOutput defines the typed output for GetActivityUseCase
type GetActivityOutput struct {
	Activity *models.Activity
}

// GetActivityUseCase handles fetching a single activity by ID
// This is a read-only operation and does NOT require a transaction
// Has access to both service and repository - decides which to use
type GetActivityUseCase struct {
	service service.ActivityServiceInterface       // For operations requiring business logic (can be nil for simple reads)
	repo    repository.ActivityRepositoryInterface // For simple read operations
}

// NewGetActivityUseCase creates a new instance with both service and repository
// For simple reads, service can be nil and use case will use repo directly
func NewGetActivityUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityRepositoryInterface,
) *GetActivityUseCase {
	return &GetActivityUseCase{
		service: svc,
		repo:    repo,
	}
}

// RequiresTransaction returns false - read operations don't need transactions
func (uc *GetActivityUseCase) RequiresTransaction() bool {
	return false
}

// Execute retrieves a single activity (typed version)
// Decision: Use repo directly for simple reads (no business logic needed)
func (uc *GetActivityUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input GetActivityInput,
) (GetActivityOutput, error) {
	// DECISION: Use repo directly for simple reads - no validation or business logic needed
	// Alternative: Could use service if we needed permission checks or data enrichment
	activity, err := uc.repo.GetByID(ctx, input.ActivityID)
	if err != nil {
		return GetActivityOutput{}, fmt.Errorf("failed to get activity: %w", err)
	}

	return GetActivityOutput{Activity: activity}, nil
}
