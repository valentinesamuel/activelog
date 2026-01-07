package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// GetTopTagsUseCase handles fetching top N most used tags
// This is a read-only operation and does NOT require a transaction
type GetTopTagsUseCase struct {
	repo repository.StatsRepositoryInterface
}

// NewGetTopTagsUseCase creates a new instance
func NewGetTopTagsUseCase(repo repository.StatsRepositoryInterface) *GetTopTagsUseCase {
	return &GetTopTagsUseCase{repo: repo}
}

// No RequiresTransaction() method = defaults to non-transactional
// Read operations don't need transaction overhead for performance

// Execute retrieves top N tags
func (uc *GetTopTagsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract user ID (required)
	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	// Extract limit (optional, defaults to 10)
	limit := 10
	if l, exists := input["limit"]; exists {
		if limitInt, ok := l.(int); ok {
			limit = limitInt
		}
	}

	// Enforce maximum limit
	if limit > 50 {
		limit = 50
	}
	if limit < 1 {
		limit = 1
	}

	// Fetch top tags
	tags, err := uc.repo.GetTopTagsByUser(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top tags: %w", err)
	}

	// Return result
	return map[string]interface{}{
		"tags":  tags,
		"limit": limit,
		"count": len(tags),
	}, nil
}
