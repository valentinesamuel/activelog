package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/service"
)

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

// No RequiresTransaction() method = defaults to non-transactional
// Read operations don't need transaction overhead for performance

// Execute retrieves activity statistics
// Decision: Use repo for simple stats, service available for enrichment
func (uc *GetActivityStatsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract user ID (required)
	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	// Extract start date (optional, defaults to 30 days ago)
	var startDate time.Time
	if sd, exists := input["start_date"]; exists {
		if t, ok := sd.(time.Time); ok {
			startDate = t
		} else if tPtr, ok := sd.(*time.Time); ok && tPtr != nil {
			startDate = *tPtr
		} else {
			return nil, fmt.Errorf("invalid start_date type")
		}
	} else {
		// Default: 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	// Extract end date (optional, defaults to now)
	var endDate time.Time
	if ed, exists := input["end_date"]; exists {
		if t, ok := ed.(time.Time); ok {
			endDate = t
		} else if tPtr, ok := ed.(*time.Time); ok && tPtr != nil {
			endDate = *tPtr
		} else {
			return nil, fmt.Errorf("invalid end_date type")
		}
	} else {
		// Default: now
		endDate = time.Now()
	}

	// DECISION: Use repo directly for simple stats retrieval - no business logic needed
	// Alternative: Could use service to enrich stats with activity level determination,
	// insights, or additional analytics (e.g., uc.service.EnrichStats(stats))
	stats, err := uc.repo.GetStats(userID, &startDate, &endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity stats: %w", err)
	}

	// Example of using both: Could use service for enrichment if needed
	// activityLevel := uc.service.DetermineActivityLevel(stats.TotalActivities)
	// insights, _ := uc.service.GenerateInsights(ctx, userID, stats)

	// Return result
	return map[string]interface{}{
		"stats":      stats,
		"start_date": startDate,
		"end_date":   endDate,
	}, nil
}
