package usecases

import (
	"context"
	"database/sql"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

type UploadActivityPhotoUseCase struct {
	service service.ActivityServiceInterface
	repo    repository.ActivityRepositoryInterface
}

func NewUploadActivityPhotoUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityRepositoryInterface,
) *UploadActivityPhotoUseCase {
	return &UploadActivityPhotoUseCase{
		service: svc,
		repo:    repo,
	}
}

func (uc *UploadActivityPhotoUseCase) RequiresTransaction() bool {
	return true
}

func (uc *UploadActivityPhotoUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input map[string]interface{},
) (map[string]interface{}, error) {
	// TODO: Implement photo upload logic
	return map[string]interface{}{
		"activityPhotos": nil,
	}, nil
}
