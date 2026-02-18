package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/valentinesamuel/activelog/internal/platform/config"
	"github.com/valentinesamuel/activelog/internal/platform/jobs"
	internalAsynq "github.com/valentinesamuel/activelog/internal/adapters/queue/asynq"
	"github.com/valentinesamuel/activelog/internal/adapters/queue/memory"
	queueTypes "github.com/valentinesamuel/activelog/internal/adapters/queue/types"
)

func main() {
	fmt.Println("Starting ActiveLog Worker...")

	if err := run(); err != nil {
		log.Fatalf("Worker error: %v", err)
	}
}

func run() error {
	config.MustLoad()

	factory := jobs.NewHandlerFactory()
	factory.Register(queueTypes.EventWelcomeEmail, jobs.HandleWelcomeEmail)
	factory.Register(queueTypes.EventWeeklySummary, jobs.HandleWeeklySummary)
	factory.Register(queueTypes.EventGenerateExport, jobs.HandleGenerateExport)
	factory.Register(queueTypes.EventRefreshRateLimitConfig, jobs.HandleRefreshRateLimitConfig)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if config.Queue.Provider == "asynq" {
		return runAsynqWorker(ctx, factory, quit)
	}

	return runMemoryWorker(ctx, factory, quit)
}

func runAsynqWorker(_ context.Context, factory *jobs.HandlerFactory, quit <-chan os.Signal) error {
	redisAddr := config.GetEnv("REDIS_ADDRESS", "localhost:6379")
	srv := internalAsynq.NewWorkerServer(redisAddr, 10)

	mux := asynq.NewServeMux()
	handler := func(_ context.Context, t *asynq.Task) error {
		var payload queueTypes.JobPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("worker: unmarshal payload: %w", err)
		}
		return factory.Dispatch(context.Background(), payload)
	}

	for _, event := range []queueTypes.EventType{
		queueTypes.EventWelcomeEmail,
		queueTypes.EventWeeklySummary,
		queueTypes.EventGenerateExport,
		queueTypes.EventSendVerificationEmail,
		queueTypes.EventActivityCreated,
		queueTypes.EventActivityDeleted,
		queueTypes.EventRefreshRateLimitConfig,
	} {
		mux.HandleFunc(string(event), handler)
	}

	log.Println("asynq worker started")
	if err := srv.Start(mux); err != nil {
		return fmt.Errorf("asynq worker failed to start: %w", err)
	}

	<-quit
	log.Println("Shutting down asynq worker...")
	srv.Shutdown()
	return nil
}

func runMemoryWorker(ctx context.Context, factory *jobs.HandlerFactory, quit <-chan os.Signal) error {
	mem := memory.New(100)

	for _, queue := range []queueTypes.QueueName{queueTypes.InboxQueue, queueTypes.OutboxQueue} {
		mem.StartWorking(ctx, queue, factory.Dispatch)
	}

	log.Println("memory worker started")
	<-quit
	log.Println("Shutting down memory worker...")
	return nil
}
