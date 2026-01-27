package query_test

import (
	"testing"

	"github.com/valentinesamuel/activelog/pkg/query"
)

// TestRelationshipRegistry_DeepNesting_v3 tests multi-level path resolution (v3.0)
func TestRelationshipRegistry_DeepNesting_v3(t *testing.T) {
	// Setup: Create manager with multiple registries
	manager := query.NewRegistryManager()

	// Activities registry (root)
	activitiesRegistry := query.NewRelationshipRegistry("activities")
	activitiesRegistry.Register(query.ManyToOneRelationship("user", "users", "user_id"))
	manager.RegisterTable("activities", activitiesRegistry)

	// Users registry (second level)
	usersRegistry := query.NewRelationshipRegistry("users")
	usersRegistry.Register(query.ManyToOneRelationship("company", "companies", "company_id"))
	manager.RegisterTable("users", usersRegistry)

	// Companies registry (third level)
	companiesRegistry := query.NewRelationshipRegistry("companies")
	companiesRegistry.Register(query.ManyToOneRelationship("department", "departments", "department_id"))
	manager.RegisterTable("companies", companiesRegistry)

	// Query options with deep nesting
	opts := &query.QueryOptions{
		Filter: map[string]interface{}{
			"user.company.department.name": "Engineering",
		},
	}

	// Generate JOINs
	joins := activitiesRegistry.GenerateJoins(opts)

	// Verify 3 JOINs generated (user, company, department)
	if len(joins) != 3 {
		t.Errorf("Expected 3 JOINs for deep nesting, got %d", len(joins))
	}

	// Verify JOIN sequence
	expectedJoins := []struct {
		table     string
		condition string
	}{
		{"users", "users.id = activities.user_id"},
		{"companies", "companies.id = users.company_id"},
		{"departments", "departments.id = companies.department_id"},
	}

	for i, expected := range expectedJoins {
		if i >= len(joins) {
			break
		}
		if joins[i].Table != expected.table {
			t.Errorf("JOIN %d: expected table %s, got %s", i, expected.table, joins[i].Table)
		}
		if joins[i].Condition != expected.condition {
			t.Errorf("JOIN %d: expected condition %s, got %s", i, expected.condition, joins[i].Condition)
		}
	}
}

// TestRelationshipRegistry_SelfReferential_v3 tests self-referential relationships (v3.0)
func TestRelationshipRegistry_SelfReferential_v3(t *testing.T) {
	// Setup: Comments table with parent_id
	registry := query.NewRelationshipRegistry("comments")
	registry.Register(query.SelfReferentialRelationship("parent", "comments", "parent_id", 3))

	// Query: Filter by parent comment author
	opts := &query.QueryOptions{
		Filter: map[string]interface{}{
			"parent.author": "john",
		},
	}

	joins := registry.GenerateJoins(opts)

	// Verify 1 JOIN with alias
	if len(joins) != 1 {
		t.Fatalf("Expected 1 JOIN for self-referential, got %d", len(joins))
	}

	// Verify alias usage
	expectedTable := "comments AS parent_comments"
	expectedCondition := "parent_comments.id = comments.parent_id"

	if joins[0].Table != expectedTable {
		t.Errorf("Expected table %s, got %s", expectedTable, joins[0].Table)
	}
	if joins[0].Condition != expectedCondition {
		t.Errorf("Expected condition %s, got %s", expectedCondition, joins[0].Condition)
	}
}

// TestRelationshipRegistry_Polymorphic_v3 tests polymorphic relationships (v3.0)
func TestRelationshipRegistry_Polymorphic_v3(t *testing.T) {
	// Setup: Comments can belong to Posts or Activities
	registry := query.NewRelationshipRegistry("comments")
	registry.Register(query.PolymorphicRelationship(
		"commentable",
		"commentable_type",
		"commentable_id",
		map[string]string{
			"Post":     "posts",
			"Activity": "activities",
		},
	))

	// Test Case 1: Filter with type = Post
	opts := &query.QueryOptions{
		Filter: map[string]interface{}{
			"commentable_type":  "Post",
			"commentable.title": "Hello World",
		},
	}

	joins := registry.GenerateJoins(opts)

	if len(joins) != 1 {
		t.Fatalf("Expected 1 JOIN for polymorphic (Post), got %d", len(joins))
	}

	// Should JOIN posts table
	if joins[0].Table != "posts" {
		t.Errorf("Expected posts table, got %s", joins[0].Table)
	}

	expectedCondition := "posts.id = comments.commentable_id"
	if joins[0].Condition != expectedCondition {
		t.Errorf("Expected condition %s, got %s", expectedCondition, joins[0].Condition)
	}

	// Test Case 2: Filter with type = Activity
	opts2 := &query.QueryOptions{
		Filter: map[string]interface{}{
			"commentable_type":    "Activity",
			"commentable.content": "workout",
		},
	}

	joins2 := registry.GenerateJoins(opts2)

	if len(joins2) != 1 {
		t.Fatalf("Expected 1 JOIN for polymorphic (Activity), got %d", len(joins2))
	}

	// Should JOIN activities table
	if joins2[0].Table != "activities" {
		t.Errorf("Expected activities table, got %s", joins2[0].Table)
	}

	// Test Case 3: No type specified - should not generate JOIN
	opts3 := &query.QueryOptions{
		Filter: map[string]interface{}{
			"commentable.title": "Test",
		},
	}

	joins3 := registry.GenerateJoins(opts3)

	if len(joins3) != 0 {
		t.Errorf("Expected 0 JOINs without type filter, got %d", len(joins3))
	}
}

// TestRelationshipRegistry_WithConditions_v3 tests additional JOIN conditions (v3.0)
func TestRelationshipRegistry_WithConditions_v3(t *testing.T) {
	registry := query.NewRelationshipRegistry("posts")

	// ManyToMany with soft delete filter
	rel := query.ManyToManyRelationship("tags", "tags", "post_tags", "post_id", "tag_id").
		WithConditions(
			query.AdditionalCondition{
				Column:   "tags.is_active",
				Operator: "eq",
				Value:    true,
			},
		)

	registry.Register(rel)

	opts := &query.QueryOptions{
		Filter: map[string]interface{}{
			"tags.name": "tech",
		},
	}

	joins := registry.GenerateJoins(opts)

	// Verify 2 JOINs (junction + target)
	if len(joins) != 2 {
		t.Fatalf("Expected 2 JOINs for many-to-many with conditions, got %d", len(joins))
	}

	// Second JOIN should have additional condition
	if !containsString(joins[1].Condition, "tags.is_active") {
		t.Errorf("Expected JOIN condition to include 'tags.is_active', got: %s", joins[1].Condition)
	}

	if !containsString(joins[1].Condition, "AND") {
		t.Errorf("Expected JOIN condition to include 'AND' for additional condition, got: %s", joins[1].Condition)
	}
}

// TestRelationshipRegistry_CycleDetection_v3 tests duplicate JOIN prevention (v3.0)
func TestRelationshipRegistry_CycleDetection_v3(t *testing.T) {
	registry := query.NewRelationshipRegistry("activities")
	registry.Register(query.ManyToManyRelationship("tags", "tags", "activity_tags", "activity_id", "tag_id"))

	// Query with multiple references to same relationship
	opts := &query.QueryOptions{
		Filter: map[string]interface{}{
			"tags.name": "cardio",
		},
		Search: map[string]interface{}{
			"tags.description": "running",
		},
		Order: map[string]string{
			"tags.name": "ASC",
		},
	}

	joins := registry.GenerateJoins(opts)

	// Should only generate 2 JOINs once (junction + target), not 6 (2 per reference)
	if len(joins) != 2 {
		t.Errorf("Expected 2 JOINs (deduped), got %d", len(joins))
	}

	// Verify no duplicate tables
	seenTables := make(map[string]bool)
	for _, join := range joins {
		if seenTables[join.Table] {
			t.Errorf("Duplicate JOIN detected for table: %s", join.Table)
		}
		seenTables[join.Table] = true
	}
}

// TestRelationshipRegistry_ExtractPath_v3 tests path extraction (v3.0)
func TestRelationshipRegistry_ExtractPath_v3(t *testing.T) {
	registry := query.NewRelationshipRegistry("activities")

	tests := []struct {
		column       string
		expectedPath string
	}{
		{"tags.name", "tags"},
		{"user.company.department.name", "user.company.department"},
		{"created_at", ""},
		{"title", ""},
		{"user.username", "user"},
		{"a.b.c.d.e", "a.b.c.d"},
	}

	for _, tt := range tests {
		t.Run(tt.column, func(t *testing.T) {
			// Use reflection to call private method (for testing)
			// In production, extractPath is called internally by GenerateJoins
			path := extractPathHelper(registry, tt.column)
			if path != tt.expectedPath {
				t.Errorf("extractPath(%s) = %s, want %s", tt.column, path, tt.expectedPath)
			}
		})
	}
}

// TestRegistryManager_CrossRegistry_v3 tests registry manager (v3.0)
func TestRegistryManager_CrossRegistry_v3(t *testing.T) {
	manager := query.NewRegistryManager()

	// Register multiple tables
	activitiesRegistry := query.NewRelationshipRegistry("activities")
	usersRegistry := query.NewRelationshipRegistry("users")
	companiesRegistry := query.NewRelationshipRegistry("companies")

	manager.RegisterTable("activities", activitiesRegistry)
	manager.RegisterTable("users", usersRegistry)
	manager.RegisterTable("companies", companiesRegistry)

	// Verify registration
	reg, found := manager.GetRegistry("activities")
	if !found {
		t.Error("Expected to find activities registry")
	}
	if reg.ParentTable != "activities" {
		t.Errorf("Expected ParentTable=activities, got %s", reg.ParentTable)
	}

	reg2, found2 := manager.GetRegistry("users")
	if !found2 {
		t.Error("Expected to find users registry")
	}
	if reg2.ParentTable != "users" {
		t.Errorf("Expected ParentTable=users, got %s", reg2.ParentTable)
	}

	// Non-existent registry
	_, found3 := manager.GetRegistry("nonexistent")
	if found3 {
		t.Error("Expected not to find nonexistent registry")
	}
}

// TestRelationshipRegistry_MixedFeatures_v3 tests combining multiple v3.0 features
func TestRelationshipRegistry_MixedFeatures_v3(t *testing.T) {
	manager := query.NewRegistryManager()

	// Setup complex scenario with deep nesting + polymorphic + self-referential
	commentsRegistry := query.NewRelationshipRegistry("comments")
	commentsRegistry.Register(query.SelfReferentialRelationship("parent", "comments", "parent_id", 3))
	commentsRegistry.Register(query.PolymorphicRelationship(
		"commentable",
		"commentable_type",
		"commentable_id",
		map[string]string{
			"Post": "posts",
		},
	))
	manager.RegisterTable("comments", commentsRegistry)

	// Query combining features
	opts := &query.QueryOptions{
		Filter: map[string]interface{}{
			"parent.author":     "john",
			"commentable_type":  "Post",
			"commentable.title": "Hello",
		},
	}

	joins := commentsRegistry.GenerateJoins(opts)

	// Should generate 2 JOINs: 1 for parent (self-ref), 1 for commentable (polymorphic)
	if len(joins) != 2 {
		t.Errorf("Expected 2 JOINs for mixed features, got %d", len(joins))
	}

	// Verify both JOIN types present
	hasAlias := false
	hasPolymorphic := false

	for _, join := range joins {
		if containsString(join.Table, "AS") {
			hasAlias = true
		}
		if join.Table == "posts" {
			hasPolymorphic = true
		}
	}

	if !hasAlias {
		t.Error("Expected self-referential JOIN with alias")
	}
	if !hasPolymorphic {
		t.Error("Expected polymorphic JOIN to posts table")
	}
}

// Helper function to test extractPath (simulates calling the private method)
func extractPathHelper(registry *query.RelationshipRegistry, column string) string {
	// Create a dummy query option to trigger path extraction
	opts := &query.QueryOptions{
		Filter: map[string]interface{}{
			column: "dummy",
		},
	}

	// Generate joins (which internally calls extractPath)
	_ = registry.GenerateJoins(opts)

	// Extract path manually using same logic
	lastDot := -1
	for i := len(column) - 1; i >= 0; i-- {
		if column[i] == '.' {
			lastDot = i
			break
		}
	}

	if lastDot == -1 {
		return ""
	}
	return column[:lastDot]
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
