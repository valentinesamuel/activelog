# Broker Architecture in Go - Event-Driven Design

## Overview

A **broker** is a messaging system that enables asynchronous communication between services or within a service. Common use cases:
- Publishing domain events (e.g., "ActivityCreated")
- Background job processing
- Service-to-service communication
- Event sourcing
- CQRS implementation

---

## Where Broker Fits in Clean Architecture

```
┌─────────────────────────────────────────┐
│   Interfaces Layer (HTTP)               │
├─────────────────────────────────────────┤
│   Application Layer (Use Cases)         │  ← Publishes events
├─────────────────────────────────────────┤
│   Domain Layer (Entities + Events)      │  ← Defines events
├─────────────────────────────────────────┤
│   Infrastructure Layer                  │
│   ├── Persistence (Database)            │
│   ├── Logging                           │
│   └── Broker (RabbitMQ/Kafka/Redis)    │  ← Implements broker
└─────────────────────────────────────────┘

Flow:
1. Use case creates domain event
2. Use case publishes event to broker
3. Broker delivers to subscribers
4. Subscribers handle events asynchronously
```

---

## Directory Structure with Broker

```
activelog/
├── internal/
│   ├── domain/
│   │   ├── activity/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   ├── events.go              # NEW: Domain events
│   │   │   └── service.go
│   │   └── shared/
│   │       └── event.go               # NEW: Base event interface
│   │
│   ├── application/
│   │   ├── activity/
│   │   │   ├── usecases/
│   │   │   │   └── create_activity.go
│   │   │   └── handlers/              # NEW: Event handlers
│   │   │       └── activity_created_handler.go
│   │   └── ports/
│   │       └── event_publisher.go     # NEW: Publisher interface (port)
│   │
│   ├── infrastructure/
│   │   ├── broker/                    # NEW: Broker implementations
│   │   │   ├── rabbitmq/
│   │   │   │   ├── publisher.go
│   │   │   │   ├── subscriber.go
│   │   │   │   ├── connection.go
│   │   │   │   └── config.go
│   │   │   ├── redis/
│   │   │   │   ├── publisher.go
│   │   │   │   └── subscriber.go
│   │   │   ├── kafka/
│   │   │   │   ├── publisher.go
│   │   │   │   └── subscriber.go
│   │   │   └── memory/                # For testing
│   │   │       ├── publisher.go
│   │   │       └── subscriber.go
│   │   └── persistence/
│   │
│   └── interfaces/
│       └── broker/                    # NEW: Broker worker/consumer
│           ├── worker.go              # Event consumer
│           └── handlers.go            # Maps events to handlers
│
└── tests/
    └── integration/
        └── broker/                    # Broker integration tests
            └── event_flow_test.go
```

---

## Domain Layer - Define Events

### Base Event Interface (`internal/domain/shared/event.go`)

```go
package shared

import (
	"time"
)

// Event is the base interface all domain events must implement
type Event interface {
	EventName() string
	OccurredAt() time.Time
	AggregateID() string
	EventVersion() int
}

// BaseEvent provides common event fields
type BaseEvent struct {
	Name        string
	Timestamp   time.Time
	AggregateId string
	Version     int
}

func (e BaseEvent) EventName() string {
	return e.Name
}

func (e BaseEvent) OccurredAt() time.Time {
	return e.Timestamp
}

func (e BaseEvent) AggregateID() string {
	return e.AggregateId
}

func (e BaseEvent) EventVersion() int {
	return e.Version
}
```

### Activity Domain Events (`internal/domain/activity/events.go`)

```go
package activity

import (
	"time"

	"github.com/valentinesamuel/activelog/internal/domain/shared"
)

// ActivityCreatedEvent is published when a new activity is created
type ActivityCreatedEvent struct {
	shared.BaseEvent
	ActivityID      int64
	UserID          int
	Type            ActivityType
	Title           string
	DurationMinutes int
	ActivityDate    time.Time
}

// NewActivityCreatedEvent creates a new ActivityCreated event
func NewActivityCreatedEvent(activity *Activity) *ActivityCreatedEvent {
	return &ActivityCreatedEvent{
		BaseEvent: shared.BaseEvent{
			Name:        "activity.created",
			Timestamp:   time.Now(),
			AggregateId: fmt.Sprintf("activity-%d", activity.ID),
			Version:     1,
		},
		ActivityID:      activity.ID,
		UserID:          activity.UserID,
		Type:            activity.Type,
		Title:           activity.Title,
		DurationMinutes: activity.DurationMinutes,
		ActivityDate:    activity.ActivityDate,
	}
}

// ActivityUpdatedEvent is published when an activity is updated
type ActivityUpdatedEvent struct {
	shared.BaseEvent
	ActivityID      int64
	UserID          int
	UpdatedFields   []string // Which fields were changed
	PreviousTitle   string
	NewTitle        string
}

// ActivityDeletedEvent is published when an activity is deleted
type ActivityDeletedEvent struct {
	shared.BaseEvent
	ActivityID int64
	UserID     int
	DeletedAt  time.Time
}

// NewActivityDeletedEvent creates a new ActivityDeleted event
func NewActivityDeletedEvent(activityID int64, userID int) *ActivityDeletedEvent {
	return &ActivityDeletedEvent{
		BaseEvent: shared.BaseEvent{
			Name:        "activity.deleted",
			Timestamp:   time.Now(),
			AggregateId: fmt.Sprintf("activity-%d", activityID),
			Version:     1,
		},
		ActivityID: activityID,
		UserID:     userID,
		DeletedAt:  time.Now(),
	}
}
```

---

## Application Layer - Event Publisher Port

### Publisher Interface (`internal/application/ports/event_publisher.go`)

```go
package ports

import (
	"context"

	"github.com/valentinesamuel/activelog/internal/domain/shared"
)

// EventPublisher is the interface for publishing domain events
// This is a PORT in hexagonal architecture
type EventPublisher interface {
	// Publish sends an event to the broker
	Publish(ctx context.Context, event shared.Event) error

	// PublishBatch sends multiple events in one operation
	PublishBatch(ctx context.Context, events []shared.Event) error

	// Close closes the connection to the broker
	Close() error
}

// EventSubscriber is the interface for subscribing to events
type EventSubscriber interface {
	// Subscribe starts listening for events
	Subscribe(ctx context.Context, eventNames []string, handler EventHandler) error

	// Close stops the subscriber
	Close() error
}

// EventHandler handles received events
type EventHandler interface {
	Handle(ctx context.Context, event shared.Event) error
}
```

---

## Infrastructure Layer - Broker Implementations

### RabbitMQ Publisher (`internal/infrastructure/broker/rabbitmq/publisher.go`)

```go
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/valentinesamuel/activelog/internal/application/ports"
	"github.com/valentinesamuel/activelog/internal/domain/shared"
)

type Publisher struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	exchange string
}

var _ ports.EventPublisher = (*Publisher)(nil) // Compile-time interface check

func NewPublisher(conn *amqp091.Connection, exchange string) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		exchange,
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &Publisher{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, event shared.Event) error {
	// Serialize event to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to RabbitMQ
	err = p.channel.PublishWithContext(
		ctx,
		p.exchange,       // exchange
		event.EventName(), // routing key
		false,            // mandatory
		false,            // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
			Timestamp:    event.OccurredAt(),
			MessageId:    event.AggregateID(),
			Type:         event.EventName(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (p *Publisher) PublishBatch(ctx context.Context, events []shared.Event) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (p *Publisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	return p.conn.Close()
}
```

### RabbitMQ Subscriber (`internal/infrastructure/broker/rabbitmq/subscriber.go`)

```go
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"github.com/valentinesamuel/activelog/internal/application/ports"
	"github.com/valentinesamuel/activelog/internal/domain/shared"
)

type Subscriber struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	exchange string
	queue    string
}

func NewSubscriber(conn *amqp091.Connection, exchange, queue string) (*Subscriber, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		queue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Subscriber{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
		queue:    queue,
	}, nil
}

func (s *Subscriber) Subscribe(
	ctx context.Context,
	eventNames []string,
	handler ports.EventHandler,
) error {
	// Bind queue to exchange for each event type
	for _, eventName := range eventNames {
		err := s.channel.QueueBind(
			s.queue,
			eventName,   // routing key
			s.exchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	// Start consuming
	msgs, err := s.channel.Consume(
		s.queue,
		"",    // consumer tag
		false, // auto-ack (we'll ack manually)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Process messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				// Deserialize event
				var event shared.BaseEvent
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					log.Printf("Failed to unmarshal event: %v", err)
					msg.Nack(false, false) // Don't requeue
					continue
				}

				// Handle event
				if err := handler.Handle(ctx, &event); err != nil {
					log.Printf("Failed to handle event: %v", err)
					msg.Nack(false, true) // Requeue
					continue
				}

				// Acknowledge
				msg.Ack(false)
			}
		}
	}()

	return nil
}

func (s *Subscriber) Close() error {
	if err := s.channel.Close(); err != nil {
		return err
	}
	return s.conn.Close()
}
```

### Connection Setup (`internal/infrastructure/broker/rabbitmq/connection.go`)

```go
package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Vhost    string
}

func NewConnection(cfg Config) (*amqp091.Connection, error) {
	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Vhost,
	)

	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return conn, nil
}
```

### In-Memory Publisher (For Testing) (`internal/infrastructure/broker/memory/publisher.go`)

```go
package memory

import (
	"context"
	"sync"

	"github.com/valentinesamuel/activelog/internal/application/ports"
	"github.com/valentinesamuel/activelog/internal/domain/shared"
)

// InMemoryBroker is a simple in-memory event broker for testing
type InMemoryBroker struct {
	mu         sync.RWMutex
	events     []shared.Event
	handlers   map[string][]ports.EventHandler
}

func NewInMemoryBroker() *InMemoryBroker {
	return &InMemoryBroker{
		events:   make([]shared.Event, 0),
		handlers: make(map[string][]ports.EventHandler),
	}
}

func (b *InMemoryBroker) Publish(ctx context.Context, event shared.Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Store event
	b.events = append(b.events, event)

	// Immediately call handlers (synchronous for testing)
	handlers, exists := b.handlers[event.EventName()]
	if !exists {
		return nil
	}

	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (b *InMemoryBroker) PublishBatch(ctx context.Context, events []shared.Event) error {
	for _, event := range events {
		if err := b.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (b *InMemoryBroker) Subscribe(
	ctx context.Context,
	eventNames []string,
	handler ports.EventHandler,
) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, eventName := range eventNames {
		b.handlers[eventName] = append(b.handlers[eventName], handler)
	}

	return nil
}

func (b *InMemoryBroker) GetPublishedEvents() []shared.Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.events
}

func (b *InMemoryBroker) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = make([]shared.Event, 0)
}

func (b *InMemoryBroker) Close() error {
	return nil
}
```

---

## Use Cases - Publishing Events

### Updated Create Activity Use Case

```go
// internal/application/activity/usecases/create_activity.go
package usecases

import (
	"context"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/application/activity/dto"
	"github.com/valentinesamuel/activelog/internal/application/ports"
	"github.com/valentinesamuel/activelog/internal/domain/activity"
)

type CreateActivityUseCase struct {
	activityRepo activity.Repository
	publisher    ports.EventPublisher // NEW: Event publisher
}

func NewCreateActivityUseCase(
	activityRepo activity.Repository,
	publisher ports.EventPublisher,
) *CreateActivityUseCase {
	return &CreateActivityUseCase{
		activityRepo: activityRepo,
		publisher:    publisher,
	}
}

func (uc *CreateActivityUseCase) Execute(
	ctx context.Context,
	req dto.CreateActivityRequest,
) (*dto.ActivityResponse, error) {
	// 1. Convert DTO to domain entity
	act := &activity.Activity{
		UserID:          req.UserID,
		Type:            activity.ActivityType(req.Type),
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		DistanceKm:      req.DistanceKm,
		CaloriesBurned:  req.CaloriesBurned,
		Notes:           req.Notes,
		ActivityDate:    req.ActivityDate,
	}

	// 2. Validate business rules
	if err := act.Validate(); err != nil {
		return nil, err
	}

	// 3. Persist
	if err := uc.activityRepo.Create(ctx, act); err != nil {
		return nil, err
	}

	// 4. Publish domain event (NEW!)
	event := activity.NewActivityCreatedEvent(act)
	if err := uc.publisher.Publish(ctx, event); err != nil {
		// Log error but don't fail the request
		// Consider using outbox pattern for guaranteed delivery
		fmt.Printf("Failed to publish event: %v\n", err)
	}

	// 5. Convert to response DTO
	return &dto.ActivityResponse{
		ID:              act.ID,
		UserID:          act.UserID,
		Type:            string(act.Type),
		Title:           act.Title,
		Description:     act.Description,
		DurationMinutes: act.DurationMinutes,
		DistanceKm:      act.DistanceKm,
		CaloriesBurned:  act.CaloriesBurned,
		Notes:           act.Notes,
		ActivityDate:    act.ActivityDate,
		CreatedAt:       act.CreatedAt,
		UpdatedAt:       act.UpdatedAt,
	}, nil
}
```

---

## Event Handlers - Subscribing to Events

### Activity Created Handler (`internal/application/activity/handlers/activity_created_handler.go`)

```go
package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/valentinesamuel/activelog/internal/domain/activity"
	"github.com/valentinesamuel/activelog/internal/domain/shared"
)

// ActivityCreatedHandler handles ActivityCreated events
type ActivityCreatedHandler struct {
	// Add dependencies (e.g., notification service, analytics service)
}

func NewActivityCreatedHandler() *ActivityCreatedHandler {
	return &ActivityCreatedHandler{}
}

func (h *ActivityCreatedHandler) Handle(ctx context.Context, event shared.Event) error {
	// Type assert to specific event
	activityEvent, ok := event.(*activity.ActivityCreatedEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", event)
	}

	log.Printf("Handling ActivityCreated: %+v", activityEvent)

	// Example actions:
	// 1. Send notification to user
	// 2. Update analytics/stats
	// 3. Trigger achievement checks
	// 4. Update leaderboard
	// 5. Send webhook to external services

	// Simulated processing
	log.Printf("Activity %d created by user %d: %s",
		activityEvent.ActivityID,
		activityEvent.UserID,
		activityEvent.Title,
	)

	return nil
}
```

### Multiple Event Handlers (`internal/application/activity/handlers/stats_updater.go`)

```go
package handlers

import (
	"context"
	"log"

	"github.com/valentinesamuel/activelog/internal/domain/activity"
	"github.com/valentinesamuel/activelog/internal/domain/shared"
	"github.com/valentinesamuel/activelog/internal/domain/stats"
)

// StatsUpdaterHandler updates user statistics when activities change
type StatsUpdaterHandler struct {
	statsRepo stats.Repository
}

func NewStatsUpdaterHandler(statsRepo stats.Repository) *StatsUpdaterHandler {
	return &StatsUpdaterHandler{
		statsRepo: statsRepo,
	}
}

func (h *StatsUpdaterHandler) Handle(ctx context.Context, event shared.Event) error {
	switch e := event.(type) {
	case *activity.ActivityCreatedEvent:
		return h.handleCreated(ctx, e)
	case *activity.ActivityDeletedEvent:
		return h.handleDeleted(ctx, e)
	default:
		log.Printf("Ignoring event type: %T", event)
		return nil
	}
}

func (h *StatsUpdaterHandler) handleCreated(
	ctx context.Context,
	event *activity.ActivityCreatedEvent,
) error {
	log.Printf("Updating stats for user %d after activity creation", event.UserID)

	// Update user's weekly/monthly stats
	// This happens asynchronously, doesn't block the main request
	return h.statsRepo.IncrementActivity(ctx, event.UserID, event.DurationMinutes)
}

func (h *StatsUpdaterHandler) handleDeleted(
	ctx context.Context,
	event *activity.ActivityDeletedEvent,
) error {
	log.Printf("Updating stats for user %d after activity deletion", event.UserID)
	return h.statsRepo.DecrementActivity(ctx, event.UserID)
}
```

---

## Worker/Consumer Setup

### Worker (`internal/interfaces/broker/worker.go`)

```go
package broker

import (
	"context"
	"log"

	"github.com/valentinesamuel/activelog/internal/application/activity/handlers"
	"github.com/valentinesamuel/activelog/internal/application/ports"
)

// Worker consumes events from the broker
type Worker struct {
	subscriber ports.EventSubscriber
	handlers   map[string]ports.EventHandler
}

func NewWorker(subscriber ports.EventSubscriber) *Worker {
	return &Worker{
		subscriber: subscriber,
		handlers:   make(map[string]ports.EventHandler),
	}
}

// RegisterHandler registers a handler for specific event types
func (w *Worker) RegisterHandler(eventName string, handler ports.EventHandler) {
	w.handlers[eventName] = handler
}

// Start begins consuming events
func (w *Worker) Start(ctx context.Context) error {
	// Get list of event names we're handling
	eventNames := make([]string, 0, len(w.handlers))
	for eventName := range w.handlers {
		eventNames = append(eventNames, eventName)
	}

	// Create a routing handler
	router := &EventRouter{handlers: w.handlers}

	// Subscribe to events
	log.Printf("Starting worker, subscribing to: %v", eventNames)
	return w.subscriber.Subscribe(ctx, eventNames, router)
}

func (w *Worker) Stop() error {
	return w.subscriber.Close()
}

// EventRouter routes events to the correct handler
type EventRouter struct {
	handlers map[string]ports.EventHandler
}

func (r *EventRouter) Handle(ctx context.Context, event shared.Event) error {
	handler, exists := r.handlers[event.EventName()]
	if !exists {
		log.Printf("No handler registered for event: %s", event.EventName())
		return nil
	}

	return handler.Handle(ctx, event)
}
```

---

## Wiring It All Together - Main Application

### Main with Broker (`cmd/api/main.go`)

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/valentinesamuel/activelog/internal/application/activity/handlers"
	"github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	"github.com/valentinesamuel/activelog/internal/infrastructure/broker/rabbitmq"
	"github.com/valentinesamuel/activelog/internal/infrastructure/persistence/postgres"
	brokerWorker "github.com/valentinesamuel/activelog/internal/interfaces/broker"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Setup Database
	db := setupDatabase()
	defer db.Close()

	// 2. Setup Broker Connection
	brokerConn, err := rabbitmq.NewConnection(rabbitmq.Config{
		Host:     "localhost",
		Port:     5672,
		User:     "guest",
		Password: "guest",
		Vhost:    "/",
	})
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer brokerConn.Close()

	// 3. Create Publisher
	publisher, err := rabbitmq.NewPublisher(brokerConn, "activelog.events")
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// 4. Create Repositories
	activityRepo := postgres.NewActivityRepository(db)
	statsRepo := postgres.NewStatsRepository(db)

	// 5. Create Use Cases (with publisher)
	createActivityUseCase := usecases.NewCreateActivityUseCase(
		activityRepo,
		publisher, // Inject publisher
	)

	// 6. Create Event Handlers
	activityCreatedHandler := handlers.NewActivityCreatedHandler()
	statsUpdater := handlers.NewStatsUpdaterHandler(statsRepo)

	// 7. Setup Subscriber & Worker
	subscriber, err := rabbitmq.NewSubscriber(
		brokerConn,
		"activelog.events",
		"activelog.worker",
	)
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}

	worker := brokerWorker.NewWorker(subscriber)
	worker.RegisterHandler("activity.created", activityCreatedHandler)
	worker.RegisterHandler("activity.created", statsUpdater) // Multiple handlers per event
	worker.RegisterHandler("activity.deleted", statsUpdater)

	// 8. Start Worker in Background
	go func() {
		if err := worker.Start(ctx); err != nil {
			log.Printf("Worker error: %v", err)
		}
	}()

	// 9. Start HTTP Server
	server := setupHTTPServer(createActivityUseCase)
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// 10. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	cancel() // Stop worker
	server.Shutdown(ctx)
}
```

---

## Testing with In-Memory Broker

### Integration Test (`tests/integration/broker/event_flow_test.go`)

```go
package broker_test

import (
	"context"
	"testing"
	"time"

	"github.com/valentinesamuel/activelog/internal/application/activity/handlers"
	"github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	"github.com/valentinesamuel/activelog/internal/domain/activity"
	"github.com/valentinesamuel/activelog/internal/infrastructure/broker/memory"
	"github.com/valentinesamuel/activelog/tests/testhelpers/builders"
)

func TestEventFlow_ActivityCreated(t *testing.T) {
	// 1. Setup in-memory broker
	broker := memory.NewInMemoryBroker()
	defer broker.Close()

	// 2. Setup handler
	handler := handlers.NewActivityCreatedHandler()
	broker.Subscribe(context.Background(), []string{"activity.created"}, handler)

	// 3. Setup use case with broker
	activityRepo := setupMockRepository()
	useCase := usecases.NewCreateActivityUseCase(activityRepo, broker)

	// 4. Execute use case
	req := dto.CreateActivityRequest{
		UserID:          1,
		Type:            "running",
		Title:           "Test Run",
		DurationMinutes: 30,
		ActivityDate:    time.Now(),
	}

	resp, err := useCase.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("UseCase failed: %v", err)
	}

	// 5. Verify event was published
	events := broker.GetPublishedEvents()
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	// 6. Verify event type
	event, ok := events[0].(*activity.ActivityCreatedEvent)
	if !ok {
		t.Fatalf("Expected ActivityCreatedEvent, got %T", events[0])
	}

	// 7. Verify event data
	if event.ActivityID != resp.ID {
		t.Errorf("Expected activity ID %d, got %d", resp.ID, event.ActivityID)
	}
	if event.Title != "Test Run" {
		t.Errorf("Expected title 'Test Run', got '%s'", event.Title)
	}
}
```

---

## Common Broker Patterns

### 1. Transactional Outbox Pattern

Ensures events are published even if broker is down:

```go
// Store events in database table first
type OutboxEvent struct {
	ID        int64
	EventName string
	EventData []byte
	Published bool
	CreatedAt time.Time
}

// Background worker publishes from outbox
func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.publishPendingEvents(ctx)
		}
	}
}
```

### 2. Event Versioning

Handle schema changes:

```go
type ActivityCreatedEventV2 struct {
	BaseEvent
	// New fields
	Tags []string
}

func (h *Handler) Handle(ctx context.Context, event shared.Event) error {
	switch event.EventVersion() {
	case 1:
		return h.handleV1(event.(*ActivityCreatedEvent))
	case 2:
		return h.handleV2(event.(*ActivityCreatedEventV2))
	default:
		return fmt.Errorf("unsupported version: %d", event.EventVersion())
	}
}
```

### 3. Dead Letter Queue

Handle failed events:

```go
type DeadLetterHandler struct {
	maxRetries int
}

func (h *DeadLetterHandler) Handle(ctx context.Context, event shared.Event) error {
	retryCount := getRetryCount(event)

	if retryCount >= h.maxRetries {
		// Move to dead letter queue
		return h.sendToDeadLetter(event)
	}

	// Retry
	incrementRetryCount(event)
	return h.republish(event)
}
```

---

## Broker Comparison

### RabbitMQ
**Best for:** Traditional message queuing, reliable delivery
- ✅ Mature, battle-tested
- ✅ Good routing capabilities
- ✅ Management UI
- ❌ Not as high throughput as Kafka

### Redis (Pub/Sub or Streams)
**Best for:** Simple pub/sub, caching + messaging
- ✅ Very fast
- ✅ Easy to set up
- ✅ Multi-purpose (cache + queue)
- ❌ Less reliable (no persistence by default)

### Kafka
**Best for:** High throughput, event sourcing, log aggregation
- ✅ Extremely high throughput
- ✅ Replay events
- ✅ Distributed, scalable
- ❌ More complex setup

### NATS
**Best for:** Microservices, low latency
- ✅ Very lightweight
- ✅ Low latency
- ✅ Simple
- ❌ Less features than RabbitMQ

---

## Summary

### Broker in Clean Architecture

```
Domain Layer:
  ├── Defines events (ActivityCreatedEvent)
  └── No broker dependencies

Application Layer:
  ├── Uses EventPublisher interface (port)
  ├── Defines event handlers
  └── Publishes events in use cases

Infrastructure Layer:
  ├── Implements EventPublisher (RabbitMQ, Redis, etc.)
  └── Connects to actual message brokers

Interfaces Layer:
  └── Worker that consumes events and routes to handlers
```

### Key Benefits

- ✅ **Async processing** - Don't block HTTP requests
- ✅ **Decoupling** - Services communicate via events
- ✅ **Scalability** - Scale consumers independently
- ✅ **Reliability** - Guaranteed delivery with queues
- ✅ **Testability** - Use in-memory broker for tests

---

## Next Steps

1. Choose your broker (RabbitMQ recommended for start)
2. Implement domain events
3. Add publisher to use cases
4. Create event handlers
5. Setup worker to consume events
6. Test with in-memory broker first
7. Deploy with real broker

**The broker pattern enables event-driven architecture, just like in your kuja_user_ms project!** 🚀
