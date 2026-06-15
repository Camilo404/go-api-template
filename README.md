# Go API Template

Production-ready Go service template. Copy the folder, change the
module name, and start adding resources. Stdlib-only — no
frameworks, no ORM, no runtime dependencies.

---

## Why this template

| Goal | How it's solved |
| --- | --- |
| Copy-paste into a new project | `cmd/`, `internal/`, `deployments/`, `Makefile` are the whole repo. No global state, no magic. |
| Swap the backing store | `store.TaskStorer` is an interface. Memory is the default; a Postgres/Mongo/Redis impl is one type away. |
| Replace the example resource | `Task` appears in 5 files. Rename to `User`, `Order`, `Product`, etc. |
| Production-grade HTTP | Graceful shutdown, request ID, structured logs, panic recovery, CORS, body size limit, paginated lists. |
| Testable | Handlers, services, and stores are independently tested. The store is a black box to handlers thanks to interfaces. |

---

## Architecture

```
cmd/api/main.go            # Composition root: wire deps, run server
internal/
  config/                  # Env-driven config with validation
  server/                  # http.Server + SIGINT/SIGTERM graceful shutdown
  router/                  # URL → handler dispatch + middleware chain
  middleware/              # request_id, logging, recovery, cors
  handlers/                # HTTP transport (thin)
  service/                 # Business logic (validation, orchestration)
  store/                   # Persistence interface + in-memory impl
  models/                  # Entities, DTOs, errors, pagination
  httpx/                   # JSON response + decode helpers
deployments/               # Dockerfile, docker-compose, .dockerignore
configs/                   # .env.example
Makefile
```

### Request lifecycle

```
Request
  → middleware.RequestID      assign/propagate X-Request-ID
  → middleware.Recovery       catch panics
  → middleware.Logging        emit access log
  → middleware.CORS           set CORS headers, short-circuit OPTIONS
  → router (http.ServeMux)
  → handler                   parse, call service, render
  → service                   validate, call store
  → store                     read/write persistence
Response
```

### Layer rules

- `handlers` know about HTTP, JSON, and `service`. They never touch a `store` directly.
- `service` knows about business rules and the store **interface**. It does not import `net/http`.
- `store` knows about persistence. It returns `*models.APIError` (e.g. `ErrNotFound`) for expected failures.
- `models` is shared by all layers. It owns validation and the APIError → HTTP status mapping.
- `httpx` is the only place that writes JSON responses or decodes request bodies.

---

## Quick start

```bash
# 1. Build and run
make run
# or
go run ./cmd/api

# 2. Hit the example endpoint
curl -s localhost:8080/health
curl -s -X POST localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"first task"}'

# 3. Run tests
make test
```

The API listens on `:8080` by default. Override with `PORT=9090`.

---

## Configuration

All settings come from environment variables. See [configs/.env.example](configs/.env.example) for the full list.

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | TCP port the server binds to |
| `APP_ENV` | `development` | Free-form, surfaced in logs |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `CORS_ORIGINS` | `*` | Comma-separated list. Use explicit origins in production. |
| `READ_TIMEOUT` | `15s` | `http.Server.ReadTimeout` |
| `WRITE_TIMEOUT` | `15s` | `http.Server.WriteTimeout` |
| `IDLE_TIMEOUT` | `60s` | `http.Server.IdleTimeout` |
| `SHUTDOWN_WAIT` | `15s` | Max time to drain in-flight requests on SIGINT/SIGTERM |
| `MAX_BODY_BYTES` | `1048576` | 1 MiB cap on request bodies |

---

## Using this as a template for a new service

### 1. Copy the folder

```bash
cp -R template-api orders-api
cd orders-api
```

### 2. Rename the module

Edit `go.mod` and every import line:

```
module github.com/yourorg/orders-api
```

```bash
# from the repo root
grep -rl 'github.com/Camilo404/go-api-template' . | xargs sed -i '' \
  -e 's|github.com/Camilo404/go-api-template|github.com/yourorg/orders-api|g'
```

### 3. Replace the `Task` example with your resource

Search for `Task` (case-sensitive) and rename it. The type appears in
exactly these files:

| File | What to rename |
| --- | --- |
| `internal/models/task.go` | `Task` → `Order`, `CreateTaskInput` → `CreateOrderInput`, `UpdateTaskInput` → `UpdateOrderInput` |
| `internal/models/errors.go` | `ErrTitleRequired` etc. → `ErrNameRequired`, `ErrPriceRequired`, etc. Update `HTTPStatus` accordingly. |
| `internal/store/store.go` | `TaskStorer` → `OrderStorer`, method types |
| `internal/store/memory.go` | `MemoryTaskStore` → `MemoryOrderStore` |
| `internal/service/tasks.go` | `TaskService` → `OrderService` |
| `internal/handlers/tasks.go` | `TaskHandler` → `OrderHandler` |
| `cmd/api/main.go` | `taskStore`, `taskSvc` → `orderStore`, `orderSvc` |
| `internal/router/router.go` | `Tasks` field on `Deps`, route registration |

### 4. Plug in a real store

Implement the interface and swap the wiring in `cmd/api/main.go`:

```go
// internal/store/postgres.go
type PostgresOrderStore struct{ db *sql.DB }
func (s *PostgresOrderStore) List(ctx context.Context, page, perPage int) ([]models.Order, int, error) { ... }
// ... implement the rest of OrderStorer
```

```go
// cmd/api/main.go
orderStore, err := store.NewPostgresOrderStore(os.Getenv("DATABASE_URL"))
if err != nil { /* fail fast */ }
orderSvc := service.NewOrderService(orderStore)
```

### 5. Add auth, metrics, tracing

These plug in cleanly as middleware in `internal/router/router.go`:

```go
return middleware.Chain(mux,
    middleware.RequestID,
    middleware.Recovery(logger),
    middleware.Logging(logger),
    middleware.CORS(cfg.CORSOrigins),
    auth.JWT(cfg.JWTSecret),   // <-- new
    metrics.Prometheus(),      // <-- new
    tracing.OpenTelemetry(),   // <-- new
)
```

---

## HTTP conventions

- All endpoints are versioned under `/api/v1/`.
- Errors are always JSON: `{"code":"...","message":"..."}`.
- IDs in URLs are strings (e.g. `/tasks/42`). Service layer is responsible for parsing.
- List endpoints accept `?page=N&per_page=M` and return `{ "data": [...], "page": { "page": N, "per_page": M, "total": T } }`.
- Bodies are capped at `MAX_BODY_BYTES` (default 1 MiB). Unknown JSON fields are rejected.

---

## Testing

```bash
make test          # race-enabled, runs every package
make cover         # writes coverage.out
go test -run TestTaskService ./internal/service
```

Tests live next to the code they cover (`*_test.go`). The example
suite covers:

- `internal/store/memory_test.go` — CRUD, pagination, not-found
- `internal/service/tasks_test.go` — validation, full flow
- `internal/handlers/tasks_test.go` — happy path, validation, method-not-allowed
- `internal/httpx/decode_test.go` — happy path, unknown fields, oversize body

---

## Docker

```bash
make docker-build         # tag: api:latest
make docker-run           # run with the example env
```

The image is multi-stage, ~20 MB, and runs as a non-root user.

---

## Project layout cheatsheet

| Path | Add a new resource here when… |
| --- | --- |
| `internal/models/<thing>.go` | You are defining a new entity or DTO. |
| `internal/store/store.go` | You are adding a new method to the storer contract. |
| `internal/store/<backend>.go` | You are implementing that contract for a new backend. |
| `internal/service/<thing>.go` | You are adding business rules that don't belong in the handler. |
| `internal/handlers/<thing>.go` | You are exposing a new resource over HTTP. |
| `internal/router/router.go` | You are registering a new route or middleware. |
| `cmd/api/main.go` | You are wiring a new dependency or starting a new binary. |

---

## License

Choose whatever fits your organisation; this template ships without one.
