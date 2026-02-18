package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/valentinesamuel/activelog/internal/adapters/queue/types"
	"github.com/valentinesamuel/activelog/internal/service"
)

// Scheduler wraps robfig/cron and wires in domain services.
type Scheduler struct {
	cron      *cron.Cron
	statsCalc *service.StatsCalculator
	cleanup   *service.CleanupService
	queue     types.QueueProvider
}

// New creates a UTC-based Scheduler.
func New(
	statsCalc *service.StatsCalculator,
	cleanup *service.CleanupService,
	queue types.QueueProvider,
) *Scheduler {
	c := cron.New(cron.WithLocation(time.UTC))
	return &Scheduler{
		cron:      c,
		statsCalc: statsCalc,
		cleanup:   cleanup,
		queue:     queue,
	}
}

// Start registers all cron jobs and starts the scheduler.
func (s *Scheduler) Start() {
	// Daily stats calculation at midnight UTC
	s.cron.AddFunc("0 0 * * *", func() {
		ctx := context.Background()
		if err := s.statsCalc.CalculateDailyStats(ctx); err != nil {
			log.Printf("[scheduler] CalculateDailyStats error: %v", err)
		}
	})

	// Weekly summary emails every Sunday at 09:00 UTC
	s.cron.AddFunc("0 9 * * 0", func() {
		s.enqueueWeeklySummaries()
	})

	// Monthly report generation on the 1st of each month at midnight UTC
	s.cron.AddFunc("0 0 1 * *", func() {
		s.enqueueMonthlyReports()
	})

	// Cleanup stale soft-deleted records every day at 02:00 UTC
	s.cron.AddFunc("0 2 * * *", func() {
		ctx := context.Background()
		if err := s.cleanup.DeleteOldData(ctx); err != nil {
			log.Printf("[scheduler] DeleteOldData error: %v", err)
		}
	})

	s.cron.Start()
	log.Println("[scheduler] started (UTC)")
}

// Stop gracefully stops the scheduler and waits for running jobs to finish.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("[scheduler] stopped")
}

// enqueueWeeklySummaries enqueues a WeeklySummary job for every active user.
// For now it enqueues a placeholder job; the real user list will come from the
// user repository once it exposes a ListActiveUsers method.
func (s *Scheduler) enqueueWeeklySummaries() {
	ctx := context.Background()
	log.Println("[scheduler] enqueue weekly summaries (placeholder – no active users yet)")

	// Example: iterate active user IDs and enqueue per user.
	// for _, userID := range activeUserIDs {
	//     s.enqueueJob(ctx, types.InboxQueue, types.EventWeeklySummary, map[string]int{"user_id": userID})
	// }
	_ = ctx
}

// enqueueMonthlyReports enqueues a GenerateExport job for every active user.
func (s *Scheduler) enqueueMonthlyReports() {
	ctx := context.Background()
	log.Println("[scheduler] enqueue monthly reports (placeholder – no active users yet)")
	_ = ctx
}

// enqueueJob is a helper that marshals data and enqueues a job.
func (s *Scheduler) enqueueJob(ctx context.Context, queue types.QueueName, event types.EventType, data any) {
	raw, err := json.Marshal(data)
	if err != nil {
		log.Printf("[scheduler] marshal error for event %s: %v", event, err)
		return
	}

	taskID, err := s.queue.Enqueue(ctx, queue, types.JobPayload{
		Event: event,
		Data:  raw,
	})
	if err != nil {
		log.Printf("[scheduler] enqueue error for event %s: %v", event, err)
		return
	}

	log.Printf("[scheduler] enqueued event=%s queue=%s taskID=%s", event, queue, taskID)
}
