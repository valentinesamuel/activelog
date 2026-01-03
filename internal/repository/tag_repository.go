package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/valentinesamuel/activelog/pkg/logger"
)

type TagRepository struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{
		db: db,
	}
}

func (tr *TagRepository) GetOrCreateTag(ctx context.Context, name string) (int, error) {
	query := `
		INSERT INTO tags (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE
		SET name = EXCLUDED.name
		RETURNING id
	`

	var id int
	err := tr.db.QueryRowContext(ctx, query, name).Scan(&id)
	if err != nil {
		return 0, err
	}

	logger.Info().Int("tag_id", id).Msg("✅ Created tag")
	return id, nil
}

func (tr *TagRepository) GetTagsForActivity(ctx context.Context, activityID int) ([]string, error) {

	query := `
		SELECT 
		 tags.name
		FROM activity_tags as at
		JOIN tags ON at.tag_id = tags.id
		WHERE activity_id = $1
	`

	rows, err := tr.db.QueryContext(ctx, query, activityID)
	if err != nil {
		return nil, fmt.Errorf("❌ Error listing activity tags: %w", err)
	}

	defer rows.Close()

	var tags []string

	for rows.Next() {
		var tagName string
		err := rows.Scan(
			&tagName,
		)

		if err != nil {
			return nil, fmt.Errorf("❌ Error scanning tags: %w", err)
		}
		tags = append(tags, tagName)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("❌ Error listing tags")

		return nil, err
	}

	fmt.Println("✅ Tags fetched successfully!")
	return tags, nil

}

func (tr *TagRepository) LinkActivityTag(ctx context.Context, activityID, tagID int) error {
	query := `
		INSERT INTO activity_tags
		(tag_id, activity_id)
		VALUES ($1, $2)
		ON CONFLICT (tag_id, activity_id) DO NOTHING;
	`

	err := tr.db.QueryRowContext(ctx, query, tagID, activityID).Scan()

	if err != nil {
		return fmt.Errorf("❌ Error creating activity tag %w", err)
	}

	fmt.Println("✅ Activity tag created successfully!")

	return nil
}
