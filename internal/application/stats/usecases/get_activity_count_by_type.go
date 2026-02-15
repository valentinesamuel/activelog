package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// GetActivityCountByTypeInput defines the typed input for GetActivityCountByTypeUseCase
type GetActivityCountByTypeInput struct {
	UserID int
}

// GetActivityCountByTypeOutput defines the typed output for GetActivityCountByTypeUseCase
type GetActivityCountByTypeOutput struct {
	CountByType map[string]int
	TotalCount  int
}

// GetActivityCountByTypeUseCase handles fetching activity count breakdown by type
// This is a read-only operation and does NOT require a transaction
type GetActivityCountByTypeUseCase struct {
	repo repository.StatsRepositoryInterface
}

// NewGetActivityCountByTypeUseCase creates a new instance
func NewGetActivityCountByTypeUseCase(repo repository.StatsRepositoryInterface) *GetActivityCountByTypeUseCase {
	return &GetActivityCountByTypeUseCase{repo: repo}
}

// RequiresTransaction returns false - read operations don't need transactions
func (uc *GetActivityCountByTypeUseCase) RequiresTransaction() bool {
	return false
}

// Execute retrieves activity count by type (typed version)
func (uc *GetActivityCountByTypeUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input GetActivityCountByTypeInput,
) (GetActivityCountByTypeOutput, error) {
	countByType, err := uc.repo.GetActivityCountByType(ctx, input.UserID)
	if err != nil {
		return GetActivityCountByTypeOutput{}, fmt.Errorf("failed to get activity count by type: %w", err)
	}

	// Calculate total count
	totalCount := 0
	for _, count := range countByType {
		totalCount += count
	}

	return GetActivityCountByTypeOutput{
		CountByType: countByType,
		TotalCount:  totalCount,
	}, nil
}
