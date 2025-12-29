package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"github.com/valentinesamuel/activelog/internal/repository"
)

// setupTestDB creates a test database connection and runs migrations
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper() // Mark this as a test helper function

	// Connect to test database
	dsn := "postgres://activelog_user:activelog@localhost:5444/activelog_test?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Create users table if it doesn't exist (simple schema)
	_, err = db.Exec(`
                CREATE TABLE IF NOT EXISTS users (
                        id SERIAL PRIMARY KEY,
                        email VARCHAR(255) UNIQUE NOT NULL,
                        password_hash VARCHAR(255) NOT NULL,
                        name VARCHAR(255),
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
        `)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	return db
}

// cleanupTestDB removes all data from tables
func cleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	// Clean up all tables before each test
	_, err := db.Exec("TRUNCATE users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Failed to clean up database: %v", err)
	}
}

func TestRegistration(t *testing.T) {
	// Setup: Create handler with real database
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	handler := NewUserHandler(userRepo) // Adjust based on your handler name

	tests := []struct {
		name       string
		email      string
		password   string
		wantStatus int
	}{
		{"valid registration", "test@example.com", "SecurePass123!", 201},
		{"duplicate email", "test@example.com", "SecurePass123!", 409},
		{"weak password", "test2@example.com", "123", 400},
		{"invalid email", "notanemail", "SecurePass123!", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean database before each test (except duplicate email test)
			if tt.name != "duplicate email" {
				cleanupTestDB(t, db)
			}

			// --- A. Prepare the Request Body ---
			payload := map[string]string{
				"email":    tt.email,
				"password": tt.password,
			}

			body, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal JSON: %v", err)
			}
			reqBody := bytes.NewReader(body)

			// --- B. Create Mock HTTP Objects ---
			req := httptest.NewRequest("POST", "/auth/register", reqBody)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			// --- C. Execute the Handler ---
			handler.CreateUser(rr, req) // Call your Register method

			// --- D. Assertions ---
			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v\nResponse body: %s",
					status, tt.wantStatus, rr.Body.String())
			}

			// Additional assertion for successful registration
			if tt.wantStatus == http.StatusCreated {
				var response map[string]interface{}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				// Check that user data is returned
				if response["user"] == nil {
					t.Error("Expected 'user' field in response")
				}
			}
		})
	}
}
