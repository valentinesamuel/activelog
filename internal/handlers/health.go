package handlers

import (
	"github.com/valentinesamuel/activelog/pkg/response"
	"net/http"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	responseData := map[string]string{
		"status":  "healthy",
		"service": "activelog-api",
	}

	response.SendJSON(w, http.StatusOK, responseData)
}
