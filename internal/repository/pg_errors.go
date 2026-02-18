package repository

import (
	"github.com/lib/pq"
	"github.com/valentinesamuel/activelog/pkg/errors"
)

// PostgreSQL error codes (SQLSTATE)
const (
	pgUniqueViolation     = pq.ErrorCode("23505")
	pgForeignKeyViolation = pq.ErrorCode("23503")
	pgNotNullViolation    = pq.ErrorCode("23502")
	pgCheckViolation      = pq.ErrorCode("23514")
)

// mapPgError converts a raw *pq.Error into an application sentinel error.
// Returns nil if err is not a *pq.Error, so callers can use it as a filter.
func mapPgError(err error) error {
	pqErr, ok := err.(*pq.Error)
	if !ok {
		return nil
	}
	switch pqErr.Code {
	case pgUniqueViolation:
		return errors.ErrAlreadyExists
	case pgForeignKeyViolation:
		return errors.ErrInvalidInput
	case pgNotNullViolation:
		return errors.ErrInvalidInput
	default:
		return nil
	}
}
