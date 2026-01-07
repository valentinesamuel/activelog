package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// GetActivityUseCase handles fetching a single activity by ID
// This is a read-only operation and does NOT require a transaction
// Has access to both service and repository - decides which to use
type GetActivityUseCase struct {
	service service.ActivityServiceInterface      // For operations requiring business logic (can be nil for simple reads)
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

// No RequiresTransaction() method = defaults to non-transactional
// Read operations don't need transaction overhead for performance

// Execute retrieves a single activity
// Decision: Use repo directly for simple reads (no business logic needed)
func (uc *GetActivityUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract input
	activityID, ok := input["activity_id"].(int64)
	if !ok {
		return nil, fmt.Errorf("activity_id is required and must be int64")
	}

	// DECISION: Use repo directly for simple reads - no validation or business logic needed
	// Alternative: Could use service if we needed permission checks or data enrichment
	activity, err := uc.repo.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	// Return result
	return map[string]interface{}{
		"activity": activity,
	}, nil
}
