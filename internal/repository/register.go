package repository

import (
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// CoreDBKey is the key for the database connection singleton
const CoreDBKey = "db"

// CoreRegistryManagerKey is the key for the registry manager singleton
const CoreRegistryManagerKey = "registryManager"

// RegisterRepositories registers all repository factories with the container
// Dependencies: Requires "db" and "registryManager" to be registered first
func RegisterRepositories(c *container.Container) {
	// Tag repository (no dependencies besides DB)
	c.Register(TagRepoKey, func(c *container.Container) (interface{}, error) {
		db := c.MustResolve(CoreDBKey).(DBConn)
		return NewTagRepository(db), nil
	})

	// Activity repository (depends on TagRepository and RegistryManager)
	c.Register(ActivityRepoKey, func(c *container.Container) (interface{}, error) {
		db := c.MustResolve(CoreDBKey).(DBConn)
		tagRepo := c.MustResolve(TagRepoKey).(*TagRepository)
		manager := c.MustResolve(CoreRegistryManagerKey).(*query.RegistryManager)

		// Create repository with manager support (v3.0)
		activityRepo := NewActivityRepository(db, tagRepo)

		// Register this repository's registry with the manager for deep nesting
		manager.RegisterTable("activities", activityRepo.GetRegistry())

		return activityRepo, nil
	})

		c.Register(ActivityPhotoRepoKey, func(c *container.Container) (interface{}, error) {
		db := c.MustResolve(CoreDBKey).(DBConn)
		activityRepo := c.MustResolve(ActivityRepoKey).(*ActivityRepository)
		manager := c.MustResolve(CoreRegistryManagerKey).(*query.RegistryManager)

		// Create repository with manager support (v3.0)
		activityPhotoRepo := NewActivityPhotoRepository(db, activityRepo)

		// Register this repository's registry with the manager for deep nesting
		manager.RegisterTable("activity_photos", activityPhotoRepo.GetRegistry())

		return activityPhotoRepo, nil
	})

	// User repository
	c.Register(UserRepoKey, func(c *container.Container) (interface{}, error) {
		db := c.MustResolve(CoreDBKey).(DBConn)
		return NewUserRepository(db), nil
	})

	// Stats repository
	c.Register(StatsRepoKey, func(c *container.Container) (interface{}, error) {
		db := c.MustResolve(CoreDBKey).(DBConn)
		return NewStatsRepository(db), nil
	})
}
