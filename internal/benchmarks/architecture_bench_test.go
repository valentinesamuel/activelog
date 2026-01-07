package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
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
		UserID:     1,
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
	activity.ID = 123
	return nil
}

type mockTagRepo struct{}

func (m *mockTagRepo) GetByID(ctx context.Context, id int64) (*models.Tag, error) {
	return &models.Tag{
		BaseEntity: models.BaseEntity{ID: id},
	}, nil
}

func (m *mockTagRepo) GetByName(ctx context.Context, userID int, name string) (*models.Tag, error) {
	return &models.Tag{Name: name}, nil
}

func (m *mockTagRepo) Create(ctx context.Context, tag *models.Tag) error {
	tag.ID = 1
	return nil
}

func (m *mockTagRepo) CreateOrGet(ctx context.Context, userID int, tagName string) (*models.Tag, error) {
	return &models.Tag{
		BaseEntity: models.BaseEntity{ID: 1},
		Name:       tagName,
	}, nil
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

// =============================================================================
// 1. BASELINE BENCHMARKS
// =============================================================================

// BenchmarkBaseline_DirectRepositoryCall measures direct repository access
func BenchmarkBaseline_DirectRepositoryCall(b *testing.B) {
	repo := &mockActivityRepo{}
	ctx := context.Background()

	activity := &models.Activity{
		UserID:          1,
		ActivityType:    "running",
		Title:           "Morning Run",
		DurationMinutes: 30,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, nil, activity)
	}
}

// =============================================================================
// 2. SERVICE LAYER BENCHMARKS
// =============================================================================

// BenchmarkServiceLayer_WithValidation measures service layer overhead
func BenchmarkServiceLayer_WithValidation(b *testing.B) {
	activityRepo := &mockActivityRepo{}
	tagRepo := &mockTagRepo{}
	svc := service.NewActivityService(activityRepo, tagRepo)

	ctx := context.Background()
	req := &models.CreateActivityRequest{
		ActivityType:    "running",
		Title:           "Morning Run",
		DurationMinutes: 30,
		ActivityDate:    time.Now().Add(-1 * time.Hour),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = svc.CreateActivity(ctx, nil, 1, req)
	}
}

// BenchmarkServiceLayer_vs_DirectRepo compares service vs. direct access
func BenchmarkServiceLayer_vs_DirectRepo(b *testing.B) {
	b.Run("Direct Repository", func(b *testing.B) {
		repo := &mockActivityRepo{}
		ctx := context.Background()

		activity := &models.Activity{
			UserID:          1,
			ActivityType:    "running",
			Title:           "Morning Run",
			DurationMinutes: 30,
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = repo.Create(ctx, nil, activity)
		}
	})

	b.Run("With Service Layer", func(b *testing.B) {
		activityRepo := &mockActivityRepo{}
		tagRepo := &mockTagRepo{}
		svc := service.NewActivityService(activityRepo, tagRepo)

		ctx := context.Background()
		req := &models.CreateActivityRequest{
			ActivityType:    "running",
			Title:           "Morning Run",
			DurationMinutes: 30,
			ActivityDate:    time.Now().Add(-1 * time.Hour),
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = svc.CreateActivity(ctx, nil, 1, req)
		}
	})
}

// =============================================================================
// 3. USE CASE BENCHMARKS
// =============================================================================

// BenchmarkUseCase_WithService measures use case with service
func BenchmarkUseCase_WithService(b *testing.B) {
	activityRepo := &mockActivityRepo{}
	tagRepo := &mockTagRepo{}
	svc := service.NewActivityService(activityRepo, tagRepo)
	uc := usecases.NewCreateActivityUseCase(svc, activityRepo)

	ctx := context.Background()
	input := map[string]interface{}{
		"user_id": 1,
		"request": &models.CreateActivityRequest{
			ActivityType:    "running",
			Title:           "Morning Run",
			DurationMinutes: 30,
			ActivityDate:    time.Now().Add(-1 * time.Hour),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = uc.Execute(ctx, nil, input)
	}
}

// BenchmarkUseCase_WithRepo measures use case with both dependencies
// Note: Use case now receives both service and repo, decides which to use
func BenchmarkUseCase_WithRepo(b *testing.B) {
	activityRepo := &mockActivityRepo{}
	tagRepo := &mockTagRepo{}
	svc := service.NewActivityService(activityRepo, tagRepo)
	uc := usecases.NewCreateActivityUseCase(svc, activityRepo)

	ctx := context.Background()
	input := map[string]interface{}{
		"user_id": 1,
		"request": &models.CreateActivityRequest{
			ActivityType:    "running",
			Title:           "Morning Run",
			DurationMinutes: 30,
			ActivityDate:    time.Now(),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = uc.Execute(ctx, nil, input)
	}
}

// =============================================================================
// 4. BROKER PATTERN BENCHMARKS
// =============================================================================

// BenchmarkBroker_SingleUseCase measures broker with one use case
func BenchmarkBroker_SingleUseCase(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	activityRepo := &mockActivityRepo{}
	tagRepo := &mockTagRepo{}
	svc := service.NewActivityService(activityRepo, tagRepo)
	uc := usecases.NewCreateActivityUseCase(svc, activityRepo)

	brokerInstance := broker.NewBroker(db)
	ctx := context.Background()

	input := map[string]interface{}{
		"user_id": 1,
		"request": &models.CreateActivityRequest{
			ActivityType:    "running",
			Title:           "Morning Run",
			DurationMinutes: 30,
			ActivityDate:    time.Now().Add(-1 * time.Hour),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()

		_, _ = brokerInstance.RunUseCases(ctx, []broker.UseCase{uc}, input)
	}
}

// BenchmarkBroker_MultipleUseCases measures broker with chained use cases
func BenchmarkBroker_MultipleUseCases(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	activityRepo := &mockActivityRepo{}
	tagRepo := &mockTagRepo{}
	svc := service.NewActivityService(activityRepo, tagRepo)

	createUC := usecases.NewCreateActivityUseCase(svc, activityRepo)
	updateUC := usecases.NewUpdateActivityUseCase(svc, activityRepo)

	brokerInstance := broker.NewBroker(db)
	ctx := context.Background()

	input := map[string]interface{}{
		"user_id": 1,
		"request": &models.CreateActivityRequest{
			ActivityType:    "running",
			Title:           "Morning Run",
			DurationMinutes: 30,
			ActivityDate:    time.Now().Add(-1 * time.Hour),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()

		_, _ = brokerInstance.RunUseCases(ctx, []broker.UseCase{createUC, updateUC}, input)
	}
}

// BenchmarkBroker_TransactionBoundaryBreaking measures mixed transaction chains
func BenchmarkBroker_TransactionBoundaryBreaking(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	activityRepo := &mockActivityRepo{}
	tagRepo := &mockTagRepo{}
	svc := service.NewActivityService(activityRepo, tagRepo)

	createUC := usecases.NewCreateActivityUseCase(svc, activityRepo) // Requires TX
	getUC := usecases.NewGetActivityUseCase(svc, activityRepo)       // Non-TX
	updateUC := usecases.NewUpdateActivityUseCase(svc, activityRepo) // Requires TX

	brokerInstance := broker.NewBroker(db)
	ctx := context.Background()

	input := map[string]interface{}{
		"user_id": 1,
		"activity_id": int64(123),
		"request": &models.CreateActivityRequest{
			ActivityType:    "running",
			Title:           "Morning Run",
			DurationMinutes: 30,
			ActivityDate:    time.Now().Add(-1 * time.Hour),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// First TX
		mock.ExpectBegin()
		mock.ExpectCommit()
		// Non-TX (boundary break)
		// Second TX
		mock.ExpectBegin()
		mock.ExpectCommit()

		_, _ = brokerInstance.RunUseCases(ctx, []broker.UseCase{createUC, getUC, updateUC}, input)
	}
}

// =============================================================================
// 5. DI CONTAINER BENCHMARKS
// =============================================================================

// BenchmarkContainer_ResolveSingleton measures singleton resolution
func BenchmarkContainer_ResolveSingleton(b *testing.B) {
	c := container.New()
	c.RegisterSingleton("service", service.NewActivityService(&mockActivityRepo{}, &mockTagRepo{}))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = c.MustResolve("service")
	}
}

// BenchmarkContainer_ResolveTransient measures transient resolution
func BenchmarkContainer_ResolveTransient(b *testing.B) {
	c := container.New()
	c.Register("service", func(c *container.Container) (interface{}, error) {
		return service.NewActivityService(&mockActivityRepo{}, &mockTagRepo{}), nil
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = c.Resolve("service")
	}
}

// BenchmarkContainer_ConcurrentResolution measures concurrent access
func BenchmarkContainer_ConcurrentResolution(b *testing.B) {
	c := container.New()
	c.RegisterSingleton("service", service.NewActivityService(&mockActivityRepo{}, &mockTagRepo{}))

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = c.MustResolve("service")
		}
	})
}

// =============================================================================
// 6. MEMORY ALLOCATION BENCHMARKS
// =============================================================================

// BenchmarkMemory_MapAccess measures map[string]interface{} overhead
func BenchmarkMemory_MapAccess(b *testing.B) {
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

// BenchmarkMemory_StructAccess measures direct struct access
func BenchmarkMemory_StructAccess(b *testing.B) {
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

// =============================================================================
// 7. END-TO-END BENCHMARKS
// =============================================================================

// BenchmarkE2E_CreateActivity_FullStack measures complete request flow
func BenchmarkE2E_CreateActivity_FullStack(b *testing.B) {
	b.Run("Direct Repository (Baseline)", func(b *testing.B) {
		repo := &mockActivityRepo{}
		ctx := context.Background()

		req := &models.CreateActivityRequest{
			ActivityType:    "running",
			Title:           "Morning Run",
			DurationMinutes: 30,
			ActivityDate:    time.Now(),
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			activity := &models.Activity{
				UserID:          1,
				ActivityType:    req.ActivityType,
				Title:           req.Title,
				DurationMinutes: req.DurationMinutes,
				ActivityDate:    req.ActivityDate,
			}
			_ = repo.Create(ctx, nil, activity)
		}
	})

	b.Run("Broker + Service (Current Architecture)", func(b *testing.B) {
		db, mock, _ := sqlmock.New()
		defer db.Close()

		activityRepo := &mockActivityRepo{}
		tagRepo := &mockTagRepo{}
		svc := service.NewActivityService(activityRepo, tagRepo)
		uc := usecases.NewCreateActivityUseCase(svc, activityRepo)
		brokerInstance := broker.NewBroker(db)

		ctx := context.Background()
		input := map[string]interface{}{
			"user_id": 1,
			"request": &models.CreateActivityRequest{
				ActivityType:    "running",
				Title:           "Morning Run",
				DurationMinutes: 30,
				ActivityDate:    time.Now().Add(-1 * time.Hour),
			},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			mock.ExpectBegin()
			mock.ExpectCommit()

			_, _ = brokerInstance.RunUseCases(ctx, []broker.UseCase{uc}, input)
		}
	})
}
