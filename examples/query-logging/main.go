package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/valentinesamuel/activelog/pkg/database"
)

func main() {
	// 1. Connect to database normally
	db, err := sql.Open("postgres", "postgres://activelog_user:password@localhost:5432/activelog_dev?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Wrap with logging (enable in development only)
	logger := log.New(os.Stdout, "[SQL] ", log.LstdFlags|log.Lshortfile)
	loggedDB := database.NewLoggingDB(db, logger)

	// 3. Use loggedDB instead of db everywhere!
	ctx := context.Background()

	// Example 1: Simple query
	log.Println("\n=== Example 1: Simple Query ===")
	var count int
	err = loggedDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM activities WHERE user_id = $1", 1).Scan(&count)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Example 2: Transaction with multiple queries
	log.Println("\n=== Example 2: Transaction ===")
	tx, err := loggedDB.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Insert activity
	_, err = tx.ExecContext(ctx, `
		INSERT INTO activities (user_id, activity_type, duration_minutes, activity_date)
		VALUES ($1, $2, $3, $4)
	`, 1, "running", 30, time.Now())
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	// Insert tag
	_, err = tx.ExecContext(ctx, `
		INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO NOTHING
	`, "morning")
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	// Commit
	tx.Commit()

	// Example 3: N+1 Query Problem (BAD)
	log.Println("\n=== Example 3: N+1 Query Problem (watch the query count!) ===")
	rows, _ := loggedDB.QueryContext(ctx, "SELECT id FROM activities WHERE user_id = $1 LIMIT 5", 1)
	defer rows.Close()

	for rows.Next() {
		var activityID int
		rows.Scan(&activityID)

		// This creates N queries! (N+1 problem)
		loggedDB.QueryContext(ctx, "SELECT name FROM tags JOIN activity_tags ON tags.id = activity_tags.tag_id WHERE activity_tags.activity_id = $1", activityID)
	}

	log.Println("\n=== Notice: 1 query for activities + N queries for tags = N+1 queries! ===")
}
