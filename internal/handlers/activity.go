package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	requestcontext "github.com/valentinesamuel/activelog/internal/requestContext"
	"github.com/valentinesamuel/activelog/internal/validator"
	appErrors "github.com/valentinesamuel/activelog/pkg/errors"
	"github.com/valentinesamuel/activelog/pkg/query"
	"github.com/valentinesamuel/activelog/pkg/response"
)

// ActivityHandler uses the broker pattern for use case orchestration
// All operations flow through broker → use cases for consistency
type ActivityHandler struct {
	broker             *broker.Broker
	repo               repository.ActivityRepositoryInterface
	createActivityUC   *usecases.CreateActivityUseCase
	getActivityUC      *usecases.GetActivityUseCase
	listActivitiesUC   *usecases.ListActivitiesUseCase
	updateActivityUC   *usecases.UpdateActivityUseCase
	deleteActivityUC   *usecases.DeleteActivityUseCase
	getActivityStatsUC *usecases.GetActivityStatsUseCase
}

type ActivityHandlerDeps struct {
	Broker             *broker.Broker
	Repo               repository.ActivityRepositoryInterface
	CreateActivityUC   *usecases.CreateActivityUseCase
	GetActivityUC      *usecases.GetActivityUseCase
	ListActivitiesUC   *usecases.ListActivitiesUseCase
	UpdateActivityUC   *usecases.UpdateActivityUseCase
	DeleteActivityUC   *usecases.DeleteActivityUseCase
	GetActivityStatsUC *usecases.GetActivityStatsUseCase
}

// NewActivityHandler creates a handler with broker pattern
func NewActivityHandler(
	deps ActivityHandlerDeps,
) *ActivityHandler {
	return &ActivityHandler{
		broker:             deps.Broker,
		repo:               deps.Repo,
		createActivityUC:   deps.CreateActivityUC,
		getActivityUC:      deps.GetActivityUC,
		listActivitiesUC:   deps.ListActivitiesUC,
		updateActivityUC:   deps.UpdateActivityUC,
		deleteActivityUC:   deps.DeleteActivityUC,
		getActivityStatsUC: deps.GetActivityStatsUC,
	}
}

// CreateActivity handles activity creation using broker pattern
// @Summary Create a new activity
// @Description Creates a new activity for the authenticated user
// @Tags Activities
// @Accept json
// @Produce json
// @Param request body models.CreateActivityRequest true "Activity creation request"
// @Success 201 {object} models.Activity "Created activity"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/activities [post]
func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestUser, _ := requestcontext.FromContext(ctx)
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

	// Execute typed use case through broker
	result, err := broker.RunUseCase(
		h.broker,
		ctx,
		h.createActivityUC,
		usecases.CreateActivityInput{
			UserID:  requestUser.Id,
			Request: &req,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create activity")
		response.Error(w, http.StatusInternalServerError, "Failed to create activity")
		return
	}

	log.Info().Int64("activityId", result.ActivityID).Msg("Activity Created")
	response.SendJSON(w, http.StatusCreated, result.Activity)
}

// GetActivity fetches a single activity using broker pattern
// @Summary Get an activity by ID
// @Description Returns a single activity by its ID
// @Tags Activities
// @Produce json
// @Param id path int true "Activity ID"
// @Success 200 {object} models.Activity "Activity found"
// @Failure 400 {object} map[string]string "Invalid activity ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Activity not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/activities/{id} [get]
func (h *ActivityHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	// Execute typed use case through broker
	result, err := broker.RunUseCase(
		h.broker,
		ctx,
		h.getActivityUC,
		usecases.GetActivityInput{
			ActivityID: int64(id),
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

	response.SendJSON(w, http.StatusOK, result.Activity)
}

// ListActivities fetches activities using dynamic filtering with QueryOptions
// @Summary List activities
// @Description Returns a paginated list of activities for the authenticated user with filtering, searching, and sorting
// @Tags Activities
// @Produce json
// @Param filter[activity_type] query string false "Filter by activity type"
// @Param filter[tags.name] query string false "Filter by tag name"
// @Param search[title] query string false "Search in title (case-insensitive)"
// @Param search[description] query string false "Search in description (case-insensitive)"
// @Param order[created_at] query string false "Sort by created_at (ASC or DESC)"
// @Param order[activity_date] query string false "Sort by activity_date (ASC or DESC)"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} map[string]interface{} "Paginated activities with metadata"
// @Failure 400 {object} map[string]string "Invalid query parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/activities [get]
func (h *ActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestUser, ok := requestcontext.FromContext(ctx)

	if !ok {
		log.Error().Msg("Failed to get user from context")
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	// Parse query parameters into QueryOptions
	queryOpts, err := query.ParseQueryParams(r.URL.Query())
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// Define whitelists for security (CRITICAL: Only allow safe columns)
	allowedFilters := []string{
		// Direct columns (main table)
		"activity_type",
		"duration_minutes",
		"distance_km",
		"calories_burned",
		"activity_date",
		"created_at",
		"updated_at",

		// Relationship columns (natural names - auto-JOINs!)
		"tags.name", // Filter by tag name - automatically JOINs tags table
		"tags.id",   // Filter by tag ID
	}

	allowedSearch := []string{
		// Direct columns
		"title",
		"description",
		"notes",

		// Relationship columns (natural names - auto-JOINs!)
		"tags.name", // Search tag names
	}

	allowedOrder := []string{
		// Direct columns
		"created_at",
		"updated_at",
		"activity_date",
		"duration_minutes",
		"distance_km",
		"calories_burned",

		// Relationship columns (natural names - auto-JOINs!)
		"tags.name", // Order by tag name alphabetically
	}

	// Operator whitelisting (v1.1.0+)
	// Define which operators are allowed for each column
	operatorWhitelists := query.OperatorWhitelist{
		// Direct columns - comparison operators
		"activity_date":    query.ComparisonOperators(), // All 6: eq, ne, gt, gte, lt, lte
		"distance_km":      query.ComparisonOperators(),
		"duration_minutes": query.ComparisonOperators(),
		"calories_burned":  query.ComparisonOperators(),
		"created_at":       query.ComparisonOperators(),

		// Direct columns - equality only
		"activity_type": query.EqualityOperators(), // eq, ne only

		// Relationship columns
		"tags.name": query.EqualityOperators(),  // eq, ne for tag names
		"tags.id":   query.StrictEqualityOnly(), // eq only for tag IDs
	}

	// Validate query options against whitelists
	if err := query.ValidateQueryOptions(queryOpts, allowedFilters, allowedSearch, allowedOrder); err != nil {
		log.Warn().Err(err).Msg("Invalid query parameters")
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate operator-based filters (v1.1.0+)
	if err := query.ValidateFilterConditions(queryOpts, allowedFilters, operatorWhitelists); err != nil {
		log.Warn().Err(err).Msg("Invalid filter operator")
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Execute typed use case through broker
	result, err := broker.RunUseCase(
		h.broker,
		ctx,
		h.listActivitiesUC,
		usecases.ListActivitiesInput{
			UserID:       requestUser.Id,
			QueryOptions: queryOpts,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to list activities")
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	// Set cache status headers
	if result.Cache.Hit {
		w.Header().Set("X-Cache-Status", "HIT")
	} else {
		w.Header().Set("X-Cache-Status", "MISS")
		w.Header().Set("X-Cache-TTL", strconv.Itoa(int(result.Cache.TTL.Seconds())))
	}

	// Return standardized response with pagination metadata
	response.SendJSON(w, http.StatusOK, map[string]interface{}{
		"data": result.Result.Data,
		"meta": result.Result.Meta,
	})
}

// UpdateActivity handles activity updates using broker pattern
// @Summary Update an activity
// @Description Updates an existing activity by ID (partial update supported)
// @Tags Activities
// @Accept json
// @Produce json
// @Param id path int true "Activity ID"
// @Param request body models.UpdateActivityRequest true "Activity update request"
// @Success 200 {object} models.Activity "Updated activity"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/activities/{id} [patch]
func (h *ActivityHandler) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestUser, _ := requestcontext.FromContext(ctx)

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

	// Execute typed use case through broker
	result, err := broker.RunUseCase(
		h.broker,
		ctx,
		h.updateActivityUC,
		usecases.UpdateActivityInput{
			UserID:     requestUser.Id,
			ActivityID: id,
			Request:    &req,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to update activity")
		response.Error(w, http.StatusInternalServerError, "Failed to update activity")
		return
	}

	response.SendJSON(w, http.StatusOK, result.Activity)
}

// DeleteActivity handles activity deletion using broker pattern
// @Summary Delete an activity
// @Description Deletes an activity by ID
// @Tags Activities
// @Param id path int true "Activity ID"
// @Success 204 "Activity deleted successfully"
// @Failure 400 {object} map[string]string "Invalid activity ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Activity not found"
// @Security BearerAuth
// @Router /api/v1/activities/{id} [delete]
func (h *ActivityHandler) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestUser, _ := requestcontext.FromContext(ctx)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	// Execute typed use case through broker
	_, err = broker.RunUseCase(
		h.broker,
		ctx,
		h.deleteActivityUC,
		usecases.DeleteActivityInput{
			UserID:     requestUser.Id,
			ActivityID: id,
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
// @Summary Get activity statistics
// @Description Returns aggregated statistics for the authenticated user's activities
// @Tags Activities
// @Produce json
// @Param startDate query string false "Start date filter (RFC3339 format)"
// @Param endDate query string false "End date filter (RFC3339 format)"
// @Success 200 {object} map[string]interface{} "Activity statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/activities/stats [get]
func (h *ActivityHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestUser, _ := requestcontext.FromContext(ctx)

	// Parse query parameters
	input := usecases.GetActivityStatsInput{
		UserID: requestUser.Id,
	}

	if startStr := r.URL.Query().Get("startDate"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			input.StartDate = &parsed
		}
	}

	if endStr := r.URL.Query().Get("endDate"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			input.EndDate = &parsed
		}
	}

	// Execute typed use case through broker
	result, err := broker.RunUseCase(
		h.broker,
		ctx,
		h.getActivityStatsUC,
		input,
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get stats")
		response.Error(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	response.SendJSON(w, http.StatusOK, result.Stats)
}
