package query

import (
	"fmt"
	"strings"
)

// RelationshipType defines how tables are related
type RelationshipType string

const (
	// OneToMany: One parent has many children (e.g., User has many Activities)
	OneToMany RelationshipType = "one_to_many"

	// ManyToOne: Many children belong to one parent (e.g., Activities belong to User)
	ManyToOne RelationshipType = "many_to_one"

	// ManyToMany: Many-to-many through junction table (e.g., Activities <-> Tags)
	ManyToMany RelationshipType = "many_to_many"

	// SelfReferential: Table references itself (e.g., Comments with parent_id)
	SelfReferential RelationshipType = "self_referential"

	// Polymorphic: Relationship to multiple table types (e.g., commentable_type/commentable_id)
	Polymorphic RelationshipType = "polymorphic"
)

// Relationship defines how to JOIN tables automatically
type Relationship struct {
	// Name is the dot-notation prefix users will use
	// Example: "tags" for activities.tags.name
	Name string

	// Type of relationship
	Type RelationshipType

	// TargetTable is the table being joined
	// Example: "tags"
	TargetTable string

	// ForeignKey is the column linking parent to child (for OneToMany/ManyToOne)
	// Example: "user_id" in activities table
	ForeignKey string

	// PrimaryKey is the column in the target table (usually "id")
	// Example: "id" in users table
	PrimaryKey string

	// JunctionTable is the intermediate table for ManyToMany
	// Example: "activity_tags"
	JunctionTable string

	// JunctionForeignKey is the column in junction table referencing parent
	// Example: "activity_id" in activity_tags
	JunctionForeignKey string

	// JunctionTargetKey is the column in junction table referencing target
	// Example: "tag_id" in activity_tags
	JunctionTargetKey string

	// TargetPrimaryKey is the primary key in the target table
	// Example: "id" in tags table
	TargetPrimaryKey string

	// v3.0: Additional configuration for advanced features
	Alias           string                // For self-referential relationships (e.g., "parent_comment")
	MaxDepth        int                   // Max nesting depth for self-referential (default: 3)
	PolymorphicType string                // Type discriminator column (e.g., "commentable_type")
	PolymorphicID   string                // ID column (e.g., "commentable_id")
	PolymorphicMap  map[string]string     // Type -> Table mapping (e.g., "Post" -> "posts")
	JoinConditions  []AdditionalCondition // Extra WHERE clauses in JOIN
}

// AdditionalCondition represents extra conditions in JOIN clauses (v3.0)
type AdditionalCondition struct {
	Column   string      // Column name (e.g., "tags.is_active")
	Operator string      // Operator (e.g., "eq", "ne")
	Value    interface{} // Value to compare
}

// RelationshipRegistry stores all relationships for an entity
type RelationshipRegistry struct {
	// ParentTable is the main table (e.g., "activities")
	ParentTable string

	// Relationships maps relationship names to their configurations
	// Example: map["tags"] = Relationship{...}
	Relationships map[string]Relationship

	// v3.0: Reference to RegistryManager for cross-registry resolution
	manager *RegistryManager
}

// RegistryManager manages multiple registries for cross-registry path resolution (v3.0)
// Example: "user.company.department.name" requires resolving across user, company, and department registries
type RegistryManager struct {
	registries map[string]*RelationshipRegistry // table name -> registry
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager() *RegistryManager {
	return &RegistryManager{
		registries: make(map[string]*RelationshipRegistry),
	}
}

// RegisterTable registers a table's relationship registry
func (rm *RegistryManager) RegisterTable(tableName string, registry *RelationshipRegistry) {
	rm.registries[tableName] = registry
	registry.manager = rm // Link back to manager
}

// GetRegistry retrieves a registry for a table
func (rm *RegistryManager) GetRegistry(tableName string) (*RelationshipRegistry, bool) {
	registry, exists := rm.registries[tableName]
	return registry, exists
}

// NewRelationshipRegistry creates a new registry for an entity
func NewRelationshipRegistry(parentTable string) *RelationshipRegistry {
	return &RelationshipRegistry{
		ParentTable:   parentTable,
		Relationships: make(map[string]Relationship),
	}
}

// Register adds a relationship to the registry
func (rr *RelationshipRegistry) Register(rel Relationship) {
	rr.Relationships[rel.Name] = rel
}

// ManyToManyRelationship is a helper to create many-to-many relationships
func ManyToManyRelationship(name, targetTable, junctionTable, junctionForeignKey, junctionTargetKey string) Relationship {
	return Relationship{
		Name:               name,
		Type:               ManyToMany,
		TargetTable:        targetTable,
		JunctionTable:      junctionTable,
		JunctionForeignKey: junctionForeignKey,
		JunctionTargetKey:  junctionTargetKey,
		TargetPrimaryKey:   "id",
	}
}

// ManyToOneRelationship is a helper to create many-to-one relationships
func ManyToOneRelationship(name, targetTable, foreignKey string) Relationship {
	return Relationship{
		Name:        name,
		Type:        ManyToOne,
		TargetTable: targetTable,
		ForeignKey:  foreignKey,
		PrimaryKey:  "id",
	}
}

// OneToManyRelationship is a helper to create one-to-many relationships
func OneToManyRelationship(name, targetTable, foreignKey string) Relationship {
	return Relationship{
		Name:        name,
		Type:        OneToMany,
		TargetTable: targetTable,
		ForeignKey:  foreignKey,
		PrimaryKey:  "id",
	}
}

// SelfReferentialRelationship creates a self-referential relationship (v3.0)
// Example: Comments table with parent_id referencing other comments
func SelfReferentialRelationship(name, table, foreignKey string, maxDepth int) Relationship {
	alias := name + "_" + table // e.g., "parent_comments"
	if maxDepth == 0 {
		maxDepth = 3 // Default max depth
	}
	return Relationship{
		Name:        name,
		Type:        SelfReferential,
		TargetTable: table,
		ForeignKey:  foreignKey,
		PrimaryKey:  "id",
		Alias:       alias,
		MaxDepth:    maxDepth,
	}
}

// PolymorphicRelationship creates a polymorphic relationship (v3.0)
// Example: Comments can belong to Posts or Activities
func PolymorphicRelationship(name, typeColumn, idColumn string, typeMap map[string]string) Relationship {
	return Relationship{
		Name:            name,
		Type:            Polymorphic,
		PolymorphicType: typeColumn,
		PolymorphicID:   idColumn,
		PolymorphicMap:  typeMap,
	}
}

// WithConditions adds extra WHERE conditions to a relationship (v3.0)
func (r Relationship) WithConditions(conditions ...AdditionalCondition) Relationship {
	r.JoinConditions = append(r.JoinConditions, conditions...)
	return r
}

// GenerateJoins analyzes query options and generates required JOINs (v3.0 enhanced)
// Supports:
//   - 1-level relationships: tags.name
//   - Deep nesting: user.company.department.name (requires RegistryManager)
//   - Self-referential: parent.author
//   - Polymorphic: commentable.title (requires type filter)
func (rr *RelationshipRegistry) GenerateJoins(opts *QueryOptions) []JoinConfig {
	neededPaths := make(map[string]bool) // Full paths like "user.company.department"
	joins := []JoinConfig{}

	// Collect all relationship paths from query options
	for column := range opts.Filter {
		if path := rr.extractPath(column); path != "" {
			neededPaths[path] = true
		}
	}

	for column := range opts.FilterOr {
		if path := rr.extractPath(column); path != "" {
			neededPaths[path] = true
		}
	}

	for _, condition := range opts.FilterConditions {
		if path := rr.extractPath(condition.Column); path != "" {
			neededPaths[path] = true
		}
	}

	for column := range opts.Search {
		if path := rr.extractPath(column); path != "" {
			neededPaths[path] = true
		}
	}

	for column := range opts.Order {
		if path := rr.extractPath(column); path != "" {
			neededPaths[path] = true
		}
	}

	// Generate JOINs for each unique path
	seenTables := make(map[string]bool) // Track which tables are already joined
	for path := range neededPaths {
		pathJoins := rr.resolvePathToJoins(path, opts, seenTables)
		joins = append(joins, pathJoins...)
	}

	return joins
}

// extractPath extracts the relationship path from a column reference (v3.0)
// Examples:
//   - "tags.name" → "tags"
//   - "user.company.department.name" → "user.company.department"
//   - "created_at" → ""
func (rr *RelationshipRegistry) extractPath(column string) string {
	lastDot := strings.LastIndex(column, ".")
	if lastDot == -1 {
		return "" // No relationship
	}
	return column[:lastDot]
}

// resolvePathToJoins resolves a relationship path to JOIN clauses (v3.0)
// Handles both single-level and multi-level paths
func (rr *RelationshipRegistry) resolvePathToJoins(path string, opts *QueryOptions, seenTables map[string]bool) []JoinConfig {
	joins := []JoinConfig{}

	// Split path into segments: "user.company.department" → ["user", "company", "department"]
	segments := strings.Split(path, ".")

	// Start with the current registry
	currentRegistry := rr
	currentTable := rr.ParentTable

	for _, segment := range segments {
		// Find relationship in current registry
		rel, exists := currentRegistry.Relationships[segment]
		if !exists {
			// Unknown relationship - skip
			break
		}

		// Generate JOIN(s) for this relationship
		segmentJoins := rr.generateJoinForRelationship(rel, currentTable, opts, seenTables)
		joins = append(joins, segmentJoins...)

		// Move to the next table for deep nesting
		currentTable = rel.TargetTable

		// For deep nesting, try to get the next registry (if manager is available)
		if rr.manager != nil {
			nextRegistry, found := rr.manager.GetRegistry(currentTable)
			if found {
				currentRegistry = nextRegistry
			} else {
				// No registry for this table - can't go deeper
				break
			}
		} else {
			// No manager - can't resolve cross-registry paths
			break
		}
	}

	return joins
}

// generateJoinForRelationship creates JOIN configs for a single relationship (v3.0)
func (rr *RelationshipRegistry) generateJoinForRelationship(rel Relationship, parentTable string, opts *QueryOptions, seenTables map[string]bool) []JoinConfig {
	joins := []JoinConfig{}

	switch rel.Type {
	case ManyToMany:
		// Junction table JOIN (if not already joined)
		if !seenTables[rel.JunctionTable] {
			joins = append(joins, JoinConfig{
				Table:     rel.JunctionTable,
				Condition: fmt.Sprintf("%s.%s = %s.id", rel.JunctionTable, rel.JunctionForeignKey, parentTable),
			})
			seenTables[rel.JunctionTable] = true
		}

		// Target table JOIN (if not already joined)
		if !seenTables[rel.TargetTable] {
			joins = append(joins, JoinConfig{
				Table:     rel.TargetTable,
				Condition: fmt.Sprintf("%s.id = %s.%s", rel.TargetTable, rel.JunctionTable, rel.JunctionTargetKey),
			})
			seenTables[rel.TargetTable] = true
		}

	case ManyToOne:
		// Single JOIN (if not already joined)
		if !seenTables[rel.TargetTable] {
			joins = append(joins, JoinConfig{
				Table:     rel.TargetTable,
				Condition: fmt.Sprintf("%s.id = %s.%s", rel.TargetTable, parentTable, rel.ForeignKey),
			})
			seenTables[rel.TargetTable] = true
		}

	case OneToMany:
		// Single JOIN (if not already joined)
		if !seenTables[rel.TargetTable] {
			joins = append(joins, JoinConfig{
				Table:     rel.TargetTable,
				Condition: fmt.Sprintf("%s.%s = %s.id", rel.TargetTable, rel.ForeignKey, parentTable),
			})
			seenTables[rel.TargetTable] = true
		}

	case SelfReferential:
		// Self-referential with alias (if not already joined)
		aliasKey := rel.Alias
		if !seenTables[aliasKey] {
			joins = append(joins, JoinConfig{
				Table:     fmt.Sprintf("%s AS %s", rel.TargetTable, rel.Alias),
				Condition: fmt.Sprintf("%s.id = %s.%s", rel.Alias, parentTable, rel.ForeignKey),
			})
			seenTables[aliasKey] = true
		}

	case Polymorphic:
		// Polymorphic JOIN requires type information
		typeValue := rr.getPolymorphicType(rel, opts)
		if typeValue == "" {
			// No type specified - can't JOIN
			return joins
		}

		targetTable, ok := rel.PolymorphicMap[typeValue]
		if !ok {
			// Unknown type - skip
			return joins
		}

		// JOIN the target table (if not already joined)
		if !seenTables[targetTable] {
			joins = append(joins, JoinConfig{
				Table:     targetTable,
				Condition: fmt.Sprintf("%s.id = %s.%s", targetTable, parentTable, rel.PolymorphicID),
			})
			seenTables[targetTable] = true
		}
	}

	// Add additional conditions if specified (v3.0 feature)
	if len(rel.JoinConditions) > 0 && len(joins) > 0 {
		lastJoin := &joins[len(joins)-1]
		for _, cond := range rel.JoinConditions {
			condSQL := rr.buildConditionSQL(cond)
			if condSQL != "" {
				lastJoin.Condition += " AND " + condSQL
			}
		}
	}

	return joins
}

// getPolymorphicType extracts the type value for polymorphic relationships (v3.0)
func (rr *RelationshipRegistry) getPolymorphicType(rel Relationship, opts *QueryOptions) string {
	// Check Filter
	if typeVal, ok := opts.Filter[rel.PolymorphicType]; ok {
		if str, isStr := typeVal.(string); isStr {
			return str
		}
	}

	// Check FilterConditions
	for _, cond := range opts.FilterConditions {
		if cond.Column == rel.PolymorphicType {
			if str, isStr := cond.Value.(string); isStr {
				return str
			}
		}
	}

	return ""
}

// buildConditionSQL converts an AdditionalCondition to SQL (v3.0)
func (rr *RelationshipRegistry) buildConditionSQL(cond AdditionalCondition) string {
	opMap := map[string]string{
		"eq":  "=",
		"ne":  "!=",
		"gt":  ">",
		"gte": ">=",
		"lt":  "<",
		"lte": "<=",
	}

	sqlOp, ok := opMap[cond.Operator]
	if !ok {
		sqlOp = "=" // Default
	}

	// Simple formatting - in production, use parameterized queries
	return fmt.Sprintf("%s %s %v", cond.Column, sqlOp, cond.Value)
}

// extractRelationship extracts relationship name from a dot-notation column
// Example: "tags.name" → ("tags", true)
// Example: "created_at" → ("", false)
func (rr *RelationshipRegistry) extractRelationship(column string) (string, bool) {
	// Find first dot
	dotIndex := -1
	for i, ch := range column {
		if ch == '.' {
			dotIndex = i
			break
		}
	}

	if dotIndex == -1 {
		return "", false
	}

	relationshipName := column[:dotIndex]

	// Check if this is a registered relationship
	if _, exists := rr.Relationships[relationshipName]; exists {
		return relationshipName, true
	}

	return "", false
}

// ValidateColumn checks if a column reference is valid
// Supports both direct columns and relationship columns
func (rr *RelationshipRegistry) ValidateColumn(column string, allowedColumns []string) error {
	// Check if it's a relationship column (contains dot)
	if rel, found := rr.extractRelationship(column); found {
		// Verify the relationship exists
		if _, exists := rr.Relationships[rel]; !exists {
			return fmt.Errorf("unknown relationship '%s' in column '%s'", rel, column)
		}
		return nil
	}

	// Check if it's in allowed direct columns
	for _, allowed := range allowedColumns {
		if column == allowed {
			return nil
		}
	}

	return fmt.Errorf("column '%s' is not in whitelist", column)
}
