package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/validator"
	appErrors "github.com/valentinesamuel/activelog/pkg/errors"
	"github.com/valentinesamuel/activelog/pkg/response"
)

type ActivityHandler struct {
	repo *repository.ActivityRepository
}

func NewActivityHandler(repo *repository.ActivityRepository) *ActivityHandler {
	return &ActivityHandler{repo: repo}
}

func (a *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateActivityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := validator.Validate(&req)
	if err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  "Validation failed",
			"fields": validationErrors,
		},
		)
		return
	}

	activity := &models.Activity{
		UserID:          1,
		ActivityType:    req.ActivityType,
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		DistanceKm:      req.DistanceKm,
		CaloriesBurned:  req.CaloriesBurned,
		Notes:           req.Notes,
		ActivityDate:    req.ActivityDate,
	}

	if err := a.repo.Create(ctx, activity); err != nil {
		log.Error().Err(err).Msg("❌ Failed to create activity")
		response.Error(w, http.StatusInternalServerError, "❌ Failed to create activity")
		return
	}

	log.Info().Int("activityId", int(activity.ID)).Msg("✅ Activity Created")
	response.SendJSON(w, http.StatusCreated, activity)
}

func (a *ActivityHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	activity, err := a.repo.GetByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Activity not found")
			return
		}

		log.Error().Err(err).Int("id", id).Msg("Failed to get activity")
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activity")
		return
	}
	response.SendJSON(w, http.StatusOK, activity)
}

func (a *ActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request) {

	UserID := 1
	ctx := r.Context()
	filters := models.ActivityFilters{
		ActivityType: r.URL.Query().Get("type"),
		Limit:        10,
		Offset:       0,
	}

	// parsedFilters := parseFilters(r, &filters)

	activities, err := a.repo.GetActivitiesWithTags(ctx, UserID)
	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to list activities")
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	total, err := a.repo.Count(UserID)
	if err != nil {
		log.Err(err).Msg("❌ Failed to count activities")
		total = len(activities)
	}

	response.SendJSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
		"count":      len(activities),
		"total":      total,
		"limit":      filters.Limit,
		"offset":     filters.Offset,
	})
}

func parseFilters(r *http.Request, filters *models.ActivityFilters) models.ActivityFilters {
	// parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > 100 {
				limit = 100
			}
			filters.Limit = limit
		}
	}
	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}
	// Parse date range
	if startStr := r.URL.Query().Get("startDate"); startStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startStr); err == nil {
			filters.StartDate = &startDate
		}
	}
	if endStr := r.URL.Query().Get("endDate"); endStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endStr); err == nil {
			filters.EndDate = &endDate
		}
	}
	return *filters
}

func (h *ActivityHandler) UpdateActivity(w http.ResponseWriter, r *http.Request) {
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

	// Fetch existing activity
	userID := 1
	activity, err := h.repo.GetByID(ctx, int64(id))
	if err != nil || activity.UserID != userID {
		response.Error(w, http.StatusNotFound, "Activity not found")
		return
	}

	// Apply updates
	if req.ActivityType != nil {
		activity.ActivityType = *req.ActivityType
	}
	if req.Title != nil {
		activity.Title = *req.Title
	}
	if req.Description != nil {
		activity.Description = *req.Description
	}
	if req.DurationMinutes != nil {
		activity.DurationMinutes = *req.DurationMinutes
	}
	if req.DistanceKm != nil {
		activity.DistanceKm = *req.DistanceKm
	}
	if req.CaloriesBurned != nil {
		activity.CaloriesBurned = *req.CaloriesBurned
	}
	if req.Notes != nil {
		activity.Notes = *req.Notes
	}
	if req.ActivityDate != nil {
		activity.ActivityDate = *req.ActivityDate
	}

	// Save
	if err := h.repo.Update(id, activity); err != nil {
		log.Error().Err(err).Msg("Failed to update activity")
		response.Error(w, http.StatusInternalServerError, "Failed to update activity")
		return
	}

	response.SendJSON(w, http.StatusOK, activity)
}

func (h *ActivityHandler) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	userID := 1

	if err := h.repo.Delete(id, userID); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to delete activity")
		response.Error(w, http.StatusNotFound, "Activity not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ActivityHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := 1

	var startDate, endDate *time.Time

	if startStr := r.URL.Query().Get("startDate"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			startDate = &parsed
		}
	}

	if endStr := r.URL.Query().Get("endDate"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			endDate = &parsed
		}
	}

	stats, err := h.repo.GetStats(userID, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to get stats")
		response.Error(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	response.SendJSON(w, http.StatusOK, stats)
}
