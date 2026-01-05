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

func NewStatsHandler(repo repository.StatsRepositoryInterface) *StatsHandler {
	return &StatsHandler{repo: repo}
}

func (sh *StatsHandler) GetWeeklyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := 1

	weeklyStats, err := sh.repo.GetWeeklyStats(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Error fetching weekly stats")
		return
	}

	response.SendJSON(w, http.StatusOK, weeklyStats)
}

func (sh *StatsHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := 1

	weeklyStats, err := sh.repo.GetMonthlyStats(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Error fetching weekly stats")
		return
	}

	response.SendJSON(w, http.StatusOK, weeklyStats)
}

func (sh *StatsHandler) GetUserActivitySummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := 1

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

	userID := 1
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
