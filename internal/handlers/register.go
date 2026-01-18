package handlers

import (
	activityUsecases "github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	photoUsecases "github.com/valentinesamuel/activelog/internal/application/activityPhoto/usecases"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/repository"
)

// RegisterHandlers registers all HTTP handler factories with the container
// Dependencies: Requires use cases, broker, and repositories to be registered first
func RegisterHandlers(c *container.Container) {
	// Health handler (no dependencies)
	c.Register(HealthHandlerKey, func(c *container.Container) (interface{}, error) {
		return NewHealthHandler(), nil
	})

	// User handler (legacy pattern for now)
	c.Register(UserHandlerKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(repository.UserRepoKey).(*repository.UserRepository)
		return NewUserHandler(repo), nil
	})

	// Activity handler (broker pattern with all use cases)
	c.Register(ActivityHandlerKey, func(c *container.Container) (interface{}, error) {
		brokerInstance := c.MustResolve(broker.BrokerKey).(*broker.Broker)
		repo := c.MustResolve(repository.ActivityRepoKey).(repository.ActivityRepositoryInterface)

		// Resolve all use cases using their exported keys
		createUC := c.MustResolve(activityUsecases.CreateActivityUCKey).(broker.UseCase)
		getUC := c.MustResolve(activityUsecases.GetActivityUCKey).(broker.UseCase)
		listUC := c.MustResolve(activityUsecases.ListActivitiesUCKey).(broker.UseCase)
		updateUC := c.MustResolve(activityUsecases.UpdateActivityUCKey).(broker.UseCase)
		deleteUC := c.MustResolve(activityUsecases.DeleteActivityUCKey).(broker.UseCase)
		getStatsUC := c.MustResolve(activityUsecases.GetActivityStatsUCKey).(broker.UseCase)

		return NewActivityHandler(ActivityHandlerDeps{
			Broker:             brokerInstance,
			Repo:               repo,
			CreateActivityUC:   createUC,
			GetActivityUC:      getUC,
			ListActivitiesUC:   listUC,
			UpdateActivityUC:   updateUC,
			DeleteActivityUC:   deleteUC,
			GetActivityStatsUC: getStatsUC,
		}), nil
	})

	// Stats handler (legacy pattern for now - will migrate to V2 later)
	c.Register(StatsHandlerKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(repository.StatsRepoKey).(repository.StatsRepositoryInterface)
		return NewStatsHandler(repo), nil
	})

	// Activity photo handler
	c.Register(ActivityPhotoHandlerKey, func(c *container.Container) (interface{}, error) {
		brokerInstance := c.MustResolve(broker.BrokerKey).(*broker.Broker)
		repo := c.MustResolve(repository.ActivityPhotoRepoKey).(repository.ActivityPhotoRepositoryInterface)

		// Resolve use case using exported key
		uploadActivityPhotoUC := c.MustResolve(photoUsecases.UploadActivityPhotosUCKey).(broker.UseCase)
		getActivityPhotoUC := c.MustResolve(photoUsecases.GetActivityPhotosUCKey).(broker.UseCase)

		return NewActivityPhotoHandler(brokerInstance, repo, uploadActivityPhotoUC, getActivityPhotoUC), nil
	})
}
