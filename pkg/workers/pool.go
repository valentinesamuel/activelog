package workers

import "sync"

// WorkerPool distributes jobs across a fixed number of goroutines and collects results.
// J is the job type, R is the result type.
type WorkerPool[J any, R any] struct {
	numWorkers int
}

// New creates a WorkerPool with the given number of workers.
// numWorkers is clamped to a minimum of 1.
func New[J any, R any](numWorkers int) *WorkerPool[J, R] {
	if numWorkers < 1 {
		numWorkers = 1
	}
	return &WorkerPool[J, R]{numWorkers: numWorkers}
}

// Submit fans out jobs to worker goroutines and returns results in completion order.
// The result slice length equals the number of jobs submitted.
func (p *WorkerPool[J, R]) Submit(jobs []J, fn func(J) R) []R {
	if len(jobs) == 0 {
		return nil
	}

	jobCh := make(chan J, len(jobs))
	resultCh := make(chan R, len(jobs))

	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < p.numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				resultCh <- fn(job)
			}
		}()
	}

	// Enqueue all jobs then close the input channel
	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)

	// Close result channel once all workers are done
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Drain results
	results := make([]R, 0, len(jobs))
	for result := range resultCh {
		results = append(results, result)
	}
	return results
}
