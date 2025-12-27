.PHONY: run build test clean migrate-up migrate-down help

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

test:
	go test ./...
	echo "Completed tests 🧪"

vuln-check:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

security:
	@echo "Running security checks..."
	@go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

format:
	go fmt ./...
	echo "Completed formatting 🫧 🧽 🧼"

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