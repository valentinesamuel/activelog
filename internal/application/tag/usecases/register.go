package usecases

import (
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/repository"
)

// RegisterTagUseCases registers all tag-related use case factories
// Dependencies: Requires repositories to be registered first
func RegisterTagUseCases(c *container.Container) {
	// Read operations (non-transactional)
	// Tags are typically read-only operations with dynamic filtering
	c.Register(ListTagsUCKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(repository.TagRepoKey).(repository.TagRepositoryInterface)
		return NewListTagsUseCase(repo), nil
	})
}
