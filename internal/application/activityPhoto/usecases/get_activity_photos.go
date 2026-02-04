package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// GetActivityPhotosInput defines the typed input for GetActivityPhotoUseCase
type GetActivityPhotosInput struct {
	ActivityID int
}

// GetActivityPhotosOutput defines the typed output for GetActivityPhotoUseCase
type GetActivityPhotosOutput struct {
	Photos []*models.ActivityPhoto
}

// GetActivityPhotoUseCase handles fetching photos for an activity
type GetActivityPhotoUseCase struct {
	service service.ActivityServiceInterface
	repo    repository.ActivityPhotoRepositoryInterface
}

// NewGetActivityPhotoUseCase creates a new instance
func NewGetActivityPhotoUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityPhotoRepositoryInterface,
) *GetActivityPhotoUseCase {
	return &GetActivityPhotoUseCase{
		service: svc,
		repo:    repo,
	}
}

// RequiresTransaction returns false - read operations don't need transactions
func (uc *GetActivityPhotoUseCase) RequiresTransaction() bool {
	return false
}

// Execute retrieves photos for an activity (typed version)
func (uc *GetActivityPhotoUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input GetActivityPhotosInput,
) (GetActivityPhotosOutput, error) {
	photos, err := uc.repo.GetByActivityID(ctx, input.ActivityID)
	if err != nil {
		return GetActivityPhotosOutput{}, fmt.Errorf("failed to get activity photos: %w", err)
	}

	return GetActivityPhotosOutput{Photos: photos}, nil
}
