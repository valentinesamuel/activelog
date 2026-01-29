package main

import (
	activityUsecases "github.com/valentinesamuel/activelog/internal/application/activity/usecases/di"
	photoUsecases "github.com/valentinesamuel/activelog/internal/application/activityPhoto/usecases/di"
	"github.com/valentinesamuel/activelog/internal/application/broker/di"
	statsUsecases "github.com/valentinesamuel/activelog/internal/application/stats/usecases/di"
	tagUsecases "github.com/valentinesamuel/activelog/internal/application/tag/usecases/di"
	di2 "github.com/valentinesamuel/activelog/internal/cache/di"
	"github.com/valentinesamuel/activelog/internal/container"
	di3 "github.com/valentinesamuel/activelog/internal/handlers/di"
	"github.com/valentinesamuel/activelog/internal/repository"
	di4 "github.com/valentinesamuel/activelog/internal/repository/di"
	di5 "github.com/valentinesamuel/activelog/internal/service/di"
	di6 "github.com/valentinesamuel/activelog/internal/storage/di"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// setupContainer creates and configures the DI container
// All dependencies are registered here following Clean Architecture layering
// Registration order: Core → Storage → Repositories → Services → Broker → UseCases → Handlers
func setupContainer(db repository.DBConn) *container.Container {
	c := container.New()

	// Register core singletons (must be first)
	registerCoreDependencies(c, db)

	// Register storage provider (uses config globals)
	di6.RegisterStorage(c)
	di2.RegisterCache(c)

	// Eagerly resolve dependedncies
	c.MustResolve(di6.StorageProviderKey)
	c.MustResolve(di2.CacheProviderKey)

	// Register layers in dependency order
	di4.RegisterRepositories(c) // Layer 1: Data access
	di5.RegisterServices(c)     // Layer 2: Business logic
	di.RegisterBroker(c)        // Layer 3: Use case orchestration

	// Register use cases by domain
	activityUsecases.RegisterActivityUseCases(c)
	tagUsecases.RegisterTagUseCases(c)
	statsUsecases.RegisterStatsUseCases(c)
	photoUsecases.RegisterActivityPhotoUseCases(c)

	// Register handlers (depends on everything above)
	di3.RegisterHandlers(c)

	return c
}

// registerCoreDependencies registers core singletons like database connection
// These must be registered before any other dependencies
func registerCoreDependencies(c *container.Container, db repository.DBConn) {
	c.RegisterSingleton(di4.CoreDBKey, db)
	c.RegisterSingleton(di.CoreRawDBKey, db.GetRawDB())
	c.RegisterSingleton(di4.CoreRegistryManagerKey, setupRegistryManager())
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
