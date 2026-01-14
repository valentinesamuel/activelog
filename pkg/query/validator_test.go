package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateQueryOptions(t *testing.T) {
	allowedFilters := []string{"status", "activity_type", "user_id"}
	allowedSearch := []string{"title", "description"}
	allowedOrder := []string{"created_at", "updated_at", "activity_date"}

	tests := []struct {
		name    string
		opts    *QueryOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid options",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"status": "active",
				},
				Search: map[string]interface{}{
					"title": "morning",
				},
				Order: map[string]string{
					"created_at": "DESC",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid filter column",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"email": "user@example.com", // Not in allowedFilters
				},
			},
			wantErr: true,
			errMsg:  "filtering on column 'email' is not allowed",
		},
		{
			name: "invalid filterOr column",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				FilterOr: map[string]interface{}{
					"password": "secret", // Not in allowedFilters
				},
			},
			wantErr: true,
			errMsg:  "filtering on column 'password' is not allowed",
		},
		{
			name: "invalid search column",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Search: map[string]interface{}{
					"email": "user", // Not in allowedSearch
				},
			},
			wantErr: true,
			errMsg:  "searching on column 'email' is not allowed",
		},
		{
			name: "invalid order column",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Order: map[string]string{
					"password": "ASC", // Not in allowedOrder
				},
			},
			wantErr: true,
			errMsg:  "ordering by column 'password' is not allowed",
		},
		{
			name: "invalid order direction",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Order: map[string]string{
					"created_at": "RANDOM", // Not ASC or DESC
				},
			},
			wantErr: true,
			errMsg:  "invalid order direction",
		},
		{
			name: "page less than 1",
			opts: &QueryOptions{
				Page:  0,
				Limit: 10,
			},
			wantErr: true,
			errMsg:  "page must be at least 1",
		},
		{
			name: "limit less than 1",
			opts: &QueryOptions{
				Page:  1,
				Limit: 0,
			},
			wantErr: true,
			errMsg:  "limit must be at least 1",
		},
		{
			name: "limit exceeds maximum",
			opts: &QueryOptions{
				Page:  1,
				Limit: 101, // Exceeds MaxPageSize of 100
			},
			wantErr: true,
			errMsg:  "limit cannot exceed 100",
		},
		{
			name: "multiple filters all valid",
			opts: &QueryOptions{
				Page:  1,
				Limit: 50,
				Filter: map[string]interface{}{
					"status":        "active",
					"activity_type": "running",
					"user_id":       123,
				},
			},
			wantErr: false,
		},
		{
			name: "empty options",
			opts: &QueryOptions{
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
			name: "case insensitive column matching",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"STATUS": "active", // Uppercase should match "status"
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQueryOptions(tt.opts, allowedFilters, allowedSearch, allowedOrder)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOrderDirection(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		wantErr   bool
	}{
		{
			name:      "ASC uppercase",
			direction: "ASC",
			wantErr:   false,
		},
		{
			name:      "DESC uppercase",
			direction: "DESC",
			wantErr:   false,
		},
		{
			name:      "asc lowercase",
			direction: "asc",
			wantErr:   false,
		},
		{
			name:      "desc lowercase",
			direction: "desc",
			wantErr:   false,
		},
		{
			name:      "mixed case Asc",
			direction: "Asc",
			wantErr:   false,
		},
		{
			name:      "invalid direction",
			direction: "ASCENDING",
			wantErr:   true,
		},
		{
			name:      "invalid direction RANDOM",
			direction: "RANDOM",
			wantErr:   true,
		},
		{
			name:      "empty string",
			direction: "",
			wantErr:   true,
		},
		{
			name:      "numeric string",
			direction: "1",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrderDirection(tt.direction)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		opts    *QueryOptions
		config  *ValidationConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with default config",
			opts: &QueryOptions{
				Page:   1,
				Limit:  10,
				Filter: map[string]interface{}{"user_id": 123},
				Order:  map[string]string{"created_at": "DESC"},
			},
			config: &ValidationConfig{
				AllowedFilters:   []string{"user_id"},
				AllowedSearch:    []string{},
				AllowedOrder:     []string{"created_at"},
				MaxPageSize:      100,
				RequireUserScope: true,
			},
			wantErr: false,
		},
		{
			name: "exceeds custom max page size",
			opts: &QueryOptions{
				Page:  1,
				Limit: 60,
			},
			config: &ValidationConfig{
				AllowedFilters: []string{},
				AllowedSearch:  []string{},
				AllowedOrder:   []string{},
				MaxPageSize:    50,
			},
			wantErr: true,
			errMsg:  "limit cannot exceed 50",
		},
		{
			name: "missing user_id when required",
			opts: &QueryOptions{
				Page:   1,
				Limit:  10,
				Filter: map[string]interface{}{"status": "active"},
			},
			config: &ValidationConfig{
				AllowedFilters:   []string{"status", "user_id"},
				AllowedSearch:    []string{},
				AllowedOrder:     []string{},
				MaxPageSize:      100,
				RequireUserScope: true,
			},
			wantErr: true,
			errMsg:  "user_id filter is required",
		},
		{
			name: "user_id not required",
			opts: &QueryOptions{
				Page:   1,
				Limit:  10,
				Filter: map[string]interface{}{"status": "active"},
			},
			config: &ValidationConfig{
				AllowedFilters:   []string{"status"},
				AllowedSearch:    []string{},
				AllowedOrder:     []string{},
				MaxPageSize:      100,
				RequireUserScope: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithConfig(tt.opts, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultValidationConfig(t *testing.T) {
	config := DefaultValidationConfig()

	assert.NotNil(t, config)
	assert.NotNil(t, config.AllowedFilters)
	assert.NotNil(t, config.AllowedSearch)
	assert.NotNil(t, config.AllowedOrder)
	assert.Equal(t, 100, config.MaxPageSize)
	assert.True(t, config.RequireUserScope)
	assert.Contains(t, config.AllowedOrder, "created_at")
	assert.Contains(t, config.AllowedOrder, "updated_at")
}

func TestSanitizeSearchTerm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "with percent sign",
			input:    "test%",
			expected: "test\\%",
		},
		{
			name:     "with underscore",
			input:    "hello_world",
			expected: "hello\\_world",
		},
		{
			name:     "with backslash",
			input:    "path\\to\\file",
			expected: "path\\\\to\\\\file",
		},
		{
			name:     "with multiple special chars",
			input:    "test%_data\\",
			expected: "test\\%\\_data\\\\",
		},
		{
			name:     "SQL injection attempt",
			input:    "%'; DROP TABLE users--",
			expected: "\\%'; DROP TABLE users--",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeSearchTerm(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateColumnName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple column",
			input:   "user_id",
			wantErr: false,
		},
		{
			name:    "valid camelCase column",
			input:   "createdAt",
			wantErr: false,
		},
		{
			name:    "valid qualified column",
			input:   "activities.user_id",
			wantErr: false,
		},
		{
			name:    "valid with alias",
			input:   "t.name",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "starts with number",
			input:   "123column",
			wantErr: true,
		},
		{
			name:    "contains spaces",
			input:   "user id",
			wantErr: true,
		},
		{
			name:    "SQL injection attempt",
			input:   "'; DROP TABLE users--",
			wantErr: true,
		},
		{
			name:    "contains semicolon",
			input:   "user_id; DELETE",
			wantErr: true,
		},
		{
			name:    "contains parentheses",
			input:   "COUNT(id)",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "this_is_a_very_long_column_name_that_exceeds_the_postgresql_limit_of_sixty_three_characters",
			wantErr: true,
		},
		{
			name:    "contains dash",
			input:   "user-id",
			wantErr: true,
		},
		{
			name:    "contains special chars",
			input:   "user@id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColumnName(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"status", "activity_type", "user_id"}

	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "exact match",
			slice:    slice,
			item:     "status",
			expected: true,
		},
		{
			name:     "not in slice",
			slice:    slice,
			item:     "email",
			expected: false,
		},
		{
			name:     "case insensitive match",
			slice:    slice,
			item:     "STATUS",
			expected: true,
		},
		{
			name:     "mixed case match",
			slice:    slice,
			item:     "Activity_Type",
			expected: true,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "status",
			expected: false,
		},
		{
			name:     "empty item",
			slice:    slice,
			item:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidation_SecurityScenarios(t *testing.T) {
	allowedFilters := []string{"status", "user_id"}
	allowedSearch := []string{"title"}
	allowedOrder := []string{"created_at"}

	tests := []struct {
		name    string
		opts    *QueryOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "SQL injection attempt in filter",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"'; DROP TABLE users--": "value",
				},
			},
			wantErr: true,
			errMsg:  "filtering on column",
		},
		{
			name: "attempt to access password column",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Filter: map[string]interface{}{
					"password": "secret",
				},
			},
			wantErr: true,
			errMsg:  "filtering on column 'password' is not allowed",
		},
		{
			name: "attempt to search sensitive data",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Search: map[string]interface{}{
					"credit_card": "1234",
				},
			},
			wantErr: true,
			errMsg:  "searching on column 'credit_card' is not allowed",
		},
		{
			name: "attempt to order by password",
			opts: &QueryOptions{
				Page:  1,
				Limit: 10,
				Order: map[string]string{
					"password_hash": "ASC",
				},
			},
			wantErr: true,
			errMsg:  "ordering by column 'password_hash' is not allowed",
		},
		{
			name: "DoS attempt with huge limit",
			opts: &QueryOptions{
				Page:  1,
				Limit: 999999,
			},
			wantErr: true,
			errMsg:  "limit cannot exceed 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQueryOptions(tt.opts, allowedFilters, allowedSearch, allowedOrder)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
