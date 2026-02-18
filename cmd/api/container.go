package main

import (
	activityUsecases "github.com/valentinesamuel/activelog/internal/application/activity/usecases/di"
	photoUsecases "github.com/valentinesamuel/activelog/internal/application/activityPhoto/usecases/di"
	"github.com/valentinesamuel/activelog/internal/application/broker/di"
	statsUsecases "github.com/valentinesamuel/activelog/internal/application/stats/usecases/di"
	tagUsecases "github.com/valentinesamuel/activelog/internal/application/tag/usecases/di"
	cacheRegister "github.com/valentinesamuel/activelog/internal/cache/di"
	"github.com/valentinesamuel/activelog/internal/container"
	emailRegister "github.com/valentinesamuel/activelog/internal/email/di"
	handlerRegister "github.com/valentinesamuel/activelog/internal/handlers/di"
	queueRegister "github.com/valentinesamuel/activelog/internal/queue/di"
	"github.com/valentinesamuel/activelog/internal/repository"
	repositoryRegister "github.com/valentinesamuel/activelog/internal/repository/di"
	serviceRegister "github.com/valentinesamuel/activelog/internal/service/di"
	storageRegister "github.com/valentinesamuel/activelog/internal/storage/di"
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
	storageRegister.RegisterStorage(c)
	cacheRegister.RegisterCache(c)
	queueRegister.RegisterQueue(c)
	emailRegister.RegisterEmail(c)

	// Eagerly resolve dependedncies
	c.MustResolve(storageRegister.StorageProviderKey)
	c.MustResolve(cacheRegister.CacheProviderKey)
	c.MustResolve(queueRegister.QueueProviderKey)
	c.MustResolve(emailRegister.EmailProviderKey)

	// Register layers in dependency order
	repositoryRegister.RegisterRepositories(c) // Layer 1: Data access
	serviceRegister.RegisterServices(c)        // Layer 2: Business logic
	di.RegisterBroker(c)                       // Layer 3: Use case orchestration

	// Register use cases by domain
	activityUsecases.RegisterActivityUseCases(c)
	tagUsecases.RegisterTagUseCases(c)
	statsUsecases.RegisterStatsUseCases(c)
	photoUsecases.RegisterActivityPhotoUseCases(c)

	// Register handlers (depends on everything above)
	handlerRegister.RegisterHandlers(c)

	return c
}

// registerCoreDependencies registers core singletons like database connection
// These must be registered before any other dependencies
func registerCoreDependencies(c *container.Container, db repository.DBConn) {
	c.RegisterSingleton(repositoryRegister.CoreDBKey, db)
	c.RegisterSingleton(di.CoreRawDBKey, db.GetRawDB())
	c.RegisterSingleton(repositoryRegister.CoreRegistryManagerKey, setupRegistryManager())
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
