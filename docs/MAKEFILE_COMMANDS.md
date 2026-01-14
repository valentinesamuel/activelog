# Makefile Commands Reference

Quick reference for all available `make` commands in ActiveLog.

## Development

| Command | Description |
|---------|-------------|
| `make help` | Display all available commands |
| `make run` | Run the application with hot reload (using air) |
| `make build` | Build the binary to `bin/activelog` |
| `make format` | Format all Go code |

## Database Migrations

| Command | Description |
|---------|-------------|
| `make migrate-up` | Run all pending migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-create NAME=<name>` | Create a new migration file |

**Example:**
```bash
make migrate-create NAME=add_user_avatar
```

## Testing

### Basic Tests

| Command | Description |
|---------|-------------|
| `make test` | Run all tests (unit + integration) |
| `make test-unit` | Run unit tests only (with `-short` flag) |
| `make test-integration` | Run integration tests only |
| `make test-verbose` | Run tests with verbose output |

### Test Coverage

| Command | Description |
|---------|-------------|
| `make test-coverage` | Generate coverage report (terminal) |
| `make test-coverage-html` | Generate and open HTML coverage report |
| `make test-coverage-by-package` | Show coverage per package |
| `make test-coverage-threshold` | Verify coverage meets 70% threshold |
| `make test-coverage-detailed` | Show detailed coverage by function |

## Benchmarking & Profiling

### Benchmarks

| Command | Description |
|---------|-------------|
| `make bench` | Run all benchmarks |
| `make bench-verbose` | Run benchmarks with verbose output |
| `make bench-compare` | Run N+1 comparison benchmark |
| `make bench-cpu` | Run benchmarks with CPU profiling |
| `make bench-mem` | Run benchmarks with memory profiling |
| `make bench-all` | Run benchmarks with CPU + memory profiling |

### Profiling

| Command | Description |
|---------|-------------|
| `make install-graphviz` | Install graphviz for profile visualization |
| `make profile-cpu` | Analyze CPU profile (web UI - requires graphviz) |
| `make profile-mem` | Analyze memory profile (web UI - requires graphviz) |
| `make profile-cpu-cli` | Analyze CPU profile in CLI mode (no graphviz) |
| `make profile-mem-cli` | Analyze memory profile in CLI mode (no graphviz) |

**Note:** Run `make bench-cpu` or `make bench-all` before analyzing profiles.

**Graphviz:** Required for web UI visualization. Install with `make install-graphviz` or use `-cli` variants for text-based analysis.

## Mocking

| Command | Description |
|---------|-------------|
| `make mocks-install` | Install mockgen tool |
| `make mocks` | Generate all mocks from interfaces |
| `make mocks-verify` | Verify mocks compile correctly |
| `make clean-mocks` | Remove generated mock files |

## Security

| Command | Description |
|---------|-------------|
| `make vuln-check` | Check for known vulnerabilities |
| `make security` | Run security checks with Nancy |

## Docker

| Command | Description |
|---------|-------------|
| `make docker-up` | Start Docker containers (postgres, redis, etc.) |
| `make docker-down` | Stop Docker containers |

## Cleanup

| Command | Description |
|---------|-------------|
| `make clean` | Clean all build artifacts and generated files |
| `make clean-mocks` | Remove generated mock files |
| `make clean-bench` | Remove benchmark and profile files |

---

## Common Workflows

### Running Tests Before Commit

```bash
make format          # Format code
make test            # Run all tests
make test-coverage   # Check coverage
```

### Full Test Suite with Coverage

```bash
make test-coverage-html
# Opens coverage report in browser
```

### Benchmarking Performance

```bash
# Quick benchmark
make bench

# Full profiling analysis
make bench-all
make profile-cpu
make profile-mem
```

### Setting Up a New Feature

```bash
make migrate-create NAME=add_new_feature
make migrate-up
make mocks
make test
```

### Pre-deployment Checks

```bash
make format
make test-coverage-threshold
make vuln-check
make build
```

---

## Tips

1. **Tab completion**: Most shells support tab completion for Makefile targets
2. **Multiple commands**: Run sequentially with `&&`:
   ```bash
   make format && make test && make build
   ```
3. **Check available commands**: Run `make help` to see all documented commands
4. **Docker**: Ensure Docker is running before `make docker-up`
5. **Coverage threshold**: Set to 70% - fails if below threshold

---

## Environment Variables

Some commands use these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `BINARY_NAME` | `activelog` | Name of the compiled binary |
| `DB_URL` | `postgres://...` | Database connection string |

Override in Makefile or export in shell:
```bash
export DB_URL="postgres://user:pass@localhost/db"
make migrate-up
```

---

## Quick Reference Card

**Most Used Commands:**
```bash
make run              # Start dev server
make test             # Run tests
make bench            # Run benchmarks
make migrate-up       # Apply migrations
make docker-up        # Start containers
make clean            # Clean up
```

**Performance Analysis:**
```bash
make bench-all        # Profile everything
make profile-cpu      # Analyze CPU
make profile-mem      # Analyze memory
```

**Quality Checks:**
```bash
make format           # Format code
make test-coverage    # Check coverage
make vuln-check       # Check security
```
