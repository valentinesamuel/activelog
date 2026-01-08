package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// DBConn is an interface that abstracts database operations
// This allows us to use either *sql.DB or *database.LoggingDB
type DBConn interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	GetRawDB() *sql.DB // For broker pattern
}

// TxConn is an interface for database transactions
// Both *sql.Tx and *database.LoggingTx implement this interface
type TxConn interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
}

//go:generate mockgen -destination=mocks/mock_stats_repository.go -package=mocks github.com/valentinesamuel/activelog/internal/repository StatsRepositoryInterface
type StatsRepositoryInterface interface {
	GetWeeklyStats(ctx context.Context, userID int) (*WeeklyStats, error)
	GetMonthlyStats(ctx context.Context, userID int) (*MonthlyStats, error)
	GetActivityCountByType(ctx context.Context, userID int) (map[string]int, error)
	GetUserActivitySummary(ctx context.Context, userID int) (*UserActivitySummary, error)
	GetTopTagsByUser(ctx context.Context, userID int, limit int) ([]TagUsage, error)
}

//go:generate mockgen -destination=mocks/mock_activity_repository.go -package=mocks github.com/valentinesamuel/activelog/internal/repository ActivityRepositoryInterface
type ActivityRepositoryInterface interface {
	Create(ctx context.Context, tx TxConn, activity *models.Activity) error
	GetByID(ctx context.Context, id int64) (*models.Activity, error)
	ListByUser(ctx context.Context, UserID int) ([]*models.Activity, error)
	ListByUserWithFilters(UserID int, filters models.ActivityFilters) ([]*models.Activity, error)
	Count(userID int) (int, error)
	Update(ctx context.Context, tx TxConn, id int, activity *models.Activity) error
	Delete(ctx context.Context, tx TxConn, id int, userID int) error
	GetStats(userID int, startDate, endDate *time.Time) (*ActivityStats, error)
	CreateWithTags(ctx context.Context, activity *models.Activity, tags []*models.Tag) error
	GetActivitiesWithTags(ctx context.Context, userID int, filters models.ActivityFilters) ([]*models.Activity, error)
	ListActivitiesWithQuery(ctx context.Context, opts *query.QueryOptions) (*query.PaginatedResult, error)
}

//go:generate mockgen -destination=mocks/mock_user_repository.go -package=mocks github.com/valentinesamuel/activelog/internal/repository UserRepositoryInterface
type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user *models.User) error
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
}

//go:generate mockgen -destination=mocks/mock_tag_repository.go -package=mocks github.com/valentinesamuel/activelog/internal/repository TagRepositoryInterface
type TagRepositoryInterface interface {
	GetOrCreateTag(ctx context.Context, tx TxConn, name string) (int, error)
	GetTagsForActivity(ctx context.Context, activityID int) ([]*models.Tag, error)
	LinkActivityTag(ctx context.Context, tx TxConn, activityID int, tagID int) error
	ListTagsWithQuery(ctx context.Context, opts *query.QueryOptions) (*query.PaginatedResult, error)
}
