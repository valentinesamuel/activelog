package di

import (
	"github.com/valentinesamuel/activelog/internal/platform/container"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/repository/di"
	"github.com/valentinesamuel/activelog/internal/service"
)

// RegisterServices registers all service-layer factories with the container
// Dependencies: Requires repositories to be registered first
func RegisterServices(c *container.Container) {
	// Activity service (handles activity business logic)
	c.Register(ActivityServiceKey, func(c *container.Container) (interface{}, error) {
		activityRepo := c.MustResolve(di.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		tagRepo := c.MustResolve(di.TagRepoKey).(repository.TagRepositoryInterface)
		return service.NewActivityService(activityRepo, tagRepo), nil
	})

	// Stats service (handles statistics and analytics logic)
	c.Register(StatsServiceKey, func(c *container.Container) (interface{}, error) {
		statsRepo := c.MustResolve(di.StatsRepoKey).(repository.StatsRepositoryInterface)
		activityRepo := c.MustResolve(di.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return service.NewStatsService(statsRepo, activityRepo), nil
	})
}
