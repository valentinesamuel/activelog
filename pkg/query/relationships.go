package query

import "fmt"

// RelationshipType defines how tables are related
type RelationshipType string

const (
	// OneToMany: One parent has many children (e.g., User has many Activities)
	OneToMany RelationshipType = "one_to_many"

	// ManyToOne: Many children belong to one parent (e.g., Activities belong to User)
	ManyToOne RelationshipType = "many_to_one"

	// ManyToMany: Many-to-many through junction table (e.g., Activities <-> Tags)
	ManyToMany RelationshipType = "many_to_many"
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
}

// RelationshipRegistry stores all relationships for an entity
type RelationshipRegistry struct {
	// ParentTable is the main table (e.g., "activities")
	ParentTable string

	// Relationships maps relationship names to their configurations
	// Example: map["tags"] = Relationship{...}
	Relationships map[string]Relationship
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
		Name:                name,
		Type:                ManyToMany,
		TargetTable:         targetTable,
		JunctionTable:       junctionTable,
		JunctionForeignKey:  junctionForeignKey,
		JunctionTargetKey:   junctionTargetKey,
		TargetPrimaryKey:    "id",
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

// GenerateJoins analyzes query options and generates required JOINs
func (rr *RelationshipRegistry) GenerateJoins(opts *QueryOptions) []JoinConfig {
	needed := make(map[string]bool)
	joins := []JoinConfig{}

	// Scan all filter columns for relationships
	for column := range opts.Filter {
		if rel, found := rr.extractRelationship(column); found {
			needed[rel] = true
		}
	}

	for column := range opts.FilterOr {
		if rel, found := rr.extractRelationship(column); found {
			needed[rel] = true
		}
	}

	// Scan FilterConditions (v1.1.0+)
	for _, condition := range opts.FilterConditions {
		if rel, found := rr.extractRelationship(condition.Column); found {
			needed[rel] = true
		}
	}

	// Scan search columns
	for column := range opts.Search {
		if rel, found := rr.extractRelationship(column); found {
			needed[rel] = true
		}
	}

	// Scan order columns
	for column := range opts.Order {
		if rel, found := rr.extractRelationship(column); found {
			needed[rel] = true
		}
	}

	// Generate JOINs for each needed relationship
	for relName := range needed {
		rel, exists := rr.Relationships[relName]
		if !exists {
			continue
		}

		switch rel.Type {
		case ManyToMany:
			// Generate two JOINs for many-to-many
			joins = append(joins,
				JoinConfig{
					Table:     rel.JunctionTable,
					Condition: fmt.Sprintf("%s.%s = %s.id", rel.JunctionTable, rel.JunctionForeignKey, rr.ParentTable),
				},
				JoinConfig{
					Table:     rel.TargetTable,
					Condition: fmt.Sprintf("%s.id = %s.%s", rel.TargetTable, rel.JunctionTable, rel.JunctionTargetKey),
				},
			)

		case ManyToOne:
			// Single JOIN for many-to-one
			joins = append(joins, JoinConfig{
				Table:     rel.TargetTable,
				Condition: fmt.Sprintf("%s.id = %s.%s", rel.TargetTable, rr.ParentTable, rel.ForeignKey),
			})

		case OneToMany:
			// Single JOIN for one-to-many
			joins = append(joins, JoinConfig{
				Table:     rel.TargetTable,
				Condition: fmt.Sprintf("%s.%s = %s.id", rel.TargetTable, rel.ForeignKey, rr.ParentTable),
			})
		}
	}

	return joins
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
