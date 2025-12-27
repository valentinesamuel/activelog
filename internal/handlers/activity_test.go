package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/valentinesamuel/activelog/internal/models"
)

func TestCreateActivity_Validation(t *testing.T) {
	// Test invalid request (missing required field)
	invalidReq := map[string]interface{}{
		"title": "Test Activity",
		// missing activity_type and activity_date
	}

	body, _ := json.Marshal(invalidReq)
	req := httptest.NewRequest("POST", "/api/v1/activities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Note: This requires mocking the repository
	// For now, this is a structural example
	// Full test would need dependency injection or mocking

	// Check that validation errors are returned
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestCreateActivity_ValidRequest(t *testing.T) {
	validReq := models.CreateActivityRequest{
		ActivityType:    "running",
		Title:           "Morning Run",
		DurationMinutes: 30,
		DistanceKm:      5.0,
		ActivityDate:    time.Now(),
	}

	body, _ := json.Marshal(validReq)
	req := httptest.NewRequest("POST", "/api/v1/activities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Would need to test with mock repository
	// This shows structure for future testing
}
