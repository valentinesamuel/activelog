# MONTH 7: Concurrency Deep Dive

**Weeks:** 25-28
**Phase:** Mastering Go's Superpower
**Theme:** Learn concurrent programming the Go way

---

## Overview

This month focuses on Go's most powerful feature: concurrency. You'll learn goroutines, channels, and synchronization primitives. By mastering these concepts, you'll be able to write highly concurrent, efficient programs that fully utilize modern multi-core processors. This is what makes Go special.

---

## Learning Path

### Week 25: Goroutines Fundamentals
- What are goroutines vs threads
- Creating and managing goroutines
- When to use goroutines
- Common goroutine patterns

### Week 26: Channels (Buffered/Unbuffered)
- Channel basics and syntax
- Buffered vs unbuffered channels
- Channel direction (send-only, receive-only)
- Closing channels properly

### Week 27: Select Statements + Sync Primitives
- Select statement for multiplexing
- Timeouts and cancellation
- sync.Mutex and sync.RWMutex
- sync.WaitGroup for coordination
- sync.Once for initialization

### Week 28: Context for Cancellation + Race Detection
- Context package deep dive
- Cancellation propagation
- Request timeouts with context
- Race detector usage
- Common race conditions

---

## Core Concepts

### Goroutines Fundamentals
```go
// Simple goroutine
func main() {
    go sayHello() // Runs concurrently

    time.Sleep(time.Second) // Wait for goroutine (bad practice, see WaitGroup below)
}

func sayHello() {
    fmt.Println("Hello from goroutine!")
}

// Goroutine with anonymous function
go func() {
    fmt.Println("Hello from anonymous goroutine!")
}()

// Pass data to goroutine
go func(name string) {
    fmt.Printf("Hello %s from goroutine!\n", name)
}("Alice")
```

**Goroutine vs Thread:**
- Goroutines are lightweight (2KB initial stack)
- Threads are heavyweight (1MB+ stack)
- Can have millions of goroutines
- OS threads are limited (thousands)

### Channels (Buffered/Unbuffered)
```go
// Unbuffered channel - blocks until sender and receiver are ready
ch := make(chan int)

go func() {
    ch <- 42 // Send (blocks until receiver is ready)
}()

value := <-ch // Receive (blocks until sender sends)

// Buffered channel - doesn't block until buffer is full
ch := make(chan int, 3) // Buffer size 3

ch <- 1 // Doesn't block
ch <- 2 // Doesn't block
ch <- 3 // Doesn't block
ch <- 4 // BLOCKS until someone receives

// Channel directions (type safety)
func send(ch chan<- int) {    // Can only send
    ch <- 42
}

func receive(ch <-chan int) { // Can only receive
    value := <-ch
}

// Closing channels
ch := make(chan int)
go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // Signal no more values
}()

// Receive until closed
for value := range ch {
    fmt.Println(value)
}
```

### Select Statements
```go
// Select multiplexes multiple channels
select {
case msg1 := <-ch1:
    fmt.Println("Received from ch1:", msg1)
case msg2 := <-ch2:
    fmt.Println("Received from ch2:", msg2)
case ch3 <- 42:
    fmt.Println("Sent to ch3")
default:
    fmt.Println("No channel ready")
}

// Timeout with select
select {
case result := <-ch:
    fmt.Println("Got result:", result)
case <-time.After(5 * time.Second):
    fmt.Println("Timeout!")
}

// Context cancellation with select
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

select {
case result := <-ch:
    return result
case <-ctx.Done():
    return ctx.Err() // Returns context.Canceled
}
```

### Sync Primitives
```go
import "sync"

// Mutex - mutual exclusion lock
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

// RWMutex - multiple readers, single writer
type Cache struct {
    mu    sync.RWMutex
    data  map[string]string
}

func (c *Cache) Get(key string) string {
    c.mu.RLock() // Multiple readers can acquire RLock
    defer c.mu.RUnlock()
    return c.data[key]
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock() // Only one writer can acquire Lock
    defer c.mu.Unlock()
    c.data[key] = value
}

// WaitGroup - wait for goroutines to finish
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1) // Increment counter
    go func(id int) {
        defer wg.Done() // Decrement counter when done
        fmt.Printf("Goroutine %d\n", id)
    }(i)
}

wg.Wait() // Block until counter is 0

// sync.Once - run initialization exactly once
var (
    instance *Database
    once     sync.Once
)

func GetDatabase() *Database {
    once.Do(func() {
        instance = &Database{} // Only runs once, thread-safe
        instance.Connect()
    })
    return instance
}
```

### Context for Cancellation
```go
// Context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel() // Always defer cancel to avoid context leak

result, err := DoWork(ctx)

func DoWork(ctx context.Context) (string, error) {
    select {
    case result := <-heavyWork():
        return result, nil
    case <-ctx.Done():
        return "", ctx.Err() // Returns "context deadline exceeded"
    }
}

// Context with cancellation
ctx, cancel := context.WithCancel(context.Background())

go func() {
    time.Sleep(2 * time.Second)
    cancel() // Cancel after 2 seconds
}()

<-ctx.Done()
fmt.Println("Context cancelled:", ctx.Err())

// Context with values (use sparingly)
ctx := context.WithValue(context.Background(), "user_id", 123)

userID := ctx.Value("user_id").(int)
```

### Race Detection
```go
// Race condition example
type Counter struct {
    value int
}

func (c *Counter) Increment() {
    c.value++ // RACE! Not thread-safe
}

// Run with race detector: go run -race main.go
// Or test: go test -race ./...

// Fix with mutex
type SafeCounter struct {
    mu    sync.Mutex
    value int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    c.value++
    c.mu.Unlock()
}
```

---

# WEEKLY TASK BREAKDOWNS

## Week 25: Goroutines Fundamentals

### 📋 Implementation Tasks

**Task 1: Learn Goroutine Basics** (60 min)
- [ ] Create `examples/goroutines/` directory for practice
- [ ] Write example: simple goroutine with `go func()`
- [ ] Write example: goroutine with parameters
- [ ] Write example: multiple goroutines
- [ ] Understand goroutine vs thread differences
- [ ] Learn about the Go scheduler

**Task 2: Implement Parallel Statistics Calculation** (90 min)
- [ ] Update `internal/services/stats_service.go`
- [ ] Implement `CalculateUserStats(ctx, userID)` with parallel queries
  - **Logic:**
    1. Create result channel: `results := make(chan StatResult, 4)`
    2. Launch 4 goroutines, each querying different stat:
       - `go func() { total := repo.GetTotalActivities(ctx, userID); results <- StatResult{Type: "total", Value: total} }()`
       - Same for distance, duration, streak
    3. Collect 4 results from channel: `for i := 0; i < 4; i++ { r := <-results; ... }`
    4. Combine results into UserStats struct
    5. Return combined stats
    - **Why:** 4 DB queries run in parallel instead of sequential, ~4x faster
- [ ] Launch 4 goroutines for different stats (total activities, distance, duration, streak)
- [ ] Use channel to collect results
- [ ] Handle errors from any goroutine
- [ ] Compare performance: sequential vs parallel

**Task 3: Add WaitGroup for Goroutine Coordination** (60 min)
- [ ] Refactor parallel stats to use `sync.WaitGroup`
- [ ] Add `wg.Add(1)` before launching goroutine
- [ ] Use `defer wg.Done()` in each goroutine
- [ ] Call `wg.Wait()` to wait for all goroutines
- [ ] Test goroutines complete before function returns

**Task 4: Implement Concurrent Photo Processing** (120 min)
- [ ] Update photo upload handler
- [ ] Process multiple photos concurrently
- [ ] Use semaphore pattern to limit concurrency (max 5)
- [ ] Resize and generate thumbnail in parallel
- [ ] Upload both versions to S3 concurrently
- [ ] Collect all results before responding

**Task 5: Handle Goroutine Errors Properly** (45 min)
- [ ] Create error channel for goroutine errors
- [ ] Send errors to channel instead of ignoring
- [ ] Collect first error and cancel remaining work
- [ ] Test error handling (simulate failure in one goroutine)

**Task 6: Write Tests for Concurrent Code** (60 min)
- [ ] Test parallel stats calculation
- [ ] Test concurrent photo processing
- [ ] Test error handling in goroutines
- [ ] Run with race detector: `go test -race ./...`
- [ ] Fix any race conditions found

### 📦 Files You'll Create/Modify

```
examples/
└── goroutines/
    ├── basic.go                   [CREATE]
    ├── waitgroup.go               [CREATE]
    └── errors.go                  [CREATE]

internal/
├── services/
│   └── stats_service.go           [MODIFY - parallel queries]
└── handlers/
    └── photo_handler.go           [MODIFY - concurrent processing]
```

### 🔄 Implementation Order

1. **Learn**: Practice goroutines basics with examples
2. **Apply**: Parallel stats calculation
3. **Coordinate**: Add WaitGroup for proper synchronization
4. **Scale**: Concurrent photo processing with semaphore
5. **Error handling**: Proper error collection from goroutines
6. **Test**: Write tests and use race detector

### ⚠️ Blockers to Watch For

- **Forgetting to wait**: Goroutines may not complete before function returns
- **Race conditions**: Accessing shared data without synchronization
- **Goroutine leaks**: Goroutines that never exit
- **Too many goroutines**: Can exhaust resources
- **Error handling**: Errors in goroutines easily lost
- **Panic handling**: Panic in goroutine crashes entire program

### ✅ Definition of Done

- [ ] Understand goroutines vs threads
- [ ] Can launch and coordinate multiple goroutines
- [ ] Parallel stats 2-3x faster than sequential
- [ ] Concurrent photo processing working
- [ ] All goroutine errors handled properly
- [ ] No race conditions (race detector passes)
- [ ] All tests passing

---

## Week 26: Channels (Buffered/Unbuffered)

### 📋 Implementation Tasks

**Task 1: Learn Channel Basics** (60 min)
- [ ] Create examples for unbuffered channels
- [ ] Create examples for buffered channels
- [ ] Understand blocking vs non-blocking behavior
- [ ] Practice channel direction (send-only, receive-only)
- [ ] Learn proper channel closing

**Task 2: Implement Worker Pool Pattern** (120 min)
- [ ] Create `pkg/workers/pool.go`
- [ ] Implement `WorkerPool` struct with job/result channels
  - **Logic:**
    1. Define: `type WorkerPool struct { jobs chan Job; results chan Result; numWorkers int }`
    2. Create N worker goroutines in `Start()` method
    3. Each worker: `for job := range p.jobs { result := job.Process(); p.results <- result }`
    4. Workers block waiting for jobs on channel
    5. Main code sends jobs to `jobs` channel, reads results from `results` channel
    6. Call `close(jobs)` when done sending - workers exit when channel closes
    - **Why:** Limits concurrency to N workers, prevents creating millions of goroutines
- [ ] Create worker goroutines that process from job channel
- [ ] Send jobs to channel, collect results from result channel
- [ ] Close channels properly when done
- [ ] Test with batch photo processing

**Task 3: Implement Pipeline Pattern** (90 min)
- [ ] Create image processing pipeline:
  - Stage 1: Read images from input channel
  - Stage 2: Resize images
  - Stage 3: Generate thumbnails
  - Stage 4: Upload to S3
- [ ] Chain stages with channels
- [ ] Test pipeline processes images correctly

**Task 4: Add Fan-Out, Fan-In Pattern** (90 min)
- [ ] Implement fan-out: distribute work to N workers
- [ ] Implement fan-in: collect results from N workers
- [ ] Use for parallel database queries
- [ ] Test performance improvement

**Task 5: Handle Channel Errors and Cancellation** (60 min)
- [ ] Add error channel to worker pool
- [ ] Handle errors from workers
- [ ] Implement graceful cancellation
- [ ] Close all channels properly on error

**Task 6: Implement Activity Feed with Channels** (75 min)
- [ ] Fetch activities from multiple friends concurrently
- [ ] Use channels to collect activities
- [ ] Merge and sort activities by date
- [ ] Limit concurrency to avoid overwhelming database

### 📦 Files You'll Create/Modify

```
examples/
└── channels/
    ├── basic.go                   [CREATE]
    ├── buffered.go                [CREATE]
    └── pipeline.go                [CREATE]

pkg/
└── workers/
    ├── pool.go                    [CREATE]
    ├── pool_test.go               [CREATE]
    └── patterns.go                [CREATE]

internal/
└── services/
    └── feed_service.go            [MODIFY - concurrent fetching]
```

### 🔄 Implementation Order

1. **Learn**: Channel basics (buffered/unbuffered)
2. **Worker pool**: Implement reusable worker pool
3. **Pipeline**: Image processing pipeline
4. **Fan-out/in**: Parallel work distribution
5. **Error handling**: Channel error handling
6. **Apply**: Activity feed with concurrent fetching

### ⚠️ Blockers to Watch For

- **Deadlock**: Sending to channel with no receiver
- **Channel leaks**: Not closing channels causes memory leaks
- **Closing twice**: Panic when closing already-closed channel
- **Send on closed**: Panic when sending to closed channel
- **Buffer size**: Too small = blocking, too large = memory waste
- **Only sender closes**: Receiver should never close channel

### ✅ Definition of Done

- [ ] Understand buffered vs unbuffered channels
- [ ] Worker pool pattern implemented and tested
- [ ] Pipeline pattern working for image processing
- [ ] Fan-out/fan-in pattern working
- [ ] Channels closed properly (no leaks)
- [ ] Activity feed faster with concurrent fetching
- [ ] All tests passing

---

## Week 27: Select Statements + Sync Primitives

### 📋 Implementation Tasks

**Task 1: Learn Select Statement** (60 min)
- [ ] Create examples for select with multiple channels
- [ ] Create example with timeout using `time.After()`
- [ ] Create example with default case (non-blocking)
- [ ] Practice select for multiplexing channels

**Task 2: Implement Request Timeouts** (90 min)
- [ ] Add timeout to database queries using select
- [ ] Add timeout to S3 uploads using select
- [ ] Add timeout to external API calls
- [ ] Test timeout triggers correctly
- [ ] Return appropriate error on timeout

**Task 3: Learn Mutex and RWMutex** (60 min)
- [ ] Create examples for `sync.Mutex`
- [ ] Create examples for `sync.RWMutex`
- [ ] Understand difference between Lock() and RLock()
- [ ] Learn when to use each type

**Task 4: Implement Thread-Safe Cache** (90 min)
- [ ] Create `pkg/cache/memory_cache.go`
- [ ] Use `sync.RWMutex` for concurrent access
- [ ] Implement `Get(key)` with RLock (multiple readers)
- [ ] Implement `Set(key, value)` with Lock (single writer)
- [ ] Implement `Delete(key)` with Lock
- [ ] Add TTL support
- [ ] Test concurrent access (no races)

**Task 5: Use sync.Once for Initialization** (45 min)
- [ ] Implement singleton pattern with `sync.Once`
- [ ] Use for one-time database connection initialization
- [ ] Use for one-time S3 client initialization
- [ ] Test thread-safety (multiple goroutines, one initialization)

**Task 6: Add Connection Pool with Sync Primitives** (75 min)
- [ ] Create connection pool for external services
- [ ] Use channel as semaphore for pool size
- [ ] Use mutex for pool state management
- [ ] Implement acquire/release pattern
- [ ] Test pool prevents resource exhaustion

### 📦 Files You'll Create/Modify

```
examples/
└── sync/
    ├── select.go                  [CREATE]
    ├── mutex.go                   [CREATE]
    ├── rwmutex.go                 [CREATE]
    └── once.go                    [CREATE]

pkg/
└── cache/
    ├── memory_cache.go            [CREATE]
    ├── memory_cache_test.go       [CREATE]
    └── connection_pool.go         [CREATE]
```

### 🔄 Implementation Order

1. **Select**: Learn and practice select statements
2. **Timeouts**: Add request timeouts with select
3. **Mutex**: Learn mutex and RWMutex
4. **Cache**: Thread-safe in-memory cache
5. **Once**: Singleton initialization
6. **Pool**: Connection pool with sync primitives

### ⚠️ Blockers to Watch For

- **Deadlock**: Waiting for channel that never sends
- **Lock contention**: Too much locking = poor performance
- **Forgetting unlock**: Always use `defer mu.Unlock()`
- **RLock vs Lock**: Using wrong lock type
- **Select default**: Default makes select non-blocking
- **Nested locks**: Can cause deadlock

### ✅ Definition of Done

- [ ] Understand select statement multiplexing
- [ ] Timeouts implemented for slow operations
- [ ] Thread-safe cache with RWMutex working
- [ ] Singleton pattern with sync.Once
- [ ] Connection pool limiting concurrent connections
- [ ] No race conditions (race detector clean)
- [ ] All tests passing

---

## Week 28: Context for Cancellation + Race Detection

### 📋 Implementation Tasks

**Task 1: Learn Context Package** (60 min)
- [ ] Create examples for `context.WithTimeout`
- [ ] Create examples for `context.WithCancel`
- [ ] Create examples for `context.WithDeadline`
- [ ] Create examples for `context.WithValue` (use sparingly)
- [ ] Understand context propagation

**Task 2: Add Context to All Repository Methods** (90 min)
- [ ] Update all repository methods to accept `context.Context`
- [ ] Pass context to all database queries
- [ ] Test query cancellation when context cancelled
- [ ] Verify context deadline respected

**Task 3: Implement Request-Scoped Cancellation** (75 min)
- [ ] Extract request context in handlers
- [ ] Pass context through service layer
- [ ] Pass context to repository layer
- [ ] Test cancellation propagates through entire stack
- [ ] Verify work stops when request cancelled

**Task 4: Add Context to Background Jobs** (60 min)
- [ ] Update job handlers to use context
- [ ] Set timeout context for each job type
- [ ] Handle `context.DeadlineExceeded` error
- [ ] Test job cancellation on timeout

**Task 5: Implement Graceful Degradation** (90 min)
- [ ] If cache query times out, fall back to database
- [ ] If S3 upload times out, retry with exponential backoff
- [ ] If external API times out, return cached data
- [ ] Log timeouts for monitoring

**Task 6: Run Race Detector on Entire Codebase** (120 min)
- [ ] Run `go test -race ./...` on all packages
- [ ] Fix all race conditions found
- [ ] Common issues to fix:
  - Concurrent map access
  - Shared variable access without mutex
  - Channel operations without sync
- [ ] Re-run until no races detected

**Task 7: Add CI Check for Race Conditions** (30 min)
- [ ] Add race detector to GitHub Actions workflow
- [ ] Make race detection required for PR merge
- [ ] Document race detector usage in README

### 📦 Files You'll Create/Modify

```
examples/
└── context/
    ├── timeout.go                 [CREATE]
    ├── cancel.go                  [CREATE]
    └── values.go                  [CREATE]

internal/
├── repository/
│   └── *.go                       [MODIFY - add context]
├── services/
│   └── *.go                       [MODIFY - pass context]
└── jobs/
    └── handlers.go                [MODIFY - use context]

.github/workflows/
└── test.yml                       [MODIFY - add race detector]
```

### 🔄 Implementation Order

1. **Learn**: Context package examples
2. **Repository**: Add context to all queries
3. **Propagation**: Pass context through layers
4. **Jobs**: Add context to background jobs
5. **Degradation**: Graceful fallback on timeout
6. **Race detection**: Fix all race conditions
7. **CI**: Automate race detection

### ⚠️ Blockers to Watch For

- **Context leaks**: Always call cancel() (use defer)
- **Ignoring context**: Not respecting context cancellation
- **Wrong context**: Using Background() instead of request context
- **Value overuse**: Don't abuse context.WithValue
- **Race conditions**: Concurrent access without synchronization
- **False positives**: Some races are benign (investigate carefully)

### ✅ Definition of Done

- [ ] All repository methods accept context
- [ ] Request cancellation propagates through stack
- [ ] Job timeouts working correctly
- [ ] Graceful degradation on timeouts
- [ ] Zero race conditions (race detector clean)
- [ ] CI enforces race-free code
- [ ] All tests passing with -race flag

---

## Patterns to Master

### Fan-Out, Fan-In
```go
// Fan-out: distribute work to multiple goroutines
// Fan-in: collect results from multiple goroutines

func FanOutFanIn(items []Item) []Result {
    numWorkers := runtime.NumCPU()

    // Fan-out: create input channel
    jobs := make(chan Item, len(items))
    results := make(chan Result, len(items))

    // Start workers (fan-out)
    for w := 0; w < numWorkers; w++ {
        go worker(jobs, results)
    }

    // Send jobs
    for _, item := range items {
        jobs <- item
    }
    close(jobs)

    // Collect results (fan-in)
    collected := make([]Result, 0, len(items))
    for i := 0; i < len(items); i++ {
        collected = append(collected, <-results)
    }

    return collected
}

func worker(jobs <-chan Item, results chan<- Result) {
    for job := range jobs {
        results <- processItem(job)
    }
}
```

### Pipeline Pattern
```go
// Pipeline: chain processing stages with channels

func pipeline() {
    // Stage 1: Generate numbers
    nums := generate(1, 2, 3, 4, 5)

    // Stage 2: Square numbers
    squared := square(nums)

    // Stage 3: Print results
    for result := range squared {
        fmt.Println(result)
    }
}

func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}
```

### Worker Pool Pattern
```go
type Job struct {
    ID   int
    Data string
}

type Result struct {
    JobID int
    Value string
}

func WorkerPool(jobs []Job, numWorkers int) []Result {
    jobsCh := make(chan Job, len(jobs))
    resultsCh := make(chan Result, len(jobs))

    // Start worker pool
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for job := range jobsCh {
                // Process job
                result := Result{
                    JobID: job.ID,
                    Value: processJob(job),
                }
                resultsCh <- result
            }
        }(i)
    }

    // Send jobs
    go func() {
        for _, job := range jobs {
            jobsCh <- job
        }
        close(jobsCh)
    }()

    // Wait for workers to finish and close results
    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    // Collect results
    results := make([]Result, 0, len(jobs))
    for result := range resultsCh {
        results = append(results, result)
    }

    return results
}
```

---

## Practical Applications

### Parallel Statistics Calculation
```go
func (s *StatsService) CalculateUserStats(ctx context.Context, userID int) (*Stats, error) {
    type result struct {
        name  string
        value interface{}
        err   error
    }

    resultsCh := make(chan result, 4)

    // Calculate different stats in parallel
    go func() {
        total, err := s.repo.GetTotalActivities(ctx, userID)
        resultsCh <- result{"total", total, err}
    }()

    go func() {
        distance, err := s.repo.GetTotalDistance(ctx, userID)
        resultsCh <- result{"distance", distance, err}
    }()

    go func() {
        duration, err := s.repo.GetTotalDuration(ctx, userID)
        resultsCh <- result{"duration", duration, err}
    }()

    go func() {
        streak, err := s.repo.GetCurrentStreak(ctx, userID)
        resultsCh <- result{"streak", streak, err}
    }()

    // Collect results
    stats := &Stats{}
    for i := 0; i < 4; i++ {
        r := <-resultsCh
        if r.err != nil {
            return nil, r.err
        }

        switch r.name {
        case "total":
            stats.TotalActivities = r.value.(int)
        case "distance":
            stats.TotalDistance = r.value.(float64)
        case "duration":
            stats.TotalDuration = r.value.(int)
        case "streak":
            stats.CurrentStreak = r.value.(int)
        }
    }

    return stats, nil
}
```

### Concurrent File Processing
```go
func ProcessPhotos(ctx context.Context, photos []Photo) error {
    sem := make(chan struct{}, 5) // Semaphore: max 5 concurrent
    errCh := make(chan error, len(photos))

    var wg sync.WaitGroup

    for _, photo := range photos {
        wg.Add(1)
        go func(p Photo) {
            defer wg.Done()

            // Acquire semaphore
            sem <- struct{}{}
            defer func() { <-sem }() // Release

            if err := processPhoto(ctx, p); err != nil {
                errCh <- err
            }
        }(photo)
    }

    // Wait for completion
    go func() {
        wg.Wait()
        close(errCh)
    }()

    // Check for errors
    for err := range errCh {
        if err != nil {
            return err
        }
    }

    return nil
}
```

---

## Common Pitfalls

1. **Goroutine leaks**
   - ❌ Creating goroutines that never exit
   - ✅ Use context for cancellation

2. **Race conditions**
   - ❌ Accessing shared data without synchronization
   - ✅ Use mutexes or channels

3. **Closing channels from receiver**
   - ❌ `close(ch)` in receiver
   - ✅ Only sender should close

4. **Not using WaitGroup**
   - ❌ `time.Sleep()` to wait for goroutines
   - ✅ Use `sync.WaitGroup`

5. **Forgetting to call cancel()**
   - ❌ Context leaks
   - ✅ Always `defer cancel()`

---

## Testing Concurrent Code

```go
func TestConcurrentCounter(t *testing.T) {
    counter := &SafeCounter{}

    var wg sync.WaitGroup
    numGoroutines := 100
    incrementsPerGoroutine := 1000

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < incrementsPerGoroutine; j++ {
                counter.Increment()
            }
        }()
    }

    wg.Wait()

    expected := numGoroutines * incrementsPerGoroutine
    assert.Equal(t, expected, counter.value)
}

// Run with race detector
// go test -race -count=100 ./...
```

---

## Resources

- [Go Concurrency Patterns (Rob Pike)](https://www.youtube.com/watch?v=f6kdp27TYZs)
- [Concurrency in Go (Book)](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/)
- [Go Blog: Share Memory By Communicating](https://go.dev/blog/codelab-share)
- [Effective Go: Concurrency](https://go.dev/doc/effective_go#concurrency)

---

## Next Steps

After completing Month 7, you'll move to **Month 8: Social Features & Real-time**, where you'll learn:
- Friend system implementation
- Activity feed with real-time updates
- WebSocket integration
- Feature flags system

**You now understand Go's most powerful feature!** 🚀
