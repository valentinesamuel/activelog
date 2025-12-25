package models

import (
	"database/sql"
	"time"
)

type BaseEntity struct {
	ID int64 `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt sql.NullTime `json:"updatedAt" db:"updated_at"`
}
