# MONTH 10: Containerization & CI/CD

**Weeks:** 37-40
**Phase:** DevOps & Automation
**Theme:** Automate everything

---

## Overview

This month focuses on modern DevOps practices. You'll containerize your application with Docker, set up local development with Docker Compose, and implement CI/CD pipelines with GitHub Actions. By the end, your code will be automatically tested and ready for deployment on every commit.

---

## Learning Path

### Week 37: Docker Basics
- Dockerfile creation
- Multi-stage builds
- Image optimization
- Container best practices

### Week 38: Docker Compose
- Compose file structure
- Service orchestration
- Networking between containers
- Volume management

### Week 39: GitHub Actions CI
- Workflow syntax
- Running tests automatically
- Code linting
- Build verification

### Week 40: GitHub Actions CD
- Automated deployments
- Docker image building and pushing
- Environment secrets
- Deployment strategies

---

## Docker

### Dockerfile (Multi-stage Build)
```dockerfile
# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Copy migrations (if needed)
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
```

### .dockerignore
```
# Ignore unnecessary files
.git
.github
.vscode
*.md
.env
.env.*
*.log
tmp/
vendor/
*.test
coverage.txt
```

### Build and Run
```bash
# Build image
docker build -t activelog:latest .

# Run container
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://... \
  -e REDIS_URL=redis://... \
  -e JWT_SECRET=secret \
  activelog:latest

# View logs
docker logs <container_id>

# Stop container
docker stop <container_id>
```

---

## Docker Compose

### docker-compose.yml
```yaml
version: '3.9'

services:
  # API Service
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://activelog:password@db:5432/activelog?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - SMTP_HOST=${SMTP_HOST}
      - SMTP_PORT=${SMTP_PORT}
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    networks:
      - activelog-network
    restart: unless-stopped

  # PostgreSQL Database
  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=activelog
      - POSTGRES_USER=activelog
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U activelog"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - activelog-network
    restart: unless-stopped

  # Redis Cache
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    networks:
      - activelog-network
    restart: unless-stopped

  # Worker (Background Jobs)
  worker:
    build:
      context: .
      dockerfile: Dockerfile
    command: ./worker  # Separate binary for worker
    environment:
      - DATABASE_URL=postgres://activelog:password@db:5432/activelog?sslmode=disable
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis
    networks:
      - activelog-network
    restart: unless-stopped

  # Prometheus (Metrics)
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - activelog-network
    restart: unless-stopped

  # Grafana (Dashboards)
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
    depends_on:
      - prometheus
    networks:
      - activelog-network
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:

networks:
  activelog-network:
    driver: bridge
```

### Development Commands
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Run migrations
docker-compose exec api ./migrate -path migrations -database $DATABASE_URL up

# Stop all services
docker-compose down

# Rebuild and restart
docker-compose up -d --build

# Run tests in container
docker-compose exec api go test ./...

# Access database
docker-compose exec db psql -U activelog -d activelog
```

---

## GitHub Actions CI/CD

### .github/workflows/ci.yml (Continuous Integration)
```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  # Lint Code
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  # Run Tests
  test:
    name: Test
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: activelog_test
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/activelog_test?sslmode=disable
          REDIS_URL: redis://localhost:6379

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.txt

  # Build Docker Image
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: false
          tags: activelog:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  # Security Scan
  security:
    name: Security Scan
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'
```

### .github/workflows/cd.yml (Continuous Deployment)
```yaml
name: CD

on:
  push:
    branches: [ main ]
    tags:
      - 'v*'

jobs:
  deploy:
    name: Build and Deploy
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: username/activelog
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Deploy to Production (example)
        run: |
          # This would trigger your deployment
          # Examples: AWS ECS, Kubernetes, etc.
          echo "Deploying version ${{ github.sha }}"
          # curl -X POST https://your-deployment-webhook.com
```

### Makefile
```makefile
.PHONY: help build test lint docker-build docker-up docker-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/activelog cmd/api/main.go

test: ## Run tests
	go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run

docker-build: ## Build Docker image
	docker build -t activelog:latest .

docker-up: ## Start Docker Compose services
	docker-compose up -d

docker-down: ## Stop Docker Compose services
	docker-compose down

docker-logs: ## View Docker Compose logs
	docker-compose logs -f

migrate-up: ## Run database migrations up
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration
	migrate -path migrations -database "$(DATABASE_URL)" down 1

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

run: ## Run the application locally
	go run cmd/api/main.go
```

---

## Best Practices

### Docker
1. **Use multi-stage builds** - Smaller final images
2. **Run as non-root user** - Better security
3. **Use .dockerignore** - Faster builds
4. **Layer caching** - Cache dependencies separately
5. **Health checks** - Container health monitoring

### CI/CD
1. **Run tests on every commit** - Catch bugs early
2. **Cache dependencies** - Faster pipeline runs
3. **Parallel jobs** - Reduce total time
4. **Security scanning** - Automated vulnerability detection
5. **Semantic versioning** - Clear version history

---

## Common Pitfalls

1. **Large Docker images**
   - ❌ Including source code in final image
   - ✅ Use multi-stage builds

2. **Hardcoded secrets**
   - ❌ Secrets in Dockerfile or code
   - ✅ Use environment variables and GitHub Secrets

3. **No health checks**
   - ❌ Can't detect unhealthy containers
   - ✅ Add HEALTHCHECK to Dockerfile

4. **Running as root**
   - ❌ Security vulnerability
   - ✅ Create and use non-root user

---

## Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint](https://golangci-lint.run/)

---

## Next Steps

After completing Month 10, you'll move to **Month 11: AWS Deployment & Monitoring**, where you'll learn:
- AWS ECS deployment
- HTTPS/TLS configuration
- Production monitoring
- Distributed tracing

**Your development workflow is now fully automated!** 🚀
