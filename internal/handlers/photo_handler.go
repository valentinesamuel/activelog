package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

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
		response.Fail(w, r, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	contentType := r.Header.Get("Content-Type")
	logger.Info().Str("content_type", contentType).Msg("Received upload request")

	err = r.ParseMultipartForm(50 << 20)
	if err != nil {
		logger.Error().Err(err).Str("content_type", contentType).Msg("Failed to parse multipart form")
		response.Fail(w, r, http.StatusBadRequest, err.Error())
		return
	}

	photos := r.MultipartForm.File["photos"]
	if len(photos) > 5 {
		response.Fail(w, r, http.StatusBadRequest, "Too many files")
		return
	}

	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	// Validate file types concurrently using a semaphore (chan struct{} with cap=5).
	// Each goroutine acquires a slot before opening/inspecting the file, then releases it.
	sem := make(chan struct{}, 5)
	type validationErr struct{ err error }
	validationCh := make(chan validationErr, len(photos))

	var wg sync.WaitGroup
	for _, photo := range photos {
		p := photo // capture loop variable
		wg.Add(1)
		sem <- struct{}{} // acquire semaphore slot
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			file, err := p.Open()
			if err != nil {
				validationCh <- validationErr{fmt.Errorf("cannot open %s: %w", p.Filename, err)}
				return
			}
			defer file.Close()

			detectedType, err := utils.DetectFileType(file)
			if err != nil {
				validationCh <- validationErr{fmt.Errorf("cannot detect type for %s: %w", p.Filename, err)}
				return
			}
			if !allowedTypes[detectedType] {
				validationCh <- validationErr{fmt.Errorf("invalid file format for %s", p.Filename)}
				return
			}

			logger.Info().Str("file", p.Filename).Str("type", detectedType).Msg("photo validated")
			validationCh <- validationErr{}
		}()
	}

	wg.Wait()
	close(validationCh)

	for v := range validationCh {
		if v.err != nil {
			response.Fail(w, r, http.StatusBadRequest, v.err.Error())
			return
		}
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
		response.Fail(w, r, http.StatusInternalServerError, "Failed to upload activity photo")
		return
	}

	log.Info().Int("activityId", result.ActivityID).Msg("Activity Photos Created")
	response.Success(w, r, http.StatusCreated, result.ActivityPhotos)
}

func (h *ActivityPhotoHandler) GetActivityPhoto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get activity photo")
		response.Fail(w, r, http.StatusBadRequest, "Invalid activity ID")
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
		response.Fail(w, r, http.StatusInternalServerError, "Failed to get activity photos")
		return
	}

	log.Info().Int("activityId", id).Int("count", len(result.Photos)).Msg("Activity Photos retrieved")
	response.Success(w, r, http.StatusOK, result.Photos)
}
