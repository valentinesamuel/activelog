package service

import (
	"context"
	"time"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// StatsService implements StatsServiceInterface
// Encapsulates business logic for statistics and analytics
type StatsService struct {
	statsRepo    repository.StatsRepositoryInterface
	activityRepo repository.ActivityRepositoryInterface
}

// NewStatsService creates a new stats service instance
func NewStatsService(
	statsRepo repository.StatsRepositoryInterface,
	activityRepo repository.ActivityRepositoryInterface,
) *StatsService {
	return &StatsService{
		statsRepo:    statsRepo,
		activityRepo: activityRepo,
	}
}

// CalculateActivityStats computes statistics for a date range
func (s *StatsService) CalculateActivityStats(
	ctx context.Context,
	userID int,
	startDate, endDate interface{},
) (*repository.ActivityStats, error) {
	// Convert interfaces to time.Time pointers
	var start, end *time.Time

	if sd, ok := startDate.(time.Time); ok {
		start = &sd
	} else if sdPtr, ok := startDate.(*time.Time); ok {
		start = sdPtr
	}

	if ed, ok := endDate.(time.Time); ok {
		end = &ed
	} else if edPtr, ok := endDate.(*time.Time); ok {
		end = edPtr
	}

	// Business logic: Default to last 30 days if not specified
	if start == nil {
		defaultStart := time.Now().AddDate(0, 0, -30)
		start = &defaultStart
	}
	if end == nil {
		defaultEnd := time.Now()
		end = &defaultEnd
	}

	// Business Rule: End date must be after start date
	if end.Before(*start) {
		// Swap if needed
		start, end = end, start
	}

	// Fetch stats from repository
	stats, err := s.activityRepo.GetStats(userID, start, end)
	if err != nil {
		return nil, err
	}

	// Business logic: Calculate additional derived metrics
	// (This is where you'd add computations that don't belong in the repository)
	// For now, just return the stats as-is
	return stats, nil
}

// GetUserSummary generates comprehensive user activity summary
func (s *StatsService) GetUserSummary(
	ctx context.Context,
	userID int,
) (*repository.UserActivitySummary, error) {
	// Fetch user summary from stats repository
	summary, err := s.statsRepo.GetUserActivitySummary(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Set TotalActivities from ActivityCount (for backward compatibility)
	summary.TotalActivities = summary.ActivityCount

	// Business logic: Enrich summary with calculated fields
	// Example: Determine activity level based on total activities
	if summary.TotalActivities == 0 {
		summary.ActivityLevel = "Inactive"
	} else if summary.TotalActivities < 10 {
		summary.ActivityLevel = "Beginner"
	} else if summary.TotalActivities < 50 {
		summary.ActivityLevel = "Intermediate"
	} else if summary.TotalActivities < 100 {
		summary.ActivityLevel = "Active"
	} else {
		summary.ActivityLevel = "Expert"
	}

	return summary, nil
}
