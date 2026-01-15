package usecases

import (
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// RegisterPhotoUseCases registers all photo-related use case factories
// Dependencies: Requires services and repositories to be registered first
func RegisterActivityPhotoUseCases(c *container.Container) {
	c.Register(UploadActivityPhotosUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(service.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repository.ActivityRepoKey).(repository.ActivityRepositoryInterface)
		return NewUploadActivityPhotoUseCase(svc, repo), nil
	})
}
