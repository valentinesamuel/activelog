package query

// FilterCondition represents a single filter condition with an operator.
// Enables comparison operations beyond simple equality.
//
// Supported operators:
//   - "eq"  : Equal (=)
//   - "ne"  : Not Equal (!=)
//   - "gt"  : Greater Than (>)
//   - "gte" : Greater Than or Equal (>=)
//   - "lt"  : Less Than (<)
//   - "lte" : Less Than or Equal (<=)
//
// Example usage:
//
//	condition := FilterCondition{
//	    Column:   "created_at",
//	    Operator: "gte",
//	    Value:    "2024-01-01",
//	}
//	// SQL: WHERE created_at >= $1
type FilterCondition struct {
	// Column is the database column name
	Column string `json:"column"`

	// Operator is the comparison operator (eq, ne, gt, gte, lt, lte)
	Operator string `json:"operator"`

	// Value is the value to compare against
	Value interface{} `json:"value"`
}

// QueryOptions represents all possible query parameters for dynamic filtering.
// This structure is universal and works for ANY entity (Activities, Tags, Users, Stats).
//
// Example usage (legacy equality syntax):
//
//	opts := &QueryOptions{
//	    Page:   1,
//	    Limit:  20,
//	    Filter: map[string]interface{}{
//	        "activity_type": "running",
//	        "status":       "completed",
//	    },
//	    Search: map[string]interface{}{
//	        "title": "morning",
//	    },
//	    Order: map[string]string{
//	        "created_at": "DESC",
//	    },
//	}
//
// Example usage (new operator syntax):
//
//	opts := &QueryOptions{
//	    Page:  1,
//	    Limit: 20,
//	    FilterConditions: []FilterCondition{
//	        {Column: "created_at", Operator: "gte", Value: "2024-01-01"},
//	        {Column: "distance", Operator: "lt", Value: 10.0},
//	    },
//	}
type QueryOptions struct {
	// Page is the current page number (1-indexed)
	Page int `json:"page"`

	// Limit is the maximum number of items per page
	Limit int `json:"limit"`

	// Filter contains AND conditions for WHERE clause (LEGACY - kept for backward compatibility)
	// Example: {"activity_type": "running", "status": "completed"}
	// SQL: WHERE activity_type = $1 AND status = $2
	// NOTE: For operator-based filtering, use FilterConditions instead
	Filter map[string]interface{} `json:"filter"`

	// FilterConditions contains operator-based filter conditions (NEW in v1.1.0)
	// Enables comparison operators: eq, ne, gt, gte, lt, lte
	// Example: []FilterCondition{{Column: "created_at", Operator: "gte", Value: "2024-01-01"}}
	// SQL: WHERE created_at >= $1
	FilterConditions []FilterCondition `json:"filterConditions"`

	// FilterOr contains OR conditions for WHERE clause
	// Example: {"type": "running", "type": "cycling"}
	// SQL: WHERE (type = $1 OR type = $2)
	FilterOr map[string]interface{} `json:"filterOr"`

	// Search contains ILIKE pattern matching conditions
	// Example: {"title": "morning", "description": "run"}
	// SQL: WHERE (title ILIKE '%morning%' OR description ILIKE '%run%')
	Search map[string]interface{} `json:"search"`

	// Order contains column → direction mappings for ORDER BY
	// Example: {"created_at": "DESC", "amount": "ASC"}
	// SQL: ORDER BY created_at DESC, amount ASC
	Order map[string]string `json:"order"`
}

// PaginatedResult represents paginated data with metadata.
// This is the standard response structure for all list endpoints.
//
// Example JSON response:
//
//	{
//	    "data": [...],
//	    "meta": {
//	        "page": 2,
//	        "limit": 20,
//	        "count": 20,
//	        "previousPage": 1,
//	        "nextPage": 3,
//	        "pageCount": 5,
//	        "totalRecords": 95
//	    }
//	}
type PaginatedResult struct {
	// Data contains the actual result items (type varies by entity)
	Data interface{} `json:"data"`

	// Meta contains pagination metadata
	Meta PaginationMeta `json:"meta"`
}

// PaginationMeta contains metadata about the pagination state.
type PaginationMeta struct {
	// Page is the current page number (1-indexed)
	Page int `json:"page"`

	// Limit is the maximum number of items per page
	Limit int `json:"limit"`

	// Count is the number of items in the current page
	Count int `json:"count"`

	// PreviousPage is the previous page number, or false if on first page
	// Type: int or bool (false)
	PreviousPage interface{} `json:"previousPage"`

	// NextPage is the next page number, or false if on last page
	// Type: int or bool (false)
	NextPage interface{} `json:"nextPage"`

	// PageCount is the total number of pages
	PageCount int `json:"pageCount"`

	// TotalRecords is the total number of records across all pages
	TotalRecords int `json:"totalRecords"`
}

// JoinConfig defines a table join configuration for relationship filtering.
// Used when filtering by related entities (e.g., filter activities by tag names).
//
// Example: Filter activities by tags
//
//	joins := []JoinConfig{
//	    {
//	        Table:     "activity_tags at",
//	        Condition: "at.activity_id = activities.id",
//	        Alias:     "at",
//	    },
//	    {
//	        Table:     "tags t",
//	        Condition: "t.id = at.tag_id",
//	        Alias:     "t",
//	    },
//	}
type JoinConfig struct {
	// Table is the table name with optional alias
	// Example: "activity_tags at", "tags t", "users u"
	Table string

	// Condition is the JOIN condition
	// Example: "at.activity_id = activities.id", "t.id = at.tag_id"
	Condition string

	// Alias is the table alias used in the condition
	// Example: "at", "t", "u"
	Alias string
}

// NewQueryOptions creates a QueryOptions with sensible defaults.
// Page defaults to 1, Limit defaults to 10, all maps are initialized.
func NewQueryOptions() *QueryOptions {
	return &QueryOptions{
		Page:             1,
		Limit:            10,
		Filter:           make(map[string]interface{}),
		FilterConditions: []FilterCondition{},
		FilterOr:         make(map[string]interface{}),
		Search:           make(map[string]interface{}),
		Order:            make(map[string]string),
	}
}
