# Implementation Plan: Month 6 + Month 7 Features

## Context

The todo.md lists 4 work areas to complete:
1. **Queue System** – fully functional inbox/outbox queues using asynq with a consumer factory pattern
2. **Email Adapter** – pluggable email provider (SMTP/SendGrid/etc) mirroring the existing storage adapter pattern
3. **Cron** – scheduled tasks per Week 23 of month6.md (daily stats, weekly emails, monthly reports, cleanup)
4. **Month 7 extras** – webhook support, CSV export, PDF export, concurrency patterns

Current state of queue code is broken: `internal/queue/di/register.go:18` returns `types.QueueProviderKey` which does not exist in the types package (compile error introduced after the rename commit). The `asynq.Provider.New()` also closes its client immediately, which is a bug.

Libraries already in go.mod: `github.com/hibiken/asynq v0.26.0`, `github.com/robfig/cron/v3 v3.0.1`, `github.com/redis/go-redis/v9`.

---

## Phase 1: Queue System (Complete & Redesign)✅

### Goal
Enable `queueAdapter.Enqueue(ctx, queue, payload)` from anywhere. Multiple provider backends selectable via `QUEUE_PROVIDER` env var. Two providers: `asynq` (Redis-backed, distributed) and `memory` (in-process, for dev/tests).

### Interface (`internal/queue/types/types.go`) [MODIFY]
- Remove `EmailTask` struct
- Add `QueueName` type + constants: `InboxQueue = "inbox"`, `OutboxQueue = "outbox"`
- Add `EventType` type + constants for inbox: `EventWelcomeEmail`, `EventWeeklySummary`, `EventGenerateExport`, `EventSendVerificationEmail`; for outbox: `EventActivityCreated`, `EventActivityDeleted`
- Add `JobPayload { Event EventType; Data json.RawMessage }`
- Redesign `QueueProvider` interface:
  ```go
  type QueueProvider interface {
      Enqueue(ctx context.Context, queue QueueName, payload JobPayload) (taskID string, err error)
  }
  ```

### Provider: Asynq (`internal/queue/asynq/provider.go`) [MODIFY]
- Fix `New()`: do not close client immediately (current bug)
- Implement `Enqueue()`: marshal payload to JSON, `client.EnqueueContext(ctx, asynq.NewTask(string(event), data), asynq.Queue(string(queue)), asynq.MaxRetry(3))`
- Add `NewWorkerServer(redisAddr string, concurrency int) *asynq.Server`

### Provider: In-Memory (`internal/queue/memory/provider.go`) [CREATE]
- Buffered channel per queue: `jobs map[QueueName]chan JobPayload`
- `New(bufferSize int) *Provider`
- `Enqueue()`: non-blocking send to channel; returns error if channel full
- `StartWorking(ctx, queue QueueName, handler func(ctx, JobPayload) error)`: background goroutine draining the channel; stops on ctx cancel
- Suitable for tests and local dev (no Redis required)

### DI (`internal/queue/di/register.go`) [MODIFY]
- Fix broken return type `types.QueueProviderKey` → `types.QueueProvider`
- Factory: `QUEUE_PROVIDER=asynq` → asynq provider; `QUEUE_PROVIDER=memory` (default) → memory provider

### Job Layer
**`internal/jobs/types.go`** [CREATE]
- Payload structs: `WelcomeEmailPayload{UserID, Email, Name}`, `WeeklySummaryPayload{UserID}`, `ExportPayload{UserID, Format}`

**`internal/jobs/handlers.go`** [CREATE]
- `HandleWelcomeEmail(ctx, payload) error` – call email provider
- `HandleWeeklySummary(ctx, payload) error` – fetch stats, send email
- `HandleGenerateExport(ctx, payload) error` – generate CSV/PDF, upload S3, update export record

**`internal/jobs/factory.go`** [CREATE]
- `HandlerFactory { handlers map[EventType]func(ctx, JobPayload) error }`
- `Register(event, handlerFn)` / `Dispatch(ctx, payload) error`
- Provider-agnostic: works with both asynq worker mux and memory provider's `StartWorking`

**`cmd/worker/main.go`** [CREATE]
- Separate worker binary; initializes DI container, builds factory, starts asynq server (`critical:6, default:3, low:1`)
- When `QUEUE_PROVIDER=memory`, skip asynq server and call `memProvider.StartWorking` per queue
- Graceful shutdown on `SIGTERM`

---

## Phase 2: Email Provider (Mirror Storage Pattern)✅

### Goal
`emailProvider.Send(ctx, input)` works regardless of backend (SMTP, SendGrid, noop).

### Changes

**`internal/config/email.go`** [CREATE]
- `EmailConfigType { Provider, SMTPHost, SMTPPort, SMTPUser, SMTPPass, From string }`
- Load from env vars: `EMAIL_PROVIDER`, `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `EMAIL_FROM`

**`internal/config/schema.go`** [MODIFY]
- Add `Email *EmailConfigType` field; call `loadEmail()` in `MustLoad()`

**`internal/email/types/types.go`** [CREATE]
- `EmailProvider` interface: `Send(ctx, input SendEmailInput) error`
- `SendEmailInput { To, From, Subject, HTMLBody, TextBody string }`

**`internal/email/smtp/provider.go`** [CREATE]
- Uses `gopkg.in/gomail.v2` (add to go.mod)
- `New(config) (*Provider, error)`, `Send(ctx, input) error`

**`internal/email/noop/provider.go`** [CREATE]
- No-op implementation for testing/dev that logs instead of sending

**`internal/email/templates/`** [CREATE]
- `welcome.html`, `weekly_summary.html`, `verification.html` (HTML templates)
- `template_data.go` – structs for template data

**`internal/email/di/register.go` + `keys.go`** [CREATE]
- Factory following storage DI pattern; selects smtp vs noop based on `config.Email.Provider`
- `EmailProviderKey = "EmailProvider"`

**`cmd/api/container.go`** [MODIFY]
- Call `emailRegister.RegisterEmail(c)` and eagerly resolve it

---

## Phase 3: Cron / Scheduler (Week 23)✅

### Goal
Scheduled background tasks using `robfig/cron/v3` (already in go.mod).

### Changes

**`migrations/007_create_daily_stats.up.sql`** [CREATE]
- Table: `daily_stats(id, user_id, date, total_activities, total_distance_km, total_duration_minutes, created_at)`

**`internal/scheduler/scheduler.go`** [CREATE]
- `Scheduler { cron *cron.Cron }` wrapping `cron.New(cron.WithLocation(time.UTC))`
- `NewScheduler(statsCalc, cleanup, emailSvc, jobQueue) *Scheduler`
- `Start()` / `Stop()` methods
- Registered jobs:
  - `"0 0 * * *"` → `statsCalc.CalculateDailyStats(ctx)`
  - `"0 9 * * 0"` → enqueue weekly summary jobs for all active users via `jobQueue.Enqueue`
  - `"0 0 1 * *"` → enqueue monthly report generation per user
  - `"0 2 * * *"` → `cleanup.DeleteOldData(ctx)`

**`internal/services/stats_calculator.go`** [CREATE]
- `CalculateDailyStats(ctx) error` – aggregates previous day's activities per user, writes to `daily_stats`

**`internal/services/cleanup_service.go`** [CREATE]
- `DeleteOldData(ctx) error` – hard-deletes soft-deleted records older than 30 days

**`cmd/api/main.go`** [MODIFY]
- Resolve scheduler dependencies from container, call `scheduler.Start()` in `run()`, call `scheduler.Stop()` in `gracefulShutdown()`

---

## Phase 4: Export Features (Week 24)✅

### Goal
CSV streaming download + async PDF export backed by S3.

### Changes

**`migrations/008_create_exports.up.sql`** [CREATE]
- Table: `exports(id uuid, user_id, format, status, s3_key, error_message, created_at, completed_at)`

**`internal/export/types.go`** [CREATE]
- `ExportFormat` enum: `CSV`, `PDF`
- `ExportStatus` enum: `Pending`, `Processing`, `Completed`, `Failed`

**`internal/export/csv_exporter.go`** [CREATE]
- `ExportActivitiesCSV(ctx, userID int, w io.Writer) error`
- Streams rows via `encoding/csv` – no full dataset in memory

**`internal/export/pdf_exporter.go`** [CREATE]
- `GenerateActivityReport(ctx, userID int) ([]byte, error)`
- Uses `github.com/jung-kurt/gofpdf` (add to go.mod if not present)
- Sections: title, user info, summary stats table, activity list

**`internal/repository/export_repository.go`** [CREATE]
- `Create`, `UpdateStatus`, `GetByID`, `ListByUser` for export records

**`internal/handlers/export_handler.go`** [CREATE]
- `GET /api/v1/activities/export/csv` – direct streaming CSV
- `POST /api/v1/activities/export/pdf` – enqueue PDF job, return `job_id`
- `GET /api/v1/jobs/{jobId}/status` – poll job status
- `GET /api/v1/jobs/{jobId}/download` – redirect to S3 presigned URL

**`cmd/api/main.go`** [MODIFY]
- Register export routes in `registerActivityRoutes`

---

## Phase 5: Concurrency Patterns (Month 7)✅

### Goal
Apply goroutines/channels/sync patterns to existing services; add batch endpoints; build reusable worker pool.

### Changes

**`pkg/workers/pool.go`** [CREATE]
- Generic `WorkerPool[J, R any]` using buffered job/result channels + `sync.WaitGroup`
- `New(numWorkers int) *WorkerPool[J, R]`
- `Submit(jobs []J, fn func(J) R) []R`

**`internal/service/stats_service.go`** [MODIFY]
- Add `CalculateUserStatsConcurrent(ctx, userID int) (*Stats, error)`
- Launch 4 goroutines (total activities, distance, duration, streak), collect via buffered channel

**`internal/handlers/activity.go`** [MODIFY]
- Add `BatchCreateActivities` handler: `POST /api/v1/activities/batch`
- Add `BatchDeleteActivities` handler: `DELETE /api/v1/activities/batch`
- Both use `pkg/workers.WorkerPool` for parallel processing

**`internal/handlers/photo_handler.go`** [MODIFY]
- Refactor multi-photo upload to use semaphore pattern (`chan struct{}` with cap=5)

**`pkg/cache/memory_cache.go`** [CREATE]
- `MemoryCache` with `sync.RWMutex`, TTL support
- `Get(key)`, `Set(key, value, ttl)`, `Delete(key)`, `Flush()`

---

## Phase 6: WebSocket Hub (Real-time Notifications)✅

### Goal
Real-time push notifications to connected clients (friend requests, likes, comments, activity events).

### Changes

**`internal/websocket/hub.go`** [CREATE]
- `Hub { clients map[int]*Client; broadcast chan Message; register chan *Client; unregister chan *Client; mu sync.RWMutex }`
- `NewHub() *Hub`
- `Run()` – event loop: registers/unregisters clients, broadcasts messages; uses `select` over the 3 channels
- `SendToUser(userID int, msgType string, payload any)` – targeted delivery; acquires RLock, writes to client's send channel

**`internal/websocket/client.go`** [CREATE]
- `Client { hub *Hub; conn *websocket.Conn; userID int; send chan []byte }`
- `readPump()` – reads from WebSocket connection; defers unregister + close on exit
- `writePump()` – reads from `send` channel and writes to WebSocket; sends ping every 54s via `time.NewTicker`; handles close gracefully

**`internal/websocket/types.go`** [CREATE]
- `Message { Type, UserID string; Payload any; Timestamp time.Time }`
- Message type constants: `MsgFriendRequest`, `MsgFriendAccepted`, `MsgActivityLiked`, `MsgActivityComment`

**`internal/websocket/handler.go`** [CREATE]
- `ServeWS(w, r)` – upgrades HTTP to WebSocket using `gorilla/websocket` upgrader (need `go get github.com/gorilla/websocket`)
- Extracts `userID` from JWT context, creates `Client`, registers with hub, launches `readPump`/`writePump` goroutines

**`cmd/api/main.go`** [MODIFY]
- Initialize `hub := websocket.NewHub()`; `go hub.Run()` in `run()`
- Register `GET /ws` route → `hub.ServeWS`
- Register hub in DI container so services can call `hub.SendToUser`

**`cmd/api/container.go`** [MODIFY]
- Register `WebSocketHubKey` singleton in container

### Note
The hub is intentionally single-instance (in-process). For multi-instance deployments, replace the in-process hub with Redis pub/sub to fan out to all instances — but this is a later concern.

---

## Phase 7: Feature Flags✅

### Goal
Env-var-driven feature toggles with middleware, enabling gradual rollouts and kill switches without deploys.

### Changes

**`internal/featureflags/flags.go`** [CREATE]
- `FeatureFlags { EnableComments, EnableLikes, EnableFriends, EnableWebhooks, EnableFeed bool }`
- `Load() *FeatureFlags` – reads from env vars (`FEATURE_COMMENTS=enabled`, etc.)
- `IsEnabled(userID int, feature string) bool` – checks flag; supports future per-user percentage rollouts via consistent hashing

**`internal/featureflags/middleware.go`** [CREATE]
- `Middleware struct { flags *FeatureFlags }`
- `Check(feature string) func(http.Handler) http.Handler` – returns 403 with `{"error": "feature_not_available"}` if flag is off

**`internal/handlers/features_handler.go`** [CREATE]
- `GET /api/v1/features` – returns current flag state for the requesting user

**`cmd/api/main.go`** [MODIFY]
- Load feature flags on startup; wire `featureflags.Middleware` into routes that require them (friends, comments, likes endpoints)

---

## Phase 8: Webhook System (Provider Pattern)

### Goal
Outgoing webhooks with a pluggable event-bus backend. The event-bus (how webhook events are fanned out internally) is provider-switchable via `WEBHOOK_PROVIDER` env var. Two providers: `memory` (in-process channels) and `redis` (Redis pub/sub for multi-instance).

### Interface (`internal/webhook/types/types.go`) [CREATE]
```go
// WebhookEvent is the internal event published when something happens
type WebhookEvent struct {
    EventType string
    UserID    int
    Payload   json.RawMessage
    Timestamp time.Time
}

// WebhookBusProvider is the event-bus that fans events to subscribers
type WebhookBusProvider interface {
    Publish(ctx context.Context, event WebhookEvent) error
    Subscribe(ctx context.Context, handler func(ctx context.Context, event WebhookEvent)) error
}
```

### Provider: In-Process (`internal/webhook/memory/provider.go`) [CREATE]
- `Provider { ch chan WebhookEvent; mu sync.RWMutex; handlers []func(ctx, WebhookEvent) }`
- `Publish()`: sends to buffered channel (non-blocking)
- `Subscribe()`: starts a goroutine reading from the channel, calls registered handlers
- Suitable for single-instance deploys and tests

### Provider: Redis Pub/Sub (`internal/webhook/redis/provider.go`) [CREATE]
- Uses `redis/go-redis/v9` (already in go.mod)
- `Publish()`: `rdb.Publish(ctx, "webhook:events", jsonMarshal(event))`
- `Subscribe()`: `rdb.Subscribe(ctx, "webhook:events")` in a goroutine; JSON-unmarshal each message and call handlers
- Suitable for multi-instance deploys (all instances share the same Redis channel)

### DI (`internal/webhook/di/register.go` + `keys.go`) [CREATE]
- `WEBHOOK_PROVIDER=memory` → memory provider; `WEBHOOK_PROVIDER=redis` → redis provider
- `WebhookBusKey = "WebhookBus"`

### Webhook Delivery (`internal/webhook/delivery.go`) [CREATE]
- Called by the subscriber goroutine after receiving an event
- Looks up registered webhooks for the event type from `webhook_repository`
- For each matching webhook: HTTP POST with HMAC-SHA256 `X-Webhook-Signature` header (using webhook's secret)
- Retries up to 3 times with exponential backoff on non-2xx responses

### Database
**`migrations/009_create_webhooks.up.sql`** [CREATE]
- `webhooks(id uuid, user_id, url, events text[], secret, active, created_at)`

**`internal/repository/webhook_repository.go`** [CREATE]
- `Create`, `Delete`, `ListByUserID`, `ListByEvent(eventType) []*Webhook`

### API
**`internal/handlers/webhook_handler.go`** [CREATE]
- `POST /api/v1/webhooks` – register webhook URL
- `GET /api/v1/webhooks` – list user's webhooks
- `DELETE /api/v1/webhooks/{id}` – delete webhook

### Integration
**`internal/application/activity/usecases/create_activity.go`** [MODIFY]
- After activity created: `webhookBus.Publish(ctx, WebhookEvent{EventType: "activity.created", ...})`

**`cmd/api/main.go`** [MODIFY]
- Initialize webhook bus from DI container
- Call `webhookBus.Subscribe(ctx, delivery.Handle)` to start the delivery goroutine

---

## Critical Files to Modify

| File                                | Change                                  |
| ----------------------------------- | --------------------------------------- |
| `internal/queue/types/types.go`     | Full redesign                           |
| `internal/queue/asynq/provider.go`  | Fix bug + implement                     |
| `internal/queue/di/register.go`     | Fix broken type reference               |
| `internal/config/schema.go`         | Add email config                        |
| `cmd/api/main.go`                   | Add scheduler start/stop, export routes |
| `cmd/api/container.go`              | Register email + scheduler              |
| `internal/service/stats_service.go` | Add concurrent stats                    |
| `internal/handlers/activity.go`     | Add batch endpoints                     |

---

## Verification

```bash
# Build everything
go build ./...

# Run with race detector
go test -race ./...

# Test queue (requires Redis)
docker-compose up -d redis
QUEUE_PROVIDER=asynq go run ./cmd/worker/main.go

# Test email (use noop provider for dev)
EMAIL_PROVIDER=noop go run ./cmd/api/main.go

# Test CSV export
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/activities/export/csv

# Test cron manually (expose a test endpoint or call directly in tests)
```

## Implementation Order

Phases should be implemented in order (1→8) as each builds on the previous:
- Phase 1 (Queue) is a prerequisite for email jobs (Phase 2) and export jobs (Phase 4)
- Phase 2 (Email) is a prerequisite for the cron weekly email job (Phase 3)
- Phase 3 (Cron) depends on the export job handler (Phase 4) for monthly reports
- Phase 5 (Concurrency) and Phase 6 (WebSocket) can be done in parallel after Phase 4
- Phase 7 (Feature Flags) is independent – can be done any time after Phase 4
- Phase 8 (Webhooks) builds on Phase 1 (queue for async delivery) and Phase 6 (WS hub for notifications)
