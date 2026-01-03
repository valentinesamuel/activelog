package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// LoggingDB wraps *sql.DB to log all queries
type LoggingDB struct {
	*sql.DB
	logger *log.Logger
}

// NewLoggingDB creates a new logging database wrapper
func NewLoggingDB(db *sql.DB, logger *log.Logger) *LoggingDB {
	if logger == nil {
		logger = log.Default()
	}
	return &LoggingDB{
		DB:     db,
		logger: logger,
	}
}

// QueryContext wraps db.QueryContext with logging
func (db *LoggingDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	db.logQuery("QUERY", query, args, duration, err)
	return rows, err
}

// QueryRowContext wraps db.QueryRowContext with logging
func (db *LoggingDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	db.logQuery("QUERY ROW", query, args, duration, nil)
	return row
}

// ExecContext wraps db.ExecContext with logging
func (db *LoggingDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	db.logQuery("EXEC", query, args, duration, err)
	return result, err
}

// BeginTx wraps db.BeginTx with logging
func (db *LoggingDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*LoggingTx, error) {
	start := time.Now()
	tx, err := db.DB.BeginTx(ctx, opts)
	duration := time.Since(start)

	if err != nil {
		db.logger.Printf("❌ BEGIN TRANSACTION failed: %v (took %v)", err, duration)
		return nil, err
	}

	db.logger.Printf("✅ BEGIN TRANSACTION (took %v)", duration)
	return &LoggingTx{Tx: tx, logger: db.logger}, nil
}

// logQuery logs the query with formatted output
func (db *LoggingDB) logQuery(queryType, query string, args []interface{}, duration time.Duration, err error) {
	status := "✅"
	if err != nil {
		status = "❌"
	}

	// Color code by duration
	durationStr := formatDuration(duration)

	db.logger.Printf("%s [%s] %s | %s | args: %v",
		status,
		queryType,
		durationStr,
		formatQuery(query),
		formatArgs(args),
	)

	if err != nil {
		db.logger.Printf("   └─ Error: %v", err)
	}
}

// LoggingTx wraps *sql.Tx to log transaction operations
type LoggingTx struct {
	*sql.Tx
	logger *log.Logger
}

// QueryContext wraps tx.QueryContext with logging
func (tx *LoggingTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := tx.Tx.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	tx.logQuery("TX QUERY", query, args, duration, err)
	return rows, err
}

// QueryRowContext wraps tx.QueryRowContext with logging
func (tx *LoggingTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := tx.Tx.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	tx.logQuery("TX QUERY ROW", query, args, duration, nil)
	return row
}

// ExecContext wraps tx.ExecContext with logging
func (tx *LoggingTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := tx.Tx.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	tx.logQuery("TX EXEC", query, args, duration, err)
	return result, err
}

// Commit wraps tx.Commit with logging
func (tx *LoggingTx) Commit() error {
	start := time.Now()
	err := tx.Tx.Commit()
	duration := time.Since(start)

	if err != nil {
		tx.logger.Printf("❌ COMMIT failed: %v (took %v)", err, duration)
		return err
	}

	tx.logger.Printf("✅ COMMIT (took %v)", duration)
	return nil
}

// Rollback wraps tx.Rollback with logging
func (tx *LoggingTx) Rollback() error {
	start := time.Now()
	err := tx.Tx.Rollback()
	duration := time.Since(start)

	if err != nil && err != sql.ErrTxDone {
		tx.logger.Printf("⚠️  ROLLBACK warning: %v (took %v)", err, duration)
		return err
	}

	tx.logger.Printf("↩️  ROLLBACK (took %v)", duration)
	return nil
}

// logQuery logs transaction queries
func (tx *LoggingTx) logQuery(queryType, query string, args []interface{}, duration time.Duration, err error) {
	status := "✅"
	if err != nil {
		status = "❌"
	}

	durationStr := formatDuration(duration)

	tx.logger.Printf("%s [%s] %s | %s | args: %v",
		status,
		queryType,
		durationStr,
		formatQuery(query),
		formatArgs(args),
	)

	if err != nil {
		tx.logger.Printf("   └─ Error: %v", err)
	}
}

func formatQuery(query string) string {
	query = fmt.Sprintf("%s", query)
	// if len(query) > 100 {
	// 	return query[:100] + "..."
	// }
	return query
}

func formatArgs(args []interface{}) string {
	if len(args) == 0 {
		return "[]"
	}
	return fmt.Sprintf("%v", args)
}

func formatDuration(d time.Duration) string {
	if d > 1*time.Second {
		return fmt.Sprintf("🐢 %v", d) // Slow query
	} else if d > 100*time.Millisecond {
		return fmt.Sprintf("⚠️  %v", d) // Warning
	}
	return fmt.Sprintf("⚡ %v", d) // Fast
}
