package service

import (
	"context"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
)

// ActivityServiceInterface defines business logic for activity operations
// Services encapsulate domain logic and coordinate multiple repositories
type ActivityServiceInterface interface {
	// CreateActivity handles activity creation with business rules
	// - Validates activity data
	// - Processes tags
	// - Enforces business constraints
	CreateActivity(ctx context.Context, tx repository.TxConn, userID int, req *models.CreateActivityRequest) (*models.Activity, error)

	// UpdateActivity handles activity updates with business rules
	// - Validates ownership
	// - Enforces update constraints
	// - Handles partial updates
	UpdateActivity(ctx context.Context, tx repository.TxConn, userID int, activityID int, req *models.UpdateActivityRequest) (*models.Activity, error)

	// DeleteActivity handles activity deletion with business rules
	// - Validates ownership
	// - Handles cascade deletions
	DeleteActivity(ctx context.Context, tx repository.TxConn, userID int, activityID int) error
}

// StatsServiceInterface defines business logic for statistics operations
// Read-only operations that aggregate and compute analytics
type StatsServiceInterface interface {
	// CalculateActivityStats computes statistics for a date range
	// - Aggregates activity data
	// - Computes averages and totals
	// - Formats output
	CalculateActivityStats(ctx context.Context, userID int, startDate, endDate interface{}) (*repository.ActivityStats, error)

	// GetUserSummary generates comprehensive user activity summary
	// - Total activities
	// - Activity streaks
	// - Achievement metrics
	GetUserSummary(ctx context.Context, userID int) (*repository.UserActivitySummary, error)
}
