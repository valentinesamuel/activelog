package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
	"github.com/valentinesamuel/activelog/internal/storage/types"
)

type UploadActivityPhotoUseCase struct {
	service service.ActivityServiceInterface
	repo    repository.ActivityRepositoryInterface
	storage types.StorageProvider
}

func NewUploadActivityPhotoUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityRepositoryInterface,
	storage types.StorageProvider,
) *UploadActivityPhotoUseCase {
	return &UploadActivityPhotoUseCase{
		service: svc,
		repo:    repo,
		storage: storage,
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

	// Check if storage provider is available
	if uc.storage == nil {
		return nil, fmt.Errorf("storage provider not configured")
	}

	// Upload each photo
	uploadedPhotos := make([]models.ActivityPhoto, 0, len(photos))
	for _, photo := range photos {
		activityPhoto, err := uc.uploadPhoto(ctx, activityID, photo)
		if err != nil {
			// If any upload fails, we should handle cleanup
			// For now, return error with partial uploads info
			return map[string]interface{}{
				"activityPhotos": uploadedPhotos,
				"activity_id":    activityID,
				"count":          len(uploadedPhotos),
				"error":          err.Error(),
			}, fmt.Errorf("failed to upload photo %s: %w", photo.Filename, err)
		}
		uploadedPhotos = append(uploadedPhotos, *activityPhoto)
	}

	return map[string]interface{}{
		"activityPhotos": uploadedPhotos,
		"activity_id":    activityID,
		"count":          len(uploadedPhotos),
	}, nil
}

// uploadPhoto uploads a single photo to storage and returns metadata
func (uc *UploadActivityPhotoUseCase) uploadPhoto(
	ctx context.Context,
	activityID int,
	fileHeader *multipart.FileHeader,
) (*models.ActivityPhoto, error) {
	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Generate unique storage key
	key := uc.generateStorageKey(activityID, fileHeader.Filename)
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to storage
	output, err := uc.storage.Upload(ctx, &types.UploadInput{
		Key:         key,
		Body:        file,
		ContentType: contentType,
		Size:        fileHeader.Size,
		Metadata: map[string]string{
			"activity_id":       fmt.Sprintf("%d", activityID),
			"original_filename": fileHeader.Filename,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to storage: %w", err)
	}

	// Create activity photo record
	activityPhoto := &models.ActivityPhoto{
		ActivityID:  activityID,
		S3Key:       output.Key,
		ContentType: contentType,
		FileSize:    fileHeader.Size,
		UploadedAt:  output.UploadedAt,
	}

	return activityPhoto, nil
}

// generateStorageKey creates a unique key for storing the photo
// Format: activities/{activityID}/photos/{uuid}{extension}
func (uc *UploadActivityPhotoUseCase) generateStorageKey(activityID int, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg" // Default extension
	}
	uniqueID := uuid.New().String()
	return fmt.Sprintf("activities/%d/photos/%s%s", activityID, uniqueID, ext)
}

// GetPresignedURL generates a presigned URL for accessing an uploaded photo
func (uc *UploadActivityPhotoUseCase) GetPresignedURL(
	ctx context.Context,
	key string,
	expiresIn time.Duration,
) (string, error) {
	if uc.storage == nil {
		return "", fmt.Errorf("storage provider not configured")
	}

	return uc.storage.GetPresignedURL(ctx, &types.PresignedURLInput{
		Key:       key,
		ExpiresIn: expiresIn,
		Operation: types.PresignGet,
	})
}
