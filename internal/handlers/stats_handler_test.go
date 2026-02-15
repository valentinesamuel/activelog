package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/valentinesamuel/activelog/internal/handlers"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/repository/mocks"
)

func TestStatsHandler_GetWeeklyStats(t *testing.T) {
	tests := []struct {
		name           string
		userID         interface{}
		setupMock      func(*mocks.MockStatsRepositoryInterface)
		expectedStatus int
		expectedBody   map[string]interface{}
		checkError     bool
	}{
		{
			name:   "success - returns weekly stats for authenticated user",
			userID: 1,
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				m.EXPECT().
					GetWeeklyStats(gomock.Any(), 1).
					Return(&repository.WeeklyStats{
						TotalActivities: 12,
						TotalDuration:   360,
						TotalDistance:   45.5,
						AvgDuration:     30.0,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"totalActivities":      float64(12),
				"totalDurationMinutes": float64(360),
				"totalDistanceKm":      45.5,
				"avgDurationMinutes":   30.0,
			},
		},
		{
			name:           "error - user not authenticated",
			userID:         nil, // No user_id in context
			setupMock:      func(m *mocks.MockStatsRepositoryInterface) {},
			expectedStatus: http.StatusUnauthorized,
			checkError:     true,
		},
		{
			name:   "error - repository returns error",
			userID: 1,
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				m.EXPECT().
					GetWeeklyStats(gomock.Any(), 1).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkError:     true,
		},
		{
			name:   "success - no activities returns zeros",
			userID: 2,
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				m.EXPECT().
					GetWeeklyStats(gomock.Any(), 2).
					Return(&repository.WeeklyStats{
						TotalActivities: 0,
						TotalDuration:   0,
						TotalDistance:   0,
						AvgDuration:     0,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"totalActivities":      float64(0),
				"totalDurationMinutes": float64(0),
				"totalDistanceKm":      float64(0),
				"avgDurationMinutes":   float64(0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockStatsRepositoryInterface(ctrl)
			tt.setupMock(mockRepo)

			handler := handlers.NewStatsHandler(mockRepo)

			// Create request with context
			req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/weekly", nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			// Execute
			handler.GetWeeklyStats(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.checkError && tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestStatsHandler_GetMonthlyStats(t *testing.T) {
	tests := []struct {
		name           string
		userID         interface{}
		setupMock      func(*mocks.MockStatsRepositoryInterface)
		expectedStatus int
		expectedBody   map[string]int
		checkError     bool
	}{
		{
			name:   "success - returns monthly stats by activity type",
			userID: 1,
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				stats := repository.MonthlyStats{
					"running":    15,
					"basketball": 8,
					"gym":        12,
				}
				m.EXPECT().
					GetMonthlyStats(gomock.Any(), 1).
					Return(&stats, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]int{
				"running":    15,
				"basketball": 8,
				"gym":        12,
			},
		},
		{
			name:           "error - user not authenticated",
			userID:         nil,
			setupMock:      func(m *mocks.MockStatsRepositoryInterface) {},
			expectedStatus: http.StatusUnauthorized,
			checkError:     true,
		},
		{
			name:   "error - database connection failed",
			userID: 1,
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				m.EXPECT().
					GetMonthlyStats(gomock.Any(), 1).
					Return(nil, errors.New("connection refused"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockStatsRepositoryInterface(ctrl)
			tt.setupMock(mockRepo)

			handler := handlers.NewStatsHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/monthly", nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handler.GetMonthlyStats(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.checkError && tt.expectedBody != nil {
				var response map[string]int
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestStatsHandler_GetTopTags(t *testing.T) {
	tests := []struct {
		name           string
		userID         interface{}
		queryParams    string
		setupMock      func(*mocks.MockStatsRepositoryInterface)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "success - returns top 10 tags by default",
			userID:      1,
			queryParams: "",
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				tags := []repository.TagUsage{
					{TagName: "outdoor", Count: 45},
					{TagName: "morning", Count: 32},
					{TagName: "cardio", Count: 28},
				}
				m.EXPECT().
					GetTopTagsByUser(gomock.Any(), 1, 10).
					Return(tags, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(3), response["total_unique_tags"])

				tags := response["tags"].([]interface{})
				assert.Len(t, tags, 3)

				firstTag := tags[0].(map[string]interface{})
				assert.Equal(t, "outdoor", firstTag["tagName"])
				assert.Equal(t, float64(45), firstTag["count"])
			},
		},
		{
			name:        "success - respects custom limit parameter",
			userID:      1,
			queryParams: "?limit=5",
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				tags := []repository.TagUsage{
					{TagName: "outdoor", Count: 45},
					{TagName: "morning", Count: 32},
				}
				m.EXPECT().
					GetTopTagsByUser(gomock.Any(), 1, 5).
					Return(tags, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(2), response["total_unique_tags"])
			},
		},
		{
			name:        "success - empty tags list",
			userID:      1,
			queryParams: "",
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				m.EXPECT().
					GetTopTagsByUser(gomock.Any(), 1, 10).
					Return([]repository.TagUsage{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(0), response["total_unique_tags"])
			},
		},
		{
			name:           "error - user not authenticated",
			userID:         nil,
			queryParams:    "",
			setupMock:      func(m *mocks.MockStatsRepositoryInterface) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "error - repository error",
			userID:      1,
			queryParams: "",
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				m.EXPECT().
					GetTopTagsByUser(gomock.Any(), 1, 10).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockStatsRepositoryInterface(ctrl)
			tt.setupMock(mockRepo)

			handler := handlers.NewStatsHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/tags/top"+tt.queryParams, nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handler.GetTopTags(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestStatsHandler_GetActivityCountByType(t *testing.T) {
	tests := []struct {
		name           string
		userID         interface{}
		setupMock      func(*mocks.MockStatsRepositoryInterface)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "success - returns activity breakdown",
			userID: 1,
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				breakdown := map[string]int{
					"running":    25,
					"cycling":    15,
					"swimming":   8,
					"basketball": 12,
				}
				m.EXPECT().
					GetActivityCountByType(gomock.Any(), 1).
					Return(breakdown, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(60), response["total_activities"])

				breakdown := response["activity_breakdown"].(map[string]interface{})
				assert.Equal(t, float64(25), breakdown["running"])
				assert.Equal(t, float64(15), breakdown["cycling"])
			},
		},
		{
			name:           "error - unauthenticated",
			userID:         nil,
			setupMock:      func(m *mocks.MockStatsRepositoryInterface) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "error - repository fails",
			userID: 1,
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				m.EXPECT().
					GetActivityCountByType(gomock.Any(), 1).
					Return(nil, errors.New("query timeout"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockStatsRepositoryInterface(ctrl)
			tt.setupMock(mockRepo)

			handler := handlers.NewStatsHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/by-type", nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handler.GetActivityCountByType(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestStatsHandler_GetUserActivitySummary(t *testing.T) {
	tests := []struct {
		name           string
		userID         interface{}
		setupMock      func(*mocks.MockStatsRepositoryInterface)
		expectedStatus int
		expectedBody   *repository.UserActivitySummary
		checkError     bool
	}{
		{
			name:   "success - returns user summary",
			userID: 1,
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				m.EXPECT().
					GetUserActivitySummary(gomock.Any(), 1).
					Return(&repository.UserActivitySummary{
						Username:       "john_doe",
						ActivityCount:  150,
						UniqueTagCount: 12,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &repository.UserActivitySummary{
				Username:       "john_doe",
				ActivityCount:  150,
				UniqueTagCount: 12,
			},
		},
		{
			name:           "error - not authenticated",
			userID:         nil,
			setupMock:      func(m *mocks.MockStatsRepositoryInterface) {},
			expectedStatus: http.StatusUnauthorized,
			checkError:     true,
		},
		{
			name:   "error - user not found",
			userID: 999,
			setupMock: func(m *mocks.MockStatsRepositoryInterface) {
				m.EXPECT().
					GetUserActivitySummary(gomock.Any(), 999).
					Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockStatsRepositoryInterface(ctrl)
			tt.setupMock(mockRepo)

			handler := handlers.NewStatsHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/summary", nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handler.GetUserActivitySummary(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.checkError && tt.expectedBody != nil {
				var response repository.UserActivitySummary
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.Username, response.Username)
				assert.Equal(t, tt.expectedBody.ActivityCount, response.ActivityCount)
				assert.Equal(t, tt.expectedBody.UniqueTagCount, response.UniqueTagCount)
			}
		})
	}
}
