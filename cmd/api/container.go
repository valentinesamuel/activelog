package main

import (
	activityUsecases "github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	photoUsecases "github.com/valentinesamuel/activelog/internal/application/activityPhoto/usecases"
	statsUsecases "github.com/valentinesamuel/activelog/internal/application/stats/usecases"
	tagUsecases "github.com/valentinesamuel/activelog/internal/application/tag/usecases"
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/handlers"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// setupContainer creates and configures the DI container
// All dependencies are registered here following Clean Architecture layering
// Registration order: Core → Repositories → Services → Broker → UseCases → Handlers
func setupContainer(db repository.DBConn) *container.Container {
	c := container.New()

	// Register core singletons (must be first)
	registerCoreDependencies(c, db)

	// Register layers in dependency order
	repository.RegisterRepositories(c) // Layer 1: Data access
	service.RegisterServices(c)        // Layer 2: Business logic
	broker.RegisterBroker(c)           // Layer 3: Use case orchestration

	// Register use cases by domain
	activityUsecases.RegisterActivityUseCases(c)
	tagUsecases.RegisterTagUseCases(c)
	statsUsecases.RegisterStatsUseCases(c)
	photoUsecases.RegisterActivityPhotoUseCases(c)

	// Register handlers (depends on everything above)
	handlers.RegisterHandlers(c)

	return c
}

// registerCoreDependencies registers core singletons like database connection
// These must be registered before any other dependencies
func registerCoreDependencies(c *container.Container, db repository.DBConn) {
	c.RegisterSingleton(repository.CoreDBKey, db)
	c.RegisterSingleton(broker.CoreRawDBKey, db.GetRawDB())
	c.RegisterSingleton(repository.CoreRegistryManagerKey, setupRegistryManager())
}

// setupRegistryManager creates and configures the global RegistryManager (v3.0)
// All table registries are registered here for deep nesting support
func setupRegistryManager() *query.RegistryManager {
	manager := query.NewRegistryManager()

	// Activities registry will be registered later (needs TagRepository)
	// See repository.RegisterRepositories() for actual registration

	// Future: Register other table registries here as needed
	// manager.RegisterTable("users", usersRegistry)
	// manager.RegisterTable("companies", companiesRegistry)

	return manager
}
