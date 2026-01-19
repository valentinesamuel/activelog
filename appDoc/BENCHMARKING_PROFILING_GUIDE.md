# Benchmarking & Profiling Guide - Week 12 Task 4 & 5

## Overview

This guide covers **benchmarking** and **profiling** your Go code to measure performance and identify optimization opportunities. You'll learn how to write benchmarks, run them, and analyze CPU and memory profiles.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Benchmark Tests](#benchmark-tests)
3. [Running Benchmarks](#running-benchmarks)
4. [CPU Profiling](#cpu-profiling)
5. [Memory Profiling](#memory-profiling)
6. [Analyzing Profiles](#analyzing-profiles)
7. [N+1 Query Problem](#n1-query-problem)
8. [Best Practices](#best-practices)
9. [Makefile Commands](#makefile-commands)

> 📖 **For detailed Web UI interpretation**, see [PROFILE_INTERPRETATION_GUIDE.md](./PROFILE_INTERPRETATION_GUIDE.md)

---

## Quick Start

### Run all benchmarks:
```bash
make bench
```

### Run benchmarks with profiling:
```bash
make bench-all
```

### Install Graphviz (required for visual analysis):
```bash
make install-graphviz
```

### Analyze CPU profile:
```bash
make profile-cpu        # Web UI (requires graphviz)
# OR
make profile-cpu-cli    # CLI mode (no graphviz needed)
```

### Analyze memory profile:
```bash
make profile-mem        # Web UI (requires graphviz)
# OR
make profile-mem-cli    # CLI mode (no graphviz needed)
```

---

## Benchmark Tests

### What is Benchmarking?

Benchmarking measures the performance of your code by running it multiple times and calculating:
- **Operations per second** (how fast)
- **Time per operation** (how long each operation takes)
- **Memory allocations** (how much memory is used)

### Benchmark File Location

```
internal/repository/activity_repository_bench_test.go
```

### Writing a Benchmark

Benchmarks follow this pattern:

```go
func BenchmarkSomething(b *testing.B) {
    // Setup (excluded from timing)
    db, cleanup := setupBenchDB(b)
    defer cleanup()

    // Reset timer to exclude setup time
    b.ResetTimer()
    b.ReportAllocs() // Track memory allocations

    // The actual benchmark loop
    for i := 0; i < b.N; i++ {
        // Code being benchmarked
        result := DoSomething()
    }
}
```

**Key Points:**
- Function name starts with `Benchmark`
- Takes `*testing.B` parameter
- Uses `b.ResetTimer()` to exclude setup
- Uses `b.ReportAllocs()` to track allocations
- Loops `b.N` times (Go determines optimal N)

---

## Running Benchmarks

### 1. Run All Benchmarks

```bash
make bench
```

**Output:**
```
Running benchmarks...
BenchmarkActivityRepository_Create-8                 100      12345678 ns/op      1024 B/op      20 allocs/op
BenchmarkActivityRepository_GetByID-8               1000       1234567 ns/op       512 B/op      10 allocs/op
...
✅ Benchmarks completed
```

**Reading the output:**
- `100` - Number of iterations run (b.N)
- `12345678 ns/op` - Nanoseconds per operation
- `1024 B/op` - Bytes allocated per operation
- `20 allocs/op` - Number of allocations per operation

### 2. Run Verbose Benchmarks

```bash
make bench-verbose
```

Shows detailed output including test names and progress.

### 3. Run Specific Benchmark

```bash
go test -bench=BenchmarkActivityRepository_Create -benchmem ./internal/repository
```

### 4. Run N+1 Comparison

```bash
make bench-compare
```

Compares JOIN approach vs N+1 problem to show performance difference.

---

## CPU Profiling

### What is CPU Profiling?

CPU profiling shows where your code spends time during execution. It helps identify:
- Slow functions
- Hot paths (frequently executed code)
- Optimization opportunities

### Generate CPU Profile

```bash
make bench-cpu
```

This runs benchmarks and saves CPU profile to `cpu.out`.

### Analyze CPU Profile

```bash
make profile-cpu
```

Opens an interactive web UI at `http://localhost:8080` with:
- **Flame graph** - Visual representation of call stack
- **Top functions** - Functions consuming most CPU time
- **Source view** - Line-by-line analysis
- **Call graph** - Function call relationships

> 💡 **New to the Web UI?** See [PROFILE_INTERPRETATION_GUIDE.md](./PROFILE_INTERPRETATION_GUIDE.md) for a complete tutorial on interpreting all views, graphs, and metrics

### CLI Analysis (Alternative)

```bash
go tool pprof cpu.out
```

**Common commands:**
```
(pprof) top          # Show top 10 CPU consumers
(pprof) top20        # Show top 20
(pprof) list <func>  # Show source code for function
(pprof) web          # Open graph in browser
(pprof) help         # Show all commands
```

**Example output:**
```
(pprof) top
Showing nodes accounting for 450ms, 90% of 500ms total
      flat  flat%   sum%        cum   cum%
     150ms 30.00% 30.00%      200ms 40.00%  database/sql.(*DB).Query
     100ms 20.00% 50.00%      150ms 30.00%  github.com/lib/pq.(*conn).Exec
      80ms 16.00% 66.00%       80ms 16.00%  runtime.mallocgc
```

**Columns:**
- `flat` - Time spent in function (excluding calls)
- `cum` - Cumulative time (including calls)
- `flat%` - Percentage of total time

---

## Memory Profiling

### What is Memory Profiling?

Memory profiling shows memory allocation patterns:
- Total memory allocated
- Number of allocations
- Where allocations occur
- Memory leaks (if any)

### Generate Memory Profile

```bash
make bench-mem
```

Saves memory profile to `mem.out`.

### Analyze Memory Profile

```bash
make profile-mem
```

Opens web UI showing:
- **Memory allocations by function**
- **Allocation sites** (where memory is allocated)
- **Object types** being allocated
- **Memory hotspots**

### CLI Analysis

```bash
go tool pprof mem.out
```

**Useful commands:**
```
(pprof) top                    # Top memory allocators
(pprof) list <func>            # Source code with allocations
(pprof) alloc_space            # Total memory allocated
(pprof) alloc_objects          # Number of allocations
(pprof) inuse_space            # Memory currently in use
```

**Example:**
```
(pprof) top
Showing nodes accounting for 512MB, 85% of 600MB total
      flat  flat%   sum%        cum   cum%
     200MB 33.33% 33.33%      250MB 41.67%  *ActivityRepository.Create
     150MB 25.00% 58.33%      150MB 25.00%  encoding/json.Marshal
     100MB 16.67% 75.00%      100MB 16.67%  strings.Builder.Grow
```

---

## N+1 Query Problem

### What is the N+1 Problem?

The N+1 problem occurs when you:
1. Make **1 query** to fetch N records
2. Make **N additional queries** to fetch related data for each record

**Example:**
```go
// BAD: N+1 Problem (1 + 20 = 21 queries)
activities, _ := repo.ListByUser(userID)  // 1 query
for _, activity := range activities {
    tags, _ := repo.GetTagsForActivity(activity.ID)  // N queries
}
```

**Better:**
```go
// GOOD: Single JOIN query (1 query total)
activities, _ := repo.GetActivitiesWithTags(userID)
// Tags are already loaded via JOIN
```

### Benchmark Comparison

Run the comparison benchmark:

```bash
make bench-compare
```

**Typical results:**
```
BenchmarkComparison/WithJOIN-8              1000      1500000 ns/op
BenchmarkComparison/N+1Problem-8             100     15000000 ns/op
```

The N+1 approach is **10x slower** because it makes 21 queries instead of 1.

---

## Best Practices

### 1. Exclude Setup from Timing

```go
func BenchmarkSomething(b *testing.B) {
    // Setup
    data := setupExpensiveData()

    b.ResetTimer()  // ✅ Reset timer AFTER setup

    for i := 0; i < b.N; i++ {
        // Benchmarked code
    }
}
```

### 2. Always Use b.ReportAllocs()

```go
func BenchmarkSomething(b *testing.B) {
    b.ResetTimer()
    b.ReportAllocs()  // ✅ Track memory allocations

    for i := 0; i < b.N; i++ {
        // Code
    }
}
```

### 3. Don't Optimize Inside Benchmark Loop

```go
// ❌ BAD
for i := 0; i < b.N; i++ {
    if i%2 == 0 {
        DoSomething()
    }
}

// ✅ GOOD
for i := 0; i < b.N; i++ {
    DoSomething()
}
```

### 4. Use Sub-Benchmarks for Comparisons

```go
func BenchmarkComparison(b *testing.B) {
    b.Run("Approach1", func(b *testing.B) {
        // Benchmark approach 1
    })

    b.Run("Approach2", func(b *testing.B) {
        // Benchmark approach 2
    })
}
```

### 5. Run Multiple Times for Accuracy

```bash
# Run 5 times to get average
go test -bench=. -benchmem -count=5 ./internal/repository
```

### 6. Benchmark on Real Hardware

- Don't benchmark in containers (skewed results)
- Close other applications
- Use consistent environment
- Run on production-like hardware

---

## Analyzing Profiles

> 📚 **Complete Web UI Guide**: For an in-depth tutorial on interpreting the pprof web UI, see [PROFILE_INTERPRETATION_GUIDE.md](./PROFILE_INTERPRETATION_GUIDE.md)

### Quick Overview

#### Flame Graph

The flame graph shows:
- **Width** = Time/memory spent in function
- **Height** = Call stack depth
- **Color** = Different packages/functions

**How to read:**
- Wide sections = Performance hotspots
- Look for unexpected wide sections
- Drill down by clicking sections

#### Top Functions

Focus on functions with:
- High `flat` time (slow function body)
- High `cum` time (slow including calls)
- Many allocations
- Large memory usage

#### Source View

Shows line-by-line breakdown:
```
         .          .     15:func (r *Repo) Create(...) {
         .          .     16:    query := "INSERT..."
    100ms      100ms     17:    r.db.Query(query)  // ← Hotspot!
         .          .     18:    return nil
         .          .     19:}
```

### Detailed Analysis

For step-by-step workflows, real-world examples, and advanced interpretation techniques:

👉 **[Read the Complete Profile Interpretation Guide](./PROFILE_INTERPRETATION_GUIDE.md)**

Topics covered:
- Understanding all Web UI views (Graph, Flame Graph, Peek, Source, Top)
- CPU vs Memory profile differences
- Step-by-step analysis workflows
- Common patterns and what they mean
- Real-world optimization examples
- Pro tips and tricks

---

## Makefile Commands

### Benchmark Commands

| Command | Description |
|---------|-------------|
| `make bench` | Run all benchmarks |
| `make bench-verbose` | Run with verbose output |
| `make bench-compare` | Compare JOIN vs N+1 |
| `make bench-cpu` | Run with CPU profiling |
| `make bench-mem` | Run with memory profiling |
| `make bench-all` | Run with both CPU and memory profiling |

### Profiling Commands

| Command | Description |
|---------|-------------|
| `make profile-cpu` | Analyze CPU profile (opens web UI) |
| `make profile-mem` | Analyze memory profile (opens web UI) |

### Cleanup Commands

| Command | Description |
|---------|-------------|
| `make clean-bench` | Remove profile files |
| `make clean` | Remove all generated files |

---

## Complete Workflow Example

### 1. Write the Benchmark

```go
// internal/repository/activity_repository_bench_test.go
func BenchmarkActivityRepository_Create(b *testing.B) {
    db, cleanup := setupBenchDB(b)
    defer cleanup()

    repo := NewActivityRepository(db, nil)
    userID := createBenchUser(b, db)

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        activity := &models.Activity{
            UserID: userID,
            Title: "Test",
        }
        repo.Create(context.Background(), nil, activity)
    }
}
```

### 2. Run the Benchmark

```bash
make bench
```

### 3. Profile It

```bash
make bench-all
```

### 4. Analyze CPU Profile

```bash
make profile-cpu
```

Look for:
- Slow database queries
- Unnecessary allocations
- Hot loops

### 5. Analyze Memory Profile

```bash
make profile-mem
```

Look for:
- Large allocations
- Many small allocations
- Memory leaks

### 6. Optimize

Make changes to improve performance.

### 7. Benchmark Again

```bash
make bench
```

Compare before/after results.

---

## Common Optimizations

### 1. Reduce Allocations

**Before:**
```go
for i := 0; i < n; i++ {
    result := fmt.Sprintf("item-%d", i)  // Allocates each time
}
```

**After:**
```go
var sb strings.Builder
for i := 0; i < n; i++ {
    sb.WriteString("item-")
    sb.WriteString(strconv.Itoa(i))
    result := sb.String()
    sb.Reset()
}
```

### 2. Use Query Batching

**Before:**
```go
for _, item := range items {
    db.Exec("INSERT...", item)  // N queries
}
```

**After:**
```go
// Build bulk insert query (1 query)
db.Exec("INSERT... VALUES ($1), ($2), ($3)...", items...)
```

### 3. Avoid N+1 Queries

Use JOINs or batch loading (see N+1 section above).

### 4. Preallocate Slices

**Before:**
```go
var results []Activity  // Allocates multiple times as it grows
for ... {
    results = append(results, activity)
}
```

**After:**
```go
results := make([]Activity, 0, expectedSize)  // Preallocate capacity
for ... {
    results = append(results, activity)
}
```

---

## Understanding Benchmark Results

### ns/op (Nanoseconds per Operation)

- **< 1,000 ns** - Very fast (microsecond range)
- **1,000 - 1,000,000 ns** - Fast (millisecond range)
- **> 1,000,000 ns** - Slow (multi-millisecond)

Database operations are typically in the **microsecond to millisecond** range.

### B/op (Bytes per Operation)

- **< 100 B** - Very efficient
- **100 - 1,000 B** - Reasonable
- **> 10,000 B** - May need optimization

### allocs/op (Allocations per Operation)

- **< 5** - Very efficient
- **5 - 20** - Reasonable
- **> 50** - May need optimization

---

## Troubleshooting

### "Failed to execute dot. Is Graphviz installed?"

**Problem:** Graphviz is not installed, which is required for visual profile analysis.

**Solution:**
```bash
# Option 1: Install graphviz
make install-graphviz

# Option 2: Use CLI mode (no graphviz needed)
make profile-cpu-cli  # Or profile-mem-cli
```

**Manual installation:**
- **macOS:** `brew install graphviz`
- **Ubuntu/Debian:** `sudo apt-get install graphviz`
- **CentOS/RHEL:** `sudo yum install graphviz`
- **Windows:** Download from https://graphviz.org/download/

### "No profile found"

**Problem:** Running `make profile-cpu` without generating profile first.

**Solution:**
```bash
make bench-cpu  # Generate profile first
make profile-cpu  # Then analyze
```

### Benchmarks are unstable

**Problem:** Results vary significantly between runs.

**Solution:**
- Close other applications
- Run multiple times: `go test -bench=. -count=10`
- Use `-benchtime=10s` for longer runs
- Check system load

### Out of memory during benchmarking

**Problem:** Creating too much test data.

**Solution:**
- Reduce `b.N` manually (not recommended)
- Clean up resources in loop
- Use smaller test datasets

---

## Resources

- [Go Benchmarking Guide](https://pkg.go.dev/testing#hdr-Benchmarks)
- [pprof Documentation](https://github.com/google/pprof/blob/master/doc/README.md)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html)

---

## Summary

✅ **Benchmarks** measure performance (speed, memory)
✅ **CPU profiling** shows where time is spent
✅ **Memory profiling** shows allocation patterns
✅ **Use Makefile commands** for easy access
✅ **Optimize based on data** not guesses
✅ **Profile before and after** changes

**Golden Rule:** *Measure first, optimize second, measure again.*

Happy optimizing! 🚀
