package handlers

import (
	"github.com/valentinesamuel/activelog/pkg/response"
	"net/http"
)

type ActivityHandler struct{}

func NewActivityHandler() *ActivityHandler {
	return &ActivityHandler{}
}

func (a *ActivityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	responseData := map[string]string{
		"status":  "healthy",
		"service": "activelog-api",
	}

	response.SendJSON(w, http.StatusOK, responseData)
}

func (a *ActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	activities := []map[string]interface{}{
		{
			"id":       1,
			"type":     "running",
			"distance": 5.2,
			"duration": 30,
		},
		{
			"id":       2,
			"type":     "basketball",
			"duration": 60,
		},
	}

	response.SendJSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
		"count":      len(activities),
	})
}

func (a *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {

	response.SendJSON(w, http.StatusOK, map[string]string{
		"message": "Activity create (mock)",
	})
}
