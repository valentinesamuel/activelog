package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"

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
	// Photos come as *[]*multipart.FileHeader from the handler
	photosPtr, ok := input["photos"].(*[]*multipart.FileHeader)
	if !ok {
		return nil, fmt.Errorf("invalid photos format: expected *[]*multipart.FileHeader")
	}
	photos := *photosPtr

	activityID, ok := input["activity_id"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid activity_id format")
	}

	// TODO: Implement actual photo upload logic
	for i, photo := range photos {
		fmt.Printf("Photo %d: %s (size: %d bytes, content-type: %s)\n",
			i+1, photo.Filename, photo.Size, photo.Header.Get("Content-Type"))
	}

	return map[string]interface{}{
		"activityPhotos": nil,
		"activity_id":    activityID,
		"count":          len(photos),
	}, nil
}
