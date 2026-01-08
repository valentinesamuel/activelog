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
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/validator"
	appErrors "github.com/valentinesamuel/activelog/pkg/errors"
	"github.com/valentinesamuel/activelog/pkg/query"
	"github.com/valentinesamuel/activelog/pkg/response"
)

// ActivityHandler uses the broker pattern for use case orchestration
// All operations flow through broker → use cases for consistency
type ActivityHandler struct {
	broker               *broker.Broker
	repo                 repository.ActivityRepositoryInterface
	createActivityUC     broker.UseCase
	getActivityUC        broker.UseCase
	listActivitiesUC     broker.UseCase
	updateActivityUC     broker.UseCase
	deleteActivityUC     broker.UseCase
	getActivityStatsUC   broker.UseCase
}

// NewActivityHandler creates a handler with broker pattern
func NewActivityHandler(
	brokerInstance *broker.Broker,
	repo repository.ActivityRepositoryInterface,
	createActivityUC broker.UseCase,
	getActivityUC broker.UseCase,
	listActivitiesUC broker.UseCase,
	updateActivityUC broker.UseCase,
	deleteActivityUC broker.UseCase,
	getActivityStatsUC broker.UseCase,
) *ActivityHandler {
	return &ActivityHandler{
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
func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
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
func (h *ActivityHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
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

// ListActivities fetches activities using dynamic filtering with QueryOptions
// Supports flexible filtering, searching, sorting, and pagination via URL parameters:
//   - filter[column]=value - Filter by exact match (e.g., filter[activity_type]=running)
//   - filter[tags]=value - Filter by tag name (automatically JOINs tags table)
//   - search[column]=value - Case-insensitive search (e.g., search[title]=morning)
//   - order[column]=ASC|DESC - Sort results (e.g., order[created_at]=DESC)
//   - page=N - Page number (default: 1)
//   - limit=N - Items per page (default: 10, max: 100)
//
// Example URLs:
//   - /activities?filter[activity_type]=running&page=1&limit=20
//   - /activities?filter[tags]=cardio&search[title]=morning&order[created_at]=DESC
//   - /activities?filter[activity_type]=[running,cycling]&limit=50
func (h *ActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	UserID := 1 // TODO: Get from auth context
	ctx := r.Context()

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
		"tags.name": query.EqualityOperators(), // eq, ne for tag names
		"tags.id":   query.StrictEqualityOnly(), // eq only for tag IDs
	}

	// Validate query options against whitelists
	if err := query.ValidateQueryOptions(queryOpts, allowedFilters, allowedSearch, allowedOrder); err != nil {
		log.Warn().Err(err).Msg("❌ Invalid query parameters")
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate operator-based filters (v1.1.0+)
	if err := query.ValidateFilterConditions(queryOpts, allowedFilters, operatorWhitelists); err != nil {
		log.Warn().Err(err).Msg("❌ Invalid filter operator")
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Execute use case through broker with QueryOptions
	result, err := h.broker.RunUseCases(
		ctx,
		[]broker.UseCase{h.listActivitiesUC},
		map[string]interface{}{
			"user_id":       UserID,
			"query_options": queryOpts, // NEW: Pass QueryOptions instead of legacy filters
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to list activities")
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	// Extract paginated result
	paginatedResult := result["result"].(*query.PaginatedResult)

	// Return standardized response with pagination metadata
	response.SendJSON(w, http.StatusOK, map[string]interface{}{
		"data": paginatedResult.Data,
		"meta": paginatedResult.Meta,
	})
}

// UpdateActivity handles activity updates using broker pattern
// Similar to kuja_user_ms: resetShopPassword method (line 373-380)
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
func (h *ActivityHandler) DeleteActivity(w http.ResponseWriter, r *http.Request) {
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
func (h *ActivityHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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
