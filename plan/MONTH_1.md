# Month 1: Foundation - Building Your First Go API

**Duration:** Weeks 1-4  
**Goal:** Go from "I only know main.go" to "I built a working REST API with database integration"  
**Total Time Investment:** 28-40 hours over 4 weeks

---

## Month 1 Overview

By the end of Month 1, you will have:

✅ Set up a proper Go development environment  
✅ Understood Go project structure and organization  
✅ Built a REST API with multiple endpoints  
✅ Integrated PostgreSQL database  
✅ Written your first Go tests  
✅ Handled errors the Go way  
✅ Created proper API documentation

**What You're Building:** The foundation of ActiveLog - a simple API where users can register and log activities.

---

## Pre-Month 1: Setup (Do this before Week 1 starts)

### Environment Setup Checklist

**Install Go:**
```bash
# Download from https://go.dev/dl/
# Verify installation
go version  # Should show 1.21 or higher
```

**Install PostgreSQL:**
```bash
# macOS
brew install postgresql@15
brew services start postgresql@15

# Ubuntu/Debian
sudo apt-get install postgresql-15

# Windows
# Download from https://www.postgresql.org/download/windows/
```

**Install Essential Tools:**
```bash
# Database migration tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Air for hot reload (optional but helpful)
go install github.com/cosmtrek/air@latest

# VSCode Go extension (if using VSCode)
# Or GoLand IDE (trial or paid)
```

**Create Project Directory:**
```bash
mkdir -p ~/projects/activelog
cd ~/projects/activelog
git init
```

**Create PostgreSQL Database:**
```bash
# Connect to PostgreSQL
psql postgres

# In psql prompt:
CREATE DATABASE activelog_dev;
CREATE USER activelog_user WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE activelog_dev TO activelog_user;
\q
```

**Setup Complete!** You're ready for Week 1.

---

# WEEK 1: Go Basics & Project Structure

## Week 1 Goal
Understand Go fundamentals and set up a proper project structure. By Sunday, you'll have a "Hello World" HTTP server running with proper organization.

### Learning Objectives
- Go syntax basics (variables, functions, types)
- Project structure conventions
- Go modules and dependency management
- Basic HTTP server with standard library
- Package organization

### What You'll Build
A basic HTTP server that responds to requests with JSON. Not connected to a database yet, just proving you can structure a Go project and handle HTTP.

---

## Week 1, Monday Evening (45 min)

**Time:** 45 minutes  
**Focus:** Go syntax fundamentals

### Tasks:

**1. Complete Go Tour Basics (30 min)**
- Go to https://go.dev/tour/basics/1
- Complete sections 1-8: Packages through Functions
- Don't take notes, just read and run examples
- Focus on: how variables work, how functions are declared

**2. Read "Effective Go" Introduction (15 min)**
- https://go.dev/doc/effective_go
- Just read the introduction and "Formatting" section
- Notice: Go is opinionous about style (unlike JavaScript)

### Expected Outcome:
You should understand:
- How to declare variables (`var name string` vs `name := "value"`)
- How functions work (return types come after parameters)
- That Go has strong opinions about formatting

### Notes Section (write these down):
```
Key differences from JavaScript I noticed:
1. Everything must be done in a certain way
2. Looks like C
3. 

Questions I have:
1. 
2. 


Observations I made:
1. You must use `var` outside a function, you can use `:=`
2. 
```

---

## Week 1, Wednesday Evening (45 min)

**Time:** 45 minutes  
**Focus:** Project structure & Go modules

### Tasks:

**1. Initialize Go Module (5 min)**
```bash
cd ~/projects/activelog
go mod init github.com/yourusername/activelog

# This creates go.mod file - like package.json
```

**2. Create Project Structure (10 min)**
```bash
# Create the directory structure
mkdir -p cmd/api 
mkdir -p internal/handlers
mkdir -p internal/models
mkdir -p pkg/response
touch cmd/api/main.go
touch internal/handlers/health.go
touch pkg/response/json.go
touch README.md
```

**3. Read About Go Project Structure (15 min)**
- Read: https://github.com/golang-standards/project-layout
- Focus on understanding:
  - `cmd/` - application entry points
  - `internal/` - private application code
  - `pkg/` - code that's OK for others to import
- Don't worry if it's not 100% clear yet

**4. Create Your First main.go (15 min)**

In `cmd/api/main.go`:
```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting ActiveLog API...")

	// Create a simple HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to ActiveLog API"}`))
	})

	// Start server on port 8080
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
```

**5. Run It! (5 min)**
```bash
cd ~/projects/activelog
go run cmd/api/main.go

# In another terminal:
curl http://localhost:8080

# You should see: {"message": "Welcome to ActiveLog API"}
```

### Expected Outcome:
- You have a proper Go project structure
- You understand go.mod (it's like package.json)
- You ran your first Go HTTP server
- You're starting to see how Go organizes code differently from JavaScript

### Notes Section:
```
What worked:
- http server

What confused me:
-  func(w http.ResponseWriter, r *http.Request){}

Observation:
- go.mod == package.json
- func(w http.ResponseWriter, r *http.Request){} === app.get('/', (req, res) => { ... })
- not to call the ListenAndServe inside the HandleFunc method if not, the server would start and close immediately

Questions for Saturday:
- 
```

---

## Week 1, Saturday Morning (3-4 hours)

**Time:** 3-4 hours  
**Focus:** Build a proper HTTP server with routing and JSON responses

### Part 1: Understanding HTTP Handling (45 min)

**Read and Experiment:**

1. **Read about http.Handler interface (15 min)**
   - https://pkg.go.dev/net/http#Handler
   - Key concept: Everything in Go HTTP is a Handler
   - In JavaScript/Express, you have middleware functions
   - In Go, you have Handlers that implement ServeHTTP method

2. **Create a custom handler (30 min)**

In `internal/handlers/health.go`:
```go
package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// ServeHTTP implements the http.Handler interface
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")
	
	// Create response
	response := map[string]string{
		"status": "healthy",
		"service": "activelog-api",
	}
	
	// Encode and send
	json.NewEncoder(w).Encode(response)
}
```

**Key Learning:** This is how Go handles HTTP. Unlike JavaScript where you'd do `res.json({...})`, in Go you:
1. Set headers manually
2. Encode data to JSON manually
3. Write to response writer

### Part 1.5: Understanding Interfaces Explicitly (30 min)

**🔴 CRITICAL CONCEPT: Interfaces are fundamental to Go**

You just implemented the `http.Handler` interface without knowing it! Let's understand what interfaces are in Go.

**What is an interface in Go?**

In `internal/handlers/interface_example.go` (create this file):
```go
package handlers

import "fmt"

// ResponseWriter is an interface - it defines behavior, not data
type ResponseWriter interface {
	Write(data []byte) error
	SetStatus(code int)
}

// JSONResponse implements the ResponseWriter interface
type JSONResponse struct {
	statusCode int
}

func (j *JSONResponse) Write(data []byte) error {
	fmt.Printf("Writing JSON: %s\n", string(data))
	return nil
}

func (j *JSONResponse) SetStatus(code int) {
	j.statusCode = code
}

// XMLResponse also implements the same interface
type XMLResponse struct {
	statusCode int
}

func (x *XMLResponse) Write(data []byte) error {
	fmt.Printf("Writing XML: %s\n", string(data))
	return nil
}

func (x *XMLResponse) SetStatus(code int) {
	x.statusCode = code
}

// SendResponse works with ANY type that implements ResponseWriter
func SendResponse(w ResponseWriter, data []byte) {
	w.SetStatus(200)
	w.Write(data)
}
```

**Key Differences from JavaScript/TypeScript:**

| Concept | JavaScript/TypeScript | Go |
|---------|----------------------|-----|
| Interface | Type annotation only | Defines behavior contract |
| Implementation | Explicit (`implements`) | **Implicit** (just match the methods!) |
| Runtime | No runtime checking | Interface satisfaction at compile time |
| Purpose | Type safety | Polymorphism + testability |

**Why HealthHandler implements http.Handler:**

```go
// http.Handler is defined in standard library as:
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}

// Your HealthHandler has this method:
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // ...
}

// Therefore, HealthHandler automatically implements Handler!
// No "implements" keyword needed!
```

**Why This Matters for Testing:**

```go
// You can create a mock repository interface
type ActivityRepository interface {
    Create(activity *Activity) error
    GetByID(id int) (*Activity, error)
}

// Real implementation
type PostgresActivityRepo struct {
    db *sql.DB
}

func (r *PostgresActivityRepo) Create(activity *Activity) error {
    // Real database logic
}

// Mock implementation for testing
type MockActivityRepo struct {
    activities []*Activity
}

func (m *MockActivityRepo) Create(activity *Activity) error {
    // Fake logic - no database needed!
    m.activities = append(m.activities, activity)
    return nil
}

// Handler works with EITHER implementation
type ActivityHandler struct {
    repo ActivityRepository  // Interface, not concrete type!
}
```

**Quick Exercise (5 min):**

1. What methods does `http.Handler` require?
   - Answer: Just `ServeHTTP(ResponseWriter, *Request)`

2. Can you have a struct implement multiple interfaces?
   - Answer: Yes! A struct can implement as many interfaces as it wants

3. What happens if you're missing a method?
   - Answer: Compile error - "does not implement interface"

**Expected Outcome:**
- ✅ You understand interfaces are implicit, not explicit
- ✅ You know why `http.Handler` is an interface
- ✅ You see how interfaces enable testing (mocks)
- ✅ You recognize this is VERY different from JavaScript

### Part 2: Add Router (1 hour)

**Why we need a router:**
- Standard library routing is basic (`/` matches everything)
- We need proper path matching (`/api/v1/health` vs `/api/v1/users`)
- We'll use `gorilla/mux` (popular, simple) or `chi` (more modern)

**Install gorilla/mux:**
```bash
go get -u github.com/gorilla/mux
```

**Update cmd/api/main.go:**
```go
package main

import (
	"log"
	"net/http"
	
	"github.com/gorilla/mux"
	"github.com/yourusername/activelog/internal/handlers"
)

func main() {
	log.Println("Starting ActiveLog API...")

	// Create router
	router := mux.NewRouter()
	
	// Create handlers
	healthHandler := handlers.NewHealthHandler()
	
	// Register routes
	router.Handle("/health", healthHandler).Methods("GET")
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "ActiveLog API v1"}`))
	}).Methods("GET")
	
	// Start server
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
```

**Test it:**
```bash
go run cmd/api/main.go

# In another terminal:
curl http://localhost:8080/health
# Should see: {"status":"healthy","service":"activelog-api"}
```

### Part 3: Create Reusable JSON Response Helper (45 min)

**Create pkg/response/json.go:**
```go
package response

import (
	"encoding/json"
	"net/http"
)

// JSON sends a JSON response
func JSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// Error sends a JSON error response
func Error(w http.ResponseWriter, statusCode int, message string) error {
	return JSON(w, statusCode, map[string]string{
		"error": message,
	})
}
```

**Update health handler to use it:**
```go
package handlers

import (
	"net/http"
	"github.com/yourusername/activelog/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "healthy",
		"service": "activelog-api",
	}
	response.JSON(w, http.StatusOK, data)
}
```

### Part 4: Add More Routes (45 min)

**Create internal/handlers/activity.go:**
```go
package handlers

import (
	"net/http"
	"github.com/yourusername/activelog/pkg/response"
)

type ActivityHandler struct{}

func NewActivityHandler() *ActivityHandler {
	return &ActivityHandler{}
}

// ListActivities handles GET /activities
func (h *ActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	// For now, return mock data
	activities := []map[string]interface{}{
		{
			"id":   1,
			"type": "running",
			"distance": 5.2,
			"duration": 30,
		},
		{
			"id":   2,
			"type": "basketball",
			"duration": 60,
		},
	}
	
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
		"count": len(activities),
	})
}

// CreateActivity handles POST /activities
func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	// For now, just acknowledge receipt
	response.JSON(w, http.StatusCreated, map[string]string{
		"message": "Activity created (mock)",
	})
}
```

**Update cmd/api/main.go to include activity routes:**
```go
package main

import (
	"log"
	"net/http"
	
	"github.com/gorilla/mux"
	"github.com/yourusername/activelog/internal/handlers"
)

func main() {
	log.Println("Starting ActiveLog API...")

	router := mux.NewRouter()
	
	// Create handlers
	healthHandler := handlers.NewHealthHandler()
	activityHandler := handlers.NewActivityHandler()
	
	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Health check
	router.Handle("/health", healthHandler).Methods("GET")
	
	// Activity routes
	api.HandleFunc("/activities", activityHandler.ListActivities).Methods("GET")
	api.HandleFunc("/activities", activityHandler.CreateActivity).Methods("POST")
	
	// Root
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "ActiveLog API v1", "version": "0.1.0"}`))
	}).Methods("GET")
	
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
```

**Test all routes:**
```bash
# Start server
go run cmd/api/main.go

# In another terminal, test each endpoint:
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/activities
curl -X POST http://localhost:8080/api/v1/activities
```

### Expected Outcome After Saturday:
- ✅ You have a structured Go project
- ✅ You understand handlers and routing
- ✅ You can create JSON responses
- ✅ You have multiple API endpoints (even if they return mock data)
- ✅ You're seeing Go patterns emerge

### Saturday Reflection (15 min):
```
What I built today:
- the reusable JSON sender
- the activity route

What made sense:
- the handlers and route flow

What's still confusing:
- why the health handler needed to have a serve http but the activity handler did not need to have
- the difference between Handle and HandleFunc

How does this compare to Express/JavaScript?
- the intention and logic is almost the same. Just a few differnces like Handle and HandleFunc
```

---

## Week 1, Sunday (2-3 hours)

**Time:** 2-3 hours  
**Focus:** Testing, documentation, and cleanup

### Part 1: Write Your First Tests (1 hour)

**Create internal/handlers/health_test.go:**
```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	// Create handler
	handler := NewHealthHandler()
	
	// Create test request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create response recorder (captures response)
	rr := httptest.NewRecorder()
	
	// Call the handler
	handler.ServeHTTP(rr, req)
	
	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	
	// Check content type
	expectedContentType := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ct, expectedContentType)
	}
	
	// Check body contains expected fields
	body := rr.Body.String()
	if !contains(body, "status") {
		t.Error("response body doesn't contain 'status' field")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		   (s == substr || len(s) >= len(substr) && 
		   containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

**Run tests:**
```bash
go test ./internal/handlers/...
```

**Understanding Go testing:**
- Test files end with `_test.go`
- Test functions start with `Test`
- Use `t.Error()` or `t.Fatal()` to fail tests
- `httptest` package lets you test HTTP handlers without starting server
- No test framework needed - it's built into Go!

### Part 2: Document Your API (45 min)

**Update README.md:**
```markdown
# ActiveLog API

Personal multi-sport activity tracking API built with Go.

## Project Status
**Week 1 Complete** - Basic API structure with mock endpoints

## Current Features
- Health check endpoint
- Mock activity listing
- Mock activity creation
- JSON response handling
- Proper project structure

## Tech Stack
- Go 1.21+
- gorilla/mux for routing
- PostgreSQL (coming in Week 2)

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/activelog.git
cd activelog
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run cmd/api/main.go
```

Server will start on `http://localhost:8080`

## API Endpoints

### Health Check
```bash
GET /health
```

Response:
```json
{
  "status": "healthy",
  "service": "activelog-api"
}
```

### List Activities (Mock)
```bash
GET /api/v1/activities
```

Response:
```json
{
  "activities": [...],
  "count": 2
}
```

### Create Activity (Mock)
```bash
POST /api/v1/activities
```

Response:
```json
{
  "message": "Activity created (mock)"
}
```

## Project Structure

```
activelog/
├── cmd/
│   └── api/              # Application entry point
│       └── main.go
├── internal/
│   ├── handlers/         # HTTP request handlers
│   └── models/           # Data models (coming soon)
├── pkg/
│   └── response/         # Reusable response utilities
├── go.mod
├── go.sum
└── README.md
```

## Development

### Running Tests
```bash
go test ./...
```

### Code Formatting
```bash
go fmt ./...
```

## Roadmap

### Week 1 ✅
- [x] Project structure
- [x] Basic HTTP server
- [x] Routing with gorilla/mux
- [x] JSON response helpers
- [x] Basic tests

### Week 2 (Next)
- [ ] PostgreSQL integration
- [ ] Database models
- [ ] Real CRUD operations
- [ ] Error handling

## Learning Notes

This is a learning project following a 12-month Go mastery plan.

### Key Learnings - Week 1
- Go project structure is more opinionated than JavaScript
- Handlers implement ServeHTTP interface
- Testing is built into Go (no Jest/Mocha needed)
- Error handling is explicit (no try-catch)

## License
MIT

## Contact
Your Name - [@yourhandle](https://twitter.com/yourhandle)
```

### Part 3: Clean Up and Commit (30 min)

**Create .gitignore:**
```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test coverage
*.out
coverage.txt

# Go workspace file
go.work

# Environment variables
.env
.env.local

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Database
*.db
```

**Format all code:**
```bash
go fmt ./...
```

**Run all tests:**
```bash
go test ./...
```

**Commit your work:**
```bash
git add .
git commit -m "Week 1: Basic API structure with routing and JSON responses"
git tag week-1
```

### Expected Outcome After Sunday:
- ✅ You have tests (and understand Go testing)
- ✅ Your code is documented
- ✅ Everything is formatted properly
- ✅ You have a clean git history
- ✅ You can show someone what you built

---

## Week 1 Wrap-Up

### What You Accomplished
You went from "I only know main.go" to having a properly structured Go API with:
- Multiple endpoints
- Handler pattern
- JSON responses
- Tests
- Documentation

### Key Go Concepts Learned
1. **Project structure** - cmd/, internal/, pkg/ organization
2. **Go modules** - go.mod is like package.json
3. **HTTP handlers** - ServeHTTP interface pattern
4. **Routing** - Using gorilla/mux for path matching
5. **Testing** - Built-in testing with httptest package
6. **No classes** - But we have types with methods

### JavaScript → Go Mental Model
| JavaScript/Express | Go Equivalent |
|-------------------|---------------|
| `app.get('/path', handler)` | `router.HandleFunc("/path", handler).Methods("GET")` |
| `res.json({...})` | `json.NewEncoder(w).Encode(...)` |
| `try { } catch (e) { }` | `if err != nil { }` (explicit) |
| Jest/Mocha | Built-in `testing` package |
| `package.json` | `go.mod` |

### Reflection Questions (Answer these honestly)

**What felt easy?**
```
Your answer:

```

**What felt hard?**
```
Your answer:

```

**What's still confusing?**
```
Your answer:

```

**How confident do you feel (1-10)?**
```
Your rating: __/10

Why this rating:

```

**Ready for Week 2?**
```
Yes / Need more practice

If more practice needed, what specifically:

```

---

## Week 1 Troubleshooting

### Common Issues

**Issue: "go: cannot find main module"**
```bash
# Solution: Make sure you're in project root and ran go mod init
cd ~/projects/activelog
go mod init github.com/yourusername/activelog
```

**Issue: "imported and not used"**
```
# Go is strict about unused imports
# Solution: Remove unused imports or use goimports
go get golang.org/x/tools/cmd/goimports
goimports -w .
```

**Issue: "undefined: mux"**
```bash
# Solution: Install dependency
go get -u github.com/gorilla/mux
go mod tidy
```

**Issue: Port 8080 already in use**
```bash
# Find what's using it
lsof -i :8080  # macOS/Linux
netstat -ano | findstr :8080  # Windows

# Kill it or use different port in main.go
```

---

## Preparation for Week 2

Before starting Week 2, make sure you have:
- ✅ Completed all Week 1 tasks
- ✅ All tests passing
- ✅ Code committed to git
- ✅ PostgreSQL installed and running
- ✅ activelog_dev database created

**Next week preview:** We'll connect to PostgreSQL, create database tables, and replace those mock endpoints with real data operations.

---

*End of Week 1 - Great work! You're not dumb, you're learning. Keep going.* 🚀

---
---

# WEEK 2: Database Integration & Models

## Week 2 Goal
Connect to PostgreSQL, create database tables, and implement real CRUD operations. By Sunday, mock data is gone - everything reads/writes to the database.

### Learning Objectives
- PostgreSQL connection and connection pooling
- Database migrations
- Structs as data models
- SQL queries in Go (using database/sql)
- Pointer receivers on methods
- Error handling with database operations

### What You'll Build
Real database-backed endpoints:
- Create users table and user registration
- Create activities table
- POST /api/v1/activities (store in DB)
- GET /api/v1/activities (fetch from DB)
- GET /api/v1/activities/:id (fetch single activity)

---

## Week 2, Monday Evening (45 min)

**Time:** 45 minutes  
**Focus:** Understanding database/sql package and postgres driver

### Tasks:

**1. Read about database/sql (20 min)**
- Go to: https://go.dev/doc/database/
- Focus on: "Opening a database", "Executing queries"
- Key concept: database/sql is the interface, you need a driver (like pq for postgres)

**2. Install PostgreSQL driver (5 min)**
```bash
cd ~/projects/activelog
go get github.com/lib/pq
```

**3. Create database configuration (20 min)**

**Create internal/config/config.go:**
```go
package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	ServerPort  string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://activelog_user:your_secure_password@localhost/activelog_dev?sslmode=disable"),
		ServerPort:  getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

**Create internal/database/postgres.go:**
```go
package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // Import postgres driver
)

// Connect creates a connection to PostgreSQL
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// 🔴 CRITICAL: Configure connection pool
	// Without this, your app will have connection problems under load!
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                  // Maximum number of idle connections in pool
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum amount of time a connection may be reused

	log.Println("Successfully connected to database")
	return db, nil
}
```

**Import time package at top:**
```go
import (
	"database/sql"
	"fmt"
	"log"
	"time"  // Add this

	_ "github.com/lib/pq"
)
```

**Key Go Concepts:**
- `_ "github.com/lib/pq"` - blank import (imports for side effects only)
- `%w` in fmt.Errorf - wraps errors (Go 1.13+)
- Pointer return (`*sql.DB`) - database connections are shared

**🔴 Understanding Connection Pooling:**

```go
// What is connection pooling?
// Opening/closing DB connections is SLOW (network + auth)
// Instead, sql.DB maintains a POOL of reusable connections

// MaxOpenConns = 25 means:
// - At most 25 concurrent database queries
// - 26th query will WAIT for a connection to be available
// - Too high = too many connections to database
// - Too low = requests wait unnecessarily

// MaxIdleConns = 5 means:
// - Keep 5 connections "warm" even when not in use
// - Avoids overhead of opening new connections
// - Too high = wasted resources
// - Too low = frequent reconnections

// ConnMaxLifetime = 5 minutes means:
// - Recycle connections every 5 minutes
// - Prevents stale connections
// - Good for databases behind load balancers
```

**Why this matters:**

Without connection pooling configuration:
- ❌ Defaults to unlimited connections (can overwhelm DB)
- ❌ Connections never recycled (stale connection issues)
- ❌ Poor performance under load

With proper configuration:
- ✅ Controlled resource usage
- ✅ Better performance
- ✅ Handles load gracefully
- ✅ Prevents connection leaks

### Expected Outcome:
- You understand Go's database/sql package
- You know why we need a driver
- You have database connection code ready

### Notes:
```
How does this compare to JavaScript database libraries?
- 

Questions:
- 
```

---

## Week 2, Wednesday Evening (45 min)

**Time:** 45 minutes  
**Focus:** Database migrations setup

### Tasks:

**1. Understand Database Migrations (15 min)**
- Migrations = version control for your database schema
- Up migrations: create/modify tables
- Down migrations: undo changes
- Like git for your database structure

**2. Create migrations directory structure (5 min)**
```bash
cd ~/projects/activelog
mkdir -p migrations
```

**3. Create first migration: users table (25 min)**

**Create migrations/000001_create_users_table.up.sql:**
```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

**Create migrations/000001_create_users_table.down.sql:**
```sql
DROP TABLE IF EXISTS users;
```

**Create migrations/000002_create_activities_table.up.sql:**
```sql
CREATE TABLE IF NOT EXISTS activities (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL,
    title VARCHAR(255),
    description TEXT,
    duration_minutes INTEGER,
    distance_km DECIMAL(10, 2),
    calories_burned INTEGER,
    notes TEXT,
    activity_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_activities_date ON activities(activity_date);
CREATE INDEX idx_activities_type ON activities(activity_type);
```

**Create migrations/000002_create_activities_table.down.sql:**
```sql
DROP TABLE IF EXISTS activities;
```

**4. Run migrations manually (for now):**
```bash
# Connect to database
psql activelog_dev -U activelog_user

# Copy and paste the content of 000001_create_users_table.up.sql
# Then copy and paste the content of 000002_create_activities_table.up.sql

# Verify tables were created:
\dt

# You should see: users, activities
```

**Alternative: Use migrate tool:**
```bash
migrate -path migrations -database "postgres://activelog_user:your_secure_password@localhost/activelog_dev?sslmode=disable" up
```

### Expected Outcome:
- You have a migrations directory
- Your database has users and activities tables
- You understand up/down migrations

### Notes:
```
Why migrations instead of just CREATE TABLE in code?
- 

What happens if I run the up migration twice?
- 
```

---

## Week 2, Thursday Evening (45 min)

**Time:** 45 minutes
**Focus:** Understanding Pointers in Go

**🔴 CRITICAL CONCEPT: Pointers are everywhere in Go**

You've been using pointers (`*sql.DB`, `&Activity{}`) without fully understanding them. Let's fix that.

### Tasks:

**1. Read about pointers (15 min)**
- Go to: https://go.dev/tour/moretypes/1
- Complete the Pointers section
- Key concept: A pointer holds the memory address of a value

**2. Understand pointer syntax (30 min)**

**Create a practice file `internal/practice/pointers.go`:**
```go
package practice

import "fmt"

// Value vs Pointer demonstration
type Counter struct {
	count int
}

// Value receiver - receives a COPY
func (c Counter) IncrementValue() {
	c.count++  // This modifies the COPY, not the original!
	fmt.Printf("Inside IncrementValue: %d\n", c.count)
}

// Pointer receiver - receives the ACTUAL struct
func (c *Counter) IncrementPointer() {
	c.count++  // This modifies the ORIGINAL!
	fmt.Printf("Inside IncrementPointer: %d\n", c.count)
}

func DemoPointers() {
	// Value semantics
	counter1 := Counter{count: 0}
	counter1.IncrementValue()
	fmt.Printf("After value increment: %d\n", counter1.count)  // Still 0!

	// Pointer semantics
	counter2 := Counter{count: 0}
	counter2.IncrementPointer()  // Go automatically takes address (&counter2)
	fmt.Printf("After pointer increment: %d\n", counter2.count)  // Now 1!
}

// Example: Why database uses pointers
func WhyDatabaseUsesPointers() {
	// sql.DB is expensive to copy (has connection pool, mutexes, etc.)
	// Always use pointer to avoid copying
	// db := sql.Open(...)  // Returns *sql.DB (pointer!)
}
```

**Run it:**
```bash
cd ~/projects/activelog
# Add to cmd/api/main.go temporarily:
// practice.DemoPointers()
go run cmd/api/main.go
```

**Key Rules:**

| Scenario | Use Value | Use Pointer |
|----------|-----------|-------------|
| Small structs (< 3 fields) | ✅ Usually | Sometimes |
| Large structs | ❌ Never | ✅ Always |
| Need to modify | ❌ No | ✅ Yes |
| Expensive to copy | ❌ No | ✅ Yes |
| Methods that change state | ❌ No | ✅ Yes |

**When to use `*` vs `&`:**

```go
// & (address-of operator) - Get pointer FROM value
activity := Activity{ID: 1}
pointer := &activity  // pointer is *Activity

// * (dereference operator) - Get value FROM pointer
value := *pointer     // value is Activity

// In function signatures:
func Create(a *Activity) error {  // Accepts a pointer
    // ...
}

// Calling it:
activity := Activity{}
Create(&activity)  // Pass address with &
```

**Why `*sql.DB` is always a pointer:**

```go
// Inside database/sql package:
type DB struct {
    connector driver.Connector
    mu        sync.Mutex      // Mutex can't be copied!
    freeConn  []*driverConn   // Connection pool
    // ... many more fields
}

// If we used DB instead of *DB:
// 1. Every function call would COPY the entire connection pool!
// 2. Mutexes would break (can't copy mutex)
// 3. Memory usage would explode

// That's why it's ALWAYS *sql.DB
```

**Common Pointer Mistakes:**

```go
// ❌ WRONG - Pointer to pointer
func Create(a **Activity) error  // No! Just *Activity

// ❌ WRONG - Forgetting &
activity := Activity{}
Create(activity)  // Won't compile if Create expects *Activity

// ✅ CORRECT
activity := Activity{}
Create(&activity)
```

### Expected Outcome:
- ✅ You understand `*` (pointer type) vs `&` (address-of)
- ✅ You know when to use pointer receivers
- ✅ You understand why `*sql.DB` is always a pointer
- ✅ You can read pointer syntax confidently

### Notes:
```
What clicked about pointers:
-

What's still confusing:
-

Mental model: In JavaScript, objects are always "reference" (like pointers).
In Go, you choose: value (copy) or pointer (reference).
```

---

## Week 2, Saturday Morning (3-4 hours)

**Time:** 3-4 hours  
**Focus:** Create models and implement database operations

### Part 1: Create Data Models (45 min)

**Create internal/models/user.go:**
```go
package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

**Create internal/models/activity.go:**
```go
package models

import (
	"time"
)

type Activity struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	ActivityType    string    `json:"activity_type"`
	Title           string    `json:"title,omitempty"`
	Description     string    `json:"description,omitempty"`
	DurationMinutes int       `json:"duration_minutes,omitempty"`
	DistanceKm      float64   `json:"distance_km,omitempty"`
	CaloriesBurned  int       `json:"calories_burned,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	ActivityDate    time.Time `json:"activity_date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateActivityRequest represents the request body for creating an activity
type CreateActivityRequest struct {
	ActivityType    string    `json:"activity_type"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	DistanceKm      float64   `json:"distance_km"`
	CaloriesBurned  int       `json:"calories_burned"`
	Notes           string    `json:"notes"`
	ActivityDate    time.Time `json:"activity_date"`
}
```

**Key Go Concepts:**
- Struct tags: `` `json:"field_name"` `` tells JSON encoder what field names to use
- `omitempty` - omit field from JSON if it's zero value
- Capital letter fields are exported (public), lowercase are private
- This is like TypeScript interfaces but actual types

### Part 2: Create Repository Layer (1 hour 15 min)

**What's a repository?** 
- Separates database logic from HTTP handlers
- Makes code testable (can mock the repository)
- In JavaScript, this is like having a separate database service file

**Create internal/repository/activity_repository.go:**
```go
package repository

import (
	"context"  // 🔴 NEW: Import context package
	"database/sql"
	"fmt"
	"time"

	"github.com/yourusername/activelog/internal/models"
)

type ActivityRepository struct {
	db *sql.DB
}

func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// Create inserts a new activity
// 🔴 NEW: ctx context.Context as first parameter (Go convention)
func (r *ActivityRepository) Create(ctx context.Context, activity *models.Activity) error {
	query := `
		INSERT INTO activities (user_id, activity_type, title, description,
			duration_minutes, distance_km, calories_burned, notes, activity_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	// 🔴 NEW: Use QueryRowContext instead of QueryRow
	err := r.db.QueryRowContext(
		ctx,  // Pass context as first argument
		query,
		activity.UserID,
		activity.ActivityType,
		activity.Title,
		activity.Description,
		activity.DurationMinutes,
		activity.DistanceKm,
		activity.CaloriesBurned,
		activity.Notes,
		activity.ActivityDate,
	).Scan(&activity.ID, &activity.CreatedAt, &activity.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error creating activity: %w", err)
	}

	return nil
}

// GetByID retrieves an activity by ID
// 🔴 NEW: Added context parameter
func (r *ActivityRepository) GetByID(ctx context.Context, id int) (*models.Activity, error) {
	query := `
		SELECT id, user_id, activity_type, title, description, duration_minutes,
			distance_km, calories_burned, notes, activity_date, created_at, updated_at
		FROM activities
		WHERE id = $1
	`

	activity := &models.Activity{}
	// 🔴 NEW: Use QueryRowContext
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&activity.ID,
		&activity.UserID,
		&activity.ActivityType,
		&activity.Title,
		&activity.Description,
		&activity.DurationMinutes,
		&activity.DistanceKm,
		&activity.CaloriesBurned,
		&activity.Notes,
		&activity.ActivityDate,
		&activity.CreatedAt,
		&activity.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("activity not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching activity: %w", err)
	}

	return activity, nil
}

// ListByUser retrieves all activities for a user
// 🔴 NEW: Added context parameter
func (r *ActivityRepository) ListByUser(ctx context.Context, userID int) ([]*models.Activity, error) {
	query := `
		SELECT id, user_id, activity_type, title, description, duration_minutes,
			distance_km, calories_burned, notes, activity_date, created_at, updated_at
		FROM activities
		WHERE user_id = $1
		ORDER BY activity_date DESC
	`

	// 🔴 NEW: Use QueryContext instead of Query
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error listing activities: %w", err)
	}
	defer rows.Close()

	var activities []*models.Activity

	for rows.Next() {
		activity := &models.Activity{}
		err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.ActivityType,
			&activity.Title,
			&activity.Description,
			&activity.DurationMinutes,
			&activity.DistanceKm,
			&activity.CaloriesBurned,
			&activity.Notes,
			&activity.ActivityDate,
			&activity.CreatedAt,
			&activity.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning activity: %w", err)
		}
		activities = append(activities, activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activities: %w", err)
	}

	return activities, nil
}
```

**🔴 CRITICAL: Understanding context.Context**

```go
// What is context.Context?
// It's a special type that carries:
// 1. Deadlines (timeouts)
// 2. Cancellation signals
// 3. Request-scoped values (like user ID, trace ID)

// Why use it?
// - If user closes browser, we can cancel the database query
// - If query takes too long, we can timeout
// - We can trace requests across services
```

**Real-world example:**

```go
// User makes request at 10:00:00
// Request has 5-second timeout
// At 10:00:03, user closes browser

// WITHOUT context:
// Database query continues for 7 more seconds
// Wastes database resources
// Result is thrown away

// WITH context:
// At 10:00:03, context is cancelled
// Database query stops immediately
// Resources freed
// Much better!
```

**Where does context come from?**

```go
// In HTTP handlers, you get it from the request:
func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // Get context from HTTP request

    // Pass it to repository
    err := h.repo.Create(ctx, activity)
}
```

**Convention: ctx is ALWAYS the first parameter**

```go
// ✅ CORRECT
func Create(ctx context.Context, activity *Activity) error

// ❌ WRONG
func Create(activity *Activity, ctx context.Context) error
```

**Key Go Concepts:**
- `(r *ActivityRepository)` - pointer receiver (can modify r)
- `$1, $2, $3` - PostgreSQL placeholders (prevents SQL injection)
- `defer rows.Close()` - ensures rows are closed when function exits
- `sql.ErrNoRows` - special error for "not found"
- Always check `rows.Err()` after iterating
- **🔴 NEW: context.Context for cancellation and timeouts**

### Part 3: Update Handlers to Use Repository (1 hour)

**Update internal/handlers/activity.go:**
```go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yourusername/activelog/internal/models"
	"github.com/yourusername/activelog/internal/repository"
	"github.com/yourusername/activelog/pkg/response"
)

type ActivityHandler struct {
	repo *repository.ActivityRepository
}

func NewActivityHandler(repo *repository.ActivityRepository) *ActivityHandler {
	return &ActivityHandler{repo: repo}
}

// CreateActivity handles POST /api/v1/activities
func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	// 🔴 NEW: Get context from request
	ctx := r.Context()

	var req models.CreateActivityRequest

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// TODO: Get real user ID from authentication (Week 4)
	// For now, hardcode user_id = 1
	activity := &models.Activity{
		UserID:          1,
		ActivityType:    req.ActivityType,
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		DistanceKm:      req.DistanceKm,
		CaloriesBurned:  req.CaloriesBurned,
		Notes:           req.Notes,
		ActivityDate:    req.ActivityDate,
	}

	// Save to database
	// 🔴 NEW: Pass context to repository
	if err := h.repo.Create(ctx, activity); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create activity")
		return
	}

	response.JSON(w, http.StatusCreated, activity)
}

// GetActivity handles GET /api/v1/activities/:id
func (h *ActivityHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	// 🔴 NEW: Get context from request
	ctx := r.Context()

	// Get ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	// Fetch from database
	// 🔴 NEW: Pass context to repository
	activity, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Activity not found")
		return
	}

	response.JSON(w, http.StatusOK, activity)
}

// ListActivities handles GET /api/v1/activities
func (h *ActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	// 🔴 NEW: Get context from request
	ctx := r.Context()

	// TODO: Get real user ID from authentication (Week 4)
	// For now, hardcode user_id = 1
	userID := 1

	// 🔴 NEW: Pass context to repository
	activities, err := h.repo.ListByUser(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
		"count":      len(activities),
	})
}
```

### Part 4: Wire Everything Together in main.go (30 min)

**Update cmd/api/main.go:**
```go
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourusername/activelog/internal/config"
	"github.com/yourusername/activelog/internal/database"
	"github.com/yourusername/activelog/internal/handlers"
	"github.com/yourusername/activelog/internal/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create repositories
	activityRepo := repository.NewActivityRepository(db)

	// Create handlers
	healthHandler := handlers.NewHealthHandler()
	activityHandler := handlers.NewActivityHandler(activityRepo)

	// Setup router
	router := mux.NewRouter()

	// Health check
	router.Handle("/health", healthHandler).Methods("GET")

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/activities", activityHandler.ListActivities).Methods("GET")
	api.HandleFunc("/activities", activityHandler.CreateActivity).Methods("POST")
	api.HandleFunc("/activities/{id}", activityHandler.GetActivity).Methods("GET")

	// Root
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "ActiveLog API v1", "version": "0.1.0"}`))
	}).Methods("GET")

	// Start server
	log.Printf("Server starting on :%s\n", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, router))
}
```

**Test the real endpoints!**

First, create a test user manually:
```bash
psql activelog_dev -U activelog_user

INSERT INTO users (email, username) VALUES ('test@example.com', 'testuser');
\q
```

Now test the API:
```bash
# Start server
go run cmd/api/main.go

# Create an activity
curl -X POST http://localhost:8080/api/v1/activities \
  -H "Content-Type: application/json" \
  -d '{
    "activity_type": "running",
    "title": "Morning Run",
    "duration_minutes": 30,
    "distance_km": 5.2,
    "activity_date": "2024-12-24T07:00:00Z"
  }'

# List activities
curl http://localhost:8080/api/v1/activities

# Get specific activity (use ID from create response)
curl http://localhost:8080/api/v1/activities/1
```

### Expected Outcome After Saturday:
- ✅ You have real database-backed endpoints
- ✅ You understand the repository pattern
- ✅ You can create, read activities from PostgreSQL
- ✅ No more mock data!

### Saturday Reflection:
```
What clicked today:
- 

What was challenging:
- 

How does Go's database code compare to JavaScript ORMs?
- 
```

---

## Week 2, Sunday (2-3 hours)

**Time:** 2-3 hours  
**Focus:** Testing database operations, error handling, documentation

### Part 1: Test Database Operations (1 hour 15 min)

Testing with databases requires special setup. We'll use a test database.

**Create test database:**
```bash
psql postgres
CREATE DATABASE activelog_test;
GRANT ALL PRIVILEGES ON DATABASE activelog_test TO activelog_user;
\q

# Run migrations on test database
migrate -path migrations -database "postgres://activelog_user:your_secure_password@localhost/activelog_test?sslmode=disable" up
```

**Create internal/repository/activity_repository_test.go:**
```go
package repository

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/yourusername/activelog/internal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", "postgres://activelog_user:your_secure_password@localhost/activelog_test?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	// Clean up test data
	_, err := db.Exec("DELETE FROM activities")
	if err != nil {
		t.Logf("Warning: failed to clean activities: %v", err)
	}
	db.Close()
}

func createTestUser(t *testing.T, db *sql.DB) int {
	var userID int
	err := db.QueryRow(`
		INSERT INTO users (email, username)
		VALUES ($1, $2)
		RETURNING id
	`, "test@example.com", "testuser").Scan(&userID)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

func TestActivityRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userID := createTestUser(t, db)
	repo := NewActivityRepository(db)

	activity := &models.Activity{
		UserID:          userID,
		ActivityType:    "running",
		Title:           "Test Run",
		DurationMinutes: 30,
		DistanceKm:      5.0,
		ActivityDate:    time.Now(),
	}

	err := repo.Create(activity)
	if err != nil {
		t.Fatalf("Failed to create activity: %v", err)
	}

	if activity.ID == 0 {
		t.Error("Activity ID should be set after creation")
	}

	if activity.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set after creation")
	}
}

func TestActivityRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userID := createTestUser(t, db)
	repo := NewActivityRepository(db)

	// Create an activity first
	activity := &models.Activity{
		UserID:       userID,
		ActivityType: "basketball",
		Title:        "Pickup Game",
		ActivityDate: time.Now(),
	}
	err := repo.Create(activity)
	if err != nil {
		t.Fatalf("Failed to create activity: %v", err)
	}

	// Now fetch it
	fetched, err := repo.GetByID(activity.ID)
	if err != nil {
		t.Fatalf("Failed to get activity: %v", err)
	}

	if fetched.ID != activity.ID {
		t.Errorf("Expected ID %d, got %d", activity.ID, fetched.ID)
	}

	if fetched.Title != "Pickup Game" {
		t.Errorf("Expected title 'Pickup Game', got '%s'", fetched.Title)
	}
}

func TestActivityRepository_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userID := createTestUser(t, db)
	repo := NewActivityRepository(db)

	// Create multiple activities
	for i := 0; i < 3; i++ {
		activity := &models.Activity{
			UserID:       userID,
			ActivityType: "running",
			Title:        "Run " + string(rune(i)),
			ActivityDate: time.Now(),
		}
		err := repo.Create(activity)
		if err != nil {
			t.Fatalf("Failed to create activity: %v", err)
		}
	}

	// List them
	activities, err := repo.ListByUser(userID)
	if err != nil {
		t.Fatalf("Failed to list activities: %v", err)
	}

	if len(activities) != 3 {
		t.Errorf("Expected 3 activities, got %d", len(activities))
	}
}
```

**Run tests:**
```bash
go test ./internal/repository/...
```

### Part 2: Improve Error Handling (45 min)

**Create pkg/errors/errors.go:**
```go
package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"` // Internal error, not exposed to client
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewBadRequest(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

func NewNotFound(message string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
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
```

**Update pkg/response/json.go to handle AppError:**
```go
package response

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/activelog/pkg/errors"
)

func JSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, statusCode int, message string) error {
	return JSON(w, statusCode, map[string]string{
		"error": message,
	})
}

func AppError(w http.ResponseWriter, err *errors.AppError) error {
	return JSON(w, err.Code, map[string]interface{}{
		"error": err.Message,
		"code":  err.Code,
	})
}
```

### Part 3: Update Documentation (30 min)

**Update README.md with Week 2 progress:**

Add to roadmap section:
```markdown
### Week 2 ✅
- [x] PostgreSQL integration
- [x] Database migrations
- [x] Repository pattern
- [x] Real CRUD operations
- [x] Database testing
```

Add database setup instructions:
```markdown
## Database Setup

1. Create database and user:
```bash
psql postgres
CREATE DATABASE activelog_dev;
CREATE USER activelog_user WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE activelog_dev TO activelog_user;
\q
```

2. Run migrations:
```bash
migrate -path migrations -database "postgres://activelog_user:your_secure_password@localhost/activelog_dev?sslmode=disable" up
```

3. Create a test user:
```bash
psql activelog_dev -U activelog_user
INSERT INTO users (email, username) VALUES ('test@example.com', 'testuser');
\q
```
```

Update API examples with real curl commands.

### Part 4: Commit Your Work (15 min)

```bash
go fmt ./...
go test ./...
git add .
git commit -m "Week 2: PostgreSQL integration with repository pattern"
git tag week-2
```

---

## Week 2 Wrap-Up

### What You Accomplished
You connected a real database and implemented proper data persistence:
- PostgreSQL integration
- Database migrations
- Repository pattern for clean architecture
- Real CRUD operations
- Database testing

### Key Go Concepts Learned
1. **database/sql** - Standard library for database operations
2. **Migrations** - Database version control
3. **Repository pattern** - Separation of concerns
4. **Pointer receivers** - `(r *Repository)` for methods
5. **defer** - Cleanup with defer rows.Close()
6. **SQL placeholders** - `$1, $2` prevent injection
7. **Testing with databases** - Setup/teardown patterns

### JavaScript → Go Mental Model
| JavaScript (Sequelize/TypeORM) | Go Equivalent |
|-------------------------------|---------------|
| `Model.create({...})` | `repo.Create(&model)` |
| `Model.findByPk(id)` | `repo.GetByID(id)` |
| `Model.findAll({...})` | `repo.ListByUser(userID)` |
| Try-catch for DB errors | Explicit `if err != nil` checks |
| Migrations in code | SQL migration files |

### Reflection Questions

**What made sense this week?**
```
Your answer:

```

**What's still unclear?**
```
Your answer:

```

**How confident with databases in Go (1-10)?**
```
Your rating: __/10
```

**Biggest challenge this week?**
```
Your answer:

```

---

## Preparation for Week 3

Next week: Input validation, better error handling, and query filtering.

Make sure you have:
- ✅ All tests passing
- ✅ Database running
- ✅ Migrations applied
- ✅ Code committed

---

*End of Week 2 - You're building real backend systems now!* 🎯

---
---

# WEEK 3: Validation, Error Handling & Filtering

## Week 3 Goal
Add proper input validation, improve error handling, and implement query filtering. Make the API production-ready for basic use.

### Learning Objectives
- Input validation in Go
- Custom error types
- Query parameters and filtering
- Time handling in Go
- Logging best practices

### What You'll Build
- Request validation
- Better error responses
- Filter activities by date range, type
- Pagination
- Structured logging

---

## Week 3, Monday Evening (45 min)

**Time:** 45 minutes  
**Focus:** Input validation

### Tasks:

**1. Install validator package (5 min)**
```bash
go get github.com/go-playground/validator/v10
```

**2. Read about validation (15 min)**
- Go to: https://github.com/go-playground/validator
- Understand struct tags for validation
- Key concept: validation happens on structs using tags

**3. Add validation to models (25 min)**

**Update internal/models/activity.go:**
```go
package models

import (
	"time"
)

type Activity struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	ActivityType    string    `json:"activity_type"`
	Title           string    `json:"title,omitempty"`
	Description     string    `json:"description,omitempty"`
	DurationMinutes int       `json:"duration_minutes,omitempty"`
	DistanceKm      float64   `json:"distance_km,omitempty"`
	CaloriesBurned  int       `json:"calories_burned,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	ActivityDate    time.Time `json:"activity_date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateActivityRequest struct {
	ActivityType    string    `json:"activity_type" validate:"required,min=2,max=50"`
	Title           string    `json:"title" validate:"max=255"`
	Description     string    `json:"description" validate:"max=1000"`
	DurationMinutes int       `json:"duration_minutes" validate:"omitempty,min=1,max=1440"`
	DistanceKm      float64   `json:"distance_km" validate:"omitempty,min=0"`
	CaloriesBurned  int       `json:"calories_burned" validate:"omitempty,min=0"`
	Notes           string    `json:"notes" validate:"max=2000"`
	ActivityDate    time.Time `json:"activity_date" validate:"required"`
}

// Validate validates the request
func (r *CreateActivityRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
```

**Create internal/validator/validator.go:**
```go
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate validates a struct
func Validate(i interface{}) error {
	return validate.Struct(i)
}

// FormatValidationErrors formats validation errors into readable messages
func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				errors[field] = fmt.Sprintf("%s is required", field)
			case "min":
				errors[field] = fmt.Sprintf("%s must be at least %s", field, e.Param())
			case "max":
				errors[field] = fmt.Sprintf("%s must be at most %s characters", field, e.Param())
			default:
				errors[field] = fmt.Sprintf("%s is invalid", field)
			}
		}
	}

	return errors
}
```

### Expected Outcome:
- You understand struct tag validation
- You have validation utilities ready
- You know how to format validation errors

### Notes:
```
How is this different from JavaScript validation libraries?
- 

Questions:
- 
```

---

## Week 3, Wednesday Evening (45 min)

**Time:** 45 minutes  
**Focus:** Structured logging

### Tasks:

**1. Understand logging in Go (15 min)**
- Standard library has `log` package (basic)
- For production, use structured logging (JSON format)
- Popular: zap, zerolog, logrus

**2. Install zerolog (5 min)**
```bash
go get github.com/rs/zerolog/log
```

**3. Create logging utilities (25 min)**

**Create pkg/logger/logger.go:**
```go
package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init() {
	// Pretty print for development
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func Info() *zerolog.Event {
	return log.Info()
}

func Error() *zerolog.Event {
	return log.Error()
}

func Debug() *zerolog.Event {
	return log.Debug()
}

func Warn() *zerolog.Event {
	return log.Warn()
}
```

**Create middleware for request logging:**

**Create internal/middleware/logger.go:**
```go
package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer that captures status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.statusCode).
			Dur("duration", time.Since(start)).
			Msg("HTTP request")
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
```

**4. Add CORS and Security Headers Middleware (20 min)**

**🔴 CRITICAL: Without CORS, your frontend cannot call your API!**

**Add to internal/middleware/cors.go:**
```go
package middleware

import "net/http"

// CORS middleware enables Cross-Origin Resource Sharing
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin (can restrict later)
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Allow specific HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")

		// Allow specific headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// How long to cache preflight requests (24 hours)
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

**🔴 CRITICAL: Security Headers protect against common attacks**

**Add to internal/middleware/security.go:**
```go
package middleware

import "net/http"

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS protection (older browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Force HTTPS for 1 year (only enable when using HTTPS!)
		// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy (basic - customize as needed)
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}
```

**Understanding CORS:**

```go
// What is CORS?
// Browsers block requests from different origins by default
// Example:
//   Frontend: http://localhost:3000
//   Backend:  http://localhost:8080
//   ^ These are DIFFERENT origins!
//
// Without CORS headers, browser blocks the request
// With CORS headers, browser allows it

// Preflight requests:
// For POST/PATCH/DELETE, browser sends OPTIONS request first
// This is the "preflight" - checking if server allows the actual request
```

**Why Security Headers Matter:**

| Header | Protects Against | Impact if Missing |
|--------|------------------|-------------------|
| X-Content-Type-Options | MIME sniffing attacks | Malicious files executed as scripts |
| X-Frame-Options | Clickjacking | Your site embedded in iframe, users tricked |
| X-XSS-Protection | Cross-site scripting | User data stolen |
| Content-Security-Policy | Script injection | Malicious scripts run on your site |

### Expected Outcome:
- ✅ You have structured logging setup
- ✅ Request/response logging middleware ready
- ✅ Understanding of logging levels
- ✅ **CORS middleware enables frontend integration**
- ✅ **Security headers protect against common attacks**

### Notes:
```
Why CORS was needed:
-

What security headers do:
-
```

---

## Week 3, Saturday Morning (3-4 hours)

**Time:** 3-4 hours  
**Focus:** Implement validation, filtering, and pagination

### Part 1: Add Validation to Handlers (45 min)

**Update internal/handlers/activity.go:**
```go
func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	var req models.CreateActivityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  "Validation failed",
			"fields": validationErrors,
		})
		return
	}

	// Rest of the handler...
	activity := &models.Activity{
		UserID:          1,
		ActivityType:    req.ActivityType,
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		DistanceKm:      req.DistanceKm,
		CaloriesBurned:  req.CaloriesBurned,
		Notes:           req.Notes,
		ActivityDate:    req.ActivityDate,
	}

	if err := h.repo.Create(activity); err != nil {
		log.Error().Err(err).Msg("Failed to create activity")
		response.Error(w, http.StatusInternalServerError, "Failed to create activity")
		return
	}

	log.Info().Int("activity_id", activity.ID).Msg("Activity created")
	response.JSON(w, http.StatusCreated, activity)
}
```

### Part 2: Add Query Filtering (1 hour 15 min)

**Create internal/models/filters.go:**
```go
package models

import "time"

type ActivityFilters struct {
	ActivityType string
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}
```

**Update internal/repository/activity_repository.go:**
```go
// ListByUserWithFilters retrieves activities with filters
func (r *ActivityRepository) ListByUserWithFilters(userID int, filters models.ActivityFilters) ([]*models.Activity, error) {
	query := `
		SELECT id, user_id, activity_type, title, description, duration_minutes,
			distance_km, calories_burned, notes, activity_date, created_at, updated_at
		FROM activities
		WHERE user_id = $1
	`

	args := []interface{}{userID}
	argCount := 1

	// Add activity type filter
	if filters.ActivityType != "" {
		argCount++
		query += fmt.Sprintf(" AND activity_type = $%d", argCount)
		args = append(args, filters.ActivityType)
	}

	// Add date range filter
	if filters.StartDate != nil {
		argCount++
		query += fmt.Sprintf(" AND activity_date >= $%d", argCount)
		args = append(args, filters.StartDate)
	}

	if filters.EndDate != nil {
		argCount++
		query += fmt.Sprintf(" AND activity_date <= $%d", argCount)
		args = append(args, filters.EndDate)
	}

	query += " ORDER BY activity_date DESC"

	// Add pagination
	if filters.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listing activities: %w", err)
	}
	defer rows.Close()

	var activities []*models.Activity

	for rows.Next() {
		activity := &models.Activity{}
		err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.ActivityType,
			&activity.Title,
			&activity.Description,
			&activity.DurationMinutes,
			&activity.DistanceKm,
			&activity.CaloriesBurned,
			&activity.Notes,
			&activity.ActivityDate,
			&activity.CreatedAt,
			&activity.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning activity: %w", err)
		}
		activities = append(activities, activity)
	}

	return activities, rows.Err()
}

// Count returns total number of activities for a user
func (r *ActivityRepository) Count(userID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM activities WHERE user_id = $1"
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}
```

### Part 3: Update Handlers with Filtering (1 hour)

**Update internal/handlers/activity.go ListActivities:**
```go
func (h *ActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	userID := 1 // TODO: from auth

	// Parse query parameters
	filters := models.ActivityFilters{
		ActivityType: r.URL.Query().Get("type"),
		Limit:        10, // default
		Offset:       0,
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > 100 {
				limit = 100 // max limit
			}
			filters.Limit = limit
		}
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	// Parse date range
	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startStr); err == nil {
			filters.StartDate = &startDate
		}
	}

	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endStr); err == nil {
			filters.EndDate = &endDate
		}
	}

	// Fetch activities
	activities, err := h.repo.ListByUserWithFilters(userID, filters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list activities")
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	// Get total count
	total, err := h.repo.Count(userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count activities")
		// Continue anyway
		total = len(activities)
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
		"count":      len(activities),
		"total":      total,
		"limit":      filters.Limit,
		"offset":     filters.Offset,
	})
}
```

### Part 4: Wire Up Middleware (45 min)

**Update cmd/api/main.go:**
```go
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourusername/activelog/internal/config"
	"github.com/yourusername/activelog/internal/database"
	"github.com/yourusername/activelog/internal/handlers"
	"github.com/yourusername/activelog/internal/middleware"
	"github.com/yourusername/activelog/internal/repository"
	"github.com/yourusername/activelog/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init()

	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create repositories
	activityRepo := repository.NewActivityRepository(db)

	// Create handlers
	healthHandler := handlers.NewHealthHandler()
	activityHandler := handlers.NewActivityHandler(activityRepo)

	// Setup router
	router := mux.NewRouter()

	// Apply middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.CORS)             // 🔴 NEW: Enable CORS
	router.Use(middleware.SecurityHeaders)  // 🔴 NEW: Add security headers

	// Health check
	router.Handle("/health", healthHandler).Methods("GET")

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/activities", activityHandler.ListActivities).Methods("GET")
	api.HandleFunc("/activities", activityHandler.CreateActivity).Methods("POST")
	api.HandleFunc("/activities/{id}", activityHandler.GetActivity).Methods("GET")

	// Root
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "ActiveLog API v1", "version": "0.2.0"}`))
	}).Methods("GET")

	// 🔴 NEW: Start server with timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,  // Max time to read request
		WriteTimeout: 15 * time.Second,  // Max time to write response
		IdleTimeout:  60 * time.Second,  // Max time for keep-alive connections
	}

	logger.Info().Str("port", cfg.ServerPort).Msg("Starting server")
	log.Fatal(srv.ListenAndServe())
}
```

**🔴 CRITICAL: Understanding Request Timeouts**

```go
// Why timeouts are essential:
// WITHOUT timeouts:
// - Slow client can hold connection forever
// - Attackers can exhaust server resources (Slowloris attack)
// - Long-running queries never cancelled
// - Server becomes unresponsive

// WITH timeouts:
// - Server protected from slow clients
// - Resources freed automatically
// - Predictable behavior under load

// ReadTimeout = 15 seconds
// - Max time from connection accepted to request fully read
// - Includes reading headers + body
// - Too low = legitimate large uploads fail
// - Too high = slow attack possible

// WriteTimeout = 15 seconds
// - Max time for writing the response
// - Includes handler execution time + response writing
// - Should be longer than longest expected handler

// IdleTimeout = 60 seconds
// - Max time for keep-alive between requests
// - HTTP/1.1 allows connection reuse
// - Too short = more TCP handshakes (overhead)
// - Too long = zombie connections waste resources
```

**Add time import:**
```go
import (
	"log"
	"net/http"
	"time"  // 🔴 ADD THIS

	"github.com/gorilla/mux"
	// ...
)
```

**Test filtering:**
```bash
# List all activities
curl "http://localhost:8080/api/v1/activities"

# Filter by type
curl "http://localhost:8080/api/v1/activities?type=running"

# Pagination
curl "http://localhost:8080/api/v1/activities?limit=5&offset=0"

# Date range
curl "http://localhost:8080/api/v1/activities?start_date=2024-12-01T00:00:00Z&end_date=2024-12-31T23:59:59Z"
```

### Expected Outcome After Saturday:
- ✅ Request validation working
- ✅ Query filtering implemented
- ✅ Pagination working
- ✅ Structured logging in place
- ✅ Better error messages

---

## Week 3, Sunday (2-3 hours)

**Time:** 2-3 hours  
**Focus:** Testing, documentation, refactoring

### Part 1: Test Validation and Filtering (1 hour)

**Create internal/handlers/activity_test.go:**
```go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/activelog/internal/models"
)

func TestCreateActivity_Validation(t *testing.T) {
	// Test invalid request (missing required field)
	invalidReq := map[string]interface{}{
		"title": "Test Activity",
		// missing activity_type and activity_date
	}

	body, _ := json.Marshal(invalidReq)
	req := httptest.NewRequest("POST", "/api/v1/activities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Note: This requires mocking the repository
	// For now, this is a structural example
	// Full test would need dependency injection or mocking

	// Check that validation errors are returned
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestCreateActivity_ValidRequest(t *testing.T) {
	validReq := models.CreateActivityRequest{
		ActivityType:    "running",
		Title:           "Morning Run",
		DurationMinutes: 30,
		DistanceKm:      5.0,
		ActivityDate:    time.Now(),
	}

	body, _ := json.Marshal(validReq)
	req := httptest.NewRequest("POST", "/api/v1/activities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Would need to test with mock repository
	// This shows structure for future testing
}
```

### Part 2: Update Documentation (1 hour)

**Update README.md:**

Add filtering examples:
```markdown
## API Endpoints

### List Activities
```bash
GET /api/v1/activities?type=running&limit=10&offset=0&start_date=2024-12-01T00:00:00Z
```

Query Parameters:
- `type` - Filter by activity type
- `limit` - Number of results (max 100, default 10)
- `offset` - Pagination offset
- `start_date` - Filter activities after this date (RFC3339 format)
- `end_date` - Filter activities before this date (RFC3339 format)

Response:
```json
{
  "activities": [...],
  "count": 10,
  "total": 45,
  "limit": 10,
  "offset": 0
}
```

### Create Activity
```bash
POST /api/v1/activities
Content-Type: application/json

{
  "activity_type": "running",
  "title": "Morning Run",
  "duration_minutes": 30,
  "distance_km": 5.2,
  "activity_date": "2024-12-24T07:00:00Z"
}
```

Validation Rules:
- `activity_type`: required, 2-50 characters
- `title`: optional, max 255 characters
- `duration_minutes`: optional, 1-1440 (max 24 hours)
- `distance_km`: optional, must be positive
- `activity_date`: required, RFC3339 format

Error Response:
```json
{
  "error": "Validation failed",
  "fields": {
    "activity_type": "activity_type is required",
    "activity_date": "activity_date is required"
  }
}
```
```

### Part 3: Clean Up and Commit (30 min)

```bash
go fmt ./...
go test ./...
git add .
git commit -m "Week 3: Validation, filtering, and structured logging"
git tag week-3
```

---

## Week 3 Wrap-Up

### What You Accomplished
- Input validation with meaningful error messages
- Query filtering and pagination
- Structured logging with middleware
- Production-ready error handling

### Key Go Concepts Learned
1. **Struct tags** - For validation and JSON
2. **Middleware pattern** - Request/response interception
3. **Query parameter parsing** - Manual but explicit
4. **Time handling** - time.Time and RFC3339
5. **Structured logging** - JSON logs for production

### Reflection

**What felt natural this week?**
```
Your answer:

```

**What's still tricky?**
```
Your answer:

```

**Confidence level (1-10)?**
```
Your rating: __/10
```

---

## Preparation for Week 4

Next week: Add update/delete operations, better API design, and prepare for Month 2 (authentication).

---

*End of Week 3 - The API is taking real shape!* 💪

---
---

# WEEK 4: Complete CRUD & API Polish

## Week 4 Goal
Complete all CRUD operations, improve API design, and wrap up Month 1 with a production-ready foundation.

### Learning Objectives
- UPDATE and DELETE operations
- HTTP method handling
- API versioning
- Code organization best practices
- Preparing for authentication (Month 2)

### What You'll Build
- Update activity endpoint (PATCH)
- Delete activity endpoint (DELETE)
- Activity statistics endpoint
- Better project organization
- Makefile for common tasks

---

## Week 4, Monday Evening (45 min)

**Time:** 45 minutes  
**Focus:** Update operations

### Tasks:

**1. Add Update method to repository (30 min)**

**Update internal/repository/activity_repository.go:**
```go
// Update updates an existing activity
func (r *ActivityRepository) Update(id int, activity *models.Activity) error {
	query := `
		UPDATE activities
		SET activity_type = $1, title = $2, description = $3,
			duration_minutes = $4, distance_km = $5, calories_burned = $6,
			notes = $7, activity_date = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9 AND user_id = $10
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		activity.ActivityType,
		activity.Title,
		activity.Description,
		activity.DurationMinutes,
		activity.DistanceKm,
		activity.CaloriesBurned,
		activity.Notes,
		activity.ActivityDate,
		id,
		activity.UserID,
	).Scan(&activity.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("activity not found")
	}

	return err
}

// Delete deletes an activity
func (r *ActivityRepository) Delete(id int, userID int) error {
	query := "DELETE FROM activities WHERE id = $1 AND user_id = $2"
	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("activity not found")
	}

	return nil
}
```

**2. Create update request model (15 min)**

**Add to internal/models/activity.go:**
```go
type UpdateActivityRequest struct {
	ActivityType    *string    `json:"activity_type" validate:"omitempty,min=2,max=50"`
	Title           *string    `json:"title" validate:"omitempty,max=255"`
	Description     *string    `json:"description" validate:"omitempty,max=1000"`
	DurationMinutes *int       `json:"duration_minutes" validate:"omitempty,min=1,max=1440"`
	DistanceKm      *float64   `json:"distance_km" validate:"omitempty,min=0"`
	CaloriesBurned  *int       `json:"calories_burned" validate:"omitempty,min=0"`
	Notes           *string    `json:"notes" validate:"omitempty,max=2000"`
	ActivityDate    *time.Time `json:"activity_date"`
}
```

**Note:** Using pointers allows partial updates (PATCH semantics)

---

## Week 4, Wednesday Evening (45 min)

**Time:** 45 minutes  
**Focus:** Update and delete handlers

### Tasks:

**Add to internal/handlers/activity.go:**
```go
// UpdateActivity handles PATCH /api/v1/activities/:id
func (h *ActivityHandler) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	var req models.UpdateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate
	if err := validator.Validate(&req); err != nil {
		validationErrors := validator.FormatValidationErrors(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  "Validation failed",
			"fields": validationErrors,
		})
		return
	}

	// Fetch existing activity
	userID := 1 // TODO: from auth
	activity, err := h.repo.GetByID(id)
	if err != nil || activity.UserID != userID {
		response.Error(w, http.StatusNotFound, "Activity not found")
		return
	}

	// Apply updates
	if req.ActivityType != nil {
		activity.ActivityType = *req.ActivityType
	}
	if req.Title != nil {
		activity.Title = *req.Title
	}
	if req.Description != nil {
		activity.Description = *req.Description
	}
	if req.DurationMinutes != nil {
		activity.DurationMinutes = *req.DurationMinutes
	}
	if req.DistanceKm != nil {
		activity.DistanceKm = *req.DistanceKm
	}
	if req.CaloriesBurned != nil {
		activity.CaloriesBurned = *req.CaloriesBurned
	}
	if req.Notes != nil {
		activity.Notes = *req.Notes
	}
	if req.ActivityDate != nil {
		activity.ActivityDate = *req.ActivityDate
	}

	// Save
	if err := h.repo.Update(id, activity); err != nil {
		log.Error().Err(err).Msg("Failed to update activity")
		response.Error(w, http.StatusInternalServerError, "Failed to update activity")
		return
	}

	response.JSON(w, http.StatusOK, activity)
}

// DeleteActivity handles DELETE /api/v1/activities/:id
func (h *ActivityHandler) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	userID := 1 // TODO: from auth

	if err := h.repo.Delete(id, userID); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to delete activity")
		response.Error(w, http.StatusNotFound, "Activity not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

---

## Week 4, Saturday Morning (3-4 hours)

**Time:** 3-4 hours  
**Focus:** Statistics, organization, and tooling

### Part 1: Activity Statistics (1 hour)

**Add to internal/repository/activity_repository.go:**
```go
type ActivityStats struct {
	TotalActivities int
	TotalDuration   int
	TotalDistance   float64
	TotalCalories   int
	ActivityTypes   map[string]int
}

// GetStats returns activity statistics for a user
func (r *ActivityRepository) GetStats(userID int, startDate, endDate *time.Time) (*ActivityStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(duration_minutes), 0) as total_duration,
			COALESCE(SUM(distance_km), 0) as total_distance,
			COALESCE(SUM(calories_burned), 0) as total_calories
		FROM activities
		WHERE user_id = $1
	`

	args := []interface{}{userID}
	argCount := 1

	if startDate != nil {
		argCount++
		query += fmt.Sprintf(" AND activity_date >= $%d", argCount)
		args = append(args, startDate)
	}

	if endDate != nil {
		argCount++
		query += fmt.Sprintf(" AND activity_date <= $%d", argCount)
		args = append(args, endDate)
	}

	stats := &ActivityStats{
		ActivityTypes: make(map[string]int),
	}

	err := r.db.QueryRow(query, args...).Scan(
		&stats.TotalActivities,
		&stats.TotalDuration,
		&stats.TotalDistance,
		&stats.TotalCalories,
	)

	if err != nil {
		return nil, err
	}

	// Get activity type breakdown
	typeQuery := `
		SELECT activity_type, COUNT(*)
		FROM activities
		WHERE user_id = $1
	`

	typeArgs := []interface{}{userID}
	typeArgCount := 1

	if startDate != nil {
		typeArgCount++
		typeQuery += fmt.Sprintf(" AND activity_date >= $%d", typeArgCount)
		typeArgs = append(typeArgs, startDate)
	}

	if endDate != nil {
		typeArgCount++
		typeQuery += fmt.Sprintf(" AND activity_date <= $%d", typeArgCount)
		typeArgs = append(typeArgs, endDate)
	}

	typeQuery += " GROUP BY activity_type"

	rows, err := r.db.Query(typeQuery, typeArgs...)
	if err != nil {
		return stats, nil // Return partial stats
	}
	defer rows.Close()

	for rows.Next() {
		var activityType string
		var count int
		if err := rows.Scan(&activityType, &count); err == nil {
			stats.ActivityTypes[activityType] = count
		}
	}

	return stats, nil
}
```

**Add stats handler:**
```go
// GetStats handles GET /api/v1/activities/stats
func (h *ActivityHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := 1 // TODO: from auth

	var startDate, endDate *time.Time

	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			startDate = &parsed
		}
	}

	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			endDate = &parsed
		}
	}

	stats, err := h.repo.GetStats(userID, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stats")
		response.Error(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	response.JSON(w, http.StatusOK, stats)
}
```

### Part 2: Create Makefile (30 min)

**Create Makefile in project root:**
```makefile
.PHONY: run build test clean migrate-up migrate-down help

# Variables
BINARY_NAME=activelog
DB_URL=postgres://activelog_user:your_secure_password@localhost/activelog_dev?sslmode=disable

## help: Display this help message
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run: Run the application
run:
	go run cmd/api/main.go

## build: Build the application
build:
	go build -o bin/${BINARY_NAME} cmd/api/main.go

## test: Run tests
test:
	go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## migrate-up: Run database migrations up
migrate-up:
	migrate -path migrations -database "${DB_URL}" up

## migrate-down: Run database migrations down
migrate-down:
	migrate -path migrations -database "${DB_URL}" down

## migrate-create: Create a new migration file (usage: make migrate-create NAME=migration_name)
migrate-create:
	migrate create -ext sql -dir migrations -seq $(NAME)

## fmt: Format code
fmt:
	go fmt ./...

## lint: Run linter
lint:
	golangci-lint run

## vuln-check: Check for vulnerabilities (Go 1.18+)
vuln-check:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

## security: Run security audit
security:
	@echo "Running security checks..."
	@go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

## clean: Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

## docker-up: Start Docker containers
docker-up:
	docker-compose up -d

## docker-down: Stop Docker containers
docker-down:
	docker-compose down
```

**Test it:**
```bash
make help
make test
make run
```

### Part 3: Update main.go with all routes (1 hour)

**Final cmd/api/main.go:**
```go
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourusername/activelog/internal/config"
	"github.com/yourusername/activelog/internal/database"
	"github.com/yourusername/activelog/internal/handlers"
	"github.com/yourusername/activelog/internal/middleware"
	"github.com/yourusername/activelog/internal/repository"
	"github.com/yourusername/activelog/pkg/logger"
)

func main() {
	logger.Init()
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Repositories
	activityRepo := repository.NewActivityRepository(db)

	// Handlers
	healthHandler := handlers.NewHealthHandler()
	activityHandler := handlers.NewActivityHandler(activityRepo)

	// Router
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware)

	// Health
	router.Handle("/health", healthHandler).Methods("GET")

	// API v1
	api := router.PathPrefix("/api/v1").Subrouter()

	// Activity routes
	api.HandleFunc("/activities", activityHandler.ListActivities).Methods("GET")
	api.HandleFunc("/activities", activityHandler.CreateActivity).Methods("POST")
	api.HandleFunc("/activities/stats", activityHandler.GetStats).Methods("GET") // Must be before /:id
	api.HandleFunc("/activities/{id}", activityHandler.GetActivity).Methods("GET")
	api.HandleFunc("/activities/{id}", activityHandler.UpdateActivity).Methods("PATCH")
	api.HandleFunc("/activities/{id}", activityHandler.DeleteActivity).Methods("DELETE")

	// Root
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "ActiveLog API v1", "version": "1.0.0"}`))
	}).Methods("GET")

	logger.Info().Str("port", cfg.ServerPort).Msg("Server started")
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, router))
}
```

### Part 4: Final Testing (1 hour)

**Test complete CRUD:**
```bash
# Create
curl -X POST http://localhost:8080/api/v1/activities \
  -H "Content-Type: application/json" \
  -d '{"activity_type":"running","title":"Test","duration_minutes":30,"activity_date":"2024-12-24T07:00:00Z"}'

# Read (list)
curl http://localhost:8080/api/v1/activities

# Read (single)
curl http://localhost:8080/api/v1/activities/1

# Update
curl -X PATCH http://localhost:8080/api/v1/activities/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title"}'

# Stats
curl http://localhost:8080/api/v1/activities/stats

# Delete
curl -X DELETE http://localhost:8080/api/v1/activities/1
```

---

## Week 4, Sunday (2-3 hours)

**Time:** 2-3 hours
**Focus:** Error patterns, documentation, Month 1 wrap-up

### Part 0: Advanced Error Handling Patterns (30 min)

**🔴 CRITICAL: Professional Go error handling**

You've been using `fmt.Errorf("error: %w", err)` but there are better patterns for production code.

**Create pkg/errors/errors.go (enhanced):**
```go
package errors

import (
	"errors"
	"fmt"
)

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
		return fmt.Sprintf("validation error on field '%s': %s (%v)", e.Field, e.Message, e.Err)
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
	return fmt.Sprintf("database error during %s on %s: %v", e.Op, e.Table, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}
```

**Understanding Error Patterns:**

```go
// 1. Sentinel Errors - Use for known error conditions
var ErrNotFound = errors.New("not found")

// Check with errors.Is()
if errors.Is(err, ErrNotFound) {
    // Handle not found case
}

// 2. Error Wrapping - Preserve error chain
err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
    // %w wraps the error, %v doesn't
}

// 3. Custom Error Types - For rich error context
type ValidationError struct {
    Field string
    Err   error
}

// Check with errors.As()
var valErr *ValidationError
if errors.As(err, &valErr) {
    // Access valErr.Field
}

// 4. Error Chains - Walk the error tree
err := someFunc()
fmt.Println(err)                  // Top-level error
fmt.Println(errors.Unwrap(err))   // Next error in chain
```

**When to use each pattern:**

| Pattern | When to Use | Example |
|---------|-------------|---------|
| Sentinel Error | Known, expected errors | `ErrNotFound`, `ErrUnauthorized` |
| Error Wrapping | Adding context to existing error | `fmt.Errorf("query failed: %w", err)` |
| Custom Type | Need error metadata | `&ValidationError{Field: "email"}` |
| Direct Error | Simple, one-off errors | `errors.New("something broke")` |

**Update repository to use sentinel errors:**

```go
func (r *ActivityRepository) GetByID(ctx context.Context, id int) (*models.Activity, error) {
	// ... existing code ...

	if err == sql.ErrNoRows {
		return nil, pkgerrors.ErrNotFound  // Use sentinel error
	}
	if err != nil {
		return nil, &pkgerrors.DatabaseError{
			Op:    "select",
			Table: "activities",
			Err:   err,
		}
	}

	return activity, nil
}
```

**Update handlers to check error types:**

```go
func (h *ActivityHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	// ... existing code ...

	activity, err := h.repo.GetByID(ctx, id)
	if err != nil {
		// Check specific error types
		if errors.Is(err, pkgerrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Activity not found")
			return
		}

		// Log full error chain for debugging
		log.Error().Err(err).Int("id", id).Msg("Failed to get activity")
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activity")
		return
	}

	response.JSON(w, http.StatusOK, activity)
}
```

### Expected Outcome:
- ✅ You understand sentinel errors vs error wrapping
- ✅ You can create custom error types
- ✅ You know when to use `errors.Is()` vs `errors.As()`
- ✅ Your error messages have context

---

### Part 1: Comprehensive Documentation (1 hour 30 min)

**Create API_DOCUMENTATION.md:**
```markdown
# ActiveLog API Documentation

Version: 1.0.0

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
Not yet implemented (coming in Month 2)

## Endpoints

### List Activities
**GET** `/activities`

Query Parameters:
- `type` (string): Filter by activity type
- `limit` (int): Results per page (1-100, default: 10)
- `offset` (int): Pagination offset (default: 0)
- `start_date` (RFC3339): Filter activities after this date
- `end_date` (RFC3339): Filter activities before this date

**Response:** 200 OK
```json
{
  "activities": [...],
  "count": 10,
  "total": 45,
  "limit": 10,
  "offset": 0
}
```

### Create Activity
**POST** `/activities`

**Request Body:**
```json
{
  "activity_type": "running",
  "title": "Morning Run",
  "description": "5K run in the park",
  "duration_minutes": 30,
  "distance_km": 5.2,
  "calories_burned": 300,
  "notes": "Felt great!",
  "activity_date": "2024-12-24T07:00:00Z"
}
```

**Response:** 201 Created

### Get Activity
**GET** `/activities/:id`

**Response:** 200 OK

### Update Activity
**PATCH** `/activities/:id`

Partial update - only send fields you want to change.

**Request Body:**
```json
{
  "title": "Updated Title",
  "duration_minutes": 35
}
```

**Response:** 200 OK

### Delete Activity
**DELETE** `/activities/:id`

**Response:** 204 No Content

### Activity Statistics
**GET** `/activities/stats`

Query Parameters:
- `start_date` (RFC3339): Optional start date
- `end_date` (RFC3339): Optional end date

**Response:** 200 OK
```json
{
  "total_activities": 45,
  "total_duration": 1350,
  "total_distance": 234.5,
  "total_calories": 15000,
  "activity_types": {
    "running": 20,
    "basketball": 15,
    "walking": 10
  }
}
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "Validation failed",
  "fields": {
    "activity_type": "activity_type is required"
  }
}
```

### 404 Not Found
```json
{
  "error": "Activity not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```
```

### Part 2: Update Main README (30 min)

**Update README.md with Month 1 completion:**
```markdown
## Project Status
**Month 1 Complete** ✅

### Completed Features
- ✅ Complete CRUD operations for activities
- ✅ Query filtering and pagination
- ✅ Input validation
- ✅ Activity statistics
- ✅ Structured logging
- ✅ Database migrations
- ✅ Comprehensive testing
- ✅ Production-ready error handling

### Next Steps (Month 2)
- [ ] User authentication (JWT)
- [ ] User registration and login
- [ ] Protected routes
- [ ] User-specific data isolation
```

### Part 3: Final Commit (30 min)

```bash
go fmt ./...
go test ./...
make test-coverage  # Review coverage
git add .
git commit -m "Week 4 & Month 1: Complete CRUD, statistics, and production-ready API"
git tag week-4
git tag month-1-complete
```

---

## Week 4 Wrap-Up

### What You Accomplished
- Complete CRUD operations
- Statistics and analytics
- Production tooling (Makefile)
- Comprehensive documentation
- **MONTH 1 COMPLETE!**

### Month 1 Achievement Summary

You went from "I only know main.go" to:
- ✅ Properly structured Go project
- ✅ REST API with 6+ endpoints
- ✅ PostgreSQL database integration
- ✅ Repository pattern
- ✅ Input validation
- ✅ Query filtering and pagination
- ✅ Structured logging
- ✅ Testing infrastructure
- ✅ Production-ready code

**Lines of Code Written:** ~2000+  
**Go Concepts Mastered:** 25+  
**Confidence Level:** Significantly higher than Week 0

---

## Month 1 Final Reflection

**What are you most proud of?**
```
Your answer:

```

**What was hardest?**
```
Your answer:

```

**How does Go feel now vs Week 1?**
```
Your answer:

```

**Are you ready for Month 2 (Authentication)?**
```
Yes / Need more practice

If more practice:
```

**Confidence rating (1-10)?**
```
Start of Month 1: __/10
End of Month 1:   __/10
```

---

## What's Next: Month 2 Preview

Next month you'll add:
1. User authentication (JWT)
2. User registration and login
3. Password hashing (bcrypt)
4. Protected routes with middleware
5. User-specific data isolation
6. Refresh tokens

**Congratulations on completing Month 1!** 🎉

You're not the person who felt dumb about Go anymore. You're a Go developer building real systems.

---

*End of Month 1 - You did it!* 🚀🎯💪