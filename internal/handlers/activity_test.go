package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	appErrors "github.com/valentinesamuel/activelog/pkg/errors"
)

// MockActivityRepository is a mock implementation using testify/mock
type MockActivityRepository struct {
	mock.Mock
}

func (m *MockActivityRepository) Create(ctx context.Context, tx repository.TxConn, activity *models.Activity) error {
	args := m.Called(ctx, tx, activity)
	return args.Error(0)
}

func (m *MockActivityRepository) GetByID(ctx context.Context, id int64) (*models.Activity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Activity), args.Error(1)
}

func (m *MockActivityRepository) GetActivitiesWithTags(ctx context.Context, userID int, filters models.ActivityFilters) ([]*models.Activity, error) {
	args := m.Called(ctx, userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Activity), args.Error(1)
}

func (m *MockActivityRepository) Count(userID int) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockActivityRepository) Update(ctx context.Context, tx repository.TxConn, id int, activity *models.Activity) error {
	args := m.Called(ctx, tx, id, activity)
	return args.Error(0)
}

func (m *MockActivityRepository) Delete(ctx context.Context, tx repository.TxConn, id int, userID int) error {
	args := m.Called(ctx, tx, id, userID)
	return args.Error(0)
}

func (m *MockActivityRepository) GetStats(userID int, startDate, endDate *time.Time) (*repository.ActivityStats, error) {
	args := m.Called(userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ActivityStats), args.Error(1)
}

// Test helper functions for cleaner test setup
func newTestActivity(id int64, activityType, title string) *models.Activity {
	return &models.Activity{
		BaseEntity: models.BaseEntity{
			ID:        id,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:          1,
		ActivityType:    activityType,
		Title:           title,
		Description:     "Test description",
		DurationMinutes: 30,
		DistanceKm:      5.0,
		ActivityDate:    time.Now(),
	}
}

func newTestRequest(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		if str, ok := body.(string); ok {
			reqBody = []byte(str)
		} else {
			reqBody, _ = json.Marshal(body)
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Add user context (simulating auth middleware)
	ctx := context.WithValue(req.Context(), "user_id", 1)
	return req.WithContext(ctx)
}

func TestActivityHandler_CreateActivity(t *testing.T) {
	tests := []struct {
		name          string
		requestBody   interface{}
		setupMock     func(*MockActivityRepository)
		expectedCode  int
		checkResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "valid request",
			requestBody: models.CreateActivityRequest{
				ActivityType:    "running",
				Title:           "Morning Run",
				Description:     "A refreshing morning run",
				DurationMinutes: 30,
				DistanceKm:      5.0,
				ActivityDate:    time.Now(),
			},
			setupMock: func(m *MockActivityRepository) {
				m.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*models.Activity")).
					Run(func(args mock.Arguments) {
						activity := args.Get(2).(*models.Activity)
						activity.ID = 1
						activity.CreatedAt = time.Now()
						activity.UpdatedAt = time.Now()
					}).
					Return(nil)
			},
			expectedCode: http.StatusCreated,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response models.Activity
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response.ID)
				assert.Equal(t, "running", response.ActivityType)
			},
		},
		{
			name:        "invalid JSON",
			requestBody: `{"invalid": json}`,
			setupMock:   func(m *MockActivityRepository) {},
			expectedCode: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "error")
			},
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"title": "Test Activity",
			},
			setupMock:    func(m *MockActivityRepository) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "repository error",
			requestBody: models.CreateActivityRequest{
				ActivityType:    "running",
				Title:           "Morning Run",
				Description:     "A refreshing morning run",
				DurationMinutes: 30,
				DistanceKm:      5.0,
				ActivityDate:    time.Now(),
			},
			setupMock: func(m *MockActivityRepository) {
				m.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*models.Activity")).
					Return(fmt.Errorf("database connection failed"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockActivityRepository)
			tt.setupMock(mockRepo)
			handler := NewActivityHandler(mockRepo)

			req := newTestRequest("POST", "/api/v1/activities", tt.requestBody)
			rr := httptest.NewRecorder()

			// Execute
			handler.CreateActivity(rr, req)

			// Assert
			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestActivityHandler_GetActivity(t *testing.T) {
	tests := []struct {
		name          string
		activityID    string
		setupMock     func(*MockActivityRepository)
		expectedCode  int
		checkResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "activity found",
			activityID: "1",
			setupMock: func(m *MockActivityRepository) {
				m.On("GetByID", mock.Anything, int64(1)).
					Return(newTestActivity(1, "running", "Morning Run"), nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response models.Activity
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response.ID)
				assert.Equal(t, "running", response.ActivityType)
			},
		},
		{
			name:       "activity not found",
			activityID: "999",
			setupMock: func(m *MockActivityRepository) {
				m.On("GetByID", mock.Anything, int64(999)).
					Return(nil, appErrors.ErrNotFound)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid activity ID",
			activityID:   "invalid",
			setupMock:    func(m *MockActivityRepository) {},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockActivityRepository)
			tt.setupMock(mockRepo)
			handler := NewActivityHandler(mockRepo)

			req := newTestRequest("GET", fmt.Sprintf("/api/v1/activities/%s", tt.activityID), nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.activityID})
			rr := httptest.NewRecorder()

			// Execute
			handler.GetActivity(rr, req)

			// Assert
			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestActivityHandler_ListActivities(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockActivityRepository)
		expectedCode  int
		checkResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "list activities success",
			setupMock: func(m *MockActivityRepository) {
				activities := []*models.Activity{
					newTestActivity(1, "running", "Morning Run"),
					newTestActivity(2, "cycling", "Evening Ride"),
				}
				m.On("GetActivitiesWithTags", mock.Anything, 1, mock.AnythingOfType("models.ActivityFilters")).
					Return(activities, nil)
				m.On("Count", 1).Return(2, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)

				activities := response["activities"].([]interface{})
				assert.Len(t, activities, 2)
				assert.Equal(t, float64(2), response["total"])
			},
		},
		{
			name: "empty list",
			setupMock: func(m *MockActivityRepository) {
				m.On("GetActivitiesWithTags", mock.Anything, 1, mock.AnythingOfType("models.ActivityFilters")).
					Return([]*models.Activity{}, nil)
				m.On("Count", 1).Return(0, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)

				activities := response["activities"].([]interface{})
				assert.Len(t, activities, 0)
			},
		},
		{
			name: "repository error",
			setupMock: func(m *MockActivityRepository) {
				m.On("GetActivitiesWithTags", mock.Anything, 1, mock.AnythingOfType("models.ActivityFilters")).
					Return(nil, fmt.Errorf("database error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockActivityRepository)
			tt.setupMock(mockRepo)
			handler := NewActivityHandler(mockRepo)

			req := newTestRequest("GET", "/api/v1/activities", nil)
			rr := httptest.NewRecorder()

			// Execute
			handler.ListActivities(rr, req)

			// Assert
			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestActivityHandler_DeleteActivity(t *testing.T) {
	tests := []struct {
		name         string
		activityID   string
		setupMock    func(*MockActivityRepository)
		expectedCode int
	}{
		{
			name:       "delete success",
			activityID: "1",
			setupMock: func(m *MockActivityRepository) {
				m.On("Delete", mock.Anything, mock.Anything, 1, 1).Return(nil)
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name:       "activity not found",
			activityID: "999",
			setupMock: func(m *MockActivityRepository) {
				m.On("Delete", mock.Anything, mock.Anything, 999, 1).Return(fmt.Errorf("activity not found"))
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid activity ID",
			activityID:   "invalid",
			setupMock:    func(m *MockActivityRepository) {},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockActivityRepository)
			tt.setupMock(mockRepo)
			handler := NewActivityHandler(mockRepo)

			req := newTestRequest("DELETE", fmt.Sprintf("/api/v1/activities/%s", tt.activityID), nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.activityID})
			rr := httptest.NewRecorder()

			// Execute
			handler.DeleteActivity(rr, req)

			// Assert
			assert.Equal(t, tt.expectedCode, rr.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
