package usecases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/pkg/query"
)

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

// No RequiresTransaction() method = defaults to non-transactional
// Read operations don't need transaction overhead for performance

// Execute retrieves tags with dynamic filtering using QueryOptions
//
// Input parameters:
//   - query_options (*query.QueryOptions) - REQUIRED: Contains filter, search, order, pagination options
//   - user_id (int) - OPTIONAL: If provided, can be used to filter tags by user (if tags are user-specific)
//
// Returns:
//   - result (*query.PaginatedResult) - Contains tags data and pagination metadata
//
// Example usage:
//   input := map[string]interface{}{
//       "query_options": &query.QueryOptions{
//           Page:   1,
//           Limit:  20,
//           Filter: map[string]interface{}{
//               "name": "cardio",
//           },
//           Search: map[string]interface{}{
//               "name": "run",
//           },
//           Order: map[string]string{
//               "name": "ASC",
//           },
//       },
//   }
//   result, err := listTagsUC.Execute(ctx, nil, input)
func (uc *ListTagsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx, // Will be nil for non-transactional use cases
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract QueryOptions (required)
	queryOpts, exists := input["query_options"]
	if !exists {
		return nil, fmt.Errorf("query_options is required")
	}

	opts, ok := queryOpts.(*query.QueryOptions)
	if !ok {
		return nil, fmt.Errorf("invalid query_options type")
	}

	// Optional: Add user_id filter if tags are user-specific
	// For now, tags are global, so we skip this
	// If you want user-specific tags, uncomment:
	// if userID, ok := input["user_id"].(int); ok {
	//     opts.Filter["user_id"] = userID
	// }

	// Use dynamic filtering method
	result, err := uc.repo.ListTagsWithQuery(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	// Return paginated result
	return map[string]interface{}{
		"result": result,
	}, nil
}
