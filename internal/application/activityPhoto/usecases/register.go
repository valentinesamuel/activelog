package usecases

import (
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
	"github.com/valentinesamuel/activelog/internal/storage"
	"github.com/valentinesamuel/activelog/internal/storage/types"
)

// RegisterPhotoUseCases registers all photo-related use case factories
// Dependencies: Requires services, repositories, and storage to be registered first
func RegisterActivityPhotoUseCases(c *container.Container) {
	c.Register(UploadActivityPhotosUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(service.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repository.ActivityPhotoRepoKey).(repository.ActivityPhotoRepositoryInterface)

		// Storage provider may be nil if not configured - handle gracefully
		var storageProvider types.StorageProvider
		if resolved := c.MustResolve(storage.StorageProviderKey); resolved != nil {
			storageProvider = resolved.(types.StorageProvider)
		}

		return NewUploadActivityPhotoUseCase(svc, repo, storageProvider), nil
	})

	c.Register(GetActivityPhotosUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(service.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(repository.ActivityPhotoRepoKey).(repository.ActivityPhotoRepositoryInterface)

		return NewGetActivityPhotoUseCase(svc, repo), nil
	})
}
