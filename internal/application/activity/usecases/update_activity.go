package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/cache/types"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

type UpdateActivityInput struct {
	UserID     int
	ActivityID int
	Request    *models.UpdateActivityRequest
}

type UpdateActivityOutput struct {
	Activity *models.Activity
	Updated  bool
}

type UpdateActivityUseCase struct {
	service service.ActivityServiceInterface
	repo    repository.ActivityRepositoryInterface
	cache   types.CacheProvider
}

func NewUpdateActivityUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityRepositoryInterface,
	cache types.CacheProvider,
) *UpdateActivityUseCase {
	return &UpdateActivityUseCase{
		service: svc,
		repo:    repo,
		cache:   cache,
	}
}

func (uc *UpdateActivityUseCase) RequiresTransaction() bool {
	return true
}

func (uc *UpdateActivityUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input UpdateActivityInput,
) (UpdateActivityOutput, error) {
	if input.Request == nil {
		return UpdateActivityOutput{}, fmt.Errorf("request is required")
	}

	activity, err := uc.service.UpdateActivity(ctx, tx, input.UserID, input.ActivityID, input.Request)

	if err != nil {
		return UpdateActivityOutput{}, fmt.Errorf("failed to update activity: %w", err)
	}

	uc.cache.Del(fmt.Sprintf("activities:user:%d", activity.UserID))
	uc.cache.Del(fmt.Sprintf("activity:%d", activity.UserID))

	return UpdateActivityOutput{
		Activity: activity,
		Updated:  true,
	}, nil
}
