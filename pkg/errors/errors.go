package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Sentinel errors - predefined errors you can compare with errors.Is()
var (
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlreadyExists = errors.New("resource already exists")
)

// Custom error type with context
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("❌ validation error on field '%s': %s (%v)", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// DatabaseError wraps database-related errors
type DatabaseError struct {
	Op    string // Operation like "insert", "update", "delete"
	Table string
	Err   error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("❌ database error during %s on %s: %v", e.Op, e.Table, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("❌ %s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewBadRequest(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}
