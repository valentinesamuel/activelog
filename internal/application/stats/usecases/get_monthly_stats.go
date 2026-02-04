package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// GetMonthlyStatsInput defines the typed input for GetMonthlyStatsUseCase
type GetMonthlyStatsInput struct {
	UserID int
}

// GetMonthlyStatsOutput defines the typed output for GetMonthlyStatsUseCase
type GetMonthlyStatsOutput struct {
	MonthlyStats *repository.MonthlyStats
}

// GetMonthlyStatsUseCase handles fetching monthly statistics (last 30 days)
// This is a read-only operation and does NOT require a transaction
type GetMonthlyStatsUseCase struct {
	repo repository.StatsRepositoryInterface
}

// NewGetMonthlyStatsUseCase creates a new instance
func NewGetMonthlyStatsUseCase(repo repository.StatsRepositoryInterface) *GetMonthlyStatsUseCase {
	return &GetMonthlyStatsUseCase{repo: repo}
}

// RequiresTransaction returns false - read operations don't need transactions
func (uc *GetMonthlyStatsUseCase) RequiresTransaction() bool {
	return false
}

// Execute retrieves monthly statistics (typed version)
func (uc *GetMonthlyStatsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input GetMonthlyStatsInput,
) (GetMonthlyStatsOutput, error) {
	stats, err := uc.repo.GetMonthlyStats(ctx, input.UserID)
	if err != nil {
		return GetMonthlyStatsOutput{}, fmt.Errorf("failed to get monthly stats: %w", err)
	}

	return GetMonthlyStatsOutput{MonthlyStats: stats}, nil
}
