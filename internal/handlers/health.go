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

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	practice.DemoPointers()

	responseData := map[string]string{
		"status":  "healthy",
		"service": "activelog-api",
	}

	response.SendJSON(w, http.StatusOK, responseData)
}
