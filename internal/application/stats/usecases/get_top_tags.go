package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
)

// GetTopTagsInput defines the typed input for GetTopTagsUseCase
type GetTopTagsInput struct {
	UserID int
	Limit  int // Optional, defaults to 10, max 50
}

// GetTopTagsOutput defines the typed output for GetTopTagsUseCase
type GetTopTagsOutput struct {
	Tags  []repository.TagUsage
	Limit int
	Count int
}

// GetTopTagsUseCase handles fetching top N most used tags
// This is a read-only operation and does NOT require a transaction
type GetTopTagsUseCase struct {
	repo repository.StatsRepositoryInterface
}

// NewGetTopTagsUseCase creates a new instance
func NewGetTopTagsUseCase(repo repository.StatsRepositoryInterface) *GetTopTagsUseCase {
	return &GetTopTagsUseCase{repo: repo}
}

// RequiresTransaction returns false - read operations don't need transactions
func (uc *GetTopTagsUseCase) RequiresTransaction() bool {
	return false
}

// Execute retrieves top N tags (typed version)
func (uc *GetTopTagsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input GetTopTagsInput,
) (GetTopTagsOutput, error) {
	// Apply defaults and limits
	limit := input.Limit
	if limit == 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if limit < 1 {
		limit = 1
	}

	tags, err := uc.repo.GetTopTagsByUser(ctx, input.UserID, limit)
	if err != nil {
		return GetTopTagsOutput{}, fmt.Errorf("failed to get top tags: %w", err)
	}

	return GetTopTagsOutput{
		Tags:  tags,
		Limit: limit,
		Count: len(tags),
	}, nil
}
