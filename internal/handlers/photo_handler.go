package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/valentinesamuel/activelog/internal/application/activityPhoto/usecases"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/repository"
	requestcontext "github.com/valentinesamuel/activelog/internal/requestContext"
	"github.com/valentinesamuel/activelog/internal/utils"
	"github.com/valentinesamuel/activelog/pkg/logger"
	"github.com/valentinesamuel/activelog/pkg/response"
)

type ActivityPhotoHandler struct {
	brokerInstance         *broker.Broker
	repo                   repository.ActivityPhotoRepositoryInterface
	uploadActivityPhotosUC *usecases.UploadActivityPhotoUseCase
	getActivityPhotosUC    *usecases.GetActivityPhotoUseCase
}

func NewActivityPhotoHandler(
	brokerInstance *broker.Broker,
	repo repository.ActivityPhotoRepositoryInterface,
	uploadActivityPhotosUC *usecases.UploadActivityPhotoUseCase,
	getActivityPhotosUC *usecases.GetActivityPhotoUseCase,
) *ActivityPhotoHandler {
	return &ActivityPhotoHandler{
		brokerInstance:         brokerInstance,
		repo:                   repo,
		uploadActivityPhotosUC: uploadActivityPhotosUC,
		getActivityPhotosUC:    getActivityPhotosUC,
	}
}

func (h *ActivityPhotoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestUser, _ := requestcontext.FromContext(ctx)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.Error().Err(err).Msg("Failed to upload activity photo")
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	contentType := r.Header.Get("Content-Type")
	logger.Info().Str("content_type", contentType).Msg("Received upload request")

	err = r.ParseMultipartForm(50 << 20)
	if err != nil {
		logger.Error().Err(err).Str("content_type", contentType).Msg("Failed to parse multipart form")
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
			return
		}

		fmt.Printf("File: %s, Type: %s\n", photo.Filename, contentType)
	}

	// Execute typed use case through broker
	result, err := broker.RunUseCase(
		h.brokerInstance,
		ctx,
		h.uploadActivityPhotosUC,
		usecases.UploadActivityPhotoInput{
			UserID:     requestUser.Id,
			ActivityID: id,
			Photos:     photos,
		},
	)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to upload activity photo")
		response.Error(w, http.StatusInternalServerError, "Failed to upload activity photo")
		return
	}

	log.Info().Int("activityId", result.ActivityID).Msg("Activity Photos Created")
	response.SendJSON(w, http.StatusCreated, result.ActivityPhotos)
}

func (h *ActivityPhotoHandler) GetActivityPhoto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get activity photo")
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	// Execute typed use case through broker
	result, err := broker.RunUseCase(
		h.brokerInstance,
		ctx,
		h.getActivityPhotosUC,
		usecases.GetActivityPhotosInput{
			ActivityID: id,
		},
	)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to get activity photos")
		response.Error(w, http.StatusInternalServerError, "Failed to get activity photos")
		return
	}

	log.Info().Int("activityId", id).Int("count", len(result.Photos)).Msg("Activity Photos retrieved")
	response.SendJSON(w, http.StatusOK, result.Photos)
}
