package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	cacheTypes "github.com/valentinesamuel/activelog/internal/adapters/cache/types"
	"github.com/valentinesamuel/activelog/internal/middleware"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
	"github.com/valentinesamuel/activelog/pkg/query"
)

type ListActivitiesInput struct {
	UserID       int
	QueryOptions *query.QueryOptions
}

// CacheMeta contains cache status information for HTTP headers
type CacheMeta struct {
	Hit bool          // true = served from cache, false = fetched from DB
	TTL time.Duration // time until cache expires (only set on MISS)
}

type ListActivitiesOutput struct {
	Result *query.PaginatedResult
	Cache  CacheMeta
}

type ListActivitiesUseCase struct {
	service service.ActivityServiceInterface
	repo    repository.ActivityRepositoryInterface
	cache   cacheTypes.CacheAdapter
}

func NewListActivitiesUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityRepositoryInterface,
	cache cacheTypes.CacheAdapter,
) *ListActivitiesUseCase {
	return &ListActivitiesUseCase{
		service: svc,
		repo:    repo,
		cache:   cache,
	}
}

func (uc *ListActivitiesUseCase) RequiresTransaction() bool {
	return false
}

const cacheTTL = 2 * time.Minute

var activityCacheOpts = cacheTypes.CacheOptions{
	DB:           cacheTypes.CacheDBActivityData,
	PartitionKey: cacheTypes.CachePartitionActivities,
}

func (uc *ListActivitiesUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // unused for cached reads, required for broker interface
	input ListActivitiesInput,
) (ListActivitiesOutput, error) {
	opts := input.QueryOptions
	if opts == nil {
		return ListActivitiesOutput{}, fmt.Errorf("query_options is required")
	}

	opts.Filter["user_id"] = input.UserID

	// Generate cache key based on user + query options
	cacheKey := uc.generateCacheKey(input.UserID, opts)

	// Try cache first
	if uc.cache != nil {
		if cached, err := uc.cache.Get(ctx, cacheKey, activityCacheOpts); err == nil && cached != "" {
			var result query.PaginatedResult
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				middleware.CacheHitsTotal.Inc()
				return ListActivitiesOutput{
					Result: &result,
					Cache:  CacheMeta{Hit: true},
				}, nil
			}
		}
		middleware.CacheMissesTotal.Inc()
	}

	// Cache miss - fetch from database
	result, err := uc.repo.ListActivitiesWithQuery(ctx, opts)
	if err != nil {
		return ListActivitiesOutput{}, fmt.Errorf("failed to list activities: %w", err)
	}

	// Store in cache
	if uc.cache != nil {
		if jsonData, err := json.Marshal(result); err == nil {
			_ = uc.cache.Set(ctx, cacheKey, string(jsonData), cacheTTL, activityCacheOpts)
		}
	}

	return ListActivitiesOutput{
		Result: result,
		Cache:  CacheMeta{Hit: false, TTL: cacheTTL},
	}, nil
}

// generateCacheKey creates a unique cache key based on user and query options
func (uc *ListActivitiesUseCase) generateCacheKey(userID int, opts *query.QueryOptions) string {
	// Include query params in key to avoid serving wrong cached data
	keyData, _ := json.Marshal(opts)
	return fmt.Sprintf("user:%d:query:%s", userID, string(keyData))
}
