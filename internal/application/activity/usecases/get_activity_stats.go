package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

// GetActivityStatsInput defines the typed input for GetActivityStatsUseCase
type GetActivityStatsInput struct {
	UserID    int
	StartDate *time.Time // Optional, defaults to 30 days ago
	EndDate   *time.Time // Optional, defaults to now
}

// GetActivityStatsOutput defines the typed output for GetActivityStatsUseCase
type GetActivityStatsOutput struct {
	Stats     *repository.ActivityStats
	StartDate time.Time
	EndDate   time.Time
}

// GetActivityStatsUseCase handles fetching activity statistics for a date range
// This is a read-only operation and does NOT require a transaction
// Has access to both service and repository - decides which to use
type GetActivityStatsUseCase struct {
	service service.StatsServiceInterface          // For operations requiring enrichment (activity level, insights)
	repo    repository.ActivityRepositoryInterface // For simple statistical queries
}

// NewGetActivityStatsUseCase creates a new instance with both service and repository
// The use case will decide which one to use based on what it needs
func NewGetActivityStatsUseCase(
	svc service.StatsServiceInterface,
	repo repository.ActivityRepositoryInterface,
) *GetActivityStatsUseCase {
	return &GetActivityStatsUseCase{
		service: svc,
		repo:    repo,
	}
}

// RequiresTransaction returns false - read operations don't need transactions
func (uc *GetActivityStatsUseCase) RequiresTransaction() bool {
	return false
}

// Execute retrieves activity statistics (typed version)
// Decision: Use repo for simple stats, service available for enrichment
func (uc *GetActivityStatsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input GetActivityStatsInput,
) (GetActivityStatsOutput, error) {
	// Calculate default date range
	startDate := input.StartDate
	endDate := input.EndDate

	if startDate == nil {
		defaultStart := time.Now().AddDate(0, 0, -30)
		startDate = &defaultStart
	}

	if endDate == nil {
		defaultEnd := time.Now()
		endDate = &defaultEnd
	}

	// DECISION: Use repo directly for simple stats retrieval - no business logic needed
	// Alternative: Could use service to enrich stats with activity level determination,
	// insights, or additional analytics (e.g., uc.service.EnrichStats(stats))
	stats, err := uc.repo.GetStats(input.UserID, startDate, endDate)
	if err != nil {
		return GetActivityStatsOutput{}, fmt.Errorf("failed to get activity stats: %w", err)
	}

	return GetActivityStatsOutput{
		Stats:     stats,
		StartDate: *startDate,
		EndDate:   *endDate,
	}, nil
}
