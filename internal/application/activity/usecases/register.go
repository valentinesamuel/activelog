package usecases

import (
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// RegisterActivityUseCases registers all activity-related use case factories
// Dependencies: Requires services and repositories to be registered first
// All use cases receive BOTH service and repository - they decide which to use at runtime
func RegisterActivityUseCases(c *container.Container) {
	// Write operations (transactional)
	// These typically use service for business logic but have repo available if needed
	c.Register(CreateActivityUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(service.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repository.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return NewCreateActivityUseCase(svc, repo), nil
	})

	c.Register(UpdateActivityUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(service.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repository.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return NewUpdateActivityUseCase(svc, repo), nil
	})

	c.Register(DeleteActivityUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(service.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repository.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return NewDeleteActivityUseCase(svc, repo), nil
	})

	// Read operations (non-transactional)
	// These typically use repo directly for performance but have service available for enrichment
	c.Register(GetActivityUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(service.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repository.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return NewGetActivityUseCase(svc, repo), nil
	})

	c.Register(ListActivitiesUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(service.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repository.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return NewListActivitiesUseCase(svc, repo), nil
	})

	c.Register(GetActivityStatsUCKey, func(c *container.Container) (interface{}, error) {
		statsSvc := c.MustResolve(service.StatsServiceKey).(service.StatsServiceInterface)
		repo := c.MustResolve(repository.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return NewGetActivityStatsUseCase(statsSvc, repo), nil
	})
}
