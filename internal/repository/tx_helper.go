package repository

import (
	"context"
	"database/sql"

	"github.com/valentinesamuel/activelog/pkg/database"
)

// WithTransaction is a helper function to execute multiple repository operations in a transaction
// Usage:
//
//	err := repository.WithTransaction(ctx, db, func(tx TxConn) error {
//	    // All operations here use the same transaction
//	    if err := repo.CreateWithTx(ctx, tx, activity); err != nil {
//	        return err
//	    }
//	    if err := repo.UpdateWithTx(ctx, tx, otherActivity); err != nil {
//	        return err
//	    }
//	    return nil
//	})
func WithTransaction(ctx context.Context, db DBConn, fn func(tx TxConn) error) error {
	// Type assert to LoggingDB to get transaction support
	loggingDB, ok := db.(*database.LoggingDB)
	if !ok {
		return sql.ErrConnDone // Return error if not LoggingDB
	}

	loggingTx, err := loggingDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Execute the function with the transaction
	err = fn(loggingTx)
	if err != nil {
		_ = loggingTx.Rollback()
		return err
	}

	// Commit if everything succeeded
	return loggingTx.Commit()
}

// ExecInTx executes a query using either the provided transaction or direct DB
func ExecInTx(ctx context.Context, tx TxConn, db DBConn, query string, args ...interface{}) (sql.Result, error) {
	if tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return db.ExecContext(ctx, query, args...)
}

// QueryRowInTx executes a query using either the provided transaction or direct DB
func QueryRowInTx(ctx context.Context, tx TxConn, db DBConn, query string, args ...interface{}) *sql.Row {
	if tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return db.QueryRowContext(ctx, query, args...)
}

// QueryInTx executes a query using either the provided transaction or direct DB
func QueryInTx(ctx context.Context, tx TxConn, db DBConn, query string, args ...interface{}) (*sql.Rows, error) {
	if tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}
	return db.QueryContext(ctx, query, args...)
}
