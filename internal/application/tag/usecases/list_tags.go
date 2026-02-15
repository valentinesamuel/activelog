package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// ListTagsInput defines the typed input for ListTagsUseCase
type ListTagsInput struct {
	QueryOptions *query.QueryOptions
	UserID       int // Optional: for user-specific tags filtering
}

// ListTagsOutput defines the typed output for ListTagsUseCase
type ListTagsOutput struct {
	Result *query.PaginatedResult
}

// ListTagsUseCase handles fetching tags with dynamic filtering
// This is a read-only operation and does NOT require a transaction
type ListTagsUseCase struct {
	repo repository.TagRepositoryInterface
}

// NewListTagsUseCase creates a new instance
func NewListTagsUseCase(repo repository.TagRepositoryInterface) *ListTagsUseCase {
	return &ListTagsUseCase{
		repo: repo,
	}
}

// RequiresTransaction returns false - read operations don't need transactions
func (uc *ListTagsUseCase) RequiresTransaction() bool {
	return false
}

// Execute retrieves tags with dynamic filtering using QueryOptions (typed version)
func (uc *ListTagsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input ListTagsInput,
) (ListTagsOutput, error) {
	opts := input.QueryOptions
	if opts == nil {
		return ListTagsOutput{}, fmt.Errorf("query_options is required")
	}

	// Optional: Add user_id filter if tags are user-specific
	// For now, tags are global, so we skip this
	// If you want user-specific tags, uncomment:
	// if input.UserID != 0 {
	//     opts.Filter["user_id"] = input.UserID
	// }

	// Use dynamic filtering method
	result, err := uc.repo.ListTagsWithQuery(ctx, opts)
	if err != nil {
		return ListTagsOutput{}, fmt.Errorf("failed to list tags: %w", err)
	}

	return ListTagsOutput{Result: result}, nil
}
