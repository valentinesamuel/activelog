package handlers

import (
	"context"
	"net/http"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/pkg/response"
)

type StatsRepositoryInterface interface {
	GetWeeklyStats(ctx context.Context, userID int) (*repository.WeeklyStats, error)
	GetMonthlyStats(ctx context.Context, userID int) (*repository.MonthlyStats, error)
	GetActivityCountByType(ctx context.Context, userID int) (map[string]int, error)
}

type StatsHandler struct {
	repo StatsRepositoryInterface
}

func NewStatsHandler(repo StatsRepositoryInterface) *StatsHandler {
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
