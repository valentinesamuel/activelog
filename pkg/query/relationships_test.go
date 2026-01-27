package query

import (
	"testing"
)

func TestRelationshipRegistry_ManyToMany(t *testing.T) {
	// Setup: activities <-> tags (many-to-many)
	registry := NewRelationshipRegistry("activities")
	registry.Register(ManyToManyRelationship(
		"tags",
		"tags",
		"activity_tags",
		"activity_id",
		"tag_id",
	))

	// Test: Filter by tags.name should generate 2 JOINs
	opts := &QueryOptions{
		Filter: map[string]interface{}{
			"tags.name": "cardio",
		},
	}

	joins := registry.GenerateJoins(opts)

	// Verify 2 JOINs (junction + target)
	if len(joins) != 2 {
		t.Fatalf("Expected 2 joins for many-to-many, got %d", len(joins))
	}

	// Verify junction table JOIN
	if joins[0].Table != "activity_tags" {
		t.Errorf("Expected junction table 'activity_tags', got '%s'", joins[0].Table)
	}
	expectedJunctionCondition := "activity_tags.activity_id = activities.id"
	if joins[0].Condition != expectedJunctionCondition {
		t.Errorf("Expected junction condition '%s', got '%s'", expectedJunctionCondition, joins[0].Condition)
	}

	// Verify target table JOIN
	if joins[1].Table != "tags" {
		t.Errorf("Expected target table 'tags', got '%s'", joins[1].Table)
	}
	expectedTargetCondition := "tags.id = activity_tags.tag_id"
	if joins[1].Condition != expectedTargetCondition {
		t.Errorf("Expected target condition '%s', got '%s'", expectedTargetCondition, joins[1].Condition)
	}
}

func TestRelationshipRegistry_ManyToOne(t *testing.T) {
	// Setup: activities -> users (many-to-one)
	registry := NewRelationshipRegistry("activities")
	registry.Register(ManyToOneRelationship(
		"user",
		"users",
		"user_id",
	))

	// Test: Filter by user.username should generate 1 JOIN
	opts := &QueryOptions{
		Filter: map[string]interface{}{
			"user.username": "john",
		},
	}

	joins := registry.GenerateJoins(opts)

	// Verify 1 JOIN
	if len(joins) != 1 {
		t.Fatalf("Expected 1 join for many-to-one, got %d", len(joins))
	}

	// Verify JOIN details
	if joins[0].Table != "users" {
		t.Errorf("Expected table 'users', got '%s'", joins[0].Table)
	}
	expectedCondition := "users.id = activities.user_id"
	if joins[0].Condition != expectedCondition {
		t.Errorf("Expected condition '%s', got '%s'", expectedCondition, joins[0].Condition)
	}
}

func TestRelationshipRegistry_MultipleRelationships(t *testing.T) {
	// Setup: Multiple relationships
	registry := NewRelationshipRegistry("activities")
	registry.Register(ManyToManyRelationship(
		"tags", "tags", "activity_tags", "activity_id", "tag_id",
	))
	registry.Register(ManyToOneRelationship(
		"user", "users", "user_id",
	))

	// Test: Query using both relationships
	opts := &QueryOptions{
		Filter: map[string]interface{}{
			"tags.name":     "cardio",
			"user.username": "john",
		},
	}

	joins := registry.GenerateJoins(opts)

	// Verify 3 JOINs (2 for tags + 1 for user)
	if len(joins) != 3 {
		t.Fatalf("Expected 3 joins, got %d", len(joins))
	}

	// Should have activity_tags, tags, and users
	tables := make(map[string]bool)
	for _, join := range joins {
		tables[join.Table] = true
	}

	if !tables["activity_tags"] {
		t.Error("Missing JOIN for activity_tags")
	}
	if !tables["tags"] {
		t.Error("Missing JOIN for tags")
	}
	if !tables["users"] {
		t.Error("Missing JOIN for users")
	}
}

func TestRelationshipRegistry_SearchAndOrder(t *testing.T) {
	// Setup
	registry := NewRelationshipRegistry("activities")
	registry.Register(ManyToManyRelationship(
		"tags", "tags", "activity_tags", "activity_id", "tag_id",
	))

	// Test: Search and Order should also trigger JOINs
	opts := &QueryOptions{
		Search: map[string]interface{}{
			"tags.name": "run",
		},
		Order: map[string]string{
			"tags.name": "ASC",
		},
	}

	joins := registry.GenerateJoins(opts)

	// Should generate JOINs even for search/order
	if len(joins) != 2 {
		t.Fatalf("Expected 2 joins for search/order, got %d", len(joins))
	}
}

func TestRelationshipRegistry_FilterConditions_v1_1(t *testing.T) {
	// Setup
	registry := NewRelationshipRegistry("activities")
	registry.Register(ManyToManyRelationship(
		"tags", "tags", "activity_tags", "activity_id", "tag_id",
	))

	// Test: v1.1.0 FilterConditions should also work
	opts := &QueryOptions{
		FilterConditions: []FilterCondition{
			{Column: "tags.name", Operator: "ne", Value: "yoga"},
		},
	}

	joins := registry.GenerateJoins(opts)

	// Should generate JOINs for operator-based filters
	if len(joins) != 2 {
		t.Fatalf("Expected 2 joins for FilterConditions, got %d", len(joins))
	}
}

func TestRelationshipRegistry_NoJoinsNeeded(t *testing.T) {
	// Setup
	registry := NewRelationshipRegistry("activities")
	registry.Register(ManyToManyRelationship(
		"tags", "tags", "activity_tags", "activity_id", "tag_id",
	))

	// Test: Query without relationship columns
	opts := &QueryOptions{
		Filter: map[string]interface{}{
			"activity_type": "running",
		},
		Order: map[string]string{
			"created_at": "DESC",
		},
	}

	joins := registry.GenerateJoins(opts)

	// Should NOT generate any JOINs
	if len(joins) != 0 {
		t.Fatalf("Expected 0 joins when no relationship columns used, got %d", len(joins))
	}
}

func TestRelationshipRegistry_ExtractRelationship(t *testing.T) {
	registry := NewRelationshipRegistry("activities")
	registry.Register(ManyToManyRelationship(
		"tags", "tags", "activity_tags", "activity_id", "tag_id",
	))

	testCases := []struct {
		column   string
		expected string
		found    bool
	}{
		{"tags.name", "tags", true},
		{"tags.id", "tags", true},
		{"user.username", "", false}, // Not registered
		{"created_at", "", false},    // No dot
		{"activity_type", "", false}, // No dot
	}

	for _, tc := range testCases {
		rel, found := registry.extractRelationship(tc.column)
		if found != tc.found {
			t.Errorf("Column '%s': expected found=%v, got %v", tc.column, tc.found, found)
		}
		if rel != tc.expected {
			t.Errorf("Column '%s': expected rel='%s', got '%s'", tc.column, tc.expected, rel)
		}
	}
}

func TestRelationshipRegistry_ValidateColumn(t *testing.T) {
	registry := NewRelationshipRegistry("activities")
	registry.Register(ManyToManyRelationship(
		"tags", "tags", "activity_tags", "activity_id", "tag_id",
	))

	allowedColumns := []string{"activity_type", "distance_km"}

	testCases := []struct {
		column      string
		shouldError bool
	}{
		{"activity_type", false}, // In whitelist
		{"distance_km", false},   // In whitelist
		{"tags.name", false},     // Valid relationship
		{"tags.id", false},       // Valid relationship
		{"user.username", true},  // Unknown relationship
		{"password_hash", true},  // Not in whitelist
		{"unknown.column", true}, // Unknown relationship
	}

	for _, tc := range testCases {
		err := registry.ValidateColumn(tc.column, allowedColumns)
		hasError := err != nil
		if hasError != tc.shouldError {
			t.Errorf("Column '%s': expected error=%v, got error=%v (err: %v)",
				tc.column, tc.shouldError, hasError, err)
		}
	}
}

func TestRelationshipRegistry_OneToMany(t *testing.T) {
	// Setup: users -> activities (one-to-many)
	// This would be in UserRepository
	registry := NewRelationshipRegistry("users")
	registry.Register(OneToManyRelationship(
		"activities",
		"activities",
		"user_id",
	))

	// Test: Filter by activities.activity_type
	opts := &QueryOptions{
		Filter: map[string]interface{}{
			"activities.activity_type": "running",
		},
	}

	joins := registry.GenerateJoins(opts)

	// Verify 1 JOIN
	if len(joins) != 1 {
		t.Fatalf("Expected 1 join for one-to-many, got %d", len(joins))
	}

	// Verify JOIN details
	if joins[0].Table != "activities" {
		t.Errorf("Expected table 'activities', got '%s'", joins[0].Table)
	}
	expectedCondition := "activities.user_id = users.id"
	if joins[0].Condition != expectedCondition {
		t.Errorf("Expected condition '%s', got '%s'", expectedCondition, joins[0].Condition)
	}
}

func TestRelationshipRegistry_DuplicateJoinsAvoided(t *testing.T) {
	// Setup
	registry := NewRelationshipRegistry("activities")
	registry.Register(ManyToManyRelationship(
		"tags", "tags", "activity_tags", "activity_id", "tag_id",
	))

	// Test: Using tags in both filter and search
	opts := &QueryOptions{
		Filter: map[string]interface{}{
			"tags.name": "cardio",
		},
		Search: map[string]interface{}{
			"tags.name": "run",
		},
		Order: map[string]string{
			"tags.name": "ASC",
		},
	}

	joins := registry.GenerateJoins(opts)

	// Should only generate JOINs once (not 3 times)
	if len(joins) != 2 {
		t.Fatalf("Expected 2 joins (no duplicates), got %d", len(joins))
	}
}
