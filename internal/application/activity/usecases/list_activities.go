package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
	"github.com/valentinesamuel/activelog/pkg/query"
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

// Execute retrieves activities with dynamic filtering using QueryOptions
// This is the NEW approach that supports flexible filtering, searching, and sorting
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

	// Extract QueryOptions (required)
	queryOpts, exists := input["query_options"]
	if !exists {
		return nil, fmt.Errorf("query_options is required")
	}

	opts, ok := queryOpts.(*query.QueryOptions)
	if !ok {
		return nil, fmt.Errorf("invalid query_options type")
	}

	// SECURITY: Add user_id filter for multi-tenancy
	// This ensures users can only see their own activities
	opts.Filter["user_id"] = userID

	// Use dynamic filtering with RelationshipRegistry v3.0
	result, err := uc.repo.ListActivitiesWithQuery(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	// Return paginated result
	return map[string]interface{}{
		"result": result,
	}, nil
}
