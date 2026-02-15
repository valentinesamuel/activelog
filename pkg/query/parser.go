package query

import (
	"net/url"
	"strconv"
	"strings"
)

// ParseQueryParams parses HTTP query parameters into a QueryOptions struct.
// Handles nested parameters like filter[status]=active, order[createdAt]=DESC.
//
// Supported parameter formats:
//   - page=1, limit=10 → Simple pagination
//   - filter[columnName]=value → AND conditions (equality)
//   - filter[columnName][operator]=value → Operator-based filtering (NEW in v1.1.0)
//   - filterOr[columnName]=value → OR conditions
//   - search[columnName]=term → ILIKE pattern matching
//   - order[columnName]=ASC|DESC → Sorting
//
// Operator-based filtering examples:
//   - filter[created_at][gte]=2024-01-01 → WHERE created_at >= '2024-01-01'
//   - filter[distance][lt]=10 → WHERE distance < 10
//   - filter[status][eq]=active → WHERE status = 'active'
//
// Example URL (legacy):
//
//	/activities?page=2&limit=20&filter[activityType]=running&filter[status]=completed&search[title]=morning&order[createdAt]=DESC
//
// Example URL (with operators):
//
//	/activities?page=1&limit=20&filter[created_at][gte]=2024-01-01&filter[distance][lt]=10&order[created_at]=DESC
//
// Returns QueryOptions with all parameters parsed and typed correctly.
func ParseQueryParams(values url.Values) (*QueryOptions, error) {
	opts := &QueryOptions{
		Page:             1,  // Default page
		Limit:            10, // Default limit
		Filter:           make(map[string]interface{}),
		FilterConditions: []FilterCondition{},
		FilterOr:         make(map[string]interface{}),
		Search:           make(map[string]interface{}),
		Order:            make(map[string]string),
	}

	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}

		// Handle simple params: page, limit
		switch key {
		case "page":
			if p, err := strconv.Atoi(vals[0]); err == nil && p > 0 {
				opts.Page = p
			}
		case "limit":
			if l, err := strconv.Atoi(vals[0]); err == nil && l > 0 {
				opts.Limit = l
			}
		default:
			// Handle nested params: filter[status], order[createdAt], filter[date][gte]
			if strings.Contains(key, "[") && strings.Contains(key, "]") {
				levels := extractBracketLevels(key)

				// Detect operator-based filtering (3+ levels)
				if len(levels) == 3 && levels[0] == "filter" {
					// Operator-based: filter[column][operator]=value
					column := levels[1]
					operator := levels[2]
					value := convertValue(vals[0])

					// Add to FilterConditions
					opts.FilterConditions = append(opts.FilterConditions, FilterCondition{
						Column:   column,
						Operator: operator,
						Value:    value,
					})

					// Also add to legacy Filter map for backward compatibility (as equality)
					// This ensures existing code that only checks Filter still works
					if operator == "eq" {
						opts.Filter[column] = value
					}

				} else if len(levels) == 2 {
					// Legacy 2-level syntax: filter[column]=value
					prefix, column := levels[0], levels[1]
					value := convertValue(vals[0])

					switch prefix {
					case "filter":
						// Add to both Filter (legacy) and FilterConditions (as eq operator)
						opts.Filter[column] = value
						opts.FilterConditions = append(opts.FilterConditions, FilterCondition{
							Column:   column,
							Operator: "eq",
							Value:    value,
						})
					case "filterOr":
						opts.FilterOr[column] = value
					case "search":
						opts.Search[column] = value
					case "order":
						// Order values should stay as strings (ASC/DESC)
						opts.Order[column] = strings.ToUpper(vals[0])
					}
				}
			}
		}
	}

	return opts, nil
}

// extractBracketLevels extracts all bracket-enclosed values from a parameter key.
// Handles multi-level bracket notation for operator-based filtering.
//
// Examples:
//   - "filter[status]" → ["filter", "status"]
//   - "filter[created_at][gte]" → ["filter", "created_at", "gte"]
//   - "filter[distance][lt]" → ["filter", "distance", "lt"]
//   - "order[createdAt]" → ["order", "createdAt"]
//
// Returns an empty slice if the format is invalid.
func extractBracketLevels(key string) []string {
	var levels []string
	remaining := key

	// Extract the prefix (before first bracket)
	firstBracket := strings.Index(remaining, "[")
	if firstBracket == -1 {
		// No brackets, return the whole key as single level
		return []string{key}
	}

	// Add prefix
	prefix := remaining[:firstBracket]
	if prefix != "" {
		levels = append(levels, prefix)
	}
	remaining = remaining[firstBracket:]

	// Iteratively extract all bracket pairs
	for len(remaining) > 0 {
		if !strings.HasPrefix(remaining, "[") {
			break
		}

		closeBracket := strings.Index(remaining, "]")
		if closeBracket == -1 {
			// Malformed: opening bracket without closing
			return []string{}
		}

		// Extract content between brackets
		content := remaining[1:closeBracket]
		if content != "" {
			levels = append(levels, content)
		}

		// Move to next bracket pair
		remaining = remaining[closeBracket+1:]
	}

	return levels
}

// parseNestedParam extracts prefix and column from nested parameter notation.
// LEGACY FUNCTION - Kept for backward compatibility with 2-level parsing.
// For new operator-based filtering, use extractBracketLevels() instead.
//
// Examples:
//   - "filter[status]" → ("filter", "status")
//   - "order[createdAt]" → ("order", "createdAt")
//   - "search[title]" → ("search", "title")
//   - "filter[tags][name]" → ("filter", "tags.name") // Future: nested notation
//
// Returns empty strings if the format is invalid.
func parseNestedParam(key string) (prefix, column string) {
	startIdx := strings.Index(key, "[")
	endIdx := strings.Index(key, "]")

	if startIdx == -1 || endIdx == -1 {
		return "", ""
	}

	prefix = key[:startIdx]
	column = key[startIdx+1 : endIdx]

	// Handle nested notation: filter[tags][name] → "tags.name"
	// This extracts everything after the first ] and checks for more brackets
	remaining := key[endIdx+1:]
	if strings.HasPrefix(remaining, "[") && strings.Contains(remaining, "]") {
		nestedEndIdx := strings.Index(remaining, "]")
		if nestedEndIdx != -1 {
			nestedColumn := remaining[1:nestedEndIdx]
			column = column + "." + nestedColumn
		}
	}

	return prefix, column
}

// convertValue converts a string value to its appropriate type.
//
// Conversion rules:
//   - "true" → bool(true)
//   - "false" → bool(false)
//   - "null" → nil
//   - "123" → int(123)
//   - "123.45" → float64(123.45)
//   - "[val1,val2]" → []string{"val1", "val2"}
//   - "other" → string("other")
//
// This is critical for proper SQL query generation, as different types
// need different handling (e.g., IN clauses for arrays).
func convertValue(val string) interface{} {
	// Trim whitespace
	val = strings.TrimSpace(val)

	// Boolean
	if val == "true" {
		return true
	}
	if val == "false" {
		return false
	}

	// Null
	if val == "null" {
		return nil
	}

	// Array: [value1,value2,value3]
	if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
		val = strings.Trim(val, "[]")
		if val == "" {
			return []string{}
		}

		// Split by comma and trim each element
		parts := strings.Split(val, ",")
		result := make([]string, len(parts))
		for i, part := range parts {
			result[i] = strings.TrimSpace(part)
		}
		return result
	}

	// Number (integer)
	if num, err := strconv.Atoi(val); err == nil {
		return num
	}

	// Number (float)
	if num, err := strconv.ParseFloat(val, 64); err == nil {
		return num
	}

	// Default: string
	return val
}

// ParseArrayValue parses comma-separated values into a string slice.
// Helper function for explicit array parsing.
//
// Example: "running,cycling,swimming" → []string{"running", "cycling", "swimming"}
func ParseArrayValue(val string) []string {
	if val == "" {
		return []string{}
	}

	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// NormalizeColumnName converts user-friendly column names to database column names.
// Handles common cases like camelCase → snake_case conversion.
//
// Examples:
//   - "activityType" → "activity_type"
//   - "createdAt" → "created_at"
//   - "userId" → "user_id"
//
// Note: This is optional and can be skipped if your API uses snake_case throughout.
func NormalizeColumnName(name string) string {
	var result []rune

	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}

	return strings.ToLower(string(result))
}
