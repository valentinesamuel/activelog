# Plan: Reorganize `internal/` Folder Structure

## Context

The current `internal/` directory has 22 top-level items all sitting at the same level with no grouping. Domain use cases (activity, stats, tag), infrastructure adapters (email, queue, storage, webhook), cross-cutting concerns (config, jobs, scheduler), and core app layers (repository, service, handlers) are all peers. This makes the codebase hard to navigate mentally and doesn't scale well as more modules are added.

The goal is to group these items into a clear, layer-based hierarchy with no import cycles.

---

## Proposed Structure

```
internal/
├── models/              # Domain entities — no internal imports (leaf)
├── application/         # Use cases via broker pattern
│   ├── broker/          # (unchanged)
│   ├── activity/        # (unchanged)
│   ├── activityPhoto/   # (unchanged)
│   ├── stats/           # (unchanged)
│   └── tag/             # (unchanged)
├── service/             # Business logic services (unchanged)
├── repository/          # Data access layer (unchanged)
├── handlers/            # HTTP handlers (unchanged)
├── middleware/          # HTTP middleware (unchanged)
│
├── adapters/            # NEW grouping — third-party / protocol integrations
│   ├── cache/           # MOVE from internal/cache/
│   ├── email/           # MOVE from internal/email/
│   ├── queue/           # MOVE from internal/queue/
│   ├── storage/         # MOVE from internal/storage/
│   ├── webhook/         # MOVE from internal/webhook/
│   └── websocket/       # MOVE from internal/websocket/
│
└── platform/            # NEW grouping — cross-cutting infrastructure
    ├── config/          # MOVE from internal/config/
    ├── container/       # MOVE from internal/container/
    ├── featureflags/    # MOVE from internal/featureflags/
    ├── jobs/            # MOVE from internal/jobs/
    ├── scheduler/       # MOVE from internal/scheduler/
    ├── requestcontext/  # MOVE from internal/requestContext/ (also rename to lowercase)
    ├── validator/       # MOVE from internal/validator/
    └── utils/           # MOVE from internal/utils/
```

The result: **8 top-level entries** with a clear mental model instead of 22 flat peers.

---

## Mental Model After Change

| Layer                 | Lives At                   | Purpose                                            |
| --------------------- | -------------------------- | -------------------------------------------------- |
| Domain entities       | `models/`                  | Pure data structures, no internal deps             |
| Use cases             | `application/`             | Business operations via broker                     |
| Business logic        | `service/`                 | Orchestrates repositories, domain rules            |
| Data access           | `repository/`              | SQL, mocks, tx helpers                             |
| HTTP delivery         | `handlers/`, `middleware/` | Request/response, auth, rate limiting              |
| External integrations | `adapters/`                | Cache, email, queue, storage, webhooks, websockets |
| Infrastructure        | `platform/`                | Config, DI container, jobs, scheduler, utils       |

---

## What Moves Where

### `adapters/` (currently all at root `internal/`)
- `internal/cache/` → `internal/adapters/cache/`
- `internal/email/` → `internal/adapters/email/`
- `internal/queue/` → `internal/adapters/queue/`
- `internal/storage/` → `internal/adapters/storage/`
- `internal/webhook/` → `internal/adapters/webhook/`
- `internal/websocket/` → `internal/adapters/websocket/`

### `platform/` (currently all at root `internal/`)
- `internal/config/` → `internal/platform/config/`
- `internal/container/` → `internal/platform/container/`
- `internal/featureflags/` → `internal/platform/featureflags/`
- `internal/jobs/` → `internal/platform/jobs/`
- `internal/scheduler/` → `internal/platform/scheduler/`
- `internal/requestContext/` → `internal/platform/requestcontext/` (rename dir to lowercase)
- `internal/validator/` → `internal/platform/validator/`
- `internal/utils/` → `internal/platform/utils/`

### `export` split (special case — fixes a structural issue)
Currently `internal/export/` contains both:
- `types.go` — `ExportRecord`, `ExportFormat`, `ExportStatus` (domain entity types)
- `csv_exporter.go`, `pdf_exporter.go` — CSV/PDF generation logic

`repository/export_repository.go` imports `internal/export` for `ExportRecord`. If we moved the whole `export/` package under `application/`, the repository would depend on the application layer — an architectural inversion.

**Fix:**
1. Move `ExportRecord`, `ExportFormat`, `ExportStatus` into `internal/models/export.go` (consistent with how `Activity`, `User`, `Tag` types live in `models/`)
2. Merge `ExportActivitiesCSV` and `ExportActivitiesPDF` as methods into `internal/service/export_service.go` — they coexist as methods on an `ExportService` (or as plain functions) alongside the existing service files
3. Update `repository/export_repository.go` to import from `models` instead of `export`
4. Delete the now-empty `internal/export/` directory

### `practice/` — DELETE
This is learning scratch code. It should be removed from `internal/` entirely (or moved outside the repo to a personal notes dir).

---

## Import Graph After Change (no cycles)

```
models                       → (nothing)
adapters/*/types             → (nothing)
repository                   → models, adapters/webhook/types
service                      → models, repository
application/broker           → (nothing)
application/*/usecases       → models, repository, service, adapters/cache/types
service/export_service       → models (ExportActivitiesCSV/PDF methods)
handlers                     → application/*, models, repository, platform/requestcontext, platform/validator
middleware                   → platform/config, platform/requestcontext, adapters/cache/types
platform/jobs                → service, repository, adapters/email
platform/scheduler           → (nothing internal)
platform/container           → everything (DI root — expected)
```

---

## Pre-existing Violation to Fix in a Follow-up

`application/activity/usecases/list_activities.go:82` calls `middleware.CacheHitsTotal.Inc()` — a use case importing HTTP middleware to increment a Prometheus counter. This is a layer violation (use case → middleware). The counters should move to a shared `platform/metrics/` package. **Out of scope for this refactor but should be tracked.**

---

## Files to Change

Every Go file that imports a moved package needs its import path updated. Key files:

- All files importing `internal/cache/types` → `internal/adapters/cache/types`
- All files importing `internal/email/...` → `internal/adapters/email/...`
- All files importing `internal/queue/...` → `internal/adapters/queue/...`
- All files importing `internal/storage/...` → `internal/adapters/storage/...`
- All files importing `internal/webhook/...` → `internal/adapters/webhook/...`
- All files importing `internal/websocket` → `internal/adapters/websocket`
- All files importing `internal/config` → `internal/platform/config`
- All files importing `internal/container` → `internal/platform/container`
- All files importing `internal/featureflags` → `internal/platform/featureflags`
- All files importing `internal/jobs` → `internal/platform/jobs`
- All files importing `internal/scheduler` → `internal/platform/scheduler`
- All files importing `internal/requestContext` → `internal/platform/requestcontext`
- All files importing `internal/validator` → `internal/platform/validator`
- All files importing `internal/utils` → `internal/platform/utils`
- All files importing `internal/export` → `internal/models` (for types) or `internal/application/export` (for logic)

The `cmd/api/` entry point and `container.go` (DI root) will have the most import path changes.

---

## Verification

After the move:
```bash
go build ./...          # must have zero errors
go vet ./...            # must pass
go test ./...           # existing tests must still pass
```
