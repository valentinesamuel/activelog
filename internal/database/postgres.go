package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Connect establishes a database connection and wraps it with logging
func Connect(databaseUrl string) (*LoggingDB, error) {
	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("❌ Error opening a connection to the db: \n🛑 %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("❌ Error connecting to the db: \n🛑 %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Always wrap with logging for consistency
	logger := log.New(os.Stdout, "[SQL] ", log.LstdFlags)
	loggingDB := NewLoggingDB(db, logger)

	log.Println("✅ Successfully connected to database")
	log.Println("🔍 Query logging enabled")

	return loggingDB, nil
}
