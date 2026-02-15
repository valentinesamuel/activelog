package models

import "time"

type ActivityPhoto struct {
	BaseEntity
	ActivityID   int       `json:"activity_id,omitempty" `
	S3Key        string    `json:"s3_key,omitempty" `
	ThumbnailKey string    `json:"thumbnail_key,omitempty" `
	ContentType  string    `json:"content_type,omitempty" `
	FileSize     int64     `json:"file_size,omitempty" validate:"required,min=2,max=2457600" `
	UploadedAt   time.Time `json:"uploaded_at" `
}
