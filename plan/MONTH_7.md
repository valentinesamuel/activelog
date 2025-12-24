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
