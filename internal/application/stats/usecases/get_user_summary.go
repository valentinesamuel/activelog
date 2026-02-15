package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// GetUserSummaryInput defines the typed input for GetUserSummaryUseCase
type GetUserSummaryInput struct {
	UserID int
}

// GetUserSummaryOutput defines the typed output for GetUserSummaryUseCase
type GetUserSummaryOutput struct {
	Summary *repository.UserActivitySummary
}

// GetUserSummaryUseCase handles fetching user activity summary
// This is a read-only operation and does NOT require a transaction
type GetUserSummaryUseCase struct {
	repo repository.StatsRepositoryInterface
}

// NewGetUserSummaryUseCase creates a new instance
func NewGetUserSummaryUseCase(repo repository.StatsRepositoryInterface) *GetUserSummaryUseCase {
	return &GetUserSummaryUseCase{repo: repo}
}

// RequiresTransaction returns false - read operations don't need transactions
func (uc *GetUserSummaryUseCase) RequiresTransaction() bool {
	return false
}

// Execute retrieves user activity summary (typed version)
func (uc *GetUserSummaryUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input GetUserSummaryInput,
) (GetUserSummaryOutput, error) {
	summary, err := uc.repo.GetUserActivitySummary(ctx, input.UserID)
	if err != nil {
		return GetUserSummaryOutput{}, fmt.Errorf("failed to get user summary: %w", err)
	}

	return GetUserSummaryOutput{Summary: summary}, nil
}
