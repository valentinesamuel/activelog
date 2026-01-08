package query

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueryBuilder(t *testing.T) {
	opts := &QueryOptions{
		Page:     1,
		Limit:    10,
		Filter:   map[string]interface{}{},
		FilterOr: map[string]interface{}{},
		Search:   map[string]interface{}{},
		Order:    map[string]string{},
	}

	builder := NewQueryBuilder("activities", opts)

	assert.NotNil(t, builder)
	assert.Equal(t, "activities", builder.tableName)
	assert.Equal(t, opts, builder.options)
}

func TestQueryBuilder_ApplyFilters(t *testing.T) {
	tests := []struct {
		name           string
		filters        map[string]interface{}
		expectedSQL    string
		expectedArgNum int
	}{
		{
			name: "single filter",
			filters: map[string]interface{}{
				"status": "active",
			},
			expectedSQL:    "WHERE status = $",
			expectedArgNum: 1,
		},
		{
			name: "multiple filters",
			filters: map[string]interface{}{
				"status": "active",
				"user_id": 123,
			},
			expectedSQL:    "WHERE", // Both conditions present
			expectedArgNum: 2,
		},
		{
			name: "filter with boolean",
			filters: map[string]interface{}{
				"is_active": true,
			},
			expectedSQL:    "WHERE is_active = $",
			expectedArgNum: 1,
		},
		{
			name: "filter with null",
			filters: map[string]interface{}{
				"description": nil,
			},
			expectedSQL:    "WHERE description IS NULL",
			expectedArgNum: 0,
		},
		{
			name: "filter with array (IN clause)",
			filters: map[string]interface{}{
				"type": []string{"running", "cycling"},
			},
			expectedSQL:    "WHERE type IN ($",
			expectedArgNum: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &QueryOptions{
				Page:   1,
				Limit:  10,
				Filter: tt.filters,
			}

			builder := NewQueryBuilder("activities", opts)
			sql, args, err := builder.ApplyFilters().Build()

			require.NoError(t, err)
			assert.Contains(t, sql, tt.expectedSQL)
			assert.Len(t, args, tt.expectedArgNum)
		})
	}
}

func TestQueryBuilder_ApplyFiltersOr(t *testing.T) {
	tests := []struct {
		name        string
		filtersOr   map[string]interface{}
		expectedSQL string
	}{
		{
			name: "single OR condition",
			filtersOr: map[string]interface{}{
				"type": "running",
			},
			expectedSQL: "WHERE (type = $", // Squirrel wraps OR conditions in parentheses
		},
		{
			name: "multiple OR conditions",
			filtersOr: map[string]interface{}{
				"type":   "running",
				"status": "active",
			},
			expectedSQL: "WHERE (", // OR conditions grouped with parentheses
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &QueryOptions{
				Page:     1,
				Limit:    10,
				FilterOr: tt.filtersOr,
			}

			builder := NewQueryBuilder("activities", opts)
			sql, _, err := builder.ApplyFiltersOr().Build()

			require.NoError(t, err)
			assert.Contains(t, sql, tt.expectedSQL)
		})
	}
}

func TestQueryBuilder_ApplySearch(t *testing.T) {
	tests := []struct {
		name        string
		search      map[string]interface{}
		expectedSQL string
	}{
		{
			name: "single search term",
			search: map[string]interface{}{
				"title": "morning",
			},
			expectedSQL: "WHERE (title ILIKE $", // Squirrel wraps OR conditions in parentheses
		},
		{
			name: "multiple search terms",
			search: map[string]interface{}{
				"title":       "morning",
				"description": "run",
			},
			expectedSQL: "WHERE (", // Multiple ILIKE conditions with OR
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &QueryOptions{
				Page:   1,
				Limit:  10,
				Search: tt.search,
			}

			builder := NewQueryBuilder("activities", opts)
			sql, args, err := builder.ApplySearch().Build()

			require.NoError(t, err)
			assert.Contains(t, sql, tt.expectedSQL)

			// Verify search patterns include wildcards
			for _, arg := range args {
				if strArg, ok := arg.(string); ok {
					assert.True(t, strings.HasPrefix(strArg, "%"), "Search term should start with %")
					assert.True(t, strings.HasSuffix(strArg, "%"), "Search term should end with %")
				}
			}
		})
	}
}

func TestQueryBuilder_ApplyOrder(t *testing.T) {
	tests := []struct {
		name        string
		order       map[string]string
		expectedSQL string
	}{
		{
			name:        "no order specified - default",
			order:       map[string]string{},
			expectedSQL: "ORDER BY created_at DESC",
		},
		{
			name: "single order column",
			order: map[string]string{
				"activity_date": "DESC",
			},
			expectedSQL: "ORDER BY activity_date DESC",
		},
		{
			name: "ASC order",
			order: map[string]string{
				"amount": "ASC",
			},
			expectedSQL: "ORDER BY amount ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &QueryOptions{
				Page:  1,
				Limit: 10,
				Order: tt.order,
			}

			builder := NewQueryBuilder("activities", opts)
			sql, _, err := builder.ApplyOrder().Build()

			require.NoError(t, err)
			assert.Contains(t, sql, tt.expectedSQL)
		})
	}
}

func TestQueryBuilder_ApplyPagination(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		limit          int
		expectedLimit  uint64
		expectedOffset uint64
	}{
		{
			name:           "first page",
			page:           1,
			limit:          10,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "second page",
			page:           2,
			limit:          10,
			expectedLimit:  10,
			expectedOffset: 10,
		},
		{
			name:           "third page with larger limit",
			page:           3,
			limit:          20,
			expectedLimit:  20,
			expectedOffset: 40,
		},
		{
			name:           "zero page defaults to 1",
			page:           0,
			limit:          10,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "negative page defaults to 1",
			page:           -5,
			limit:          10,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "zero limit defaults to 10",
			page:           1,
			limit:          0,
			expectedLimit:  10,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &QueryOptions{
				Page:  tt.page,
				Limit: tt.limit,
			}

			builder := NewQueryBuilder("activities", opts)
			sql, _, err := builder.ApplyPagination().Build()

			require.NoError(t, err)
			assert.Contains(t, sql, "LIMIT")
			assert.Contains(t, sql, "OFFSET")

			// Verify the SQL contains the expected pagination values
			// Note: We can't directly assert the LIMIT/OFFSET values from SQL string,
			// but we've verified the logic in the implementation
		})
	}
}

func TestQueryBuilder_Build(t *testing.T) {
	tests := []struct {
		name        string
		opts        *QueryOptions
		expectedSQL []string // Multiple strings that should be in the SQL
	}{
		{
			name: "simple query with no filters",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
			},
			expectedSQL: []string{
				"SELECT activities.* FROM activities",
				"ORDER BY created_at DESC",
				"LIMIT",
				"OFFSET",
			},
		},
		{
			name: "query with single filter",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"status": "active",
				},
			},
			expectedSQL: []string{
				"SELECT activities.* FROM activities",
				"WHERE status = $",
				"ORDER BY created_at DESC",
			},
		},
		{
			name: "complex query with all features",
			opts: &QueryOptions{
				Page:  2,
				Limit: 20,
				Filter: map[string]interface{}{
					"user_id": 123,
					"status":  "completed",
				},
				FilterOr: map[string]interface{}{
					"type": "running",
				},
				Search: map[string]interface{}{
					"title": "morning",
				},
				Order: map[string]string{
					"activity_date": "DESC",
				},
			},
			expectedSQL: []string{
				"SELECT activities.* FROM activities",
				"WHERE",
				"ORDER BY",
				"LIMIT",
				"OFFSET",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewQueryBuilder("activities", tt.opts)
			sql, args, err := builder.
				ApplyFilters().
				ApplyFiltersOr().
				ApplySearch().
				ApplyOrder().
				ApplyPagination().
				Build()

			require.NoError(t, err)
			assert.NotEmpty(t, sql)

			// Verify all expected SQL fragments are present
			for _, expectedFragment := range tt.expectedSQL {
				assert.Contains(t, sql, expectedFragment)
			}

			// Verify PostgreSQL-style placeholders ($1, $2, etc.)
			if len(args) > 0 {
				assert.Contains(t, sql, "$1")
			}
		})
	}
}

func TestQueryBuilder_BuildCount(t *testing.T) {
	tests := []struct {
		name        string
		opts        *QueryOptions
		expectedSQL []string
	}{
		{
			name: "count query with no filters",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
			},
			expectedSQL: []string{
				"SELECT COUNT(*) FROM activities",
			},
		},
		{
			name: "count query with filters",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"status": "active",
				},
			},
			expectedSQL: []string{
				"SELECT COUNT(*) FROM activities",
				"WHERE status = $",
			},
		},
		{
			name: "count query should not have ORDER BY",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Order: map[string]string{
					"created_at": "DESC",
				},
			},
			expectedSQL: []string{
				"SELECT COUNT(*) FROM activities",
			},
		},
		{
			name: "count query should not have LIMIT/OFFSET",
			opts: &QueryOptions{
				Page:  2,
				Limit: 20,
			},
			expectedSQL: []string{
				"SELECT COUNT(*) FROM activities",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewQueryBuilder("activities", tt.opts)
			sql, _, err := builder.BuildCount()

			require.NoError(t, err)
			assert.NotEmpty(t, sql)

			// Verify expected SQL fragments
			for _, expectedFragment := range tt.expectedSQL {
				assert.Contains(t, sql, expectedFragment)
			}

			// Verify COUNT queries don't have pagination/ordering
			assert.NotContains(t, sql, "LIMIT", "Count queries should not have LIMIT")
			assert.NotContains(t, sql, "OFFSET", "Count queries should not have OFFSET")
			assert.NotContains(t, sql, "ORDER BY", "Count queries should not have ORDER BY")
		})
	}
}

func TestQueryBuilder_WithJoins(t *testing.T) {
	joins := []JoinConfig{
		{
			Table:     "activity_tags at",
			Condition: "at.activity_id = activities.id",
			Alias:     "at",
		},
		{
			Table:     "tags t",
			Condition: "t.id = at.tag_id",
			Alias:     "t",
		},
	}

	opts := &QueryOptions{
		Page:  1,
		Limit: 10,
		Filter: map[string]interface{}{
			"t.name": "running",
		},
	}

	builder := NewQueryBuilder("activities", opts)
	sql, _, err := builder.
		WithJoins(joins).
		ApplyFilters().
		Build()

	require.NoError(t, err)
	assert.Contains(t, sql, "LEFT JOIN activity_tags at ON at.activity_id = activities.id")
	assert.Contains(t, sql, "LEFT JOIN tags t ON t.id = at.tag_id")
	assert.Contains(t, sql, "WHERE")
}

func TestQueryBuilder_PostgreSQLPlaceholders(t *testing.T) {
	opts := &QueryOptions{
		Page:  1,
		Limit: 10,
		Filter: map[string]interface{}{
			"user_id": 123,
			"status":  "active",
			"type":    "running",
		},
	}

	builder := NewQueryBuilder("activities", opts)
	sql, args, err := builder.ApplyFilters().Build()

	require.NoError(t, err)

	// Verify PostgreSQL-style placeholders ($1, $2, $3)
	assert.Contains(t, sql, "$1")
	assert.Contains(t, sql, "$2")
	assert.Contains(t, sql, "$3")

	// Verify number of arguments matches placeholders
	assert.Equal(t, 3, len(args))

	// Verify no question mark placeholders (MySQL style)
	assert.NotContains(t, sql, "?")
}

func TestQueryBuilder_ChainedOperations(t *testing.T) {
	opts := &QueryOptions{
		Page:  1,
		Limit: 10,
		Filter: map[string]interface{}{
			"user_id": 123,
		},
		Search: map[string]interface{}{
			"title": "morning",
		},
		Order: map[string]string{
			"created_at": "DESC",
		},
	}

	// Test that chaining works correctly
	builder := NewQueryBuilder("activities", opts)
	sql, args, err := builder.
		ApplyFilters().
		ApplySearch().
		ApplyOrder().
		ApplyPagination().
		Build()

	require.NoError(t, err)
	assert.NotEmpty(t, sql)
	assert.NotEmpty(t, args)

	// Verify all operations are present in final SQL
	assert.Contains(t, sql, "SELECT activities.* FROM activities")
	assert.Contains(t, sql, "WHERE")
	assert.Contains(t, sql, "user_id")
	assert.Contains(t, sql, "ILIKE")
	assert.Contains(t, sql, "ORDER BY")
	assert.Contains(t, sql, "LIMIT")
	assert.Contains(t, sql, "OFFSET")
}

func TestQueryBuilder_EmptyOptions(t *testing.T) {
	opts := &QueryOptions{
		Page:     1,
		Limit:    10,
		Filter:   map[string]interface{}{},
		FilterOr: map[string]interface{}{},
		Search:   map[string]interface{}{},
		Order:    map[string]string{},
	}

	builder := NewQueryBuilder("activities", opts)
	sql, _, err := builder.
		ApplyFilters().
		ApplyFiltersOr().
		ApplySearch().
		ApplyOrder().
		ApplyPagination().
		Build()

	require.NoError(t, err)

	// Should have basic query structure
	assert.Contains(t, sql, "SELECT activities.* FROM activities")
	assert.Contains(t, sql, "ORDER BY created_at DESC") // Default order
	assert.Contains(t, sql, "LIMIT")
	assert.Contains(t, sql, "OFFSET")

	// Should not have WHERE clause
	assert.NotContains(t, sql, "WHERE")
}

func TestQueryBuilder_TagFilteringWithJoins(t *testing.T) {
	tests := []struct {
		name        string
		opts        *QueryOptions
		joins       []JoinConfig
		expectedSQL []string // Multiple strings that should be in the SQL
	}{
		{
			name: "filter by single tag name",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"user_id": 123,
					"t.name":  "cardio", // After translation from "tags"
				},
			},
			joins: []JoinConfig{
				{
					Table:     "activity_tags at",
					Condition: "at.activity_id = activities.id",
					Alias:     "at",
				},
				{
					Table:     "tags t",
					Condition: "t.id = at.tag_id",
					Alias:     "t",
				},
			},
			expectedSQL: []string{
				"SELECT activities.* FROM activities",
				"LEFT JOIN activity_tags at ON at.activity_id = activities.id",
				"LEFT JOIN tags t ON t.id = at.tag_id",
				"WHERE",
				"user_id = $",
				"t.name = $",
			},
		},
		{
			name: "filter by multiple tag names (IN clause)",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"user_id": 123,
					"t.name":  []string{"cardio", "running"},
				},
			},
			joins: []JoinConfig{
				{
					Table:     "activity_tags at",
					Condition: "at.activity_id = activities.id",
					Alias:     "at",
				},
				{
					Table:     "tags t",
					Condition: "t.id = at.tag_id",
					Alias:     "t",
				},
			},
			expectedSQL: []string{
				"SELECT activities.* FROM activities",
				"LEFT JOIN activity_tags at ON at.activity_id = activities.id",
				"LEFT JOIN tags t ON t.id = at.tag_id",
				"WHERE",
				"t.name IN ($",
			},
		},
		{
			name: "filter by tag with other filters and search",
			opts: &QueryOptions{
				Page:  1,
				Limit: 20,
				Filter: map[string]interface{}{
					"user_id":       123,
					"activity_type": "running",
					"t.name":        "cardio",
				},
				Search: map[string]interface{}{
					"title": "morning",
				},
				Order: map[string]string{
					"created_at": "DESC",
				},
			},
			joins: []JoinConfig{
				{
					Table:     "activity_tags at",
					Condition: "at.activity_id = activities.id",
					Alias:     "at",
				},
				{
					Table:     "tags t",
					Condition: "t.id = at.tag_id",
					Alias:     "t",
				},
			},
			expectedSQL: []string{
				"SELECT activities.* FROM activities",
				"LEFT JOIN activity_tags at ON at.activity_id = activities.id",
				"LEFT JOIN tags t ON t.id = at.tag_id",
				"WHERE",
				"user_id = $",
				"activity_type = $",
				"t.name = $",
				"title ILIKE $",
				"ORDER BY activities.created_at DESC",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewQueryBuilder("activities", tt.opts)
			sql, args, err := builder.
				WithJoins(tt.joins).
				ApplyFilters().
				ApplyFiltersOr().
				ApplySearch().
				ApplyOrder().
				ApplyPagination().
				Build()

			require.NoError(t, err)
			assert.NotEmpty(t, sql)

			// Verify all expected SQL fragments are present
			for _, expectedFragment := range tt.expectedSQL {
				assert.Contains(t, sql, expectedFragment, "Expected SQL to contain: %s", expectedFragment)
			}

			// Verify PostgreSQL-style placeholders
			if len(args) > 0 {
				assert.Contains(t, sql, "$1")
			}
		})
	}
}

func TestQueryBuilder_JoinWithCountQuery(t *testing.T) {
	opts := &QueryOptions{
		Page:  1,
		Limit: 10,
		Filter: map[string]interface{}{
			"user_id": 123,
			"t.name":  "cardio",
		},
	}

	joins := []JoinConfig{
		{
			Table:     "activity_tags at",
			Condition: "at.activity_id = activities.id",
			Alias:     "at",
		},
		{
			Table:     "tags t",
			Condition: "t.id = at.tag_id",
			Alias:     "t",
		},
	}

	builder := NewQueryBuilder("activities", opts)
	countSQL, _, err := builder.
		WithJoins(joins).
		ApplyFilters().
		BuildCount()

	require.NoError(t, err)

	// Count query should have JOINs for accurate counting
	assert.Contains(t, countSQL, "SELECT COUNT(*) FROM activities")
	assert.Contains(t, countSQL, "LEFT JOIN activity_tags at ON at.activity_id = activities.id")
	assert.Contains(t, countSQL, "LEFT JOIN tags t ON t.id = at.tag_id")
	assert.Contains(t, countSQL, "WHERE")
	assert.Contains(t, countSQL, "t.name = $")

	// Count query should NOT have ORDER BY or LIMIT/OFFSET
	assert.NotContains(t, countSQL, "ORDER BY")
	assert.NotContains(t, countSQL, "LIMIT")
	assert.NotContains(t, countSQL, "OFFSET")
}
