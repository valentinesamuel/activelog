package query

import (
	"fmt"
	"strings"
)

// OperatorWhitelist defines which operators are allowed for specific columns.
// This is CRITICAL for security - prevents malicious operator abuse.
//
// Different column types should allow different operators:
//   - Date/timestamp columns: eq, ne, gt, gte, lt, lte
//   - Numeric columns: eq, ne, gt, gte, lt, lte
//   - ID columns: eq only (comparisons don't make sense)
//   - String columns: eq, ne (comparisons may not be meaningful)
//
// Example usage:
//
//	operatorWhitelists := OperatorWhitelist{
//	    "created_at": []string{"eq", "gte", "lte", "gt", "lt"},
//	    "distance":   []string{"eq", "gte", "lte", "gt", "lt", "ne"},
//	    "user_id":    []string{"eq"}, // IDs only support equality
//	    "status":     []string{"eq", "ne"},
//	}
type OperatorWhitelist map[string][]string

// AllOperators returns all supported operators.
// Useful for columns that should allow all comparison types.
func AllOperators() []string {
	return []string{"eq", "ne", "gt", "gte", "lt", "lte"}
}

// ComparisonOperators returns operators for numeric/date comparisons.
// Useful for dates, timestamps, and numeric columns.
func ComparisonOperators() []string {
	return []string{"eq", "ne", "gt", "gte", "lt", "lte"}
}

// EqualityOperators returns operators for equality checks.
// Useful for string columns where ordering doesn't make sense.
func EqualityOperators() []string {
	return []string{"eq", "ne"}
}

// StrictEqualityOnly returns only the equality operator.
// Useful for ID columns where only exact matches are meaningful.
func StrictEqualityOnly() []string {
	return []string{"eq"}
}

// ValidateFilterConditions validates operator-based filter conditions.
// Checks that:
//  1. The column is in the allowed filters whitelist
//  2. The operator is valid for that specific column
//
// Example usage in handler:
//
//	allowedFilters := []string{"created_at", "distance", "user_id", "status"}
//	operatorWhitelists := OperatorWhitelist{
//	    "created_at": ComparisonOperators(),
//	    "distance":   ComparisonOperators(),
//	    "user_id":    StrictEqualityOnly(),
//	    "status":     EqualityOperators(),
//	}
//
//	if err := ValidateFilterConditions(opts, allowedFilters, operatorWhitelists); err != nil {
//	    return http.StatusBadRequest, err
//	}
//
// Returns an error if the column is not allowed or the operator is not whitelisted for that column.
func ValidateFilterConditions(
	opts *QueryOptions,
	allowedFilters []string,
	operatorWhitelists OperatorWhitelist,
) error {
	for _, condition := range opts.FilterConditions {
		// Check if column is in the allowed filters
		if !contains(allowedFilters, condition.Column) {
			return fmt.Errorf("filtering on column '%s' is not allowed", condition.Column)
		}

		// Check if operator is valid for this column
		allowedOperators, exists := operatorWhitelists[condition.Column]
		if !exists {
			// No specific whitelist for this column - default to all operators
			allowedOperators = AllOperators()
		}

		if !contains(allowedOperators, condition.Operator) {
			return fmt.Errorf(
				"operator '%s' is not allowed for column '%s' (allowed: %v)",
				condition.Operator,
				condition.Column,
				allowedOperators,
			)
		}

		// Validate that the operator is a known/supported operator
		validOperators := []string{"eq", "ne", "gt", "gte", "lt", "lte"}
		if !contains(validOperators, condition.Operator) {
			return fmt.Errorf("unknown operator '%s'", condition.Operator)
		}
	}
	return nil
}

// ValidateQueryOptions validates that only allowed columns are queried.
// This is CRITICAL for security - it prevents unauthorized column access and SQL injection.
//
// Each endpoint must define explicit whitelists for what columns can be:
//   - Filtered (filter[column])
//   - Searched (search[column])
//   - Ordered (order[column])
//
// Example usage in handler:
//
//	allowedFilters := []string{"activity_type", "status", "user_id"}
//	allowedSearch := []string{"title", "description"}
//	allowedOrder := []string{"created_at", "activity_date", "amount"}
//
//	if err := ValidateQueryOptions(opts, allowedFilters, allowedSearch, allowedOrder); err != nil {
//	    return http.StatusBadRequest, err
//	}
//
// Returns an error if any column is not in the appropriate whitelist.
func ValidateQueryOptions(
	opts *QueryOptions,
	allowedFilters []string,
	allowedSearch []string,
	allowedOrder []string,
) error {
	// Validate filter columns (AND conditions)
	for column := range opts.Filter {
		if !contains(allowedFilters, column) {
			return fmt.Errorf("filtering on column '%s' is not allowed", column)
		}
	}

	// Validate filterOr columns (OR conditions)
	for column := range opts.FilterOr {
		if !contains(allowedFilters, column) {
			return fmt.Errorf("filtering on column '%s' is not allowed", column)
		}
	}

	// Validate search columns
	for column := range opts.Search {
		if !contains(allowedSearch, column) {
			return fmt.Errorf("searching on column '%s' is not allowed", column)
		}
	}

	// Validate order columns
	for column := range opts.Order {
		if !contains(allowedOrder, column) {
			return fmt.Errorf("ordering by column '%s' is not allowed", column)
		}
	}

	// Validate order directions
	for column, direction := range opts.Order {
		if err := ValidateOrderDirection(direction); err != nil {
			return fmt.Errorf("invalid order direction for column '%s': %w", column, err)
		}
	}

	// Validate page and limit ranges
	if opts.Page < 1 {
		return fmt.Errorf("page must be at least 1")
	}

	if opts.Limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}

	// Enforce maximum page size to prevent DoS
	const MaxPageSize = 100
	if opts.Limit > MaxPageSize {
		return fmt.Errorf("limit cannot exceed %d", MaxPageSize)
	}

	return nil
}

// ValidateOrderDirection ensures the order direction is either ASC or DESC.
// Case-insensitive comparison.
//
// Valid values: "ASC", "DESC", "asc", "desc"
// Invalid values: "ASCENDING", "random", "1", etc.
func ValidateOrderDirection(direction string) error {
	upper := strings.ToUpper(direction)
	if upper != "ASC" && upper != "DESC" {
		return fmt.Errorf("invalid order direction: %s (must be ASC or DESC)", direction)
	}
	return nil
}

// contains checks if a slice contains a specific string (case-insensitive).
// Used for whitelist checking.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// ValidateWithConfig validates query options using a validation configuration.
// This is a more structured approach for complex validation rules.
type ValidationConfig struct {
	AllowedFilters     []string
	AllowedSearch      []string
	AllowedOrder       []string
	OperatorWhitelists OperatorWhitelist // NEW in v1.1.0 - per-column operator restrictions
	MaxPageSize        int
	RequireUserScope   bool // Whether user_id must be in filters (for multi-tenancy)
}

// DefaultValidationConfig returns a validation config with sensible defaults.
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		AllowedFilters:     []string{},
		AllowedSearch:      []string{},
		AllowedOrder:       []string{"created_at", "updated_at"},
		OperatorWhitelists: make(OperatorWhitelist),
		MaxPageSize:        100,
		RequireUserScope:   true,
	}
}

// ValidateWithConfig validates query options using the provided configuration.
// More flexible than ValidateQueryOptions for complex scenarios.
//
// Example (v1.1.0+ with operator whitelists):
//
//	config := &ValidationConfig{
//	    AllowedFilters:   []string{"activity_type", "status", "created_at", "distance"},
//	    AllowedSearch:    []string{"title"},
//	    AllowedOrder:     []string{"created_at", "activity_date"},
//	    OperatorWhitelists: OperatorWhitelist{
//	        "created_at": ComparisonOperators(),
//	        "distance":   ComparisonOperators(),
//	        "user_id":    StrictEqualityOnly(),
//	    },
//	    MaxPageSize:      50,
//	    RequireUserScope: true,
//	}
//	if err := ValidateWithConfig(opts, config); err != nil {
//	    return err
//	}
func ValidateWithConfig(opts *QueryOptions, config *ValidationConfig) error {
	// Standard validation
	if err := ValidateQueryOptions(opts, config.AllowedFilters, config.AllowedSearch, config.AllowedOrder); err != nil {
		return err
	}

	// Validate operator-based filter conditions (NEW in v1.1.0)
	if len(opts.FilterConditions) > 0 {
		if err := ValidateFilterConditions(opts, config.AllowedFilters, config.OperatorWhitelists); err != nil {
			return err
		}
	}

	// Check max page size
	if opts.Limit > config.MaxPageSize {
		return fmt.Errorf("limit cannot exceed %d", config.MaxPageSize)
	}

	// Check user scope requirement (for multi-tenancy)
	if config.RequireUserScope {
		if _, hasUserID := opts.Filter["user_id"]; !hasUserID {
			return fmt.Errorf("user_id filter is required for multi-tenant queries")
		}
	}

	return nil
}

// SanitizeSearchTerm sanitizes a search term to prevent SQL injection in pattern matching.
// Escapes special characters used in LIKE/ILIKE patterns.
//
// Special characters in LIKE patterns:
//   - % (matches any number of characters)
//   - _ (matches a single character)
//   - \ (escape character)
//
// Example:
//   - "test%" → "test\\%"
//   - "hello_world" → "hello\\_world"
//
// Note: This is defense-in-depth. Parameterized queries (which we use) already prevent SQL injection,
// but this prevents users from abusing LIKE wildcards.
func SanitizeSearchTerm(term string) string {
	// Escape backslashes first
	term = strings.ReplaceAll(term, "\\", "\\\\")

	// Escape LIKE wildcards
	term = strings.ReplaceAll(term, "%", "\\%")
	term = strings.ReplaceAll(term, "_", "\\_")

	return term
}

// ValidateColumnName validates that a column name is safe to use in SQL.
// Prevents SQL injection through column names.
//
// Rules:
//   - Must start with a letter
//   - Can contain letters, numbers, underscores
//   - Can contain dots for qualified names (table.column)
//   - Max length: 63 characters (PostgreSQL limit)
//
// Valid: "user_id", "created_at", "t.name", "activities.activity_type"
// Invalid: "'; DROP TABLE users--", "user id", "123column"
func ValidateColumnName(column string) error {
	if column == "" {
		return fmt.Errorf("column name cannot be empty")
	}

	if len(column) > 63 {
		return fmt.Errorf("column name too long (max 63 characters)")
	}

	// Allow letters, numbers, underscores, and dots for qualified names
	for i, r := range column {
		if i == 0 && !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return fmt.Errorf("column name must start with a letter: %s", column)
		}

		valid := (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' ||
			r == '.'

		if !valid {
			return fmt.Errorf("column name contains invalid character '%c': %s", r, column)
		}
	}

	return nil
}
