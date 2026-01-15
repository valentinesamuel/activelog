package main

import (
	"database/sql"

	"github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	activityPhotoUsecases "github.com/valentinesamuel/activelog/internal/application/photo/usecases"
	statsUsecases "github.com/valentinesamuel/activelog/internal/application/stats/usecases"
	tagUsecases "github.com/valentinesamuel/activelog/internal/application/tag/usecases"
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/handlers"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// setupContainer creates and configures the DI container
// All dependencies are registered here in a centralized location
// Follows Clean Architecture: Handlers → Broker → Use Cases → Services → Repositories
func setupContainer(db repository.DBConn) *container.Container {
	c := container.New()

	// Register core singletons
	registerCoreDependencies(c, db)

	// Register repositories
	registerRepositories(c)

	// Register services (business logic layer)
	registerServices(c)

	// Register broker (use case orchestrator)
	registerBroker(c)

	// Register use cases
	registerActivityUseCases(c)
	registerTagUseCases(c)
	registerStatsUseCases(c)
	registerActivityPhotoUseCases(c)

	// Register handlers
	registerHandlers(c)

	return c
}

// registerCoreDependencies registers core singletons like database connection
func registerCoreDependencies(c *container.Container, db repository.DBConn) {
	c.RegisterSingleton("db", db)
	c.RegisterSingleton("rawDB", db.GetRawDB())

	// Register RegistryManager (v3.0) for cross-registry path resolution
	// This enables deep nesting like: filter[user.company.department.country]=USA
	c.RegisterSingleton("registryManager", setupRegistryManager())
}

// setupRegistryManager creates and configures the global RegistryManager (v3.0)
// All table registries are registered here for deep nesting support
func setupRegistryManager() *query.RegistryManager {
	manager := query.NewRegistryManager()

	// Activities registry will be registered later (needs TagRepository)
	// See registerRepositories() for actual registration

	// Future: Register other table registries here as needed
	// manager.RegisterTable("users", usersRegistry)
	// manager.RegisterTable("companies", companiesRegistry)

	return manager
}

// registerRepositories registers all repository factories
func registerRepositories(c *container.Container) {
	// Tag repository (no dependencies besides DB)
	c.Register("tagRepo", func(c *container.Container) (interface{}, error) {
		db := c.MustResolve("db").(repository.DBConn)
		return repository.NewTagRepository(db), nil
	})

	// Activity repository (depends on TagRepository and RegistryManager)
	c.Register("activityRepo", func(c *container.Container) (interface{}, error) {
		db := c.MustResolve("db").(repository.DBConn)
		tagRepo := c.MustResolve("tagRepo").(*repository.TagRepository)
		manager := c.MustResolve("registryManager").(*query.RegistryManager)

		// Create repository with manager support (v3.0)
		activityRepo := repository.NewActivityRepository(db, tagRepo)

		// Register this repository's registry with the manager for deep nesting
		manager.RegisterTable("activities", activityRepo.GetRegistry())

		return activityRepo, nil
	})

	// User repository
	c.Register("userRepo", func(c *container.Container) (interface{}, error) {
		db := c.MustResolve("db").(repository.DBConn)
		return repository.NewUserRepository(db), nil
	})

	// Stats repository
	c.Register("statsRepo", func(c *container.Container) (interface{}, error) {
		db := c.MustResolve("db").(repository.DBConn)
		return repository.NewStatsRepository(db), nil
	})
}

// registerServices registers all service-layer dependencies
// Services encapsulate business logic and coordinate repository operations
func registerServices(c *container.Container) {
	// Activity service (handles activity business logic)
	c.Register("activityService", func(c *container.Container) (interface{}, error) {
		activityRepo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
		tagRepo := c.MustResolve("tagRepo").(repository.TagRepositoryInterface)
		return service.NewActivityService(activityRepo, tagRepo), nil
	})

	// Stats service (handles statistics and analytics logic)
	c.Register("statsService", func(c *container.Container) (interface{}, error) {
		statsRepo := c.MustResolve("statsRepo").(repository.StatsRepositoryInterface)
		activityRepo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
		return service.NewStatsService(statsRepo, activityRepo), nil
	})
}

// registerBroker registers the use case orchestrator
func registerBroker(c *container.Container) {
	c.Register("broker", func(c *container.Container) (interface{}, error) {
		rawDB := c.MustResolve("rawDB").(*sql.DB)
		return broker.NewBroker(rawDB), nil
	})
}

// registerActivityUseCases registers all activity-related use cases
// All use cases receive BOTH service and repository - they decide which to use at runtime
func registerActivityUseCases(c *container.Container) {
	// Write operations (transactional)
	// These typically use service for business logic but have repo available if needed
	c.Register("createActivityUC", func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve("activityService").(service.ActivityServiceInterface)
		repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
		return usecases.NewCreateActivityUseCase(svc, repo), nil
	})

	c.Register("updateActivityUC", func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve("activityService").(service.ActivityServiceInterface)
		repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
		return usecases.NewUpdateActivityUseCase(svc, repo), nil
	})

	c.Register("deleteActivityUC", func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve("activityService").(service.ActivityServiceInterface)
		repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
		return usecases.NewDeleteActivityUseCase(svc, repo), nil
	})

	// Read operations (non-transactional)
	// These typically use repo directly for performance but have service available for enrichment
	c.Register("getActivityUC", func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve("activityService").(service.ActivityServiceInterface)
		repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
		return usecases.NewGetActivityUseCase(svc, repo), nil
	})

	c.Register("listActivitiesUC", func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve("activityService").(service.ActivityServiceInterface)
		repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
		return usecases.NewListActivitiesUseCase(svc, repo), nil
	})

	c.Register("getActivityStatsUC", func(c *container.Container) (interface{}, error) {
		statsSvc := c.MustResolve("statsService").(service.StatsServiceInterface)
		repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
		return usecases.NewGetActivityStatsUseCase(statsSvc, repo), nil
	})
}

// registerTagUseCases registers all tag-related use cases
func registerTagUseCases(c *container.Container) {
	// Read operations (non-transactional)
	// Tags are typically read-only operations with dynamic filtering
	c.Register("listTagsUC", func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve("tagRepo").(repository.TagRepositoryInterface)
		return tagUsecases.NewListTagsUseCase(repo), nil
	})
}

// registerStatsUseCases registers all stats-related use cases
func registerStatsUseCases(c *container.Container) {
	// All stats operations are read-only (non-transactional)
	c.Register("getWeeklyStatsUC", func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve("statsRepo").(repository.StatsRepositoryInterface)
		return statsUsecases.NewGetWeeklyStatsUseCase(repo), nil
	})

	c.Register("getMonthlyStatsUC", func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve("statsRepo").(repository.StatsRepositoryInterface)
		return statsUsecases.NewGetMonthlyStatsUseCase(repo), nil
	})

	c.Register("getUserSummaryUC", func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve("statsRepo").(repository.StatsRepositoryInterface)
		return statsUsecases.NewGetUserSummaryUseCase(repo), nil
	})

	c.Register("getTopTagsUC", func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve("statsRepo").(repository.StatsRepositoryInterface)
		return statsUsecases.NewGetTopTagsUseCase(repo), nil
	})

	c.Register("getActivityCountByTypeUC", func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve("statsRepo").(repository.StatsRepositoryInterface)
		return statsUsecases.NewGetActivityCountByTypeUseCase(repo), nil
	})
}

// registerPhotoUseCases registers all photo-related use cases
func registerActivityPhotoUseCases(c *container.Container) {
	c.Register("uploadActivityPhotosUC", func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve("activityService").(service.ActivityServiceInterface)
		repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
		return activityPhotoUsecases.NewUploadActivityPhotoUseCase(svc, repo), nil
	})
}

// registerHandlers registers all HTTP handlers
func registerHandlers(c *container.Container) {
	// Health handler (no dependencies)
	c.Register("healthHandler", func(c *container.Container) (interface{}, error) {
		return handlers.NewHealthHandler(), nil
	})

	// User handler (legacy pattern for now)
	c.Register("userHandler", func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve("userRepo").(*repository.UserRepository)
		return handlers.NewUserHandler(repo), nil
	})

	// Activity handler (broker pattern with all use cases)
	c.Register("activityHandler", func(c *container.Container) (interface{}, error) {
		brokerInstance := c.MustResolve("broker").(*broker.Broker)
		repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)

		// Resolve all use cases
		createUC := c.MustResolve("createActivityUC").(broker.UseCase)
		getUC := c.MustResolve("getActivityUC").(broker.UseCase)
		listUC := c.MustResolve("listActivitiesUC").(broker.UseCase)
		updateUC := c.MustResolve("updateActivityUC").(broker.UseCase)
		deleteUC := c.MustResolve("deleteActivityUC").(broker.UseCase)
		getStatsUC := c.MustResolve("getActivityStatsUC").(broker.UseCase)

		return handlers.NewActivityHandler(handlers.ActivityHandlerDeps{
			Broker:             brokerInstance,
			Repo:               repo,
			CreateActivityUC:   createUC,
			GetActivityUC:      getUC,
			ListActivitiesUC:   listUC,
			UpdateActivityUC:   updateUC,
			DeleteActivityUC:   deleteUC,
			GetActivityStatsUC: getStatsUC},
		), nil
	})

	// Stats handler (legacy pattern for now - will migrate to V2 later)
	c.Register("statsHandler", func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve("statsRepo").(repository.StatsRepositoryInterface)
		return handlers.NewStatsHandler(repo), nil
	})

	// Activity photo handler
	c.Register("activityPhotoHandler", func(c *container.Container) (interface{}, error) {
		brokerInstance := c.MustResolve("broker").(*broker.Broker)
		repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)

		// Resolve all use cases
		uploadActivityPhotosUC := c.MustResolve("uploadActivityPhotosUC").(broker.UseCase)

		return handlers.NewActivityPhotoHandler(
			brokerInstance,
			repo,
			uploadActivityPhotosUC,
		), nil
	})
}
