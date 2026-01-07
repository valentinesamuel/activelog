package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// ListActivitiesUseCase handles fetching activities with filters
// This is a read-only operation and does NOT require a transaction
// Has access to both service and repository - decides which to use
type ListActivitiesUseCase struct {
	service service.ActivityServiceInterface      // For operations requiring business logic (can be nil for simple reads)
	repo    repository.ActivityRepositoryInterface // For simple read operations
}

// NewListActivitiesUseCase creates a new instance with both service and repository
// For simple reads, service can be nil and use case will use repo directly
func NewListActivitiesUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityRepositoryInterface,
) *ListActivitiesUseCase {
	return &ListActivitiesUseCase{
		service: svc,
		repo:    repo,
	}
}

// No RequiresTransaction() method = defaults to non-transactional
// Read operations don't need transaction overhead for performance

// Execute retrieves activities with optional filters
// Decision: Use repo directly for simple list operations (no business logic needed)
func (uc *ListActivitiesUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract user ID (required)
	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	// Extract filters (optional)
	var filters models.ActivityFilters
	if filtersInput, exists := input["filters"]; exists {
		if f, ok := filtersInput.(models.ActivityFilters); ok {
			filters = f
		} else {
			return nil, fmt.Errorf("invalid filters type")
		}
	}

	// DECISION: Use repo directly for simple list queries - no validation or business logic needed
	// Alternative: Could use service if we needed permission filtering or data enrichment
	activities, err := uc.repo.ListByUserWithFilters(userID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	// Get total count for pagination
	totalCount, err := uc.repo.Count(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count activities: %w", err)
	}

	// Example of using both: Could use service for analytics enrichment if needed
	// stats, _ := uc.service.CalculateUserStats(ctx, userID)

	// Return result
	return map[string]interface{}{
		"activities": activities,
		"total":      totalCount,
		"count":      len(activities),
	}, nil
}
