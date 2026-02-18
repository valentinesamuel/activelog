package di

import (
	"github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	cacheDI "github.com/valentinesamuel/activelog/internal/adapters/cache/di"
	cacheTypes "github.com/valentinesamuel/activelog/internal/adapters/cache/types"
	"github.com/valentinesamuel/activelog/internal/platform/container"
	"github.com/valentinesamuel/activelog/internal/repository"
	repoDI "github.com/valentinesamuel/activelog/internal/repository/di"
	"github.com/valentinesamuel/activelog/internal/service"
	serviceDI "github.com/valentinesamuel/activelog/internal/service/di"
)

// RegisterActivityUseCases registers all activity-related use case factories
// Dependencies: Requires services and repositories to be registered first
// All use cases receive BOTH service and repository - they decide which to use at runtime
func RegisterActivityUseCases(c *container.Container) {
	// Write operations (transactional)
	// These typically use service for business logic but have repo available if needed
	c.Register(CreateActivityUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(serviceDI.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repoDI.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return usecases.NewCreateActivityUseCase(svc, repo), nil
	})

	c.Register(UpdateActivityUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(serviceDI.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repoDI.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		var cacheAdapter cacheTypes.CacheAdapter
		if resolved := c.MustResolve(cacheDI.CacheAdapterKey); resolved != nil {
			cacheAdapter = resolved.(cacheTypes.CacheAdapter)
		}
		return usecases.NewUpdateActivityUseCase(svc, repo, cacheAdapter), nil
	})

	c.Register(DeleteActivityUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(serviceDI.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repoDI.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return usecases.NewDeleteActivityUseCase(svc, repo), nil
	})

	// Read operations (non-transactional)
	// These typically use repo directly for performance but have service available for enrichment
	c.Register(GetActivityUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(serviceDI.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repoDI.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return usecases.NewGetActivityUseCase(svc, repo), nil
	})

	c.Register(ListActivitiesUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(serviceDI.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repoDI.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		// Cache adapter may be nil if not configured — handle gracefully
		var cacheAdapter cacheTypes.CacheAdapter
		if resolved := c.MustResolve(cacheDI.CacheAdapterKey); resolved != nil {
			cacheAdapter = resolved.(cacheTypes.CacheAdapter)
		}
		return usecases.NewListActivitiesUseCase(svc, repo, cacheAdapter), nil
	})

	c.Register(GetActivityStatsUCKey, func(c *container.Container) (interface{}, error) {
		statsSvc := c.MustResolve(serviceDI.StatsServiceKey).(service.StatsServiceInterface)
		repo := c.MustResolve(repoDI.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return usecases.NewGetActivityStatsUseCase(statsSvc, repo), nil
	})
}
