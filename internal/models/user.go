package models

type User struct {
	BaseEntity
	Email    string `json:"email" db:"email"`
	Username string `json:"username" db:"username"`

	Activities []Activity `json:"activities,omitempty"`
}
