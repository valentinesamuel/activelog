package di

import (
	activityUsecases "github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	activityUsecasesDI "github.com/valentinesamuel/activelog/internal/application/activity/usecases/di"
	photoUsecases "github.com/valentinesamuel/activelog/internal/application/activityPhoto/usecases"
	photoUsecasesDI "github.com/valentinesamuel/activelog/internal/application/activityPhoto/usecases/di"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/application/broker/di"
	"github.com/valentinesamuel/activelog/internal/platform/container"
	"github.com/valentinesamuel/activelog/internal/handlers"
	queueDI "github.com/valentinesamuel/activelog/internal/adapters/queue/di"
	queueTypes "github.com/valentinesamuel/activelog/internal/adapters/queue/types"
	"github.com/valentinesamuel/activelog/internal/repository"
	di2 "github.com/valentinesamuel/activelog/internal/repository/di"
	storageDI "github.com/valentinesamuel/activelog/internal/adapters/storage/di"
	storageTypes "github.com/valentinesamuel/activelog/internal/adapters/storage/types"
)

// RegisterHandlers registers all HTTP handler factories with the container
// Dependencies: Requires use cases, broker, and repositories to be registered first
func RegisterHandlers(c *container.Container) {
	// Health handler (no dependencies)
	c.Register(HealthHandlerKey, func(c *container.Container) (interface{}, error) {
		return handlers.NewHealthHandler(), nil
	})

	// User handler (legacy pattern for now)
	c.Register(UserHandlerKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(di2.UserRepoKey).(*repository.UserRepository)
		return handlers.NewUserHandler(repo), nil
	})

	// Activity handler (broker pattern with typed use cases)
	c.Register(ActivityHandlerKey, func(c *container.Container) (interface{}, error) {
		brokerInstance := c.MustResolve(di.BrokerKey).(*broker.Broker)
		repo := c.MustResolve(di2.ActivityRepoKey).(repository.ActivityRepositoryInterface)

		// Resolve all typed use cases
		createUC := c.MustResolve(activityUsecasesDI.CreateActivityUCKey).(*activityUsecases.CreateActivityUseCase)
		getUC := c.MustResolve(activityUsecasesDI.GetActivityUCKey).(*activityUsecases.GetActivityUseCase)
		listUC := c.MustResolve(activityUsecasesDI.ListActivitiesUCKey).(*activityUsecases.ListActivitiesUseCase)
		updateUC := c.MustResolve(activityUsecasesDI.UpdateActivityUCKey).(*activityUsecases.UpdateActivityUseCase)
		deleteUC := c.MustResolve(activityUsecasesDI.DeleteActivityUCKey).(*activityUsecases.DeleteActivityUseCase)
		getStatsUC := c.MustResolve(activityUsecasesDI.GetActivityStatsUCKey).(*activityUsecases.GetActivityStatsUseCase)

		return handlers.NewActivityHandler(handlers.ActivityHandlerDeps{
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
		repo := c.MustResolve(di2.StatsRepoKey).(repository.StatsRepositoryInterface)
		return handlers.NewStatsHandler(repo), nil
	})

	// Activity photo handler (typed use cases)
	c.Register(ActivityPhotoHandlerKey, func(c *container.Container) (interface{}, error) {
		brokerInstance := c.MustResolve(di.BrokerKey).(*broker.Broker)
		repo := c.MustResolve(di2.ActivityPhotoRepoKey).(repository.ActivityPhotoRepositoryInterface)

		// Resolve typed use cases
		uploadActivityPhotoUC := c.MustResolve(photoUsecasesDI.UploadActivityPhotosUCKey).(*photoUsecases.UploadActivityPhotoUseCase)
		getActivityPhotoUC := c.MustResolve(photoUsecasesDI.GetActivityPhotosUCKey).(*photoUsecases.GetActivityPhotoUseCase)

		return handlers.NewActivityPhotoHandler(brokerInstance, repo, uploadActivityPhotoUC, getActivityPhotoUC), nil
	})

	// Webhook handler
	c.Register(WebhookHandlerKey, func(c *container.Container) (interface{}, error) {
		webhookRepo := c.MustResolve(di2.WebhookRepoKey).(*repository.WebhookRepository)
		return handlers.NewWebhookHandler(webhookRepo), nil
	})

	// Export handler
	c.Register(ExportHandlerKey, func(c *container.Container) (interface{}, error) {
		activityRepo := c.MustResolve(di2.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		exportRepo := c.MustResolve(di2.ExportRepoKey).(*repository.ExportRepository)
		queueProvider := c.MustResolve(queueDI.QueueProviderKey).(queueTypes.QueueProvider)
		storage := c.MustResolve(storageDI.StorageProviderKey).(storageTypes.StorageProvider)
		return handlers.NewExportHandler(handlers.ExportHandlerDeps{
			ActivityRepo:  activityRepo,
			ExportRepo:    exportRepo,
			QueueProvider: queueProvider,
			Storage:       storage,
		}), nil
	})
}
