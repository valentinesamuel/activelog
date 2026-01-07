package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/validator"
	appErrors "github.com/valentinesamuel/activelog/pkg/errors"
	"github.com/valentinesamuel/activelog/pkg/response"
)

// ActivityHandlerV2 uses the broker pattern for use case orchestration
// Similar to kuja_user_ms auth controller
// All operations now flow through broker → use cases for consistency
type ActivityHandlerV2 struct {
	broker               *broker.Broker
	repo                 ActivityRepositoryInterface
	createActivityUC     broker.UseCase
	getActivityUC        broker.UseCase // NEW: Read single activity
	listActivitiesUC     broker.UseCase // NEW: List with filters
	updateActivityUC     broker.UseCase
	deleteActivityUC     broker.UseCase
	getActivityStatsUC   broker.UseCase // NEW: Activity stats
}

// NewActivityHandlerV2 creates a handler with broker pattern
// Follows the dependency injection pattern from kuja_user_ms
func NewActivityHandlerV2(
	brokerInstance *broker.Broker,
	repo ActivityRepositoryInterface,
	createActivityUC broker.UseCase,
	getActivityUC broker.UseCase,
	listActivitiesUC broker.UseCase,
	updateActivityUC broker.UseCase,
	deleteActivityUC broker.UseCase,
	getActivityStatsUC broker.UseCase,
) *ActivityHandlerV2 {
	return &ActivityHandlerV2{
		broker:             brokerInstance,
		repo:               repo,
		createActivityUC:   createActivityUC,
		getActivityUC:      getActivityUC,
		listActivitiesUC:   listActivitiesUC,
		updateActivityUC:   updateActivityUC,
		deleteActivityUC:   deleteActivityUC,
		getActivityStatsUC: getActivityStatsUC,
	}
}

// CreateActivity handles activity creation using broker pattern
// Similar to kuja_user_ms: shopSignin method (line 316-336)
func (h *ActivityHandlerV2) CreateActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateActivityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	err := validator.Validate(&req)
	if err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  "Validation failed",
			"fields": validationErrors,
		})
		return
	}

	// Execute use case through broker
	// Pattern: broker.runUsecases([useCase], input)
	result, err := h.broker.RunUseCases(
		ctx,
		[]broker.UseCase{h.createActivityUC},
		map[string]interface{}{
			"user_id": 1, // TODO: Get from auth context
			"request": &req,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to create activity")
		response.Error(w, http.StatusInternalServerError, "Failed to create activity")
		return
	}

	// Extract activity from result
	activity := result["activity"]
	log.Info().Interface("activityId", result["activity_id"]).Msg("✅ Activity Created")
	response.SendJSON(w, http.StatusCreated, activity)
}

// GetActivity fetches a single activity using broker pattern
func (h *ActivityHandlerV2) GetActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	// Execute use case through broker (consistent with other operations)
	result, err := h.broker.RunUseCases(
		ctx,
		[]broker.UseCase{h.getActivityUC},
		map[string]interface{}{
			"activity_id": int64(id),
		},
	)

	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Activity not found")
			return
		}

		log.Error().Err(err).Int("id", id).Msg("Failed to get activity")
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activity")
		return
	}

	activity := result["activity"]
	response.SendJSON(w, http.StatusOK, activity)
}

// ListActivities fetches activities using broker pattern
func (h *ActivityHandlerV2) ListActivities(w http.ResponseWriter, r *http.Request) {
	UserID := 1 // TODO: Get from auth context
	ctx := r.Context()

	filters := models.ActivityFilters{
		ActivityType: r.URL.Query().Get("type"),
		Limit:        10,
		Offset:       0,
	}

	parsedFilters := parseFilters(r, &filters)

	// Execute use case through broker
	result, err := h.broker.RunUseCases(
		ctx,
		[]broker.UseCase{h.listActivitiesUC},
		map[string]interface{}{
			"user_id": UserID,
			"filters": parsedFilters,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to list activities")
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	// Extract results from use case output
	activities := result["activities"]
	total := result["total"]
	count := result["count"]

	response.SendJSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
		"count":      count,
		"total":      total,
		"limit":      parsedFilters.Limit,
		"offset":     parsedFilters.Offset,
	})
}

// UpdateActivity handles activity updates using broker pattern
// Similar to kuja_user_ms: resetShopPassword method (line 373-380)
func (h *ActivityHandlerV2) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	var req models.UpdateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate
	if err := validator.Validate(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  "Validation failed",
			"fields": validationErrors,
		})
		return
	}

	// Execute use case through broker
	result, err := h.broker.RunUseCases(
		ctx,
		[]broker.UseCase{h.updateActivityUC},
		map[string]interface{}{
			"user_id":     1, // TODO: Get from auth context
			"activity_id": id,
			"request":     &req,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to update activity")
		response.Error(w, http.StatusInternalServerError, "Failed to update activity")
		return
	}

	activity := result["activity"]
	response.SendJSON(w, http.StatusOK, activity)
}

// DeleteActivity handles activity deletion using broker pattern
// Similar to kuja_user_ms: logout method (line 413-434)
func (h *ActivityHandlerV2) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	userID := 1 // TODO: Get from auth context

	// Execute use case through broker
	_, err = h.broker.RunUseCases(
		ctx,
		[]broker.UseCase{h.deleteActivityUC},
		map[string]interface{}{
			"user_id":     userID,
			"activity_id": id,
		},
	)

	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to delete activity")
		response.Error(w, http.StatusNotFound, "Activity not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetStats fetches activity statistics using broker pattern
func (h *ActivityHandlerV2) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := 1 // TODO: Get from auth context

	// Parse query parameters
	input := map[string]interface{}{
		"user_id": userID,
	}

	if startStr := r.URL.Query().Get("startDate"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			input["start_date"] = parsed
		}
	}

	if endStr := r.URL.Query().Get("endDate"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			input["end_date"] = parsed
		}
	}

	// Execute use case through broker
	result, err := h.broker.RunUseCases(
		ctx,
		[]broker.UseCase{h.getActivityStatsUC},
		input,
	)

	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to get stats")
		response.Error(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	stats := result["stats"]

	response.SendJSON(w, http.StatusOK, stats)
}
