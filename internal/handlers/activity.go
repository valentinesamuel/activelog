package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
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
		response.Error(w, http.StatusInternalServerError, "Failed to create activity")
		return
	}

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
		response.Error(w, http.StatusNotFound, "Activity not found")
		return
	}
	response.SendJSON(w, http.StatusOK, activity)
}

func (a *ActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	UserID := 1

	activities, err := a.repo.ListByUser(ctx, UserID)
	if err != nil {
		fmt.Println("🛑 Error fetching activities: \n🛑", err)
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	response.SendJSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
		"count":      len(activities),
	})
}
