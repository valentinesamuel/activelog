package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// GetMonthlyStatsUseCase handles fetching monthly statistics (last 30 days)
// This is a read-only operation and does NOT require a transaction
type GetMonthlyStatsUseCase struct {
	repo repository.StatsRepositoryInterface
}

// NewGetMonthlyStatsUseCase creates a new instance
func NewGetMonthlyStatsUseCase(repo repository.StatsRepositoryInterface) *GetMonthlyStatsUseCase {
	return &GetMonthlyStatsUseCase{repo: repo}
}

// No RequiresTransaction() method = defaults to non-transactional
// Read operations don't need transaction overhead for performance

// Execute retrieves monthly statistics
func (uc *GetMonthlyStatsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract user ID (required)
	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	// Fetch monthly stats
	stats, err := uc.repo.GetMonthlyStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly stats: %w", err)
	}

	// Return result
	return map[string]interface{}{
		"monthly_stats": stats,
	}, nil
}
