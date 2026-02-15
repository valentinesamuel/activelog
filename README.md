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

## Dynamic Filtering (v1.1.0+)

ActiveLog supports powerful TypeORM-style dynamic filtering with natural column names for relationships.

### Quick Examples

**Filter by activity type:**
```bash
GET /api/v1/activities?filter[activity_type]=running
```

**Filter by tag (auto-JOIN):**
```bash
GET /api/v1/activities?filter[tags.name]=cardio
```

**Date range filtering:**
```bash
GET /api/v1/activities?filter[activity_date][gte]=2024-01-01&filter[activity_date][lte]=2024-12-31
```

**Search and sort:**
```bash
GET /api/v1/activities?search[title]=morning&order[distance_km]=DESC&page=1&limit=20
```

**Complex query with relationships:**
```bash
GET /api/v1/activities?filter[tags.name]=cardio&filter[user.username]=john&search[title]=run&order[activity_date]=DESC
```

### Key Features

- ✅ **Natural column names** - Use `tags.name` instead of aliases
- ✅ **Auto-JOIN detection** - Relationships detected from dot notation
- ✅ **Operator filtering** - Support for `eq`, `ne`, `gt`, `gte`, `lt`, `lte`
- ✅ **Secure** - Column whitelisting prevents unauthorized access
- ✅ **Performant** - Parameterized queries with automatic indexing

### Supported Operators (v1.1.0+)

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equal (default) | `filter[distance_km][eq]=10` |
| `ne` | Not equal | `filter[activity_type][ne]=running` |
| `gt` | Greater than | `filter[distance_km][gt]=5.0` |
| `gte` | Greater than or equal | `filter[activity_date][gte]=2024-01-01` |
| `lt` | Less than | `filter[duration_minutes][lt]=60` |
| `lte` | Less than or equal | `filter[activity_date][lte]=2024-12-31` |

### Documentation

For complete filtering and querying documentation, see:
- [Query Guide](docs/QUERY_GUIDE.md) - Complete guide to filtering, ordering, and relationships
- [Architecture Documentation](docs/ARCHITECTURE.md) - System architecture overview

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

### Week 2✅
Add database setup instructions:

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

## Learning Notes

This is a learning project following a 12-month Go mastery plan.

### Key Learnings - Week 1
- Go project structure is more opinionated than JavaScript
- Handlers implement ServeHTTP interface
- Testing is built into Go (no Jest/Mocha needed)
- Error handling is explicit (no try-catch)

