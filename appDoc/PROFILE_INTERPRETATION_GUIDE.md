# Go pprof Web UI - Complete Interpretation Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Accessing the Web UI](#accessing-the-web-ui)
3. [Web UI Views Overview](#web-ui-views-overview)
4. [CPU Profile Interpretation](#cpu-profile-interpretation)
5. [Memory Profile Interpretation](#memory-profile-interpretation)
6. [Step-by-Step Analysis Workflow](#step-by-step-analysis-workflow)
7. [Common Patterns & What They Mean](#common-patterns--what-they-mean)
8. [Real-World Examples](#real-world-examples)
9. [Optimization Strategies](#optimization-strategies)
10. [Pro Tips & Tricks](#pro-tips--tricks)

---

## Introduction

The Go pprof web UI is an interactive visualization tool that helps you understand where your program spends time (CPU) or allocates memory. It opens in your browser at `http://localhost:8080` and provides multiple views for analyzing performance data.

### What You'll Learn

- How to navigate the web UI
- What each graph and metric means
- How to identify performance bottlenecks
- How to prioritize optimization efforts

---

## Accessing the Web UI

### For CPU Profiles

```bash
# 1. Generate CPU profile
make bench-cpu

# 2. Open web UI
make profile-cpu

# Your browser opens at: http://localhost:8080
```

### For Memory Profiles

```bash
# 1. Generate memory profile
make bench-mem

# 2. Open web UI
make profile-mem

# Your browser opens at: http://localhost:8080
```

### Manual Access

```bash
# CPU
go tool pprof -http=:8080 cpu.out

# Memory
go tool pprof -http=:8080 mem.out

# Custom port
go tool pprof -http=:9090 cpu.out
```

---

## Web UI Views Overview

When you open the web UI, you'll see a navigation bar with these views:

| View | Purpose | Best For |
|------|---------|----------|
| **Graph** | Call graph showing function relationships | Understanding program flow |
| **Flame Graph** | Visual representation of call stacks | Identifying hotspots quickly |
| **Peek** | Quick overview of top functions | Initial assessment |
| **Source** | Line-by-line code analysis | Pinpointing exact problem lines |
| **Disassemble** | Assembly code view | Advanced low-level optimization |
| **Top** | Ranked list of functions | Identifying worst offenders |

### Navigation Tips

- **Click nodes** in graphs to drill down
- **Use the search box** to find specific functions
- **Switch between views** using the top menu
- **Use VIEW dropdown** to change metrics (e.g., alloc_space vs inuse_space)

---

## CPU Profile Interpretation

### Understanding CPU Metrics

When viewing CPU profiles, you'll see these key metrics:

#### 1. **flat** (Self Time)
- Time spent **directly in the function** (excluding calls to other functions)
- High `flat` = Function body is slow
- **Example:** Function with expensive calculations

#### 2. **cum** (Cumulative Time)
- Time spent in function **including all functions it calls**
- High `cum` but low `flat` = Function calls slow children
- **Example:** HTTP handler that calls slow database functions

#### 3. **flat%** and **cum%**
- Percentage of total execution time
- **Focus on functions with > 5% flat or cum**

### Graph View (CPU)

The graph view shows:

```
┌─────────────────────────────┐
│ main.HandleRequest          │  ← Top of call stack
│ 10ms (2%) | 500ms (100%)   │  ← (flat | cum)
└─────────────────────────────┘
              │
              ▼
┌─────────────────────────────┐
│ repository.GetActivities    │
│ 50ms (10%) | 450ms (90%)   │  ← Calls slow functions
└─────────────────────────────┘
              │
              ▼
┌─────────────────────────────┐
│ database.Query              │
│ 400ms (80%) | 400ms (80%)  │  ← Actual bottleneck!
└─────────────────────────────┘
```

**How to Read:**
- **Boxes** = Functions
- **Arrows** = Call relationships
- **Box size** = Relative time spent (bigger = slower)
- **Red/orange boxes** = Hot paths (most time spent)
- **Light colors** = Less significant

### Flame Graph (CPU)

The flame graph is read **bottom to top**:

```
                    ┌──────────┐
                    │ Function │ ← Top of stack
                    └──────────┘
              ┌──────────────────────┐
              │  Parent Function     │ ← Called by bottom
              └──────────────────────┘
┌────────────────────────────────────────────┐
│           Main Function                    │ ← Entry point
└────────────────────────────────────────────┘
```

**Reading the Flame Graph:**

1. **Width = Time spent**
   - Wider sections = More time consumed
   - Look for unexpectedly wide sections

2. **X-axis = Alphabetical** (not time!)
   - Functions are ordered alphabetically, not chronologically
   - This stacks common call paths together

3. **Y-axis = Call stack depth**
   - Bottom = Entry point
   - Top = Deepest function calls

4. **Colors**
   - Different colors distinguish packages/functions
   - **Red/Warm colors** often indicate hotspots (configurable)

**Example Analysis:**

```
If you see:
┌─────────────────────────────────────────────────┐
│        database.Query (very wide)               │
└─────────────────────────────────────────────────┘
     ↑
This means database.Query is consuming
a large portion of your program's time
```

### Top View (CPU)

The Top view shows a ranked list:

```
Showing nodes accounting for 450ms, 90.00% of 500ms total
Dropped 15 nodes (cum <= 2.5ms)

      flat  flat%   sum%        cum   cum%
     200ms 40.00% 40.00%      300ms 60.00%  database/sql.(*DB).Query
     100ms 20.00% 60.00%      100ms 20.00%  encoding/json.Marshal
      80ms 16.00% 76.00%       80ms 16.00%  runtime.mallocgc
      70ms 14.00% 90.00%      450ms 90.00%  main.HandleRequest
```

**How to Read:**

1. **Focus on top 5-10 functions**
2. **High flat% = Direct performance issue**
   - `database/sql.(*DB).Query` at 40% is a major bottleneck
3. **High cum% but low flat% = Calls expensive children**
   - `main.HandleRequest`: 14% flat, 90% cum → delegates work

### Source View (CPU)

Shows line-by-line CPU usage:

```go
         .          .      1: func (r *Repository) GetActivities(userID int) {
         .          .      2:     query := "SELECT * FROM activities WHERE user_id = $1"
         .          .      3:
    10ms       10ms      4:     rows, err := r.db.Query(query, userID)
         .          .      5:     if err != nil {
         .          .      6:         return nil, err
         .          .      7:     }
         .          .      8:
   150ms      150ms      9:     for rows.Next() {
    80ms       80ms     10:         var activity Activity
    50ms       50ms     11:         err := rows.Scan(&activity.ID, &activity.Title, ...)
         .          .     12:         activities = append(activities, activity)
         .          .     13:     }
         .          .     14:     return activities, nil
         .          .     15: }
```

**Columns:**
- **First column (flat):** Time in this line only
- **Second column (cum):** Time in this line + functions it calls
- **Line 9-11 are hotspots** (150ms + 80ms + 50ms)

---

## Memory Profile Interpretation

### Understanding Memory Metrics

Memory profiles show allocation patterns. Use the **VIEW** dropdown to switch between:

#### 1. **alloc_space** (Total Allocated)
- **Total memory allocated** since program start
- Use to find functions that allocate a lot (even if freed)
- **Best for:** Finding allocation hotspots

#### 2. **alloc_objects** (Total Allocations)
- **Number of allocations**
- Use to find functions making many small allocations
- **Best for:** Reducing GC pressure

#### 3. **inuse_space** (Currently In Use)
- Memory **currently held** (not freed)
- Use to find memory leaks
- **Best for:** Debugging memory leaks

#### 4. **inuse_objects** (Objects In Use)
- Number of objects **currently held**
- **Best for:** Finding object retention issues

### Graph View (Memory)

Similar to CPU, but shows memory allocations:

```
┌─────────────────────────────┐
│ main.ProcessData            │
│ 50MB (5%) | 500MB (50%)    │  ← Allocates through children
└─────────────────────────────┘
              │
              ▼
┌─────────────────────────────┐
│ encoding/json.Marshal       │
│ 450MB (45%) | 450MB (45%)  │  ← Actual allocator!
└─────────────────────────────┘
```

**Box interpretation:**
- **Large boxes** = Heavy allocators
- **Red/orange** = Allocation hotspots
- **cum > flat** = Allocates through dependencies

### Flame Graph (Memory)

Same structure as CPU, but width = memory allocated:

```
                    ┌───────┐
                    │ Small │ ← Small allocation
                    └───────┘
              ┌────────────────────┐
              │  Medium Alloc      │
              └────────────────────┘
┌──────────────────────────────────────────────────┐
│           Large Memory Allocator                 │ ← Look here!
└──────────────────────────────────────────────────┘
```

**What to look for:**
- Very wide sections at bottom = Large allocations
- Many narrow spikes = Many small allocations (GC pressure)

### Top View (Memory)

```
Showing nodes accounting for 512MB, 85.33% of 600MB total

      flat  flat%   sum%        cum   cum%
     200MB 33.33% 33.33%      250MB 41.67%  *ActivityRepository.Create
     150MB 25.00% 58.33%      150MB 25.00%  encoding/json.Marshal
     100MB 16.67% 75.00%      100MB 16.67%  strings.Builder.Grow
      62MB 10.33% 85.33%       62MB 10.33%  runtime.makeslice
```

**Optimization priorities:**
1. **ActivityRepository.Create** - 200MB allocated (33%)
2. **encoding/json.Marshal** - 150MB (25%)
3. **strings.Builder.Grow** - 100MB (16%)

### Source View (Memory)

Shows allocations per line:

```go
         .          .      1: func (r *Repository) Create(activity *Activity) {
         .          .      2:     query := "INSERT INTO..."
         .          .      3:
    50MB       50MB      4:     data, _ := json.Marshal(activity)  // ← Large allocation!
         .          .      5:
   100MB      100MB      6:     result := make([]byte, 1024*1024)  // ← Very large!
         .          .      7:
    10MB       10MB      8:     r.db.Exec(query, data)
         .          .      9:     return nil
         .          .     10: }
```

**Line 6 is suspicious:**
- 100MB allocation from `make([]byte, 1024*1024)`
- This might be unnecessary or could be reused

---

## Step-by-Step Analysis Workflow

### Workflow 1: Finding CPU Bottlenecks

```
1. Start with Flame Graph
   └─> Identify widest sections

2. Switch to Top View
   └─> Find functions with highest flat%

3. Click function name in Top
   └─> Opens Source view for that function

4. Analyze source code
   └─> Look for:
       - Slow loops
       - Repeated calculations
       - Inefficient algorithms

5. Go to Graph view
   └─> Understand call relationships
   └─> See if problem is in caller or callee
```

### Workflow 2: Finding Memory Leaks

```
1. Set VIEW to "inuse_space"
   └─> Shows memory NOT freed

2. Open Top view
   └─> Functions with high inuse = potential leaks

3. Check Graph view
   └─> Trace where objects are created

4. Switch to Source view
   └─> Find exact allocation lines

5. Verify in code
   └─> Are these objects properly released?
   └─> Are defer statements missing?
   └─> Are goroutines leaking?
```

### Workflow 3: Reducing Allocations

```
1. Set VIEW to "alloc_objects"
   └─> Shows number of allocations

2. Top view
   └─> Find functions with many allocations

3. Source view
   └─> Identify allocation-heavy lines

4. Optimize
   └─> Preallocate slices
   └─> Reuse buffers
   └─> Use sync.Pool
   └─> Avoid unnecessary conversions
```

---

## Common Patterns & What They Mean

### Pattern 1: Wide Database Query Section

**What you see:**
```
Flame Graph:
┌────────────────────────────────────────────┐
│     database/sql.(*DB).Query (very wide)   │
└────────────────────────────────────────────┘
```

**What it means:**
- Database queries are slow
- Possibly missing indexes
- N+1 query problem
- Inefficient queries

**Solutions:**
- Add database indexes
- Use JOINs instead of multiple queries
- Optimize SQL queries
- Add query result caching

### Pattern 2: Many Narrow Spikes (Memory)

**What you see:**
```
Flame Graph:
┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐┌┐
││││││││││││││││││││││││││││││││││││
└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘└┘
```

**What it means:**
- Many small allocations
- High GC pressure
- Inefficient memory usage

**Solutions:**
- Preallocate slices with capacity
- Use object pooling (sync.Pool)
- Reuse buffers
- Reduce string concatenations

### Pattern 3: Deep Call Stack

**What you see:**
```
Flame Graph:
           ┌──┐
         ┌─┴──┴─┐
       ┌─┴──────┴─┐
     ┌─┴──────────┴─┐
   ┌─┴────────────────┴─┐
 ┌─┴──────────────────────┴─┐
└──────────────────────────────┘
```

**What it means:**
- Deep function call hierarchy
- Potential stack overhead
- Might indicate recursion issues

**Solutions:**
- Consider flattening call hierarchy
- Check for unnecessary abstraction layers
- Optimize recursion to iteration if possible

### Pattern 4: Wide JSON Marshal/Unmarshal

**What you see:**
```
encoding/json.Marshal or Unmarshal taking 30%+ time
```

**What it means:**
- Large objects being serialized
- Inefficient JSON handling
- Reflection overhead

**Solutions:**
- Use faster JSON libraries (jsoniter, easyjson)
- Reduce data size before marshaling
- Cache marshaled results
- Use protobuf or msgpack for internal APIs

### Pattern 5: Wide `runtime.mallocgc`

**What you see:**
```
Top view:
runtime.mallocgc    150ms  30%
```

**What it means:**
- Excessive memory allocations
- Garbage collector working hard
- Memory churn

**Solutions:**
- Reduce allocations overall
- Preallocate data structures
- Use sync.Pool for temporary objects
- Profile with alloc_objects to find sources

---

## Real-World Examples

### Example 1: Slow API Endpoint

**Profile Data:**
```
Top view (CPU):
      flat  flat%   sum%        cum   cum%
     180ms 36.00% 36.00%      400ms 80.00%  database/sql.(*DB).Query
      80ms 16.00% 52.00%       80ms 16.00%  encoding/json.Marshal
      70ms 14.00% 66.00%      490ms 98.00%  handlers.GetActivities
```

**Analysis:**
1. **handlers.GetActivities** has 98% cum, only 14% flat
   - The handler itself is fast
   - It's calling slow dependencies

2. **database/sql.Query** has 36% flat
   - This is the actual bottleneck
   - Database queries are slow

**Action Items:**
```
1. Click database/sql.Query in Top view
2. Go to Source view
3. Find which queries are slow
4. Check Graph view to see calling patterns
5. Likely fix: Add indexes or optimize queries
```

### Example 2: Memory Leak

**Profile Data:**
```
VIEW: inuse_space

Top view:
      flat  flat%   sum%        cum   cum%
     500MB 50.00% 50.00%      500MB 50.00%  *CacheManager.Add
     200MB 20.00% 70.00%      200MB 20.00%  *EventProcessor.process
```

**Analysis:**
1. **CacheManager.Add** holds 500MB
   - Objects are being cached but never evicted
   - Potential memory leak

**Investigation:**
```
1. Open Source view for CacheManager.Add
2. Check if cache has eviction logic
3. Look for missing defer statements
4. Check if cleanup is called
```

**Likely Fix:**
- Implement cache size limits
- Add TTL-based eviction
- Use LRU cache instead

### Example 3: High Allocation Rate

**Profile Data:**
```
VIEW: alloc_objects

Top view:
      flat    flat%       cum    cum%
     50000   25.00%     50000  25.00%  strings.Builder.Grow
     40000   20.00%     40000  20.00%  fmt.Sprintf
     30000   15.00%     30000  15.00%  append
```

**Analysis:**
1. **50,000 allocations** from strings.Builder
   - Growing buffer many times
   - Not preallocating capacity

2. **40,000 allocations** from fmt.Sprintf
   - String formatting in hot path
   - Should use faster alternatives

**Optimization:**
```go
// Before:
for i := 0; i < 10000; i++ {
    var sb strings.Builder  // Allocates every iteration
    sb.WriteString("item: ")
    sb.WriteString(fmt.Sprintf("%d", i))  // Another allocation
}

// After:
sb := strings.Builder{}
sb.Grow(100000)  // Preallocate once
for i := 0; i < 10000; i++ {
    sb.WriteString("item: ")
    sb.WriteString(strconv.Itoa(i))  // Faster than fmt.Sprintf
    result := sb.String()
    sb.Reset()
}
```

---

## Optimization Strategies

### Strategy 1: The 80/20 Rule

**Focus on functions that account for >20% of time/memory**

```
If Top view shows:
- Function A: 50%
- Function B: 30%
- Function C: 10%
- Others: 10%

Optimize Function A first, then B.
Ignore C unless A and B are optimized.
```

### Strategy 2: Bottom-Up Optimization

```
Start from leaves (bottom of flame graph):
1. Optimize leaf functions first
2. Work your way up the call stack
3. Don't optimize callers before callees

Example:
If database.Query is slow, don't optimize
the HTTP handler that calls it first.
```

### Strategy 3: Memory Before CPU

```
Often, reducing memory allocations also improves CPU:

Fewer allocations → Less GC → Faster execution

So prioritize:
1. Memory profile optimization first
2. CPU profile afterward
```

### Strategy 4: Measure, Don't Guess

```
Before optimizing:
1. Profile current state
2. Save results

After optimizing:
3. Profile again
4. Compare results

Don't optimize without measuring!
```

---

## Pro Tips & Tricks

### Tip 1: Compare Profiles

```bash
# Generate baseline
make bench-all
cp cpu.out cpu_baseline.out
cp mem.out mem_baseline.out

# Make changes...

# Generate new profile
make bench-all

# Compare
go tool pprof -base=cpu_baseline.out cpu.out
```

This shows what **changed** between runs.

### Tip 2: Focus Mode

In the web UI, use the **Sample** dropdown to filter:

```
Sample: alloc_space    # All allocations
Sample: inuse_space    # Current memory
Sample: alloc_objects  # Allocation count
Sample: inuse_objects  # Current objects
```

Different views reveal different issues.

### Tip 3: Search Filtering

Use the search box to filter functions:

```
Search: "database"     # Show only database-related
Search: "json"         # Show only JSON operations
Search: "^main\."      # Show only main package (regex)
```

### Tip 4: Save Interesting Views

Right-click on graphs and save as SVG for documentation:

```
1. Find interesting hotspot in flame graph
2. Right-click → Save image as SVG
3. Include in optimization PRs/docs
```

### Tip 5: CLI One-Liners

```bash
# Top 20 CPU consumers
go tool pprof -top cpu.out | head -20

# Top 20 memory allocators
go tool pprof -top mem.out | head -20

# Show specific function
go tool pprof -list=CreateActivity cpu.out

# Generate call graph SVG
go tool pprof -svg cpu.out > cpu_graph.svg
```

### Tip 6: Sample Rate Adjustment

For memory profiles, adjust sample rate:

```bash
# Default: 512KB sample rate
go test -bench=. -memprofile=mem.out

# Higher resolution (slower, more detail)
go test -bench=. -memprofilerate=1 -memprofile=mem.out

# Lower resolution (faster, less detail)
go test -bench=. -memprofilerate=1048576 -memprofile=mem.out
```

### Tip 7: Interactive CLI Mode

```bash
go tool pprof cpu.out

# Inside pprof:
(pprof) top               # Top functions
(pprof) top -cum          # Sort by cumulative
(pprof) list CreateActivity   # Source for function
(pprof) web               # Open graph in browser
(pprof) traces            # Show call traces
(pprof) peek CreateActivity   # Quick function overview
```

---

## Common Questions

### Q: What's a "good" percentage in Top view?

**A:**
- **<5%**: Probably not worth optimizing
- **5-10%**: Consider if easy to fix
- **10-20%**: Definitely investigate
- **>20%**: High priority optimization target

### Q: Should I optimize `runtime.*` functions?

**A:**
Usually **no**. Runtime functions (runtime.mallocgc, runtime.scanobject, etc.) are consequences of your code.

**Exception:** If runtime.mallocgc is >30%, you have an allocation problem in *your* code.

### Q: Flame graph vs Call graph - which to use?

**A:**
- **Flame graph**: Quick overview, find hotspots fast
- **Call graph**: Understand relationships, see who calls whom

Use flame graph first, then call graph for deeper analysis.

### Q: How do I know if I have a memory leak?

**A:**
1. Set VIEW to "inuse_space"
2. If you see growing memory over time
3. And functions holding large amounts
4. That's likely a leak

Compare profiles taken at different times.

### Q: Can I profile production?

**A:**
**Yes**, but carefully:
- Profiling has ~5% overhead
- Use short duration (30s max)
- Don't run continuously
- Consider off-peak hours

```bash
# Profile production for 30s
curl http://localhost:6060/debug/pprof/profile?seconds=30 > prod.prof
go tool pprof -http=:8080 prod.prof
```

---

## Summary Checklist

Before closing the profile:

- [ ] Identified top 3 CPU hotspots
- [ ] Identified top 3 memory hotspots
- [ ] Checked for N+1 query patterns
- [ ] Looked for excessive allocations
- [ ] Examined hot paths in source view
- [ ] Prioritized optimizations by impact
- [ ] Saved baseline profile for comparison

---

## Next Steps

1. **Make small changes** - Optimize one thing at a time
2. **Measure again** - Run benchmarks after each change
3. **Compare results** - Use `-base` flag to compare
4. **Document findings** - Note what worked/didn't work
5. **Repeat** - Keep iterating until satisfied

---

## Additional Resources

- **pprof Documentation**: https://github.com/google/pprof
- **Go Blog - Profiling**: https://go.dev/blog/pprof
- **Flame Graph Guide**: http://www.brendangregg.com/flamegraphs.html
- **Go Performance Book**: https://dave.cheney.net/high-performance-go-workshop

---

## Quick Reference

### View Descriptions
| View | Shows | Use When |
|------|-------|----------|
| Graph | Call relationships | Understanding flow |
| Flame Graph | Stacked time usage | Quick hotspot ID |
| Peek | Function summary | Initial triage |
| Source | Line-by-line metrics | Finding exact lines |
| Top | Ranked function list | Prioritizing work |

### Sample Types (Memory)
| Type | Shows | Use For |
|------|-------|---------|
| alloc_space | Total allocated | Finding allocators |
| alloc_objects | Allocation count | Reducing GC pressure |
| inuse_space | Current memory | Finding leaks |
| inuse_objects | Current objects | Object retention |

### Color Meanings
- **Red/Orange**: Hot paths (most time/memory)
- **Yellow**: Moderate usage
- **Green/Blue**: Low usage
- **Gray**: Minimal impact

---

**Remember:** Profile-guided optimization beats guessing every time. Let the data guide you! 🔍🚀
