package query

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

// QueryBuilder builds SQL queries from QueryOptions using Squirrel.
// Provides a fluent API for chaining query modifications.
//
// Example usage:
//
//	builder := NewQueryBuilder("activities", queryOptions)
//	sql, args, err := builder.
//	    ApplyFilters().
//	    ApplyFiltersOr().
//	    ApplySearch().
//	    ApplyOrder().
//	    ApplyPagination().
//	    Build()
type QueryBuilder struct {
	baseQuery sq.SelectBuilder
	options   *QueryOptions
	tableName string
	joins     []JoinConfig
}

// resolveColumnForSQL translates a multi-level dot-notation path to a valid SQL column.
// For paths with 3+ segments (e.g., "tags.parent.name"), the last 2 segments are used
// so the result maps to the aliased JOIN table (e.g., "parent.name").
// Shorter paths (e.g., "tags.name", "activity_type") are returned unchanged.
//
// Examples:
//
//	"tags.parent.name"           → "parent.name"
//	"users.company.dept.title"   → "dept.title"
//	"tags.name"                  → "tags.name"
//	"activity_type"              → "activity_type"
func resolveColumnForSQL(column string) string {
	segments := strings.Split(column, ".")
	if len(segments) >= 3 {
		return segments[len(segments)-2] + "." + segments[len(segments)-1]
	}
	return column
}

// NewQueryBuilder creates a new query builder for the specified table.
//
// Parameters:
//   - tableName: The main table to query (e.g., "activities", "tags", "users")
//   - opts: The parsed query options from the HTTP request
//
// Returns a QueryBuilder ready for method chaining.
func NewQueryBuilder(tableName string, opts *QueryOptions) *QueryBuilder {
	// Use table.* to select only columns from the main table
	// This prevents issues when JOINs are added later
	selectExpr := fmt.Sprintf("%s.*", tableName)
	return &QueryBuilder{
		baseQuery: sq.Select(selectExpr).From(tableName),
		options:   opts,
		tableName: tableName,
		joins:     []JoinConfig{},
	}
}

// WithJoins adds JOIN clauses to the query for relationship filtering.
// This must be called before ApplyFilters if you want to filter on joined columns.
//
// Example:
//
//	joins := []JoinConfig{
//	    {Table: "activity_tags at", Condition: "at.activity_id = activities.id"},
//	    {Table: "tags t", Condition: "t.id = at.tag_id"},
//	}
//	builder.WithJoins(joins).ApplyFilters()
func (qb *QueryBuilder) WithJoins(joins []JoinConfig) *QueryBuilder {
	qb.joins = joins
	for _, join := range joins {
		qb.baseQuery = qb.baseQuery.LeftJoin(fmt.Sprintf("%s ON %s", join.Table, join.Condition))
	}
	return qb
}

// ApplyFilterConditions applies WHERE conditions with operator support.
// Handles comparison operators: eq, ne, gt, gte, lt, lte.
// This is the NEW method (v1.1.0+) that enables date ranges and numeric comparisons.
//
// Examples:
//   - {Column: "created_at", Operator: "gte", Value: "2024-01-01"} → WHERE created_at >= $1
//   - {Column: "distance", Operator: "lt", Value: 10} → WHERE distance < $1
//   - {Column: "status", Operator: "eq", Value: "active"} → WHERE status = $1
//   - {Column: "amount", Operator: "ne", Value: 0} → WHERE amount != $1
//
// Supported operators:
//   - "eq"  : Equal (=)
//   - "ne"  : Not Equal (!=)
//   - "gt"  : Greater Than (>)
//   - "gte" : Greater Than or Equal (>=)
//   - "lt"  : Less Than (<)
//   - "lte" : Less Than or Equal (<=)
func (qb *QueryBuilder) ApplyFilterConditions() *QueryBuilder {
	for _, condition := range qb.options.FilterConditions {
		column := resolveColumnForSQL(condition.Column)
		value := condition.Value

		switch condition.Operator {
		case "eq":
			qb.baseQuery = qb.baseQuery.Where(sq.Eq{column: value})
		case "ne":
			qb.baseQuery = qb.baseQuery.Where(sq.NotEq{column: value})
		case "gt":
			qb.baseQuery = qb.baseQuery.Where(sq.Gt{column: value})
		case "gte":
			qb.baseQuery = qb.baseQuery.Where(sq.GtOrEq{column: value})
		case "lt":
			qb.baseQuery = qb.baseQuery.Where(sq.Lt{column: value})
		case "lte":
			qb.baseQuery = qb.baseQuery.Where(sq.LtOrEq{column: value})
		default:
			// Unknown operator - skip (validation should catch this earlier)
			continue
		}
	}
	return qb
}

// ApplyFilters applies WHERE conditions with AND logic.
// LEGACY METHOD - Kept for backward compatibility.
// For new code, FilterConditions with ApplyFilterConditions() is preferred.
//
// Handles single values, arrays (IN clause), and proper type conversion.
//
// Examples:
//   - {"status": "active"} → WHERE status = $1
//   - {"type": []string{"running", "cycling"}} → WHERE type IN ($1, $2)
//   - {"user_id": 123, "status": "active"} → WHERE user_id = $1 AND status = $2
func (qb *QueryBuilder) ApplyFilters() *QueryBuilder {
	for rawColumn, value := range qb.options.Filter {
		column := resolveColumnForSQL(rawColumn)
		switch v := value.(type) {
		case []interface{}:
			// WHERE column IN (val1, val2, val3)
			qb.baseQuery = qb.baseQuery.Where(sq.Eq{column: v})

		case []string:
			// Convert []string to []interface{} for Squirrel
			vals := make([]interface{}, len(v))
			for i, s := range v {
				vals[i] = s
			}
			qb.baseQuery = qb.baseQuery.Where(sq.Eq{column: vals})

		case nil:
			// WHERE column IS NULL
			qb.baseQuery = qb.baseQuery.Where(sq.Eq{column: nil})

		default:
			// WHERE column = value
			qb.baseQuery = qb.baseQuery.Where(sq.Eq{column: v})
		}
	}
	return qb
}

// ApplyFiltersOr applies WHERE conditions with OR logic.
// These conditions are grouped with parentheses: (col1 = val1 OR col2 = val2).
//
// Examples:
//   - {"type": "running", "type": "cycling"} → WHERE (type = $1 OR type = $2)
//   - Combined with AND filters → WHERE status = 'active' AND (type = $1 OR type = $2)
func (qb *QueryBuilder) ApplyFiltersOr() *QueryBuilder {
	if len(qb.options.FilterOr) == 0 {
		return qb
	}

	orConditions := sq.Or{}
	for rawColumn, value := range qb.options.FilterOr {
		column := resolveColumnForSQL(rawColumn)
		switch v := value.(type) {
		case []interface{}:
			orConditions = append(orConditions, sq.Eq{column: v})
		case []string:
			vals := make([]interface{}, len(v))
			for i, s := range v {
				vals[i] = s
			}
			orConditions = append(orConditions, sq.Eq{column: vals})
		case nil:
			orConditions = append(orConditions, sq.Eq{column: nil})
		default:
			orConditions = append(orConditions, sq.Eq{column: v})
		}
	}

	qb.baseQuery = qb.baseQuery.Where(orConditions)
	return qb
}

// ApplySearch applies ILIKE pattern matching for fuzzy search.
// Multiple search terms are combined with OR logic.
//
// Examples:
//   - {"title": "morning"} → WHERE title ILIKE '%morning%'
//   - {"title": "morning", "description": "run"} → WHERE (title ILIKE '%morning%' OR description ILIKE '%run%')
//
// Note: ILIKE is PostgreSQL-specific case-insensitive LIKE.
// For other databases, use LIKE or LOWER(column) LIKE LOWER(pattern).
func (qb *QueryBuilder) ApplySearch() *QueryBuilder {
	if len(qb.options.Search) == 0 {
		return qb
	}

	searchConditions := sq.Or{}
	for rawColumn, value := range qb.options.Search {
		column := resolveColumnForSQL(rawColumn)
		pattern := fmt.Sprintf("%%%v%%", value)
		// Use ILike for PostgreSQL case-insensitive search
		searchConditions = append(searchConditions, sq.ILike{column: pattern})
	}

	qb.baseQuery = qb.baseQuery.Where(searchConditions)
	return qb
}

// ApplyOrder applies ORDER BY clause for sorting.
// Multiple order columns are applied in the order specified.
//
// Examples:
//   - {"created_at": "DESC"} → ORDER BY created_at DESC
//   - {"amount": "ASC", "created_at": "DESC"} → ORDER BY amount ASC, created_at DESC
//
// If no order is specified, defaults to "created_at DESC".
func (qb *QueryBuilder) ApplyOrder() *QueryBuilder {
	if len(qb.options.Order) == 0 {
		// Default order - qualify with table name if there are JOINs
		defaultColumn := "created_at"
		if len(qb.joins) > 0 && !strings.Contains(defaultColumn, ".") {
			defaultColumn = fmt.Sprintf("%s.%s", qb.tableName, defaultColumn)
		}
		qb.baseQuery = qb.baseQuery.OrderBy(fmt.Sprintf("%s DESC", defaultColumn))
		return qb
	}

	// Apply each order clause
	// Note: map iteration order is not guaranteed in Go, but for sorting
	// this usually doesn't matter. For strict ordering, consider using
	// a slice of structs instead.
	for rawColumn, direction := range qb.options.Order {
		column := resolveColumnForSQL(rawColumn)
		// Validate direction (should be done in validator, but double-check here)
		upperDir := strings.ToUpper(direction)
		if upperDir != "ASC" && upperDir != "DESC" {
			upperDir = "ASC" // Default to ASC if invalid
		}

		// Qualify column with table name if there are JOINs and column isn't already qualified
		qualifiedColumn := column
		if len(qb.joins) > 0 && !strings.Contains(column, ".") {
			qualifiedColumn = fmt.Sprintf("%s.%s", qb.tableName, column)
		}

		orderClause := fmt.Sprintf("%s %s", qualifiedColumn, upperDir)
		qb.baseQuery = qb.baseQuery.OrderBy(orderClause)
	}

	return qb
}

// ApplyPagination applies LIMIT and OFFSET for pagination.
//
// Formula: OFFSET = (Page - 1) * Limit
//
// Examples:
//   - Page=1, Limit=10 → LIMIT 10 OFFSET 0
//   - Page=2, Limit=10 → LIMIT 10 OFFSET 10
//   - Page=3, Limit=20 → LIMIT 20 OFFSET 40
//
// Defaults:
//   - Minimum page: 1
//   - Minimum limit: 1
//   - Default limit: 10
func (qb *QueryBuilder) ApplyPagination() *QueryBuilder {
	limit := qb.options.Limit
	if limit <= 0 {
		limit = 10
	}

	page := qb.options.Page
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	qb.baseQuery = qb.baseQuery.Limit(uint64(limit)).Offset(uint64(offset))
	return qb
}

// Build generates the final SQL query with PostgreSQL-style placeholders ($1, $2, ...).
//
// Returns:
//   - sql: The SQL query string with placeholders
//   - args: The argument values for the placeholders
//   - error: Any error during query building
//
// Example output:
//
//	sql: "SELECT * FROM activities WHERE activity_type = $1 AND user_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4"
//	args: []interface{}{"running", 123, 10, 0}
func (qb *QueryBuilder) Build() (string, []interface{}, error) {
	return qb.baseQuery.PlaceholderFormat(sq.Dollar).ToSql()
}

// BuildCount generates a COUNT query for pagination metadata.
// Uses the same WHERE conditions as the data query but without ORDER BY, LIMIT, OFFSET.
//
// Returns:
//   - sql: The COUNT query string
//   - args: The argument values for the placeholders
//   - error: Any error during query building
//
// Example output:
//
//	sql: "SELECT COUNT(*) FROM activities WHERE activity_type = $1 AND user_id = $2"
//	args: []interface{}{"running", 123}
func (qb *QueryBuilder) BuildCount() (string, []interface{}, error) {
	countQuery := sq.Select("COUNT(*)").From(qb.tableName)

	// Add JOINs if present (needed for filtering on joined tables)
	for _, join := range qb.joins {
		countQuery = countQuery.LeftJoin(fmt.Sprintf("%s ON %s", join.Table, join.Condition))
	}

	// Apply FilterConditions (operator-based filtering - NEW in v1.1.0)
	for _, condition := range qb.options.FilterConditions {
		column := resolveColumnForSQL(condition.Column)
		value := condition.Value

		switch condition.Operator {
		case "eq":
			countQuery = countQuery.Where(sq.Eq{column: value})
		case "ne":
			countQuery = countQuery.Where(sq.NotEq{column: value})
		case "gt":
			countQuery = countQuery.Where(sq.Gt{column: value})
		case "gte":
			countQuery = countQuery.Where(sq.GtOrEq{column: value})
		case "lt":
			countQuery = countQuery.Where(sq.Lt{column: value})
		case "lte":
			countQuery = countQuery.Where(sq.LtOrEq{column: value})
		}
	}

	// Apply Filter (AND conditions - LEGACY, kept for backward compatibility)
	for rawColumn, value := range qb.options.Filter {
		column := resolveColumnForSQL(rawColumn)
		switch v := value.(type) {
		case []interface{}:
			countQuery = countQuery.Where(sq.Eq{column: v})
		case []string:
			vals := make([]interface{}, len(v))
			for i, s := range v {
				vals[i] = s
			}
			countQuery = countQuery.Where(sq.Eq{column: vals})
		case nil:
			countQuery = countQuery.Where(sq.Eq{column: nil})
		default:
			countQuery = countQuery.Where(sq.Eq{column: v})
		}
	}

	// Apply FilterOr (OR conditions)
	if len(qb.options.FilterOr) > 0 {
		orConditions := sq.Or{}
		for rawColumn, value := range qb.options.FilterOr {
			column := resolveColumnForSQL(rawColumn)
			switch v := value.(type) {
			case []interface{}:
				orConditions = append(orConditions, sq.Eq{column: v})
			case []string:
				vals := make([]interface{}, len(v))
				for i, s := range v {
					vals[i] = s
				}
				orConditions = append(orConditions, sq.Eq{column: vals})
			case nil:
				orConditions = append(orConditions, sq.Eq{column: nil})
			default:
				orConditions = append(orConditions, sq.Eq{column: v})
			}
		}
		countQuery = countQuery.Where(orConditions)
	}

	// Apply Search conditions
	if len(qb.options.Search) > 0 {
		searchConditions := sq.Or{}
		for rawColumn, value := range qb.options.Search {
			column := resolveColumnForSQL(rawColumn)
			pattern := fmt.Sprintf("%%%v%%", value)
			searchConditions = append(searchConditions, sq.ILike{column: pattern})
		}
		countQuery = countQuery.Where(searchConditions)
	}

	return countQuery.PlaceholderFormat(sq.Dollar).ToSql()
}
