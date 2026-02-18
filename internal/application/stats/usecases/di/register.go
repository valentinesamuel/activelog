package di

import (
	"github.com/valentinesamuel/activelog/internal/application/stats/usecases"
	"github.com/valentinesamuel/activelog/internal/platform/container"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/repository/di"
)

// RegisterStatsUseCases registers all stats-related use case factories
// Dependencies: Requires repositories to be registered first
func RegisterStatsUseCases(c *container.Container) {
	// All stats operations are read-only (non-transactional)
	c.Register(GetWeeklyStatsUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(di.StatsRepoKey).(repository.StatsRepositoryInterface)
		return usecases.NewGetWeeklyStatsUseCase(repo), nil
	})

	c.Register(GetMonthlyStatsUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(di.StatsRepoKey).(repository.StatsRepositoryInterface)
		return usecases.NewGetMonthlyStatsUseCase(repo), nil
	})

	c.Register(GetUserSummaryUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(di.StatsRepoKey).(repository.StatsRepositoryInterface)
		return usecases.NewGetUserSummaryUseCase(repo), nil
	})

	c.Register(GetTopTagsUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(di.StatsRepoKey).(repository.StatsRepositoryInterface)
		return usecases.NewGetTopTagsUseCase(repo), nil
	})

	c.Register(GetActivityCountByTypeUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(di.StatsRepoKey).(repository.StatsRepositoryInterface)
		return usecases.NewGetActivityCountByTypeUseCase(repo), nil
	})
}
