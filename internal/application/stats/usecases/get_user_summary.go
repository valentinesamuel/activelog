package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// GetUserSummaryUseCase handles fetching user activity summary
// This is a read-only operation and does NOT require a transaction
type GetUserSummaryUseCase struct {
	repo repository.StatsRepositoryInterface
}

// NewGetUserSummaryUseCase creates a new instance
func NewGetUserSummaryUseCase(repo repository.StatsRepositoryInterface) *GetUserSummaryUseCase {
	return &GetUserSummaryUseCase{repo: repo}
}

// No RequiresTransaction() method = defaults to non-transactional
// Read operations don't need transaction overhead for performance

// Execute retrieves user activity summary
func (uc *GetUserSummaryUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract user ID (required)
	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	// Fetch user summary
	summary, err := uc.repo.GetUserActivitySummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user summary: %w", err)
	}

	// Return result
	return map[string]interface{}{
		"summary": summary,
	}, nil
}
