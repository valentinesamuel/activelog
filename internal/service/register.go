package service

import (
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/repository"
)

// RegisterServices registers all service-layer factories with the container
// Dependencies: Requires repositories to be registered first
func RegisterServices(c *container.Container) {
	// Activity service (handles activity business logic)
	c.Register(ActivityServiceKey, func(c *container.Container) (interface{}, error) {
		activityRepo := c.MustResolve(repository.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		tagRepo := c.MustResolve(repository.TagRepoKey).(repository.TagRepositoryInterface)
		return NewActivityService(activityRepo, tagRepo), nil
	})

	// Stats service (handles statistics and analytics logic)
	c.Register(StatsServiceKey, func(c *container.Container) (interface{}, error) {
		statsRepo := c.MustResolve(repository.StatsRepoKey).(repository.StatsRepositoryInterface)
		activityRepo := c.MustResolve(repository.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return NewStatsService(statsRepo, activityRepo), nil
	})
}
