package usecases

import (
	"bytes"
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
	"github.com/valentinesamuel/activelog/internal/adapters/storage/types"
	"github.com/valentinesamuel/activelog/pkg/imageutil"
)

// UploadActivityPhotoInput defines the typed input for UploadActivityPhotoUseCase
type UploadActivityPhotoInput struct {
	UserID     int
	ActivityID int
	Photos     []*multipart.FileHeader
}

// UploadActivityPhotoOutput defines the typed output for UploadActivityPhotoUseCase
type UploadActivityPhotoOutput struct {
	ActivityPhotos []models.ActivityPhoto
	ActivityID     int
	Count          int
}

// UploadActivityPhotoUseCase handles photo uploads for activities
type UploadActivityPhotoUseCase struct {
	service service.ActivityServiceInterface
	repo    repository.ActivityPhotoRepositoryInterface
	storage types.StorageProvider
}

// NewUploadActivityPhotoUseCase creates a new instance
func NewUploadActivityPhotoUseCase(
	svc service.ActivityServiceInterface,
	repo repository.ActivityPhotoRepositoryInterface,
	storage types.StorageProvider,
) *UploadActivityPhotoUseCase {
	return &UploadActivityPhotoUseCase{
		service: svc,
		repo:    repo,
		storage: storage,
	}
}

// RequiresTransaction returns true - write operations need transactions
func (uc *UploadActivityPhotoUseCase) RequiresTransaction() bool {
	return true
}

// Execute uploads photos for an activity (typed version)
func (uc *UploadActivityPhotoUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input UploadActivityPhotoInput,
) (UploadActivityPhotoOutput, error) {
	// Check if storage provider is available
	if uc.storage == nil {
		return UploadActivityPhotoOutput{}, fmt.Errorf("storage provider not configured")
	}

	// Upload each photo
	uploadedPhotos := make([]models.ActivityPhoto, 0, len(input.Photos))
	for _, photo := range input.Photos {
		activityPhoto, err := uc.uploadPhoto(ctx, input.ActivityID, photo, tx)
		if err != nil {
			// If any upload fails, we should handle cleanup
			// For now, return error with partial uploads info
			return UploadActivityPhotoOutput{
				ActivityPhotos: uploadedPhotos,
				ActivityID:     input.ActivityID,
				Count:          len(uploadedPhotos),
			}, fmt.Errorf("failed to upload photo %s: %w", photo.Filename, err)
		}
		uploadedPhotos = append(uploadedPhotos, *activityPhoto)
	}

	return UploadActivityPhotoOutput{
		ActivityPhotos: uploadedPhotos,
		ActivityID:     input.ActivityID,
		Count:          len(uploadedPhotos),
	}, nil
}

// uploadPhoto uploads a single photo to storage and returns metadata
func (uc *UploadActivityPhotoUseCase) uploadPhoto(
	ctx context.Context,
	activityID int,
	fileHeader *multipart.FileHeader,
	tx *sql.Tx,
) (*models.ActivityPhoto, error) {
	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode image
	imgFile, err := imageutil.DecodeImage(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Generate and encode thumbnail
	thumbImage := imageutil.GenerateThumbnail(imgFile)
	thumbBytes, err := imageutil.ConvertToJPEG(thumbImage, "jpeg")
	if err != nil {
		return nil, fmt.Errorf("failed to convert thumbnail to JPEG: %w", err)
	}

	// Generate unique storage key
	thumbKey := uc.generateStorageKey(activityID, fileHeader.Filename+"thumb")
	mainKey := uc.generateStorageKey(activityID, fileHeader.Filename)
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload main image to storage
	output, err := uc.storage.Upload(ctx, &types.UploadInput{
		Key:         mainKey,
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

	// Upload thumbnail to storage
	thumbOutput, err := uc.storage.Upload(ctx, &types.UploadInput{
		Key:         thumbKey,
		Body:        bytes.NewReader(thumbBytes),
		ContentType: "image/jpeg",
		Size:        int64(len(thumbBytes)),
		Metadata: map[string]string{
			"activity_id":       fmt.Sprintf("%d", activityID),
			"original_filename": fileHeader.Filename,
			"type":              "thumbnail",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload thumbnail to storage: %w", err)
	}

	// Create activity photo record
	activityPhoto := &models.ActivityPhoto{
		ActivityID:   activityID,
		S3Key:        output.Key,
		ThumbnailKey: thumbOutput.Key,
		ContentType:  contentType,
		FileSize:     fileHeader.Size,
		UploadedAt:   output.UploadedAt,
	}

	dbError := uc.repo.Create(ctx, tx, activityPhoto)

	if dbError != nil {
		return nil, dbError
	}

	return activityPhoto, nil
}

// generateStorageKey creates a unique key for storing the photo
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
