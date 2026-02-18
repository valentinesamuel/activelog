package di

import (
	"github.com/valentinesamuel/activelog/internal/application/activityPhoto/usecases"
	"github.com/valentinesamuel/activelog/internal/platform/container"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/repository/di"
	"github.com/valentinesamuel/activelog/internal/service"
	di2 "github.com/valentinesamuel/activelog/internal/service/di"
	di3 "github.com/valentinesamuel/activelog/internal/adapters/storage/di"
	"github.com/valentinesamuel/activelog/internal/adapters/storage/types"
)

// RegisterPhotoUseCases registers all photo-related use case factories
// Dependencies: Requires services, repositories, and storage to be registered first
func RegisterActivityPhotoUseCases(c *container.Container) {
	c.Register(UploadActivityPhotosUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(di2.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(di.ActivityPhotoRepoKey).(repository.ActivityPhotoRepositoryInterface)

		// Storage provider may be nil if not configured - handle gracefully
		var storageProvider types.StorageProvider
		if resolved := c.MustResolve(di3.StorageProviderKey); resolved != nil {
			storageProvider = resolved.(types.StorageProvider)
		}

		return usecases.NewUploadActivityPhotoUseCase(svc, repo, storageProvider), nil
	})

	c.Register(GetActivityPhotosUCKey, func(c *container.Container) (interface{}, error) {
		svc := c.MustResolve(di2.ActivityServiceKey).(service.ActivityServiceInterface)
		repo := c.MustResolve(di.ActivityPhotoRepoKey).(repository.ActivityPhotoRepositoryInterface)

		return usecases.NewGetActivityPhotoUseCase(svc, repo), nil
	})
}
