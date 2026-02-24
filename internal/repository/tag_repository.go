package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/pkg/logger"
	"github.com/valentinesamuel/activelog/pkg/query"
)

type TagRepository struct {
	db       DBConn
	registry *query.RelationshipRegistry
}

func NewTagRepository(db DBConn) *TagRepository {
	registry := query.NewRelationshipRegistry("tags")

	// Self-referential: tags can have a parent tag (parent_tag_id)
	// Alias = "parent" so filter key "tags.parent.name" maps to SQL "parent.name"
	registry.Register(query.SelfReferentialRelationship("parent", "tags", "parent_tag_id", 3))

	return &TagRepository{
		db:       db,
		registry: registry,
	}
}

// GetRegistry returns the RelationshipRegistry for this repository (v3.0)
// Used by RegistryManager for cross-registry deep nesting (e.g., activities→tags→parent)
func (tr *TagRepository) GetRegistry() *query.RelationshipRegistry {
	return tr.registry
}

func (tr *TagRepository) GetOrCreateTag(ctx context.Context, tx TxConn, name string) (int, error) {
	query := `
		INSERT INTO tags (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE
		SET name = EXCLUDED.name
		RETURNING id
	`

	var id int
	var err error

	// Use transaction if provided, otherwise use db connection
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, name).Scan(&id)
	} else {
		err = tr.db.QueryRowContext(ctx, query, name).Scan(&id)
	}

	if err != nil {
		return 0, err
	}

	logger.Info().Int("tag_id", id).Msg("✅ Created tag")
	return id, nil
}

func (tr *TagRepository) GetTagsForActivity(ctx context.Context, activityID int) ([]*models.Tag, error) {

	query := `
		SELECT
		 tags.id,
		 tags.name,
		 tags.created_at
		FROM activity_tags as at
		JOIN tags ON at.tag_id = tags.id
		WHERE activity_id = $1
	`

	rows, err := tr.db.QueryContext(ctx, query, activityID)
	if err != nil {
		return nil, fmt.Errorf("❌ Error listing activity tags: %w", err)
	}

	defer rows.Close()

	var tags []*models.Tag

	for rows.Next() {
		tag := &models.Tag{}
		err := rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("❌ Error scanning tags: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("❌ Error listing tags")

		return nil, err
	}

	fmt.Println("✅ Tags fetched successfully!")
	return tags, nil

}

func (tr *TagRepository) LinkActivityTag(ctx context.Context, tx TxConn, activityID int, tagID int) error {
	query := `
		INSERT INTO activity_tags
		(tag_id, activity_id)
		VALUES ($1, $2)
		ON CONFLICT (tag_id, activity_id) DO NOTHING;
	`

	var err error

	// Use transaction if provided, otherwise use db connection
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, tagID, activityID)
	} else {
		_, err = tr.db.ExecContext(ctx, query, tagID, activityID)
	}

	if err != nil {
		return fmt.Errorf("❌ Error creating activity tag %w", err)
	}

	fmt.Println("✅ Activity tag created successfully!")

	return nil
}

// scanTag is a reusable function to scan a single tag row
// Scans all columns from SELECT tags.*: id, name, created_at, deleted_at, parent_tag_id
func (tr *TagRepository) scanTag(rows *sql.Rows) (*models.Tag, error) {
	tag := &models.Tag{}
	var parentTagID sql.NullInt64 // parent_tag_id is nullable; not exposed on model yet
	err := rows.Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
		&tag.DeletedAt,
		&parentTagID,
	)
	return tag, err
}

// ListTagsWithQuery uses the new dynamic filtering pattern with QueryOptions
// This method leverages the generic FindAndPaginate function for flexible, type-safe queries.
//
// Example usage in handler:
//
//	opts := &query.QueryOptions{
//	    Page: 1,
//	    Limit: 20,
//	    Filter: map[string]interface{}{
//	        "name": "cardio",
//	    },
//	    Search: map[string]interface{}{
//	        "name": "run",
//	    },
//	    Order: map[string]string{
//	        "name": "ASC",
//	    },
//	}
//	result, err := repo.ListTagsWithQuery(ctx, opts)
func (tr *TagRepository) ListTagsWithQuery(
	ctx context.Context,
	opts *query.QueryOptions,
) (*query.PaginatedResult, error) {
	// Use the generic FindAndPaginate function with our scanTag function
	return FindAndPaginate[models.Tag](
		ctx,
		tr.db,
		"tags",
		opts,
		tr.scanTag,
	)
}
