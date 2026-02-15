package handlers

import (
	"net/http"

	"github.com/valentinesamuel/activelog/internal/practice"
	"github.com/valentinesamuel/activelog/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// ServeHTTP handles health check requests
// @Summary Health check
// @Description Returns the health status of the API service
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "Service is healthy"
// @Router /health [get]
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	practice.DemoPointers()

	responseData := map[string]string{
		"status":  "healthy",
		"service": "activelog-api",
	}

	response.SendJSON(w, http.StatusOK, responseData)
}
