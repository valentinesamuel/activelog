package usecases

import (
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/repository"
)

// RegisterStatsUseCases registers all stats-related use case factories
// Dependencies: Requires repositories to be registered first
func RegisterStatsUseCases(c *container.Container) {
	// All stats operations are read-only (non-transactional)
	c.Register(GetWeeklyStatsUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(repository.StatsRepoKey).(repository.StatsRepositoryInterface)
		return NewGetWeeklyStatsUseCase(repo), nil
	})

	c.Register(GetMonthlyStatsUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(repository.StatsRepoKey).(repository.StatsRepositoryInterface)
		return NewGetMonthlyStatsUseCase(repo), nil
	})

	c.Register(GetUserSummaryUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(repository.StatsRepoKey).(repository.StatsRepositoryInterface)
		return NewGetUserSummaryUseCase(repo), nil
	})

	c.Register(GetTopTagsUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(repository.StatsRepoKey).(repository.StatsRepositoryInterface)
		return NewGetTopTagsUseCase(repo), nil
	})

	c.Register(GetActivityCountByTypeUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(repository.StatsRepoKey).(repository.StatsRepositoryInterface)
		return NewGetActivityCountByTypeUseCase(repo), nil
	})
}
