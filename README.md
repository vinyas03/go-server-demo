# Go API Service (Clean Architecture with net/http & CubeAPM Tracing)

A production-grade, idiomatic RESTful API service built in Go using standard library `net/http`, clean layered architecture, and **CubeAPM** distributed tracing.

---

## 🏛️ Architecture Overview

The project adheres to clean architecture principles with full distributed observability:

```
go-learn/
├── cmd/
│   └── api/
│       └── main.go              # App bootstrap, CubeAPM tracer init & HTTP server
├── internal/
│   ├── handler/                 # Controller / Transport Layer (HTTP)
│   │   ├── task_handler.go      # Request decoding, status mapping & response formatting
│   │   ├── task_handler_test.go # Controller tests with CubeAPM instrumentation
│   │   └── response.go          # Standardized JSON response & error utilities
│   ├── service/                 # Business Logic Layer (Domain Services)
│   │   ├── task_service.go      # Business rules, context checks, child span tracing
│   │   └── task_service_test.go # Unit tests with child span support
│   ├── repository/              # Persistence / Data Access Layer
│   │   ├── repository.go        # Storage interface (swappable with SQL, Mongo, etc.)
│   │   └── memory.go            # Thread-safe in-memory store (sync.RWMutex)
│   ├── model/                   # Domain Models & Context Abstraction
│   │   ├── context.go           # RequestContext, header accessors, context keys
│   │   ├── task.go              # Task struct, DTOs (CreateTaskInput, UpdateTaskInput)
│   │   └── errors.go            # Domain error sentinels (ErrNotFound, ErrForbidden, etc.)
│   └── middleware/              # Cross-cutting HTTP middlewares
│       └── middleware.go        # CubeAPM transaction labels, logging, recovery, CORS
├── go.mod
└── README.md
```

---

## 🔍 CubeAPM Distributed Tracing Integration

**CubeAPM** natively ingests distributed traces via the standard OpenTelemetry/Elastic APM protocol. Every incoming HTTP request and internal service operation is automatically traced:

1. **HTTP Transactions (`apmhttp.Wrap`)**:
   In [`cmd/api/main.go`](cmd/api/main.go), the root handler is wrapped with `apmhttp.Wrap(wrappedHandler)`. This records every API endpoint as a transaction with HTTP method, path, response status code, and latency.
2. **Context & Metadata Enrichment**:
   [`internal/middleware/middleware.go`](internal/middleware/middleware.go) enriches active CubeAPM transactions with user identifiers (`SetUsername`), `tenant_id`, and `request_id` labels.
3. **Business Logic Child Spans**:
   [`internal/service/task_service.go`](internal/service/task_service.go) creates child spans (`apm.StartSpan(ctx, "TaskService.CreateTask", "app.service")`) to trace execution time inside domain logic and automatically sends errors using `apm.CaptureError`.
4. **Panic Tracking**:
   The recovery middleware captures unhandled runtime panics and forwards them to CubeAPM before returning a 500 error.

### CubeAPM Configuration Variables

Configure these environment variables to connect to your CubeAPM server:

| Variable | Description | Default |
|---|---|---|
| `ELASTIC_APM_SERVICE_NAME` | The service name displayed in CubeAPM | `go-learn-api` |
| `ELASTIC_APM_SERVER_URL` | CubeAPM server ingestion URL | `http://localhost:3130` |
| `ELASTIC_APM_ENVIRONMENT` | Environment name (e.g. `production`, `staging`, `dev`) | `development` |
| `ELASTIC_APM_LOG_LEVEL` | Agent log level (`error`, `warning`, `info`, `debug`) | `info` |

---

## 🚀 Getting Started

### 1. Run the API Server with CubeAPM

```bash
# Optional: Set CubeAPM server URL (defaults to http://localhost:3130)
export ELASTIC_APM_SERVER_URL="http://localhost:3130"
export ELASTIC_APM_SERVICE_NAME="go-learn-api"

go run ./cmd/api
```

### 2. Run All Tests

```bash
go test -v ./internal/...
```

---

## 📡 API Endpoints & Traced Headers

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/tasks` | Create a new task (attaches user & tenant from headers) |
| `GET` | `/api/v1/tasks` | List all tasks |
| `GET` | `/api/v1/tasks/{id}` | Get a task by ID |
| `PUT` | `/api/v1/tasks/{id}` | Update a task by ID |
| `DELETE` | `/api/v1/tasks/{id}` | Delete a task by ID |
| `GET` | `/health` | Health check endpoint |

#### Example: Create Task with Tracing & Identity Headers
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "X-User-ID: usr_456" \
  -H "X-Tenant-ID: acme" \
  -d '{"title": "Deploy to Production", "description": "Verify traces in CubeAPM"}'
```
