# Troubleshooting Guide

Common issues and their solutions.

## Migration Issues

### "no migration files found in /path/to/migrations"

**Problem:** Tests can't find migration files when running from different directories.

**Solution:** ✅ **FIXED!** The code now automatically finds the project root by looking for `go.mod` and resolves the migrations path from there.

**How it works:**
- Walks up directory tree looking for `go.mod`
- Once found, constructs path to `migrations/` folder
- Works regardless of where tests are run from

**If still having issues:**
1. Verify `go.mod` exists at project root
2. Verify `migrations/` folder exists at project root
3. Verify migration files have `.up.sql` extension

---

## Profiling Issues

### "Failed to execute dot. Is Graphviz installed?"

**Problem:** Graphviz is not installed, which `pprof` needs for visual graphs.

**Solution Option 1: Install Graphviz**
```bash
make install-graphviz
```

**Solution Option 2: Use CLI Mode (No Graphviz Needed)**
```bash
# Instead of:
make profile-cpu

# Use:
make profile-cpu-cli
```

**Manual Installation:**
- **macOS:** `brew install graphviz`
- **Ubuntu/Debian:** `sudo apt-get install graphviz`
- **CentOS/RHEL:** `sudo yum install graphviz`
- **Windows:** Download from https://graphviz.org/download/

### Profile file not found

**Error:** `❌ cpu.out not found`

**Solution:** Generate profile first:
```bash
make bench-cpu   # Generate cpu.out
make profile-cpu # Then analyze
```

---

## Docker Issues

### Container startup fails

**Error:** `Failed to start postgres container`

**Solutions:**
1. **Check Docker is running:**
   ```bash
   docker ps
   ```

2. **Pre-pull images:**
   ```bash
   docker pull postgres:latest
   docker pull testcontainers/ryuk:0.13.0
   ```

3. **Check Docker resources:**
   - Ensure Docker has enough memory allocated (4GB+)
   - Check disk space

4. **Network issues:**
   - Check internet connectivity
   - Check if corporate firewall blocks Docker registry

---

## Test Issues

### Tests pass first time, fail on second run

**Problem:** Database state not cleaned between test runs.

**Solution:** Use `testhelpers.SetupTestDB()` which creates fresh container per test.

```go
func TestSomething(t *testing.T) {
    db, cleanup := testhelpers.SetupTestDB(t)
    defer cleanup()  // ← Always call cleanup!

    // Test code...
}
```

### Integration tests are slow

**Problem:** Tests take 30+ seconds to run.

**Solutions:**
1. **Skip integration tests during development:**
   ```bash
   go test -short ./...  # Skips integration tests
   ```

2. **Pre-pull Docker images:**
   ```bash
   docker pull postgres:latest
   docker pull testcontainers/ryuk:0.13.0
   ```

3. **Run specific test:**
   ```bash
   go test -run TestIntegration_CreateActivity ./internal/repository
   ```

---

## Benchmark Issues

### Benchmark results vary significantly

**Problem:** Inconsistent benchmark results between runs.

**Solutions:**
1. **Close other applications**
2. **Run multiple times:**
   ```bash
   go test -bench=. -count=10 ./internal/repository
   ```
3. **Increase benchmark time:**
   ```bash
   go test -bench=. -benchtime=10s ./internal/repository
   ```
4. **Check system load:** Benchmarks are sensitive to CPU usage

### Out of memory during benchmarking

**Problem:** System runs out of memory.

**Solutions:**
1. **Reduce test data size**
2. **Close other applications**
3. **Increase Docker memory limit**
4. **Run fewer benchmarks:**
   ```bash
   go test -bench=BenchmarkActivityRepository_Create ./internal/repository
   ```

---

## Build Issues

### "undefined: fmt" in testhelpers

**Problem:** Missing import.

**Solution:** Already included in the fixed version. If you see this, update `testhelpers/container.go`.

### Mock generation fails

**Problem:** `mockgen` not installed or not in PATH.

**Solution:**
```bash
make mocks-install  # Install mockgen
make mocks          # Generate mocks
```

---

## Common Command Issues

### "make: command not found"

**Problem:** Make is not installed.

**Solutions:**
- **macOS:** Pre-installed, or `xcode-select --install`
- **Ubuntu:** `sudo apt-get install build-essential`
- **Windows:** Use WSL or install Make for Windows

### Database connection fails

**Error:** `connection refused` or `timeout`

**Solutions:**
1. **Start Docker containers:**
   ```bash
   make docker-up
   ```

2. **Check database is running:**
   ```bash
   docker ps | grep postgres
   ```

3. **Verify port is correct:**
   - Default: `5444`
   - Check `Makefile` DB_URL variable

4. **Run migrations:**
   ```bash
   make migrate-up
   ```

---

## Quick Diagnostic Commands

### Check Environment
```bash
# Check Docker
docker --version
docker ps

# Check Go
go version

# Check Graphviz
dot -V

# Check Make
make --version
```

### Check Project Health
```bash
# Compile all code
go build ./...

# Format code
make format

# Run unit tests only (fast)
go test -short ./...

# Check dependencies
go mod tidy
go mod verify
```

### Reset Everything
```bash
# Clean all generated files
make clean
make clean-mocks
make clean-bench

# Stop all Docker containers
make docker-down

# Restart fresh
make docker-up
make migrate-up
make test
```

---

## Getting Help

### Check Documentation
- **Benchmarking:** `docs/BENCHMARKING_PROFILING_GUIDE.md`
- **Integration Tests:** `docs/INTEGRATION_TESTS_GUIDE.md`
- **Makefile Commands:** `docs/MAKEFILE_COMMANDS.md`

### Common Commands
```bash
make help              # Show all make commands
go test -h             # Go test help
go tool pprof -h       # pprof help
```

### Still Having Issues?

1. **Check error message carefully** - Often contains the solution
2. **Search documentation** - Use Ctrl+F in guide files
3. **Check GitHub issues** - Similar problems may be documented
4. **Run with verbose output:**
   ```bash
   go test -v ./...
   make bench-verbose
   ```

---

## Prevention Tips

### Before Committing
```bash
make format              # Format code
make test                # Run all tests
make test-coverage       # Check coverage
go mod tidy              # Clean dependencies
```

### Before Deploying
```bash
make test-coverage-threshold  # Ensure 70%+ coverage
make vuln-check              # Check for vulnerabilities
make build                   # Verify it builds
```

### Regular Maintenance
```bash
# Weekly
go mod tidy              # Clean dependencies
make clean               # Remove build artifacts

# After pulling changes
make migrate-up          # Run new migrations
make mocks               # Regenerate mocks
make test                # Verify tests pass
```
