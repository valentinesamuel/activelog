package handlers

import (
	"fmt"
	"net/http"

	"github.com/valentinesamuel/activelog/internal/repository"
	requestcontext "github.com/valentinesamuel/activelog/internal/requestContext"
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

	requestUser, _ := requestcontext.FromContext(ctx)

	userID := requestUser.Id

	weeklyStats, err := sh.repo.GetWeeklyStats(ctx, userID)
	if err != nil {
		response.Fail(w, r, http.StatusInternalServerError, "Error fetching weekly stats")
		return
	}

	response.Success(w, r, http.StatusOK, weeklyStats)
}

func (sh *StatsHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestUser, _ := requestcontext.FromContext(ctx)

	monthlyStats, err := sh.repo.GetMonthlyStats(ctx, requestUser.Id)
	if err != nil {
		response.Fail(w, r, http.StatusInternalServerError, "Error fetching monthly stats")
		return
	}

	response.Success(w, r, http.StatusOK, monthlyStats)
}

func (sh *StatsHandler) GetUserActivitySummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestUser, _ := requestcontext.FromContext(ctx)

	activitySummary, err := sh.repo.GetUserActivitySummary(ctx, requestUser.Id)
	if err != nil {
		fmt.Println(err)
		response.Fail(w, r, http.StatusInternalServerError, "Error fetching user activity stats")
		return
	}

	response.Success(w, r, http.StatusOK, activitySummary)
}

func (sh *StatsHandler) GetTopTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestUser, _ := requestcontext.FromContext(ctx)

	limit := 10 // Default limit

	// Parse limit from query params if provided
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		var parsedLimit int
		if _, err := fmt.Sscanf(limitParam, "%d", &parsedLimit); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	topTags, err := sh.repo.GetTopTagsByUser(ctx, requestUser.Id, limit)
	if err != nil {
		fmt.Println(err)
		response.Fail(w, r, http.StatusInternalServerError, "Error fetching top tags")
		return
	}

	// Create response with tags and total count
	responseData := map[string]interface{}{
		"tags":              topTags,
		"total_unique_tags": len(topTags),
	}

	response.Success(w, r, http.StatusOK, responseData)
}

func (sh *StatsHandler) GetActivityCountByType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestUser, _ := requestcontext.FromContext(ctx)

	activityBreakdown, err := sh.repo.GetActivityCountByType(ctx, requestUser.Id)
	if err != nil {
		fmt.Println(err)
		response.Fail(w, r, http.StatusInternalServerError, "Error fetching activity breakdown")
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

	response.Success(w, r, http.StatusOK, responseData)
}
