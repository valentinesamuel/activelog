package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// Mock repository for benchmarking
type mockActivityRepo struct{}

func (m *mockActivityRepo) Create(ctx context.Context, tx repository.TxConn, activity *models.Activity) error {
	activity.ID = 123
	return nil
}

func (m *mockActivityRepo) GetByID(ctx context.Context, id int64) (*models.Activity, error) {
	return &models.Activity{
		BaseEntity: models.BaseEntity{ID: id},
	}, nil
}

func (m *mockActivityRepo) Update(ctx context.Context, tx repository.TxConn, id int, activity *models.Activity) error {
	return nil
}

func (m *mockActivityRepo) Delete(ctx context.Context, tx repository.TxConn, id int, userID int) error {
	return nil
}

func (m *mockActivityRepo) GetActivitiesWithTags(ctx context.Context, userID int, filters models.ActivityFilters) ([]*models.Activity, error) {
	return nil, nil
}

func (m *mockActivityRepo) Count(userID int) (int, error) {
	return 0, nil
}

func (m *mockActivityRepo) GetStats(userID int, startDate, endDate *time.Time) (*repository.ActivityStats, error) {
	return &repository.ActivityStats{}, nil
}

func (m *mockActivityRepo) ListByUser(ctx context.Context, UserID int) ([]*models.Activity, error) {
	return nil, nil
}

func (m *mockActivityRepo) ListByUserWithFilters(UserID int, filters models.ActivityFilters) ([]*models.Activity, error) {
	return nil, nil
}

func (m *mockActivityRepo) CreateWithTags(ctx context.Context, activity *models.Activity, tags []*models.Tag) error {
	return nil
}

func (m *mockActivityRepo) ListActivitiesWithQuery(ctx context.Context, opts *query.QueryOptions) (*query.PaginatedResult, error) {
	return &query.PaginatedResult{
		Data: []*models.Activity{},
		Meta: query.PaginationMeta{},
	}, nil
}

// Mock tag repository for benchmarking
type mockTagRepo struct{}

func (m *mockTagRepo) GetByID(ctx context.Context, id int64) (*models.Tag, error) {
	return &models.Tag{BaseEntity: models.BaseEntity{ID: id}}, nil
}

func (m *mockTagRepo) GetByName(ctx context.Context, userID int, name string) (*models.Tag, error) {
	return &models.Tag{Name: name}, nil
}

func (m *mockTagRepo) Create(ctx context.Context, tag *models.Tag) error {
	tag.ID = 1
	return nil
}

func (m *mockTagRepo) CreateOrGet(ctx context.Context, userID int, tagName string) (*models.Tag, error) {
	return &models.Tag{BaseEntity: models.BaseEntity{ID: 1}, Name: tagName}, nil
}

func (m *mockTagRepo) GetOrCreateTag(ctx context.Context, tx repository.TxConn, name string) (int, error) {
	return 1, nil
}

func (m *mockTagRepo) GetTagsForActivity(ctx context.Context, activityID int) ([]*models.Tag, error) {
	return nil, nil
}

func (m *mockTagRepo) LinkActivityTag(ctx context.Context, tx repository.TxConn, activityID int, tagID int) error {
	return nil
}

func (m *mockTagRepo) ListByUser(ctx context.Context, userID int) ([]*models.Tag, error) {
	return nil, nil
}

func (m *mockTagRepo) ListTagsWithQuery(ctx context.Context, opts *query.QueryOptions) (*query.PaginatedResult, error) {
	return &query.PaginatedResult{
		Data: []*models.Tag{},
		Meta: query.PaginationMeta{},
	}, nil
}

// Benchmark: Direct repository call (baseline)
func BenchmarkDirectRepositoryCall_CreateActivity(b *testing.B) {
	repo := &mockActivityRepo{}

	ctx := context.Background()
	req := &models.CreateActivityRequest{
		ActivityType:    "running",
		Title:           "Morning Run",
		DurationMinutes: 30,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		activity := &models.Activity{
			UserID:          1,
			ActivityType:    req.ActivityType,
			Title:           req.Title,
			DurationMinutes: req.DurationMinutes,
		}
		_ = repo.Create(ctx, nil, activity)
	}
}

// Benchmark: Broker approach (with use case)
func BenchmarkBrokerHandler_CreateActivity(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	activityRepo := &mockActivityRepo{}
	tagRepo := &mockTagRepo{}
	svc := service.NewActivityService(activityRepo, tagRepo)
	brokerInstance := broker.NewBroker(db)
	createUC := usecases.NewCreateActivityUseCase(svc, activityRepo)

	ctx := context.Background()
	req := &models.CreateActivityRequest{
		ActivityType:    "running",
		Title:           "Morning Run",
		DurationMinutes: 30,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()

		_, _ = brokerInstance.RunUseCases(
			ctx,
			[]broker.UseCase{createUC},
			map[string]interface{}{
				"user_id": 1,
				"request": req,
			},
		)
	}
}

// Benchmark: Just the use case (without broker)
func BenchmarkUseCase_CreateActivity(b *testing.B) {
	activityRepo := &mockActivityRepo{}
	tagRepo := &mockTagRepo{}
	svc := service.NewActivityService(activityRepo, tagRepo)
	createUC := usecases.NewCreateActivityUseCase(svc, activityRepo)

	ctx := context.Background()
	req := &models.CreateActivityRequest{
		ActivityType:    "running",
		Title:           "Morning Run",
		DurationMinutes: 30,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = createUC.Execute(
			ctx,
			nil,
			map[string]interface{}{
				"user_id": 1,
				"request": req,
			},
		)
	}
}

// Benchmark: Map access overhead
func BenchmarkMapAccess(b *testing.B) {
	input := map[string]interface{}{
		"user_id": 1,
		"request": &models.CreateActivityRequest{Title: "test"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		userID := input["user_id"].(int)
		req := input["request"].(*models.CreateActivityRequest)
		_ = userID
		_ = req
	}
}

// Benchmark: Direct struct access
func BenchmarkStructAccess(b *testing.B) {
	type Input struct {
		UserID  int
		Request *models.CreateActivityRequest
	}

	input := Input{
		UserID:  1,
		Request: &models.CreateActivityRequest{Title: "test"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		userID := input.UserID
		req := input.Request
		_ = userID
		_ = req
	}
}
