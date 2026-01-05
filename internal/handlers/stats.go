package handlers

import (
	"fmt"
	"net/http"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/pkg/response"
)

type StatsHandler struct {
	repo repository.StatsRepositoryInterface
}

// getUserIDFromContext extracts the authenticated user ID from request context
func getUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("user_id").(int)
	return userID, ok
}

func NewStatsHandler(repo repository.StatsRepositoryInterface) *StatsHandler {
	return &StatsHandler{repo: repo}
}

func (sh *StatsHandler) GetWeeklyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := getUserIDFromContext(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	weeklyStats, err := sh.repo.GetWeeklyStats(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Error fetching weekly stats")
		return
	}

	response.SendJSON(w, http.StatusOK, weeklyStats)
}

func (sh *StatsHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := getUserIDFromContext(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	monthlyStats, err := sh.repo.GetMonthlyStats(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Error fetching monthly stats")
		return
	}

	response.SendJSON(w, http.StatusOK, monthlyStats)
}

func (sh *StatsHandler) GetUserActivitySummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := getUserIDFromContext(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	activitySummary, err := sh.repo.GetUserActivitySummary(ctx, userID)
	if err != nil {
		fmt.Println(err)
		response.Error(w, http.StatusInternalServerError, "Error fetching user activity stats")
		return
	}

	response.SendJSON(w, http.StatusOK, activitySummary)
}

func (sh *StatsHandler) GetTopTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := getUserIDFromContext(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	limit := 10 // Default limit

	// Parse limit from query params if provided
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		var parsedLimit int
		if _, err := fmt.Sscanf(limitParam, "%d", &parsedLimit); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	topTags, err := sh.repo.GetTopTagsByUser(ctx, userID, limit)
	if err != nil {
		fmt.Println(err)
		response.Error(w, http.StatusInternalServerError, "Error fetching top tags")
		return
	}

	// Create response with tags and total count
	responseData := map[string]interface{}{
		"tags":              topTags,
		"total_unique_tags": len(topTags),
	}

	response.SendJSON(w, http.StatusOK, responseData)
}

func (sh *StatsHandler) GetActivityCountByType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := getUserIDFromContext(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	activityBreakdown, err := sh.repo.GetActivityCountByType(ctx, userID)
	if err != nil {
		fmt.Println(err)
		response.Error(w, http.StatusInternalServerError, "Error fetching activity breakdown")
		return
	}

	// Calculate total activities
	totalActivities := 0
	for _, count := range activityBreakdown {
		totalActivities += count
	}

	// Create response with breakdown and total
	responseData := map[string]interface{}{
		"activity_breakdown": activityBreakdown,
		"total_activities":   totalActivities,
	}

	response.SendJSON(w, http.StatusOK, responseData)
}
