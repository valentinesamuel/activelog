.PHONY: run build test clean migrate-up migrate-down migrate-force migrate-version help mocks mocks-install mocks-verify clean-mocks test-unit test-integration test-verbose test-coverage test-coverage-html test-coverage-by-package test-coverage-threshold test-coverage-detailed bench bench-verbose bench-compare bench-cpu bench-mem bench-all profile-cpu profile-mem profile-cpu-cli profile-mem-cli install-graphviz clean-bench vuln-check security format docker-up docker-down

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
	@if [ -z "$(NAME)" ]; then \
		echo "❌ NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir migrations -seq $(NAME)
	@UP_FILE=$$(ls -t migrations/*.up.sql | head -1); \
	DOWN_FILE=$$(ls -t migrations/*.down.sql | head -1); \
	echo "BEGIN;\n\n-- Add your migration SQL here\n\nCOMMIT;" > "$$UP_FILE"; \
	echo "BEGIN;\n\n-- Add your rollback SQL here\n\nCOMMIT;" > "$$DOWN_FILE"; \
	echo "✅ Created migration files with transaction wrappers:"; \
	echo "   $$UP_FILE"; \
	echo "   $$DOWN_FILE"


## migrate-version: Show current migration version
migrate-version:
	@migrate -path migrations -database "${DB_URL}" version

## migrate-force: Force set migration version without running migrations (usage: make migrate-force VERSION=3)
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION is required. Usage: make migrate-force VERSION=3"; \
		echo "💡 Run 'make migrate-version' to see current version"; \
		exit 1; \
	fi
	@echo "⚠️  Forcing migration version to $(VERSION)..."
	migrate -path migrations -database "${DB_URL}" force $(VERSION)
	@echo "✅ Migration version forced to $(VERSION)"

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

## bench: Run benchmarks (skip regular tests)
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -run=^$$ ./internal/repository
	@echo "✅ Benchmarks completed"

## bench-verbose: Run benchmarks with verbose output (skip regular tests)
bench-verbose:
	@echo "Running benchmarks with verbose output..."
	go test -bench=. -benchmem -run=^$$ -v ./internal/repository
	@echo "✅ Verbose benchmarks completed"

## bench-compare: Run N+1 comparison benchmark (skip regular tests)
bench-compare:
	@echo "Running N+1 comparison benchmark..."
	go test -bench=BenchmarkComparison -benchmem -run=^$$ ./internal/repository
	@echo "✅ Comparison benchmark completed"

## bench-cpu: Run benchmarks with CPU profiling (skip regular tests)
bench-cpu:
	@echo "Running benchmarks with CPU profiling..."
	go test -bench=. -benchmem -run=^$$ -cpuprofile=cpu.out ./internal/repository
	@echo "✅ CPU profile saved to cpu.out"
	@echo "💡 Analyze with: make profile-cpu"

## bench-mem: Run benchmarks with memory profiling (skip regular tests)
bench-mem:
	@echo "Running benchmarks with memory profiling..."
	go test -bench=. -benchmem -run=^$$ -memprofile=mem.out ./internal/repository
	@echo "✅ Memory profile saved to mem.out"
	@echo "💡 Analyze with: make profile-mem"

## bench-all: Run all benchmarks with CPU and memory profiling (skip regular tests)
bench-all:
	@echo "Running all benchmarks with profiling..."
	go test -bench=. -benchmem -run=^$$ -cpuprofile=cpu.out -memprofile=mem.out ./internal/repository
	@echo "✅ All benchmarks completed"
	@echo "✅ CPU profile saved to cpu.out"
	@echo "✅ Memory profile saved to mem.out"
	@echo "💡 Analyze CPU: make profile-cpu"
	@echo "💡 Analyze Memory: make profile-mem"

## profile-cpu: Analyze CPU profile
profile-cpu:
	@echo "Opening CPU profile analysis..."
	@if [ ! -f cpu.out ]; then \
		echo "❌ cpu.out not found. Run 'make bench-cpu' or 'make bench-all' first"; \
		exit 1; \
	fi
	@echo "💡 If graphviz error occurs, run: make install-graphviz"
	@go tool pprof -http=:8080 cpu.out || (echo "" && echo "⚠️  Graphviz not installed. Install with: make install-graphviz" && echo "📊 Alternative: Use CLI mode with: go tool pprof cpu.out")

## profile-mem: Analyze memory profile
profile-mem:
	@echo "Opening memory profile analysis..."
	@if [ ! -f mem.out ]; then \
		echo "❌ mem.out not found. Run 'make bench-mem' or 'make bench-all' first"; \
		exit 1; \
	fi
	@echo "💡 If graphviz error occurs, run: make install-graphviz"
	@go tool pprof -http=:8080 mem.out || (echo "" && echo "⚠️  Graphviz not installed. Install with: make install-graphviz" && echo "📊 Alternative: Use CLI mode with: go tool pprof mem.out")

## profile-cpu-cli: Analyze CPU profile in CLI mode (no graphviz required)
profile-cpu-cli:
	@echo "Opening CPU profile in CLI mode..."
	@if [ ! -f cpu.out ]; then \
		echo "❌ cpu.out not found. Run 'make bench-cpu' or 'make bench-all' first"; \
		exit 1; \
	fi
	@echo "📊 Profile ready. Common commands: top, top20, list <func>, web"
	go tool pprof cpu.out

## profile-mem-cli: Analyze memory profile in CLI mode (no graphviz required)
profile-mem-cli:
	@echo "Opening memory profile in CLI mode..."
	@if [ ! -f mem.out ]; then \
		echo "❌ mem.out not found. Run 'make bench-mem' or 'make bench-all' first"; \
		exit 1; \
	fi
	@echo "📊 Profile ready. Common commands: top, top20, list <func>, web"
	go tool pprof mem.out

## install-graphviz: Install graphviz for profile visualization
install-graphviz:
	@echo "Installing graphviz..."
	@if command -v brew >/dev/null 2>&1; then \
		echo "📦 Using Homebrew to install graphviz..."; \
		brew install graphviz; \
		echo "✅ Graphviz installed successfully"; \
	elif command -v apt-get >/dev/null 2>&1; then \
		echo "📦 Using apt-get to install graphviz..."; \
		sudo apt-get update && sudo apt-get install -y graphviz; \
		echo "✅ Graphviz installed successfully"; \
	elif command -v yum >/dev/null 2>&1; then \
		echo "📦 Using yum to install graphviz..."; \
		sudo yum install -y graphviz; \
		echo "✅ Graphviz installed successfully"; \
	else \
		echo "❌ Package manager not found"; \
		echo ""; \
		echo "Please install graphviz manually:"; \
		echo "  macOS:   brew install graphviz"; \
		echo "  Ubuntu:  sudo apt-get install graphviz"; \
		echo "  CentOS:  sudo yum install graphviz"; \
		echo "  Windows: Download from https://graphviz.org/download/"; \
		exit 1; \
	fi

## clean-bench: Clean benchmark and profile files
clean-bench:
	@echo "Cleaning benchmark files..."
	rm -f cpu.out mem.out *.test
	@echo "✅ Benchmark files cleaned"

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
	rm -f coverage.out cpu.out mem.out *.test
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