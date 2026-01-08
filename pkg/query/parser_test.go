package query

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		input    url.Values
		expected *QueryOptions
		wantErr  bool
	}{
		{
			name:  "empty query parameters",
			input: url.Values{},
			expected: &QueryOptions{
				Page:     1,
				Limit:    10,
				Filter:   map[string]interface{}{},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "simple pagination only",
			input: url.Values{
				"page":  []string{"2"},
				"limit": []string{"20"},
			},
			expected: &QueryOptions{
				Page:     2,
				Limit:    20,
				Filter:   map[string]interface{}{},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "single filter",
			input: url.Values{
				"filter[status]": []string{"active"},
			},
			expected: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"status": "active",
				},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "multiple filters",
			input: url.Values{
				"filter[status]":        []string{"active"},
				"filter[activity_type]": []string{"running"},
				"filter[user_id]":       []string{"123"},
			},
			expected: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"status":        "active",
					"activity_type": "running",
					"user_id":       123, // Converted to int
				},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "filter with boolean value",
			input: url.Values{
				"filter[is_active]": []string{"true"},
				"filter[is_deleted]": []string{"false"},
			},
			expected: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"is_active":  true,
					"is_deleted": false,
				},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "filter with null value",
			input: url.Values{
				"filter[description]": []string{"null"},
			},
			expected: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"description": nil,
				},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "filter with array value",
			input: url.Values{
				"filter[activity_type]": []string{"[running,cycling,swimming]"},
			},
			expected: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"activity_type": []string{"running", "cycling", "swimming"},
				},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "filter with float value",
			input: url.Values{
				"filter[distance]": []string{"5.5"},
			},
			expected: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"distance": 5.5,
				},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "filterOr conditions",
			input: url.Values{
				"filterOr[type]":   []string{"running"},
				"filterOr[status]": []string{"active"},
			},
			expected: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{},
				FilterOr: map[string]interface{}{
					"type":   "running",
					"status": "active",
				},
				Search: map[string]interface{}{},
				Order:  map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "search parameters",
			input: url.Values{
				"search[title]":       []string{"morning"},
				"search[description]": []string{"run"},
			},
			expected: &QueryOptions{
				Page:     1,
				Limit:    10,
				Filter:   map[string]interface{}{},
				FilterOr: map[string]interface{}{},
				Search: map[string]interface{}{
					"title":       "morning",
					"description": "run",
				},
				Order: map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "order parameters",
			input: url.Values{
				"order[created_at]": []string{"DESC"},
				"order[amount]":     []string{"asc"},
			},
			expected: &QueryOptions{
				Page:     1,
				Limit:    10,
				Filter:   map[string]interface{}{},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order: map[string]string{
					"created_at": "DESC",
					"amount":     "ASC", // Converted to uppercase
				},
			},
			wantErr: false,
		},
		{
			name: "complex query with all parameter types",
			input: url.Values{
				"page":                  []string{"3"},
				"limit":                 []string{"50"},
				"filter[user_id]":       []string{"456"},
				"filter[status]":        []string{"completed"},
				"filterOr[type]":        []string{"running"},
				"search[title]":         []string{"morning"},
				"order[activity_date]":  []string{"DESC"},
			},
			expected: &QueryOptions{
				Page:  3,
				Limit: 50,
				Filter: map[string]interface{}{
					"user_id": 456,
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
			wantErr: false,
		},
		{
			name: "invalid page number defaults to 1",
			input: url.Values{
				"page": []string{"invalid"},
			},
			expected: &QueryOptions{
				Page:     1, // Defaults to 1 on parse error
				Limit:    10,
				Filter:   map[string]interface{}{},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "negative page number defaults to 1",
			input: url.Values{
				"page": []string{"-5"},
			},
			expected: &QueryOptions{
				Page:     1, // Negative not allowed, defaults to 1
				Limit:    10,
				Filter:   map[string]interface{}{},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "zero limit defaults to 10",
			input: url.Values{
				"limit": []string{"0"},
			},
			expected: &QueryOptions{
				Page:     1,
				Limit:    10, // Zero not allowed, defaults to 10
				Filter:   map[string]interface{}{},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "empty array value",
			input: url.Values{
				"filter[tags]": []string{"[]"},
			},
			expected: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"tags": []string{},
				},
				FilterOr: map[string]interface{}{},
				Search:   map[string]interface{}{},
				Order:    map[string]string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseQueryParams(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Page, result.Page, "Page mismatch")
			assert.Equal(t, tt.expected.Limit, result.Limit, "Limit mismatch")
			assert.Equal(t, tt.expected.Filter, result.Filter, "Filter mismatch")
			assert.Equal(t, tt.expected.FilterOr, result.FilterOr, "FilterOr mismatch")
			assert.Equal(t, tt.expected.Search, result.Search, "Search mismatch")
			assert.Equal(t, tt.expected.Order, result.Order, "Order mismatch")
		})
	}
}

func TestParseNestedParam(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedPrefix string
		expectedColumn string
	}{
		{
			name:           "simple filter",
			input:          "filter[status]",
			expectedPrefix: "filter",
			expectedColumn: "status",
		},
		{
			name:           "simple search",
			input:          "search[title]",
			expectedPrefix: "search",
			expectedColumn: "title",
		},
		{
			name:           "simple order",
			input:          "order[created_at]",
			expectedPrefix: "order",
			expectedColumn: "created_at",
		},
		{
			name:           "filterOr",
			input:          "filterOr[type]",
			expectedPrefix: "filterOr",
			expectedColumn: "type",
		},
		{
			name:           "nested notation - two levels",
			input:          "filter[tags][name]",
			expectedPrefix: "filter",
			expectedColumn: "tags.name",
		},
		{
			name:           "nested notation - profile",
			input:          "search[profile][firstname]",
			expectedPrefix: "search",
			expectedColumn: "profile.firstname",
		},
		{
			name:           "invalid format - no brackets",
			input:          "filter",
			expectedPrefix: "",
			expectedColumn: "",
		},
		{
			name:           "invalid format - missing closing bracket",
			input:          "filter[status",
			expectedPrefix: "",
			expectedColumn: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, column := parseNestedParam(tt.input)
			assert.Equal(t, tt.expectedPrefix, prefix, "Prefix mismatch")
			assert.Equal(t, tt.expectedColumn, column, "Column mismatch")
		})
	}
}

func TestConvertValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "boolean true",
			input:    "true",
			expected: true,
		},
		{
			name:     "boolean false",
			input:    "false",
			expected: false,
		},
		{
			name:     "null value",
			input:    "null",
			expected: nil,
		},
		{
			name:     "integer",
			input:    "123",
			expected: 123,
		},
		{
			name:     "negative integer",
			input:    "-456",
			expected: -456,
		},
		{
			name:     "float",
			input:    "123.45",
			expected: 123.45,
		},
		{
			name:     "negative float",
			input:    "-67.89",
			expected: -67.89,
		},
		{
			name:     "array with multiple values",
			input:    "[running,cycling,swimming]",
			expected: []string{"running", "cycling", "swimming"},
		},
		{
			name:     "array with single value",
			input:    "[running]",
			expected: []string{"running"},
		},
		{
			name:     "empty array",
			input:    "[]",
			expected: []string{},
		},
		{
			name:     "array with spaces",
			input:    "[running, cycling, swimming]",
			expected: []string{"running", "cycling", "swimming"},
		},
		{
			name:     "simple string",
			input:    "active",
			expected: "active",
		},
		{
			name:     "string with spaces",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "string that looks like number but isn't",
			input:    "123abc",
			expected: "123abc",
		},
		{
			name:     "email address",
			input:    "user@example.com",
			expected: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseArrayValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "multiple values",
			input:    "running,cycling,swimming",
			expected: []string{"running", "cycling", "swimming"},
		},
		{
			name:     "single value",
			input:    "running",
			expected: []string{"running"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "values with spaces",
			input:    "running, cycling, swimming",
			expected: []string{"running", "cycling", "swimming"},
		},
		{
			name:     "values with extra spaces",
			input:    "  running  ,  cycling  ,  swimming  ",
			expected: []string{"running", "cycling", "swimming"},
		},
		{
			name:     "empty values filtered out",
			input:    "running,,cycling",
			expected: []string{"running", "cycling"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseArrayValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeColumnName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "camelCase to snake_case",
			input:    "activityType",
			expected: "activity_type",
		},
		{
			name:     "PascalCase to snake_case",
			input:    "ActivityType",
			expected: "activity_type",
		},
		{
			name:     "multiple words",
			input:    "createdAt",
			expected: "created_at",
		},
		{
			name:     "three words",
			input:    "userIdNumber",
			expected: "user_id_number",
		},
		{
			name:     "already snake_case",
			input:    "activity_type",
			expected: "activity_type",
		},
		{
			name:     "single word lowercase",
			input:    "status",
			expected: "status",
		},
		{
			name:     "single word uppercase",
			input:    "STATUS",
			expected: "s_t_a_t_u_s", // Each uppercase letter gets underscore
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeColumnName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewQueryOptions(t *testing.T) {
	opts := NewQueryOptions()

	assert.Equal(t, 1, opts.Page, "Default page should be 1")
	assert.Equal(t, 10, opts.Limit, "Default limit should be 10")
	assert.NotNil(t, opts.Filter, "Filter map should be initialized")
	assert.NotNil(t, opts.FilterConditions, "FilterConditions should be initialized")
	assert.NotNil(t, opts.FilterOr, "FilterOr map should be initialized")
	assert.NotNil(t, opts.Search, "Search map should be initialized")
	assert.NotNil(t, opts.Order, "Order map should be initialized")
	assert.Equal(t, 0, len(opts.Filter), "Filter map should be empty")
	assert.Equal(t, 0, len(opts.FilterConditions), "FilterConditions should be empty")
	assert.Equal(t, 0, len(opts.FilterOr), "FilterOr map should be empty")
	assert.Equal(t, 0, len(opts.Search), "Search map should be empty")
	assert.Equal(t, 0, len(opts.Order), "Order map should be empty")
}

// ==================== NEW TESTS FOR OPERATOR-BASED FILTERING (v1.1.0) ====================

func TestExtractBracketLevels(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple filter - 2 levels",
			input:    "filter[status]",
			expected: []string{"filter", "status"},
		},
		{
			name:     "Operator filter - 3 levels (gte)",
			input:    "filter[created_at][gte]",
			expected: []string{"filter", "created_at", "gte"},
		},
		{
			name:     "Operator filter - 3 levels (lt)",
			input:    "filter[distance][lt]",
			expected: []string{"filter", "distance", "lt"},
		},
		{
			name:     "Operator filter - 3 levels (eq)",
			input:    "filter[status][eq]",
			expected: []string{"filter", "status", "eq"},
		},
		{
			name:     "Order parameter - 2 levels",
			input:    "order[created_at]",
			expected: []string{"order", "created_at"},
		},
		{
			name:     "Search parameter - 2 levels",
			input:    "search[title]",
			expected: []string{"search", "title"},
		},
		{
			name:     "No brackets",
			input:    "page",
			expected: []string{"page"},
		},
		{
			name:     "Empty brackets",
			input:    "filter[]",
			expected: []string{"filter"},
		},
		{
			name:     "Unclosed bracket",
			input:    "filter[status",
			expected: []string{},
		},
		{
			name:     "Multiple levels - 4 levels",
			input:    "filter[user][address][city]",
			expected: []string{"filter", "user", "address", "city"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBracketLevels(tt.input)
			assert.Equal(t, tt.expected, result, "Bracket levels mismatch")
		})
	}
}

func TestParseQueryParams_OperatorSyntax(t *testing.T) {
	tests := []struct {
		name                    string
		input                   url.Values
		expectedFilterCount     int
		expectedConditionsCount int
		validateConditions      func(t *testing.T, conditions []FilterCondition)
	}{
		{
			name: "GTE operator on created_at",
			input: url.Values{
				"filter[created_at][gte]": []string{"2024-01-01"},
			},
			expectedFilterCount:     0, // Not an eq, so Filter map should be empty
			expectedConditionsCount: 1,
			validateConditions: func(t *testing.T, conditions []FilterCondition) {
				require.Len(t, conditions, 1)
				assert.Equal(t, "created_at", conditions[0].Column)
				assert.Equal(t, "gte", conditions[0].Operator)
				assert.Equal(t, "2024-01-01", conditions[0].Value)
			},
		},
		{
			name: "LT operator on distance",
			input: url.Values{
				"filter[distance][lt]": []string{"10"},
			},
			expectedFilterCount:     0,
			expectedConditionsCount: 1,
			validateConditions: func(t *testing.T, conditions []FilterCondition) {
				require.Len(t, conditions, 1)
				assert.Equal(t, "distance", conditions[0].Column)
				assert.Equal(t, "lt", conditions[0].Operator)
				assert.Equal(t, 10, conditions[0].Value) // Parser converts to int
			},
		},
		{
			name: "EQ operator (explicit)",
			input: url.Values{
				"filter[status][eq]": []string{"active"},
			},
			expectedFilterCount:     1, // eq should also populate legacy Filter
			expectedConditionsCount: 1,
			validateConditions: func(t *testing.T, conditions []FilterCondition) {
				require.Len(t, conditions, 1)
				assert.Equal(t, "eq", conditions[0].Operator)
			},
		},
		{
			name: "Multiple operator filters",
			input: url.Values{
				"filter[created_at][gte]": []string{"2024-01-01"},
				"filter[distance][lt]":    []string{"10"},
				"filter[status][eq]":      []string{"active"},
			},
			expectedFilterCount:     1, // Only status[eq] should be in Filter
			expectedConditionsCount: 3,
			validateConditions: func(t *testing.T, conditions []FilterCondition) {
				require.Len(t, conditions, 3)
				// Verify all three conditions exist
				operators := make(map[string]string)
				for _, cond := range conditions {
					operators[cond.Column] = cond.Operator
				}
				assert.Equal(t, "gte", operators["created_at"])
				assert.Equal(t, "lt", operators["distance"])
				assert.Equal(t, "eq", operators["status"])
			},
		},
		{
			name: "NE operator",
			input: url.Values{
				"filter[amount][ne]": []string{"0"},
			},
			expectedFilterCount:     0,
			expectedConditionsCount: 1,
			validateConditions: func(t *testing.T, conditions []FilterCondition) {
				require.Len(t, conditions, 1)
				assert.Equal(t, "ne", conditions[0].Operator)
			},
		},
		{
			name: "GT and LTE operators on same column",
			input: url.Values{
				"filter[price][gt]":  []string{"100"},
				"filter[price][lte]": []string{"500"},
			},
			expectedFilterCount:     0,
			expectedConditionsCount: 2,
			validateConditions: func(t *testing.T, conditions []FilterCondition) {
				require.Len(t, conditions, 2)
				// Both conditions should be on 'price' column
				for _, cond := range conditions {
					assert.Equal(t, "price", cond.Column)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseQueryParams(tt.input)
			require.NoError(t, err)

			assert.Len(t, opts.Filter, tt.expectedFilterCount, "Filter map count mismatch")
			assert.Len(t, opts.FilterConditions, tt.expectedConditionsCount, "FilterConditions count mismatch")

			if tt.validateConditions != nil {
				tt.validateConditions(t, opts.FilterConditions)
			}
		})
	}
}

func TestParseQueryParams_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name                    string
		input                   url.Values
		expectedFilterCount     int
		expectedConditionsCount int
		validateResult          func(t *testing.T, opts *QueryOptions)
	}{
		{
			name: "Legacy 2-level syntax (filter[status]=active)",
			input: url.Values{
				"filter[status]": []string{"active"},
			},
			expectedFilterCount:     1,
			expectedConditionsCount: 1,
			validateResult: func(t *testing.T, opts *QueryOptions) {
				// Should be in both Filter and FilterConditions
				assert.Equal(t, "active", opts.Filter["status"], "Legacy Filter map should contain status=active")
				require.Len(t, opts.FilterConditions, 1)
				assert.Equal(t, "eq", opts.FilterConditions[0].Operator, "Legacy syntax should default to 'eq' operator")
			},
		},
		{
			name: "Multiple legacy filters",
			input: url.Values{
				"filter[status]": []string{"active"},
				"filter[type]":   []string{"running"},
			},
			expectedFilterCount:     2,
			expectedConditionsCount: 2,
			validateResult: func(t *testing.T, opts *QueryOptions) {
				assert.Len(t, opts.Filter, 2)
				assert.Len(t, opts.FilterConditions, 2)
				// All should have 'eq' operator
				for _, cond := range opts.FilterConditions {
					assert.Equal(t, "eq", cond.Operator)
				}
			},
		},
		{
			name: "Mixed legacy and operator syntax",
			input: url.Values{
				"filter[status]":          []string{"active"},
				"filter[created_at][gte]": []string{"2024-01-01"},
			},
			expectedFilterCount:     1, // Only legacy filter[status]
			expectedConditionsCount: 2, // Both status and created_at
			validateResult: func(t *testing.T, opts *QueryOptions) {
				// status should be in Filter (legacy)
				assert.Equal(t, "active", opts.Filter["status"])
				// created_at should NOT be in Filter (it's not eq)
				_, exists := opts.Filter["created_at"]
				assert.False(t, exists, "created_at should not be in Filter map (not eq operator)")
				// Both should be in FilterConditions
				require.Len(t, opts.FilterConditions, 2)
			},
		},
		{
			name: "Complex mix with pagination",
			input: url.Values{
				"page":                    []string{"2"},
				"limit":                   []string{"20"},
				"filter[user_id]":         []string{"123"},
				"filter[created_at][gte]": []string{"2024-01-01"},
				"filter[distance][lt]":    []string{"10.5"},
			},
			expectedFilterCount:     1, // Only user_id (legacy)
			expectedConditionsCount: 3, // user_id, created_at, distance
			validateResult: func(t *testing.T, opts *QueryOptions) {
				assert.Equal(t, 2, opts.Page)
				assert.Equal(t, 20, opts.Limit)
				assert.Equal(t, 123, opts.Filter["user_id"])
				require.Len(t, opts.FilterConditions, 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := ParseQueryParams(tt.input)
			require.NoError(t, err)

			if tt.validateResult != nil {
				tt.validateResult(t, opts)
			}
		})
	}
}
