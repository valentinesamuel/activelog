# Webhook Delivery: Async + Persistence + Retry (Redis Streams + NATS)

## Context

The current webhook system uses Redis pub/sub (no durability — messages are lost if no subscriber is active at publish time) and delivers webhooks **synchronously** inside the subscriber goroutine, blocking event processing while HTTP calls are in flight. There is no audit trail; if the app crashes mid-delivery, all in-flight work is lost. Retry logic only exists in memory (3 short attempts: 1s/2s/4s).

**Goals:**
1. Replace Redis pub/sub with **two durable providers**: Redis Streams (XADD/XREADGROUP/XACK) and **NATS JetStream** (pull consumer with Ack/Nak).
2. Make HTTP delivery **asynchronous** — subscriber returns immediately after creating DB records.
3. Persist every delivery attempt in a `webhook_deliveries` table (audit trail + crash recovery).
4. Retry with **exponential backoff** (1m → 5m → 30m → 2h → 24h, max 5 attempts) via a polling retry worker.

---

## New Architecture

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                         WEBHOOK DELIVERY SYSTEM                            ║
╚══════════════════════════════════════════════════════════════════════════════╝

  Activity Handler
       │
       │  broker.Publish(event)
       ▼
  ┌─────────────────────────────────────────────────────────────────────────┐
  │                       WebhookBusProvider interface                      │
  │                  Publish(ctx, event) / Subscribe(ctx, handler)          │
  └──────────┬──────────────────────────────────┬───────────────────────────┘
             │                                  │
    WEBHOOK_PROVIDER=redis             WEBHOOK_PROVIDER=nats
             │                                  │
             ▼                                  ▼
  ┌──────────────────────┐          ┌───────────────────────────┐
  │   Redis Streams      │          │   NATS JetStream          │
  │                      │          │                           │
  │  XADD webhook:events │          │  js.Publish("webhook.     │
  │  XREADGROUP GROUP    │          │    events", data)         │
  │    webhook-delivery  │          │  sub.Fetch(10)            │
  │  XACK <msg-id>       │          │  msg.Ack()                │
  └──────────┬───────────┘          └───────────┬───────────────┘
             │                                  │
             └───────────────┬──────────────────┘
                             │
                             │  handler(ctx, event)
                             ▼
               ┌─────────────────────────────┐
               │         Delivery.Handle()   │
               │                             │
               │  1. ListByEvent(eventType)  │
               │  2. For each webhook:       │
               │     CreateDelivery(DB)  ────┼──► webhook_deliveries table
               │     go dispatchAsync()  ────┼──► goroutine (non-blocking)
               │  3. Return immediately      │
               │     (XACK / msg.Ack sent)   │
               └─────────────────────────────┘
                             │
                ┌────────────┴────────────┐
                │                         │
                ▼                         ▼
          HTTP 2xx                   HTTP error / timeout
                │                         │
     MarkDeliverySucceeded       MarkDeliveryFailed
      (status=succeeded)          (status=failed,
                                   next_retry_at=now+delay)
                                         │
                             ┌───────────┴───────────┐
                             │     RetryWorker        │
                             │                        │
                             │  Polls every 30s       │
                             │  ListPendingRetries()  │
                             │   WHERE next_retry_at  │
                             │     <= NOW()           │
                             │  go retryDelivery()    │
                             └───────────────────────┘

  BACKOFF SCHEDULE:
  ┌──────────┬────────────────────┬──────────────────────────────────┐
  │ Attempt  │  Delay             │  Status after failure            │
  ├──────────┼────────────────────┼──────────────────────────────────┤
  │ 1st      │ 1 minute           │ failed, next_retry_at = now+1m   │
  │ 2nd      │ 5 minutes          │ failed, next_retry_at = now+5m   │
  │ 3rd      │ 30 minutes         │ failed, next_retry_at = now+30m  │
  │ 4th      │ 2 hours            │ failed, next_retry_at = now+2h   │
  │ 5th      │ 24 hours           │ failed, next_retry_at = now+24h  │
  │ after 5  │ —                  │ exhausted (terminal, no retry)   │
  └──────────┴────────────────────┴──────────────────────────────────┘

  PROVIDERS (selected via WEBHOOK_PROVIDER env var):
  ┌──────────────┬─────────────────┬──────────────┬───────────────────────┐
  │ Provider     │ Durability      │ Multi-node   │ Dev dependency        │
  ├──────────────┼─────────────────┼──────────────┼───────────────────────┤
  │ memory       │ None (dev only) │ No           │ None                  │
  │ redis        │ Yes (streams)   │ Yes (groups) │ Redis already present │
  │ nats         │ Yes (JetStream) │ Yes (pull)   │ Add nats.go           │
  └──────────────┴─────────────────┴──────────────┴───────────────────────┘
```

---

## Key Design Decisions

**XACK / msg.Ack() timing:** Sent after `Handle()` returns (after DB delivery records are written, before HTTP completes). The DB is the source of truth for retry state. The stream stays clean; crashes are recovered by the retry worker polling `status IN ('pending','failed')`.

**Malformed messages:** ACK'd and dropped immediately (non-retryable at stream level).

**Retry worker vs. asynq:** The polling worker reads DB state directly, making crash recovery automatic — no separate queue message needed. 30-second poll interval gives adequate freshness given delays are in minutes/hours.

**Goroutine management:** One goroutine per webhook endpoint per event. Acceptable at current scale. Future: bounded semaphore `make(chan struct{}, N)`.

---

## New Dependency

```
github.com/nats-io/nats.go v1.38.0   (add to go.mod)
```

---

## Database Migration

### `migrations/010_create_webhook_deliveries.up.sql`
```sql
CREATE TYPE webhook_delivery_status AS ENUM ('pending', 'succeeded', 'failed', 'exhausted');

CREATE TABLE webhook_deliveries (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id       UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type       TEXT NOT NULL,
    payload          JSONB NOT NULL,
    status           webhook_delivery_status NOT NULL DEFAULT 'pending',
    attempt_count    INTEGER NOT NULL DEFAULT 0,
    max_attempts     INTEGER NOT NULL DEFAULT 5,
    last_http_status INTEGER,
    last_error       TEXT,
    next_retry_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_status_retry ON webhook_deliveries(status, next_retry_at)
    WHERE status IN ('pending', 'failed');
```

### `migrations/010_create_webhook_deliveries.down.sql`
```sql
DROP TABLE IF EXISTS webhook_deliveries;
DROP TYPE IF EXISTS webhook_delivery_status;
```

---

## Files to Create

### `internal/webhook/nats/provider.go`
NATS JetStream pull-consumer-based provider:
- `New(url string) (*Provider, error)` — `nats.Connect(url)`, get JetStream context, create/bind stream `WEBHOOK_EVENTS` on subject `webhook.events`, create durable pull consumer `webhook-delivery`
- `Publish(ctx, event)` — `js.PublishMsg` with JSON-serialized event
- `Subscribe(ctx, handler)` — starts `go readLoop()`
- `readLoop()` — `sub.Fetch(10, nats.MaxWait(5s))` in a loop; for each message: deserialize → call `handler(ctx, event)` → `msg.Ack()`; on deserialize error: `msg.Ack()` (drop malformed); on context cancel: return

### `internal/webhook/retry_worker.go` (`webhook` package)
- `RetryWorker{webhookRepo, delivery}` struct
- `NewRetryWorker(repo, delivery) *RetryWorker`
- `Start(ctx)` — `go run(ctx)`
- `run(ctx)` — `time.NewTicker(retryPollInterval)` loop, calls `poll(ctx)` on each tick
- `poll(ctx)` — `ListPendingRetries(100)`, for each: `GetByID(webhookID)` + deserialize payload + `go retryDelivery()`
- `retryDelivery(wh, delivery, event)` — `delivery.attemptHTTP()` + write result to DB (same logic as `dispatchAsync`)

### `migrations/010_create_webhook_deliveries.up.sql` — see above
### `migrations/010_create_webhook_deliveries.down.sql` — see above

---

## Files to Modify

### `internal/webhook/types/types.go`
Add `DeliveryStatus` enum constants and `WebhookDelivery` struct with all DB fields.

### `internal/webhook/redis/provider.go` — REWRITE (pub/sub → streams)
- Constants: `streamName="webhook:events"`, `groupName="webhook-delivery"`, `consumerName="activelog-worker-1"`, `blockDuration=5s`
- `ensureConsumerGroup(ctx)` — `XGroupCreateMkStream(..., "$")` idempotent
- `Publish(ctx, event)` — `XADD MaxLen=10000 Approx=true field="event" value=<json>`
- `Subscribe(ctx, handler)` — `ensureConsumerGroup` + `go readLoop()`
- `readLoop()` — `XREADGROUP Block=5s Count=10 Streams=[stream, ">"]`; on timeout: continue; on error: sleep 1s + continue
- `processMessage(ctx, msg, handler)` — deserialize → `handler()` → `XACK`; on bad message: `XACK` + drop

### `internal/webhook/delivery.go` — REWRITE (sync → async)
```go
var retryDelays = []time.Duration{1*time.Minute, 5*time.Minute, 30*time.Minute, 2*time.Hour, 24*time.Hour}
const maxAttempts = 5
```
- `Handle(ctx, event)` — creates DB records + `go dispatchAsync()` per webhook, returns immediately
- `dispatchAsync(wh, delivery, event)` — uses `context.Background()`, calls `attemptHTTP()`, writes outcome
- `attemptHTTP(ctx, url, eventType, sig, body) (int, error)` — single HTTP POST, 10s timeout, HMAC headers; returns `(statusCode, error)`
- `computeSignature(secret, body) string` — HMAC-SHA256, unchanged logic

### `internal/repository/webhook_repository.go`
Add 4 methods (existing 5 methods unchanged):
- `CreateDelivery(ctx, *WebhookDelivery) error` — INSERT RETURNING id/timestamps
- `MarkDeliverySucceeded(ctx, id, httpStatus) error`
- `MarkDeliveryFailed(ctx, id, *httpStatus, errMsg, *nextRetryAt) error` — status=failed or exhausted
- `ListPendingRetries(ctx, limit) ([]*WebhookDelivery, error)` — JOIN webhooks WHERE active AND status IN ('pending','failed') AND next_retry_at <= NOW() ORDER BY next_retry_at LIMIT $1

### `internal/config/webhook.go`
Add fields to `WebhookConfigType`:
```go
StreamMaxLen   int64  // WEBHOOK_STREAM_MAX_LEN, default 10000
RetryPollSecs  int    // WEBHOOK_RETRY_POLL_SECONDS, default 30
NATSUrl        string // NATS_URL, default "nats://localhost:4222"
```

### `internal/config/schema.go`
Add optional env var entries: `WEBHOOK_STREAM_MAX_LEN` (int), `WEBHOOK_RETRY_POLL_SECONDS` (int), `NATS_URL` (string). Add `"nats"` to `ValidValues` for `WEBHOOK_PROVIDER`.

### `internal/webhook/di/keys.go`
Add constants:
```go
WebhookDeliveryKey = "WebhookDelivery"
RetryWorkerKey     = "WebhookRetryWorker"
```

### `internal/webhook/di/register.go`
- `RegisterWebhookBus(c)` — factory updated with `case "nats": webhookNATS.New(config.Webhook.NATSUrl)`
- Add `RegisterWebhookDelivery(c)` — resolves `WebhookRepoKey`, constructs `webhook.NewDelivery(repo)`
- Add `RegisterRetryWorker(c)` — resolves `WebhookRepoKey` + `WebhookDeliveryKey`, constructs `webhook.NewRetryWorker(repo, delivery)`
- `createProvider()` switch: `"redis"` → Redis Streams, `"nats"` → NATS JetStream, default → memory

### `cmd/api/container.go`
After `RegisterWebhookBus(c)`:
```go
webhookDI.RegisterWebhookDelivery(c)
webhookDI.RegisterRetryWorker(c)
```

### `cmd/api/main.go`
- Add `WebhookRetryWorker *webhook.RetryWorker` field to `Application` struct
- Resolve from container in `setupDependencies()`
- After existing subscribe block in `serve()`:
  ```go
  app.WebhookRetryWorker.Start(webhookCtx)  // same ctx as bus subscription
  ```

---

## Complete File Summary

| File                                                | Action                                           |
| --------------------------------------------------- | ------------------------------------------------ |
| `migrations/010_create_webhook_deliveries.up.sql`   | CREATE                                           |
| `migrations/010_create_webhook_deliveries.down.sql` | CREATE                                           |
| `internal/webhook/nats/provider.go`                 | CREATE                                           |
| `internal/webhook/retry_worker.go`                  | CREATE                                           |
| `internal/webhook/types/types.go`                   | MODIFY — add `DeliveryStatus`, `WebhookDelivery` |
| `internal/webhook/redis/provider.go`                | REWRITE — pub/sub → streams                      |
| `internal/webhook/delivery.go`                      | REWRITE — sync → async + persistence             |
| `internal/repository/webhook_repository.go`         | MODIFY — 4 new delivery methods                  |
| `internal/config/webhook.go`                        | MODIFY — add 3 new config fields                 |
| `internal/config/schema.go`                         | MODIFY — add 3 new env var entries               |
| `internal/webhook/di/keys.go`                       | MODIFY — add 2 new keys                          |
| `internal/webhook/di/register.go`                   | MODIFY — add nats case + 2 new registrations     |
| `cmd/api/container.go`                              | MODIFY — register delivery + retry worker        |
| `cmd/api/main.go`                                   | MODIFY — add retry worker field + start          |

**Total:** 4 new files, 10 modified files. All changes contained within webhook subsystem and its DI wiring.

---

## Verification

1. **Build:** `go build ./...` — no compile errors after `go get github.com/nats-io/nats.go`
2. **Migration:** Run `010_create_webhook_deliveries.up.sql`; verify table + partial index exist
3. **Redis Streams path:**
   - `WEBHOOK_PROVIDER=redis` — register webhook, trigger activity event
   - `redis-cli XLEN webhook:events` — should show stream entries
   - `redis-cli XPENDING webhook:events webhook-delivery - + 10` — empty after ACK
   - Check `webhook_deliveries` table for new rows
4. **NATS path:**
   - `WEBHOOK_PROVIDER=nats` + `NATS_URL=nats://localhost:4222` (run `nats-server -js`)
   - Same trigger; verify delivery record + HTTP call
5. **Retry:** Point webhook at URL returning 500; check `status=failed` rows with advancing `next_retry_at`
6. **Exhaustion:** After 5 failures, verify `status=exhausted`, retry worker skips row
7. **Crash recovery:** Kill app after `CreateDelivery` but before XACK/Ack; restart; verify message reprocessed (delivery record may duplicate — acceptable for now)
