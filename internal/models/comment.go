package models

// Comment represents a polymorphic comment that can belong to any commentable entity.
// The CommentableType field (e.g., "Activity", "Tag") determines which table is JOINed
// when filtering via the polymorphic relationship.
type Comment struct {
	BaseEntity
	UserID          int    `json:"user_id"`
	CommentableType string `json:"commentable_type"`
	CommentableID   int    `json:"commentable_id"`
	Content         string `json:"content"`
}
