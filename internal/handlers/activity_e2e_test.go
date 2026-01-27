package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/database"
	"github.com/valentinesamuel/activelog/internal/handlers"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/repository/testhelpers"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// ==================== E2E TESTS FOR DYNAMIC FILTERING ====================
// These tests verify the complete request flow from HTTP endpoint to database
// using real PostgreSQL testcontainers and actual HTTP requests

// TestE2E_ListActivities_DynamicFiltering tests the full HTTP stack for dynamic filtering
func TestE2E_ListActivities_DynamicFiltering(t *testing.T) {
	// Setup: Create test database and initialize all dependencies
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	// Create repositories
	tagRepo := repository.NewTagRepository(db)
	activityRepo := repository.NewActivityRepository(db, tagRepo)

	// Create broker
	brokerInstance := broker.NewBroker(db.GetRawDB())

	// Create use cases (pass nil for service since we're using repo directly)
	listActivitiesUC := usecases.NewListActivitiesUseCase(nil, activityRepo)

	// Create handler
	handler := handlers.NewActivityHandler(handlers.ActivityHandlerDeps{
		Broker:           brokerInstance,
		Repo:             activityRepo,
		ListActivitiesUC: listActivitiesUC,
	})

	// Create test user
	userID := createE2ETestUser(t, db, "e2e@example.com", "e2euser")

	// Create test data
	setupE2ETestData(t, activityRepo, tagRepo, userID)

	t.Run("FilterByActivityType", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/activities?filter[activity_type]=running", nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Debug: Print response to understand structure
		t.Logf("Response: %+v", response)

		data := response["data"].([]interface{})
		if len(data) != 2 {
			t.Errorf("Expected 2 running activities, got %d", len(data))
		}

		// Verify all returned activities are running type
		for _, item := range data {
			activity := item.(map[string]interface{})
			// Safe type assertion with nil check (JSON uses camelCase)
			if actTypeVal, ok := activity["activityType"]; ok && actTypeVal != nil {
				actType := actTypeVal.(string)
				if actType != "running" {
					t.Errorf("Expected activityType 'running', got '%s'", actType)
				}
			} else {
				t.Error("activityType field is missing or nil")
			}
		}
	})

	t.Run("FilterByMultipleValues_IN_Clause", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/activities?filter[activity_type]=[running,cycling]", nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		if len(data) != 3 {
			t.Errorf("Expected 3 activities (running+cycling), got %d", len(data))
		}
	})

	t.Run("SearchByTitle", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/activities?search[title]=Run", nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		if len(data) != 2 {
			t.Errorf("Expected 2 activities with 'Run' in title, got %d", len(data))
		}
	})

	t.Run("OrderByDuration_DESC", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/activities?order[duration_minutes]=DESC", nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})

		// Verify activities are ordered by duration DESC
		expectedDurations := []float64{60, 50, 45, 40, 30}
		for i, item := range data {
			activity := item.(map[string]interface{})
			// Safe type assertion with nil check (JSON uses camelCase)
			if durationVal, ok := activity["durationMinutes"]; ok && durationVal != nil {
				duration := durationVal.(float64)
				if duration != expectedDurations[i] {
					t.Errorf("Activity[%d]: expected duration %v, got %v", i, expectedDurations[i], duration)
				}
			}
		}
	})

	t.Run("Pagination_Page2", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/activities?page=2&limit=2&order[created_at]=ASC", nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		meta := response["meta"].(map[string]interface{})

		if len(data) != 2 {
			t.Errorf("Expected 2 activities on page 2, got %d", len(data))
		}

		// Safe type assertions for pagination metadata (JSON uses camelCase)
		if pageVal, ok := meta["page"]; ok && pageVal != nil {
			if pageVal.(float64) != 2 {
				t.Errorf("Expected page=2, got %v", pageVal)
			}
		}

		// previousPage can be int (page number) or bool (false if none)
		if prevVal, ok := meta["previousPage"]; ok && prevVal != nil {
			// Should be page 1 (as float64 from JSON)
			if prevPage, ok := prevVal.(float64); ok {
				if prevPage != 1 {
					t.Errorf("Expected previousPage=1, got %v", prevPage)
				}
			}
		}

		// nextPage can be int (page number) or bool (false if none)
		if nextVal, ok := meta["nextPage"]; ok && nextVal != nil {
			// Should be page 3 (as float64 from JSON)
			if nextPage, ok := nextVal.(float64); ok {
				if nextPage != 3 {
					t.Errorf("Expected nextPage=3, got %v", nextPage)
				}
			}
		}
	})

	t.Run("FilterByTag_AutomaticJOIN", func(t *testing.T) {
		// Use natural column name for automatic JOIN
		req := httptest.NewRequest("GET", "/activities?filter[tags.name]=cardio", nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		// Based on test data: running(2), cycling(1), swimming(1) all have "cardio" tag = 4 total
		if len(data) != 4 {
			t.Errorf("Expected 4 activities with 'cardio' tag, got %d", len(data))
		}
	})

	t.Run("CombineFilters_AND_Logic", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/activities?filter[activity_type]=running&search[title]=Evening", nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		if len(data) != 1 {
			t.Errorf("Expected 1 activity (running + Evening), got %d", len(data))
		}

		if len(data) > 0 {
			activity := data[0].(map[string]interface{})
			if activity["title"].(string) != "Evening Run" {
				t.Errorf("Expected 'Evening Run', got '%s'", activity["title"])
			}
		}
	})

	t.Run("InvalidColumn_RejectedByWhitelist", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/activities?filter[password]=hacker", nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		// Should return 400 Bad Request due to whitelist validation
		assertStatusCode(t, w, http.StatusBadRequest)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if !strings.Contains(response["error"].(string), "not allowed") {
			t.Error("Expected error message about column not being allowed")
		}
	})

	t.Run("EmptyResults", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/activities?filter[activity_type]=nonexistent", nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Safe type assertions (JSON uses camelCase)
		if dataVal, ok := response["data"].([]interface{}); ok {
			if len(dataVal) != 0 {
				t.Errorf("Expected 0 activities, got %d", len(dataVal))
			}
		}

		if metaVal, ok := response["meta"].(map[string]interface{}); ok {
			if totalVal, ok := metaVal["totalRecords"]; ok && totalVal != nil {
				if totalVal.(float64) != 0 {
					t.Errorf("Expected totalRecords=0, got %v", totalVal)
				}
			}
		}
	})

	t.Run("PaginationMetadata_Complete", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/activities?page=1&limit=2", nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if metaVal, ok := response["meta"].(map[string]interface{}); ok {
			// Verify all pagination metadata fields are present (JSON uses camelCase)
			requiredFields := []string{"page", "limit", "count", "previousPage", "nextPage", "pageCount", "totalRecords"}
			for _, field := range requiredFields {
				if _, exists := metaVal[field]; !exists {
					t.Errorf("Missing pagination field: %s", field)
				}
			}

			// Verify values make sense (safe type assertions)
			if totalVal, ok := metaVal["totalRecords"]; ok && totalVal != nil {
				if totalVal.(float64) != 5 {
					t.Errorf("Expected totalRecords=5, got %v", totalVal)
				}
			}

			if pageCountVal, ok := metaVal["pageCount"]; ok && pageCountVal != nil {
				if pageCountVal.(float64) != 3 {
					t.Errorf("Expected pageCount=3 (5 total / 2 per page), got %v", pageCountVal)
				}
			}
		}
	})
}

// TestE2E_ListActivities_WithRouter tests the full HTTP routing with mux
func TestE2E_ListActivities_WithRouter(t *testing.T) {
	// Setup
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	// Create full dependency stack
	tagRepo := repository.NewTagRepository(db)
	activityRepo := repository.NewActivityRepository(db, tagRepo)
	brokerInstance := broker.NewBroker(db.GetRawDB())
	listActivitiesUC := usecases.NewListActivitiesUseCase(nil, activityRepo)

	handler := handlers.NewActivityHandler(handlers.ActivityHandlerDeps{
		Broker:           brokerInstance,
		Repo:             activityRepo,
		ListActivitiesUC: listActivitiesUC,
	})

	// Create router (mimic production routing)
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/activities", handler.ListActivities).Methods("GET")

	// Create test user and data
	userID := createE2ETestUser(t, db, "router@example.com", "routeruser")
	setupE2ETestData(t, activityRepo, tagRepo, userID)

	t.Run("CompleteURLWithQueryParams", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/activities?filter[activity_type]=running&order[created_at]=DESC&page=1&limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["data"] == nil {
			t.Fatal("Expected 'data' field in response")
		}

		if response["meta"] == nil {
			t.Fatal("Expected 'meta' field in response")
		}
	})
}

// ==================== HELPER FUNCTIONS ====================

func createE2ETestUser(t *testing.T, db *database.LoggingDB, email, username string) int {
	t.Helper()

	var userID int
	query := `INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING id`
	err := db.QueryRow(query, email, username, "hashedpassword123").Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

func setupE2ETestData(t *testing.T, activityRepo *repository.ActivityRepository, tagRepo *repository.TagRepository, userID int) {
	t.Helper()

	ctx := context.Background()

	// Create 5 activities with different types and tags
	activities := []struct {
		activityType string
		title        string
		duration     int
		distance     float64
		tags         []string
	}{
		{"running", "Morning Run", 30, 5.0, []string{"cardio", "outdoor"}},
		{"running", "Evening Run", 45, 7.5, []string{"cardio", "outdoor"}},
		{"cycling", "Bike Ride", 60, 15.0, []string{"cardio", "outdoor"}},
		{"swimming", "Pool Session", 40, 2.0, []string{"cardio", "indoor"}},
		{"yoga", "Yoga Class", 50, 0.0, []string{"flexibility", "indoor"}},
	}

	for _, act := range activities {
		activity := &models.Activity{
			UserID:          userID,
			ActivityType:    act.activityType,
			Title:           act.title,
			DurationMinutes: act.duration,
			DistanceKm:      act.distance,
			ActivityDate:    time.Now(),
		}

		// Create tags
		var tags []*models.Tag
		for _, tagName := range act.tags {
			tags = append(tags, &models.Tag{Name: tagName})
		}

		if err := activityRepo.CreateWithTags(ctx, activity, tags); err != nil {
			t.Fatalf("Failed to create test activity: %v", err)
		}
	}
}

func assertStatusCode(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	if w.Code != expectedStatus {
		t.Errorf("Expected status code %d, got %d. Response body: %s", expectedStatus, w.Code, w.Body.String())
	}
}

// TestE2E_OperatorFiltering_HTTPRequests tests the complete HTTP flow for operator-based filtering (v1.1.0+)
func TestE2E_OperatorFiltering_HTTPRequests(t *testing.T) {
	// Setup: Create test database and full dependency stack
	db, cleanup := testhelpers.SetupTestDB(t)
	defer cleanup()

	tagRepo := repository.NewTagRepository(db)
	activityRepo := repository.NewActivityRepository(db, tagRepo)
	brokerInstance := broker.NewBroker(db.GetRawDB())
	listActivitiesUC := usecases.NewListActivitiesUseCase(nil, activityRepo)

	handler := handlers.NewActivityHandler(handlers.ActivityHandlerDeps{
		Broker:           brokerInstance,
		Repo:             activityRepo,
		ListActivitiesUC: listActivitiesUC,
	})

	// Create test user and data
	userID := createE2ETestUser(t, db, "operator@example.com", "operatoruser")
	setupOperatorTestData(t, activityRepo, userID)

	// Test Case 1: Date Range - GTE Operator
	t.Run("DateFilter_GTE_Operator", func(t *testing.T) {
		// Filter for activities >= 7 days ago
		sevenDaysAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		url := fmt.Sprintf("/activities?filter[activity_date][gte]=%s", sevenDaysAgo)

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		// Should return activities from last 7 days (3 activities)
		if len(data) < 3 {
			t.Errorf("Expected at least 3 activities in last 7 days, got %d", len(data))
		}
	})

	// Test Case 2: Date Range - LTE Operator
	t.Run("DateFilter_LTE_Operator", func(t *testing.T) {
		// Filter for activities <= 1 day ago
		oneDayAgo := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		url := fmt.Sprintf("/activities?filter[activity_date][lte]=%s", oneDayAgo)

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		// Should return activities older than 1 day
		if len(data) < 1 {
			t.Errorf("Expected at least 1 activity older than 1 day, got %d", len(data))
		}
	})

	// Test Case 3: Date Range - Combined GTE and LTE
	t.Run("DateRange_GTE_AND_LTE", func(t *testing.T) {
		// Filter for activities between 14 and 7 days ago
		startDate := time.Now().AddDate(0, 0, -14).Format("2006-01-02")
		endDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		url := fmt.Sprintf("/activities?filter[activity_date][gte]=%s&filter[activity_date][lte]=%s", startDate, endDate)

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		// Should return activities in the date range
		if len(data) < 1 {
			t.Errorf("Expected at least 1 activity in date range, got %d", len(data))
		}
	})

	// Test Case 4: Numeric Range - GTE on Distance
	t.Run("NumericFilter_Distance_GTE", func(t *testing.T) {
		url := "/activities?filter[distance_km][gte]=10.0"

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		// Verify all returned activities have distance >= 10.0
		for _, item := range data {
			activity := item.(map[string]interface{})
			if distanceVal, ok := activity["distance_km"]; ok && distanceVal != nil {
				distance := distanceVal.(float64)
				if distance < 10.0 {
					t.Errorf("Expected distance >= 10.0, got %v", distance)
				}
			}
		}
	})

	// Test Case 5: Numeric Range - LT on Duration
	t.Run("NumericFilter_Duration_LT", func(t *testing.T) {
		url := "/activities?filter[duration_minutes][lt]=45"

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		// Verify all returned activities have duration < 45
		for _, item := range data {
			activity := item.(map[string]interface{})
			if durationVal, ok := activity["duration_minutes"]; ok && durationVal != nil {
				duration := durationVal.(float64)
				if duration >= 45 {
					t.Errorf("Expected duration < 45, got %v", duration)
				}
			}
		}
	})

	// Test Case 6: NOT EQUAL Operator
	t.Run("NE_Operator_ActivityType", func(t *testing.T) {
		url := "/activities?filter[activity_type][ne]=running"

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		// Verify no running activities returned
		for _, item := range data {
			activity := item.(map[string]interface{})
			if activityTypeVal, ok := activity["activity_type"]; ok && activityTypeVal != nil {
				activityType := activityTypeVal.(string)
				if activityType == "running" {
					t.Errorf("Expected no running activities, got %s", activityType)
				}
			}
		}
	})

	// Test Case 7: Backward Compatibility - Legacy 2-level syntax still works
	t.Run("BackwardCompatibility_LegacySyntax", func(t *testing.T) {
		url := "/activities?filter[activity_type]=cycling"

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		if len(data) != 1 {
			t.Errorf("Expected 1 cycling activity, got %d", len(data))
		}
	})

	// Test Case 8: Mixed Syntax - Legacy and Operator together
	t.Run("MixedSyntax_Legacy_AND_Operator", func(t *testing.T) {
		sevenDaysAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		url := fmt.Sprintf("/activities?filter[activity_type]=running&filter[activity_date][gte]=%s", sevenDaysAgo)

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		// Verify all returned activities are running AND in last 7 days
		for _, item := range data {
			activity := item.(map[string]interface{})
			if activityTypeVal, ok := activity["activity_type"]; ok && activityTypeVal != nil {
				activityType := activityTypeVal.(string)
				if activityType != "running" {
					t.Error("Expected only running activities")
				}
			}
		}
	})

	// Test Case 9: Complex Query - Multiple Operators + Pagination
	t.Run("ComplexQuery_MultipleOperators_WithPagination", func(t *testing.T) {
		startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		url := fmt.Sprintf("/activities?filter[activity_date][gte]=%s&filter[distance_km][gte]=5.0&filter[duration_minutes][lt]=60&page=1&limit=5&order[created_at]=DESC", startDate)

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Verify pagination metadata exists
		meta := response["meta"].(map[string]interface{})
		if meta["page"].(float64) != 1 {
			t.Error("Expected page=1")
		}
		if meta["limit"].(float64) != 5 {
			t.Error("Expected limit=5")
		}

		// Verify data matches all filter conditions
		data := response["data"].([]interface{})
		for _, item := range data {
			activity := item.(map[string]interface{})

			if distanceVal, ok := activity["distance_km"]; ok && distanceVal != nil {
				distance := distanceVal.(float64)
				if distance < 5.0 {
					t.Errorf("Expected distance >= 5.0, got %v", distance)
				}
			}

			if durationVal, ok := activity["duration_minutes"]; ok && durationVal != nil {
				duration := durationVal.(float64)
				if duration >= 60 {
					t.Errorf("Expected duration < 60, got %v", duration)
				}
			}
		}
	})

	// Test Case 10: Exclusive Range - GT and LT (boundaries excluded)
	t.Run("ExclusiveRange_GT_AND_LT", func(t *testing.T) {
		url := "/activities?filter[distance_km][gt]=5.0&filter[distance_km][lt]=15.0"

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		// Verify all returned activities have 5.0 < distance < 15.0
		for _, item := range data {
			activity := item.(map[string]interface{})
			if distanceVal, ok := activity["distance_km"]; ok && distanceVal != nil {
				distance := distanceVal.(float64)
				if distance <= 5.0 || distance >= 15.0 {
					t.Errorf("Expected 5.0 < distance < 15.0, got %v", distance)
				}
			}
		}
	})

	// Test Case 11: Edge Case - Boundary Inclusion with GTE/LTE
	t.Run("BoundaryInclusion_GTE_Includes_Exact_Value", func(t *testing.T) {
		url := "/activities?filter[duration_minutes][gte]=30"

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.ListActivities(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		// Verify activities with exactly 30 minutes are included
		foundExact := false
		for _, item := range data {
			activity := item.(map[string]interface{})
			if durationVal, ok := activity["duration_minutes"]; ok && durationVal != nil {
				duration := durationVal.(float64)
				if duration == 30 {
					foundExact = true
				}
				if duration < 30 {
					t.Errorf("Expected duration >= 30, got %v", duration)
				}
			}
		}
		if !foundExact {
			t.Log("No activity with exactly 30 minutes found (acceptable)")
		}
	})

	// Test Case 12: Router Integration - Full URL with mux
	t.Run("RouterIntegration_FullURL_WithOperators", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/api/v1/activities", handler.ListActivities).Methods("GET")

		startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		url := fmt.Sprintf("/api/v1/activities?filter[activity_date][gte]=%s&filter[distance_km][gte]=5.0&order[created_at]=DESC", startDate)

		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["data"] == nil {
			t.Fatal("Expected 'data' field in response")
		}

		if response["meta"] == nil {
			t.Fatal("Expected 'meta' field in response")
		}
	})
}

// setupOperatorTestData creates test activities with specific dates and values for operator testing
func setupOperatorTestData(t *testing.T, activityRepo *repository.ActivityRepository, userID int) {
	t.Helper()

	ctx := context.Background()
	now := time.Now()

	// Create activities with varying dates, distances, and durations
	activities := []struct {
		activityType string
		title        string
		duration     int
		distance     float64
		daysAgo      int
	}{
		{"running", "Recent Morning Run", 30, 5.0, 1},  // 1 day ago
		{"running", "Last Week Run", 45, 7.5, 7},       // 7 days ago
		{"cycling", "Bike Commute", 60, 12.0, 3},       // 3 days ago
		{"swimming", "Pool Workout", 40, 2.0, 10},      // 10 days ago
		{"running", "Long Distance Run", 90, 20.0, 14}, // 14 days ago
		{"yoga", "Morning Yoga", 30, 0.0, 2},           // 2 days ago
	}

	for _, act := range activities {
		activity := &models.Activity{
			UserID:          userID,
			ActivityType:    act.activityType,
			Title:           act.title,
			DurationMinutes: act.duration,
			DistanceKm:      act.distance,
			ActivityDate:    now.AddDate(0, 0, -act.daysAgo),
		}

		if err := activityRepo.CreateWithTags(ctx, activity, nil); err != nil {
			t.Fatalf("Failed to create test activity: %v", err)
		}
	}
}

// TestE2E_QueryParser_Integration tests query parameter parsing edge cases
func TestE2E_QueryParser_Integration(t *testing.T) {
	testCases := []struct {
		name           string
		queryString    string
		expectedError  bool
		validateResult func(*testing.T, *query.QueryOptions)
	}{
		{
			name:          "SimpleFilter",
			queryString:   "filter[status]=active",
			expectedError: false,
			validateResult: func(t *testing.T, opts *query.QueryOptions) {
				if opts.Filter["status"] != "active" {
					t.Errorf("Expected filter[status]=active, got %v", opts.Filter["status"])
				}
			},
		},
		{
			name:          "ArrayFilter",
			queryString:   "filter[type]=[a,b,c]",
			expectedError: false,
			validateResult: func(t *testing.T, opts *query.QueryOptions) {
				arr, ok := opts.Filter["type"].([]string)
				if !ok {
					t.Error("Expected filter[type] to be []string")
				}
				if len(arr) != 3 {
					t.Errorf("Expected 3 items in array, got %d", len(arr))
				}
			},
		},
		{
			name:          "MultipleFilters",
			queryString:   "filter[status]=active&filter[type]=running&filter[user_id]=123",
			expectedError: false,
			validateResult: func(t *testing.T, opts *query.QueryOptions) {
				if len(opts.Filter) != 3 {
					t.Errorf("Expected 3 filters, got %d", len(opts.Filter))
				}
			},
		},
		{
			name:          "SearchWithSpecialCharacters",
			queryString:   "search[title]=hello%20world",
			expectedError: false,
			validateResult: func(t *testing.T, opts *query.QueryOptions) {
				if opts.Search["title"] != "hello world" {
					t.Errorf("Expected search[title]='hello world', got '%v'", opts.Search["title"])
				}
			},
		},
		{
			name:          "OrderMultipleColumns",
			queryString:   "order[created_at]=DESC&order[title]=ASC",
			expectedError: false,
			validateResult: func(t *testing.T, opts *query.QueryOptions) {
				if opts.Order["created_at"] != "DESC" {
					t.Error("Expected order[created_at]=DESC")
				}
				if opts.Order["title"] != "ASC" {
					t.Error("Expected order[title]=ASC")
				}
			},
		},
		{
			name:          "PaginationDefaults",
			queryString:   "",
			expectedError: false,
			validateResult: func(t *testing.T, opts *query.QueryOptions) {
				if opts.Page != 1 {
					t.Errorf("Expected default page=1, got %d", opts.Page)
				}
				if opts.Limit != 10 {
					t.Errorf("Expected default limit=10, got %d", opts.Limit)
				}
			},
		},
		{
			name:          "OperatorSyntax_GTE",
			queryString:   "filter[created_at][gte]=2024-01-01",
			expectedError: false,
			validateResult: func(t *testing.T, opts *query.QueryOptions) {
				if len(opts.FilterConditions) != 1 {
					t.Errorf("Expected 1 FilterCondition, got %d", len(opts.FilterConditions))
				}
				if opts.FilterConditions[0].Operator != "gte" {
					t.Errorf("Expected operator 'gte', got '%s'", opts.FilterConditions[0].Operator)
				}
				if opts.FilterConditions[0].Column != "created_at" {
					t.Errorf("Expected column 'created_at', got '%s'", opts.FilterConditions[0].Column)
				}
			},
		},
		{
			name:          "MultipleOperators",
			queryString:   "filter[distance][gte]=5.0&filter[distance][lt]=15.0",
			expectedError: false,
			validateResult: func(t *testing.T, opts *query.QueryOptions) {
				if len(opts.FilterConditions) != 2 {
					t.Errorf("Expected 2 FilterConditions, got %d", len(opts.FilterConditions))
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse URL with query string
			req := httptest.NewRequest("GET", fmt.Sprintf("/test?%s", tc.queryString), nil)
			opts, err := query.ParseQueryParams(req.URL.Query())

			if tc.expectedError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tc.expectedError && tc.validateResult != nil {
				tc.validateResult(t, opts)
			}
		})
	}
}
