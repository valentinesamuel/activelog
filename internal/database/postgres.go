package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func Connect(databaseUrl string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("❌ Error opening a coonnection to the db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("❌ Error connecting to the db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("✅ Successfully connected to database")
	return db, nil
}
