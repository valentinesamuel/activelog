package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/utils"
	"github.com/valentinesamuel/activelog/pkg/logger"
	"github.com/valentinesamuel/activelog/pkg/response"
)

type ActivityPhotoHandler struct {
	brokerInstance         *broker.Broker
	repo                   repository.ActivityRepositoryInterface
	uploadActivityPhotosUC broker.UseCase
}

func NewActivityPhotoHandler(brokerInstance *broker.Broker, repo repository.ActivityRepositoryInterface, uploadActivityPhotosUC broker.UseCase) *ActivityPhotoHandler {
	return &ActivityPhotoHandler{
		brokerInstance:         brokerInstance,
		repo:                   repo,
		uploadActivityPhotosUC: uploadActivityPhotosUC,
	}
}

func (h *ActivityPhotoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.Error().Err(err).Msg("❌ Failed to upload activity photo")
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	contentType := r.Header.Get("Content-Type")
	logger.Info().Str("content_type", contentType).Msg("📤 Received upload request")

	err = r.ParseMultipartForm(50 << 20)
	if err != nil {
		logger.Error().Err(err).Str("content_type", contentType).Msg("❌ Failed to parse multipart form")
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	photos := r.MultipartForm.File["photos"]
	if len(photos) > 5 {
		response.Error(w, http.StatusBadRequest, "Too many files")
		return
	}

	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	for _, photo := range photos {
		file, err := photo.Open()
		if err != nil {
			continue
		}

		contentType, err := utils.DetectFileType(file)
		file.Close()

		if !allowedTypes[contentType] {
			response.Error(w, http.StatusBadRequest, "Invalid file format")
		}

		fmt.Printf("File: %s, Type: %s\n", photo.Filename, contentType)
	}

	result, err := h.brokerInstance.RunUseCases(
		ctx,
		[]broker.UseCase{h.uploadActivityPhotosUC},
		map[string]interface{}{
			"user_id":     1,
			"photos":      &photos,
			"activity_id": id,
		},
	)

	if err != nil {
		logger.Error().Err(err).Msg("❌ Failed to upload activity photo")
		response.Error(w, http.StatusInternalServerError, "Failed to upload activity photo")
		return
	}
	// Extract activity from result
	activityPhotos := result["activityPhotos"]
	log.Info().Interface("activityId", result["activity_id"]).Msg("✅ Activity PhotosCreated")
	response.SendJSON(w, http.StatusCreated, activityPhotos)

}
