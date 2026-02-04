package broker

import (
	"context"
	"database/sql"
)

// TypedUseCase is the generic interface for type-safe use cases.
// Each use case defines its own Input and Output types for compile-time safety.
//
// Example:
//
//	type ListActivitiesInput struct {
//	    UserID       int
//	    QueryOptions *query.QueryOptions
//	}
//
//	type ListActivitiesOutput struct {
//	    Result *query.PaginatedResult
//	}
//
//	func (uc *ListActivitiesUseCase) Execute(
//	    ctx context.Context,
//	    tx *sql.Tx,
//	    input ListActivitiesInput,
//	) (ListActivitiesOutput, error) { ... }
type TypedUseCase[I, O any] interface {
	Execute(ctx context.Context, tx *sql.Tx, input I) (O, error)
}

// TransactionalTypedUseCase is an optional marker interface for typed use cases
// that require transactions. Use cases that implement this interface can declare
// whether they need to run in a transaction.
//
// Default behavior: Use cases WITHOUT this method are assumed to be NON-TRANSACTIONAL.
// This optimizes read-heavy operations which don't need transaction overhead.
type TransactionalTypedUseCase[I, O any] interface {
	TypedUseCase[I, O]
	RequiresTransaction() bool
}
