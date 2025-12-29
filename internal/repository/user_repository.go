package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/pkg/errors"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (ar *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users
		(email, username, password_hash) 
		VALUES ($1, $2, $3)
		RETURNING email, created_at, updated_at;
	`

	err := ar.db.QueryRowContext(ctx, query, user.Email, user.Username, user.PasswordHash).Scan(&user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("❌ Error creating user %w", err)
	}

	fmt.Println("✅ User created successfully!")

	return nil
}

func (ar *UserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT 
		username, email, 
		FROM users
		WHERE email = $1
	`

	user := &models.User{}

	err := ar.db.QueryRowContext(ctx, query, email).Scan(&user.Username, &user.Email)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}

	if err != nil {
		return nil, &errors.DatabaseError{
			Op:    "SELECT",
			Table: "user",
			Err:   err,
		}
	}

	fmt.Println("✅ User created successfully!")

	return user, nil
}
