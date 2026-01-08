package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/pkg/query"
)

// FindAndPaginate is a generic function for executing paginated queries on any entity.
//
// Type Parameters:
//   - T: The entity type to be returned (e.g., models.Activity, models.Tag)
//
// Parameters:
//   - ctx: Context for query cancellation and tracing
//   - db: Database connection (can be *sql.DB or transaction)
//   - tableName: The name of the table to query
//   - opts: QueryOptions containing filter, search, sort, and pagination parameters
//   - scanFunc: Function to scan a single row into type T
//   - joins: Optional JOIN configurations for relationship filtering
//
// Returns:
//   - *query.PaginatedResult: Contains the data and pagination metadata
//   - error: Any error that occurred during query execution
//
// Example Usage:
//
//	// For Activities
//	result, err := FindAndPaginate[models.Activity](
//	    ctx, db, "activities", opts,
//	    func(rows *sql.Rows) (*models.Activity, error) {
//	        var activity models.Activity
//	        err := rows.Scan(&activity.ID, &activity.UserID, ...)
//	        return &activity, err
//	    },
//	)
//
//	// For Tags
//	result, err := FindAndPaginate[models.Tag](
//	    ctx, db, "tags", opts,
//	    func(rows *sql.Rows) (*models.Tag, error) {
//	        var tag models.Tag
//	        err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt, &tag.UpdatedAt)
//	        return &tag, err
//	    },
//	)
func FindAndPaginate[T any](
	ctx context.Context,
	db DBConn,
	tableName string,
	opts *query.QueryOptions,
	scanFunc func(*sql.Rows) (*T, error),
	joins ...query.JoinConfig,
) (*query.PaginatedResult, error) {
	// Step 1: Build and execute COUNT query for pagination metadata
	totalRecords, err := executeCountQuery(ctx, db, tableName, opts, joins...)
	if err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Step 2: Calculate pagination metadata
	meta := calculatePaginationMeta(opts.Page, opts.Limit, totalRecords)

	// Step 3: Build and execute data query
	data, err := executeDataQuery[T](ctx, db, tableName, opts, scanFunc, joins...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}

	// Step 4: Return paginated result
	return &query.PaginatedResult{
		Data: data,
		Meta: meta,
	}, nil
}

// executeCountQuery builds and executes the COUNT query for pagination
func executeCountQuery(
	ctx context.Context,
	db DBConn,
	tableName string,
	opts *query.QueryOptions,
	joins ...query.JoinConfig,
) (int, error) {
	// Build COUNT query (without ORDER BY and LIMIT/OFFSET)
	builder := query.NewQueryBuilder(tableName, opts)

	// Apply JOINs if provided
	if len(joins) > 0 {
		builder = builder.WithJoins(joins)
	}

	// Apply filters and search (but not ORDER BY or pagination)
	// Use ApplyFilterConditions() for operator support (v1.1.0+)
	// Note: Parser populates FilterConditions when parsing HTTP requests
	// ApplyFilters() handles direct Filter map usage (tests, manual QueryOptions)
	countSQL, countArgs, err := builder.
		ApplyFilterConditions().
		ApplyFilters().
		ApplyFiltersOr().
		ApplySearch().
		BuildCount()

	if err != nil {
		return 0, fmt.Errorf("failed to build count query: %w", err)
	}

	// Execute COUNT query
	var totalRecords int
	err = db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&totalRecords)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	return totalRecords, nil
}

// executeDataQuery builds and executes the main SELECT query
func executeDataQuery[T any](
	ctx context.Context,
	db DBConn,
	tableName string,
	opts *query.QueryOptions,
	scanFunc func(*sql.Rows) (*T, error),
	joins ...query.JoinConfig,
) ([]*T, error) {
	// Build SELECT query with all filters, order, and pagination
	builder := query.NewQueryBuilder(tableName, opts)

	// Apply JOINs if provided
	if len(joins) > 0 {
		builder = builder.WithJoins(joins)
	}

	// Use ApplyFilterConditions() for operator support (v1.1.0+)
	// Note: Parser populates FilterConditions when parsing HTTP requests
	// ApplyFilters() handles direct Filter map usage (tests, manual QueryOptions)
	dataSQL, dataArgs, err := builder.
		ApplyFilterConditions().
		ApplyFilters().
		ApplyFiltersOr().
		ApplySearch().
		ApplyOrder().
		ApplyPagination().
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build data query: %w", err)
	}

	// Execute SELECT query
	rows, err := db.QueryContext(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute data query: %w", err)
	}
	defer rows.Close()

	// Scan results using provided scanFunc
	var results []*T
	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, item)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// calculatePaginationMeta computes pagination metadata from query results
func calculatePaginationMeta(page, limit, totalRecords int) query.PaginationMeta {
	// Ensure page is at least 1
	if page < 1 {
		page = 1
	}

	// Ensure limit is at least 1
	if limit < 1 {
		limit = 10
	}

	// Calculate total pages
	pageCount := 0
	if totalRecords > 0 {
		pageCount = (totalRecords + limit - 1) / limit // Ceiling division
	}

	// Calculate count of items in current page
	offset := (page - 1) * limit
	count := limit
	if offset+limit > totalRecords {
		count = totalRecords - offset
	}
	if count < 0 {
		count = 0
	}

	// Calculate previous and next page numbers
	var previousPage interface{} = false
	if page > 1 {
		previousPage = page - 1
	}

	var nextPage interface{} = false
	if page < pageCount {
		nextPage = page + 1
	}

	return query.PaginationMeta{
		Page:         page,
		Limit:        limit,
		Count:        count,
		PreviousPage: previousPage,
		NextPage:     nextPage,
		PageCount:    pageCount,
		TotalRecords: totalRecords,
	}
}
