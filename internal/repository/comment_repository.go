package repository

import (
	"context"
	"database/sql"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/pkg/query"
)

// CommentRepository handles data access for comments.
// It supports polymorphic relationships so comments can belong to any entity
// (Activity, Tag, etc.) and can be filtered by the commentable entity's columns.
type CommentRepository struct {
	db       DBConn
	registry *query.RelationshipRegistry
}

// NewCommentRepository creates a new CommentRepository with a polymorphic registry.
// The "commentable" relationship auto-JOINs the correct table based on commentable_type:
//   - filter[commentable_type]=Activity → JOINs activities table
//   - filter[commentable_type]=Tag      → JOINs tags table
func NewCommentRepository(db DBConn) *CommentRepository {
	registry := query.NewRelationshipRegistry("comments")

	registry.Register(query.PolymorphicRelationship(
		"commentable",
		"commentable_type",
		"commentable_id",
		map[string]string{
			"Activity": "activities",
			"Tag":      "tags",
		},
	))

	return &CommentRepository{
		db:       db,
		registry: registry,
	}
}

// GetRegistry returns the RelationshipRegistry for this repository (v3.0)
// Registered with RegistryManager to enable cross-registry deep nesting through comments
func (cr *CommentRepository) GetRegistry() *query.RelationshipRegistry {
	return cr.registry
}

// scanComment scans a single comment row from SELECT comments.*
func (cr *CommentRepository) scanComment(rows *sql.Rows) (*models.Comment, error) {
	comment := &models.Comment{}
	err := rows.Scan(
		&comment.ID,
		&comment.UserID,
		&comment.CommentableType,
		&comment.CommentableID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	return comment, err
}

// ListCommentsWithQuery returns a paginated list of comments with dynamic filtering.
// The polymorphic relationship is resolved automatically based on commentable_type:
//
//	GET /comments?filter[commentable_type]=Activity&filter[commentable.title]=Morning+Run
//	→ LEFT JOIN activities ON activities.id = comments.commentable_id
//	→ WHERE activities.title = $1
//
//	GET /comments?filter[commentable_type]=Tag&filter[commentable.name]=cardio
//	→ LEFT JOIN tags ON tags.id = comments.commentable_id
//	→ WHERE tags.name = $1
func (cr *CommentRepository) ListCommentsWithQuery(
	ctx context.Context,
	opts *query.QueryOptions,
) (*query.PaginatedResult, error) {
	joins := cr.registry.GenerateJoins(opts)

	return FindAndPaginate[models.Comment](
		ctx,
		cr.db,
		"comments",
		opts,
		cr.scanComment,
		joins...,
	)
}
