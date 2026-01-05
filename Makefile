.PHONY: run build test clean migrate-up migrate-down help mocks mocks-install mocks-verify clean-mocks test-unit test-integration test-verbose test-coverage test-coverage-html test-coverage-by-package test-coverage-threshold test-coverage-detailed vuln-check security format docker-up docker-down

# Variables
BINARY_NAME=activelog
DB_URL=postgres://activelog_user:activelog@localhost:5444/activelog?sslmode=disable

## help: Display this help message
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

run:
	air

build:
	go build -o bin/${BINARY_NAME} cmd/api/main.go

migrate-down:
	migrate -path migrations -database "${DB_URL}" down

migrate-up:
	migrate -path migrations -database "${DB_URL}" up

## migrate-create: Create a new migration file (usage: make migrate-create NAME=migration_name)
migrate-create:
	migrate create -ext sql -dir migrations -seq $(NAME)

int:
	golangci-lint run

## mocks-install: Install mockgen tool
mocks-install:
	@echo "Installing mockgen..."
	go install go.uber.org/mock/mockgen@latest
	go get go.uber.org/mock/gomock
	@echo "✅ Mockgen installed successfully"

## mocks: Generate all mocks from interfaces
mocks:
	@echo "Generating mocks..."
	go generate ./internal/repository
	@echo "✅ Mocks generated successfully"

## mocks-verify: Verify mocks compile correctly
mocks-verify:
	@echo "Verifying mocks compile..."
	go build ./internal/repository/mocks
	@echo "✅ Mocks compile successfully"

## test: Run all tests
test:
	@echo "Running all tests..."
	go test ./...
	@echo "✅ Completed tests 🧪"

## test-unit: Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -short ./...
	@echo "✅ Unit tests completed"

## test-integration: Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -run Integration ./...
	@echo "✅ Integration tests completed"

## test-verbose: Run tests with verbose output
test-verbose:
	go test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "✅ Coverage report generated"

## test-coverage-html: Generate HTML coverage report
test-coverage-html:
	@echo "Generating HTML coverage report..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	@echo "✅ HTML coverage report opened in browser"

## test-coverage-by-package: Show coverage by package
test-coverage-by-package:
	@echo "Running coverage by package..."
	@go test -cover ./internal/handlers/... | grep coverage
	@go test -cover ./internal/repository/... | grep coverage
	@go test -cover ./pkg/... | grep coverage
	@echo "✅ Coverage by package complete"

## test-coverage-threshold: Check if coverage meets 70% threshold
test-coverage-threshold:
	@echo "Checking coverage threshold (70%)..."
	@go test -coverprofile=coverage.out ./... > /dev/null
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$(echo "$$coverage >= 70" | bc -l)" -eq 1 ]; then \
		echo "✅ Coverage $$coverage% meets threshold (≥70%)"; \
	else \
		echo "❌ Coverage $$coverage% below threshold (≥70%)"; \
		exit 1; \
	fi

## test-coverage-detailed: Show detailed coverage by function
test-coverage-detailed:
	@echo "Generating detailed coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | sort -k3 -n
	@echo ""
	@echo "Total coverage:"
	@go tool cover -func=coverage.out | grep total
	@echo "✅ Detailed coverage report complete"

## vuln-check: Check for vulnerabilities
vuln-check:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

## security: Run security checks
security:
	@echo "Running security checks..."
	@go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

## format: Format Go code
format:
	go fmt ./...
	@echo "✅ Completed formatting 🫧 🧽 🧼"

## clean: Clean build artifacts and generated files
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out
	@echo "✅ Clean completed"

## clean-mocks: Remove generated mock files
clean-mocks:
	@echo "Removing generated mocks..."
	rm -rf internal/repository/mocks/
	@echo "✅ Mocks removed"

## docker-up: Start Docker containers
docker-up:
	docker-compose up -d

## docker-down: Stop Docker containers
docker-down:
	docker-compose down