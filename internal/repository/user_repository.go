package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/pkg/errors"
	"github.com/valentinesamuel/activelog/pkg/query"
)

type UserRepository struct {
	db       DBConn
	registry *query.RelationshipRegistry
}

func NewUserRepository(db DBConn) *UserRepository {
	registry := query.NewRelationshipRegistry("users")
	// Users registry is minimal; extend here when cross-registry paths through users are needed
	return &UserRepository{
		db:       db,
		registry: registry,
	}
}

// GetRegistry returns the RelationshipRegistry for this repository (v3.0)
// Used by RegistryManager so paths like "users.company.name" can be resolved
func (ur *UserRepository) GetRegistry() *query.RelationshipRegistry {
	return ur.registry
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
		if mapped := mapPgError(err); mapped != nil {
			return mapped
		}
		return &errors.DatabaseError{Op: "INSERT", Table: "users", Err: err}
	}

	fmt.Println("✅ User created successfully!")

	return nil
}

func (ar *UserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT 
		id, username, email, password_hash
		FROM users
		WHERE email = $1
	`

	user := &models.User{}

	err := ar.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}

	if err != nil {
		if mapped := mapPgError(err); mapped != nil {
			return nil, mapped
		}
		return nil, &errors.DatabaseError{
			Op:    "SELECT",
			Table: "user",
			Err:   err,
		}
	}

	fmt.Println("✅ User found successfully!")

	return user, nil
}
