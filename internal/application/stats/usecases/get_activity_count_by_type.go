package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// GetActivityCountByTypeUseCase handles fetching activity count breakdown by type
// This is a read-only operation and does NOT require a transaction
type GetActivityCountByTypeUseCase struct {
	repo repository.StatsRepositoryInterface
}

// NewGetActivityCountByTypeUseCase creates a new instance
func NewGetActivityCountByTypeUseCase(repo repository.StatsRepositoryInterface) *GetActivityCountByTypeUseCase {
	return &GetActivityCountByTypeUseCase{repo: repo}
}

// No RequiresTransaction() method = defaults to non-transactional
// Read operations don't need transaction overhead for performance

// Execute retrieves activity count by type
func (uc *GetActivityCountByTypeUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract user ID (required)
	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	// Fetch activity count by type
	countByType, err := uc.repo.GetActivityCountByType(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity count by type: %w", err)
	}

	// Calculate total count
	totalCount := 0
	for _, count := range countByType {
		totalCount += count
	}

	// Return result
	return map[string]interface{}{
		"count_by_type": countByType,
		"total_count":   totalCount,
	}, nil
}
