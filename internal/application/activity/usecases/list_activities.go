package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/valentinesamuel/activelog/internal/cache/types"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// ListActivitiesInput defines the typed input for ListActivitiesUseCase
type ListActivitiesInput struct {
	UserID       int
	QueryOptions *query.QueryOptions
}

// ListActivitiesOutput defines the typed output for ListActivitiesUseCase
type ListActivitiesOutput struct {
	Result *query.PaginatedResult
}

// ListActivitiesUseCase handles fetching activities with filters
// This is a read-only operation and does NOT require a transaction
// Has access to both service and repository - decides which to use
type ListActivitiesUseCase struct {
	service service.ActivityServiceInterface       // For operations requiring business logic (can be nil for simple reads)
	repo    repository.ActivityRepositoryInterface // For simple read operations
	cache   types.CacheProvider
}

// NewListActivitiesUseCase creates a new instance with both service and repository
// For simple reads, service can be nil and usecase will use repo directly
func NewListActivitiesUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityRepositoryInterface,
	cache types.CacheProvider,
) *ListActivitiesUseCase {
	return &ListActivitiesUseCase{
		service: svc,
		repo:    repo,
		cache:   cache,
	}
}

// RequiresTransaction returns false - read operations don't need transactions
func (uc *ListActivitiesUseCase) RequiresTransaction() bool {
	return false
}

// Execute retrieves activities with dynamic filtering using QueryOptions (typed version)
// This is the NEW approach that supports flexible filtering, searching, and sorting
// Decision: Use repo directly for simple list operations (no business logic needed)
func (uc *ListActivitiesUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input ListActivitiesInput,
) (ListActivitiesOutput, error) {
	opts := input.QueryOptions
	if opts == nil {
		return ListActivitiesOutput{}, fmt.Errorf("query_options is required")
	}

	// SECURITY: Add user_id filter for multi-tenancy
	// This ensures users can only see their own activities
	opts.Filter["user_id"] = input.UserID

	// Use dynamic filtering with RelationshipRegistry v3.0
	result, err := uc.repo.ListActivitiesWithQuery(ctx, opts)
	if err != nil {
		return ListActivitiesOutput{}, fmt.Errorf("failed to list activities: %w", err)
	}

	if uc.cache != nil {
		jsonActivities, err := json.Marshal(result)
		if err != nil {
			fmt.Printf("Oops! Couldn't shrink-wrap for cache: %v\n", err)
		} else {
			_ = uc.cache.Set("user_activities:"+strconv.Itoa(input.UserID), string(jsonActivities))
		}
	}

	return ListActivitiesOutput{Result: result}, nil
}
