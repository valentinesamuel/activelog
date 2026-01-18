package usecases

import (
	"context"
	"database/sql"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

type GetActivityPhotoUseCase struct {
	service service.ActivityServiceInterface
	repo    repository.ActivityPhotoRepositoryInterface
}

func NewGetActivityPhotoUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityPhotoRepositoryInterface,
) *GetActivityPhotoUseCase {
	return &GetActivityPhotoUseCase{
		service: svc,
		repo:    repo,
	}
}

func (uc *GetActivityPhotoUseCase) RequiresTransaction() bool {
	return false
}

func (uc *GetActivityPhotoUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input map[string]interface{},
) (map[string]interface{}, error) {
	activityId := input["activityId"].(int)

	photos, err := uc.repo.GetByActivityID(ctx, activityId)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"photos": photos,
	}, nil
}
