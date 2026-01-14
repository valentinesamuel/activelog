package models

type User struct {
	BaseEntity
	Email        string `json:"email,omitempty" `
	Username     string `json:"username,omitempty" `
	PasswordHash string `json:"password_hash,omitempty" `

	Activities []Activity `json:"activities,omitempty"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,max=20,min=4"`
	Password string `json:"password" validate:"required,min=4"`
	Email    string `json:"email" validate:"required,min=4"`
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,min=4"`
	Password string `json:"password" validate:"required,min=4"`
}
