package models

import (
	"database/sql"
	"time"
)

type BaseEntity struct {
	ID        int64        `json:"id"  `
	CreatedAt time.Time    `json:"created_at" `
	UpdatedAt sql.NullTime `json:"updated_at" `
}
