package models

type User struct {
	BaseEntity
	Email        string `json:"email" `
	Username     string `json:"username" `
	PasswordHash string `json:"password_hash" `

	Activities []Activity `json:"activities,omitempty"`
}
