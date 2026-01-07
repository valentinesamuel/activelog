package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// GetWeeklyStatsUseCase handles fetching weekly statistics (last 7 days)
// This is a read-only operation and does NOT require a transaction
type GetWeeklyStatsUseCase struct {
	repo repository.StatsRepositoryInterface
}

// NewGetWeeklyStatsUseCase creates a new instance
func NewGetWeeklyStatsUseCase(repo repository.StatsRepositoryInterface) *GetWeeklyStatsUseCase {
	return &GetWeeklyStatsUseCase{repo: repo}
}

// No RequiresTransaction() method = defaults to non-transactional
// Read operations don't need transaction overhead for performance

// Execute retrieves weekly statistics
func (uc *GetWeeklyStatsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract user ID (required)
	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	// Fetch weekly stats
	stats, err := uc.repo.GetWeeklyStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly stats: %w", err)
	}

	// Return result
	return map[string]interface{}{
		"weekly_stats": stats,
	}, nil
}
