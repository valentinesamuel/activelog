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

