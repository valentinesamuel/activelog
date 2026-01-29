package di

import (
	"github.com/valentinesamuel/activelog/internal/application/tag/usecases"
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/repository/di"
)

// RegisterTagUseCases registers all tag-related use case factories
// Dependencies: Requires repositories to be registered first
func RegisterTagUseCases(c *container.Container) {
	// Read operations (non-transactional)
	// Tags are typically read-only operations with dynamic filtering
	c.Register(ListTagsUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(di.TagRepoKey).(repository.TagRepositoryInterface)
		return usecases.NewListTagsUseCase(repo), nil
	})
}
