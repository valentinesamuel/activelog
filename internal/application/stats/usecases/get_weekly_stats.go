package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// GetWeeklyStatsInput defines the typed input for GetWeeklyStatsUseCase
type GetWeeklyStatsInput struct {
	UserID int
}

// GetWeeklyStatsOutput defines the typed output for GetWeeklyStatsUseCase
type GetWeeklyStatsOutput struct {
	WeeklyStats *repository.WeeklyStats
}

// GetWeeklyStatsUseCase handles fetching weekly statistics (last 7 days)
// This is a read-only operation and does NOT require a transaction
type GetWeeklyStatsUseCase struct {
	repo repository.StatsRepositoryInterface
}

// NewGetWeeklyStatsUseCase creates a new instance
func NewGetWeeklyStatsUseCase(repo repository.StatsRepositoryInterface) *GetWeeklyStatsUseCase {
	return &GetWeeklyStatsUseCase{repo: repo}
}

// RequiresTransaction returns false - read operations don't need transactions
func (uc *GetWeeklyStatsUseCase) RequiresTransaction() bool {
	return false
}

// Execute retrieves weekly statistics (typed version)
func (uc *GetWeeklyStatsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input GetWeeklyStatsInput,
) (GetWeeklyStatsOutput, error) {
	stats, err := uc.repo.GetWeeklyStats(ctx, input.UserID)
	if err != nil {
		return GetWeeklyStatsOutput{}, fmt.Errorf("failed to get weekly stats: %w", err)
	}

	return GetWeeklyStatsOutput{WeeklyStats: stats}, nil
}
