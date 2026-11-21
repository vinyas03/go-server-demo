# Go API Service (Clean Architecture with net/http)

A simple, idiomatic RESTful API service built in Go using standard library `net/http` and clean layered architecture.

---

## 🏛️ Architecture Overview

The project follows clean architecture principles, separating concerns into distinct layers:

```
go-learn/
├── cmd/
│   └── api/
│       └── main.go              # Application entrypoint & dependency wiring
├── internal/
│   ├── handler/                 # Controller / Transport Layer (HTTP)
│   │   ├── task_handler.go      # Request decoding, status mapping & response formatting
│   │   ├── task_handler_test.go # Controller tests via httptest
│   │   └── response.go          # Standardized JSON response & error utilities
│   ├── service/                 # Business Logic Layer (Domain Services)
│   │   ├── task_service.go      # Business rules, context checks, validation
│   │   └── task_service_test.go # Business logic unit tests (independent of HTTP)
│   ├── repository/              # Persistence / Data Access Layer
│   │   ├── repository.go        # Storage interface (swappable with SQL, Mongo, etc.)
│   │   └── memory.go            # Thread-safe in-memory store (sync.RWMutex)
│   ├── model/                   # Domain Models & Context Abstraction
│   │   ├── context.go           # RequestContext, header accessors, context keys
│   │   ├── task.go              # Task struct, DTOs (CreateTaskInput, UpdateTaskInput)
│   │   └── errors.go            # Domain error sentinels (ErrNotFound, ErrForbidden, etc.)
│   └── middleware/              # Cross-cutting HTTP middlewares
│       └── middleware.go        # RequestContext, logging, recovery, CORS
├── go.mod
└── README.md
```

### Context & Request Header Flow

```
[ HTTP Request ] (e.g. X-User-ID: usr_123, X-Role: admin)
       │
       ▼
[ RequestContextMiddleware ] ── Extracts HTTP headers into model.RequestContext & embeds into r.Context()
       │
       ▼
[ Controller / Handler ] ────── Passes r.Context() directly to Service: svc.CreateTask(r.Context(), input)
       │
       ▼
[ Business Service ] ────────── Inspects context:
                                - model.GetUserID(ctx)
                                - model.GetHeader(ctx, "X-Role")
                                - ctx.Err() (cancellation/timeout checks)
```

---

## 🚀 Getting Started

### 1. Run the API Server

```bash
go run ./cmd/api
```
The server will start on port `8080`:
```
Server starting on http://localhost:8080
```

### 2. Run the Unit & Handler Tests

```bash
go test -v ./internal/...
```

---

## 📡 API Endpoints

### Task Management (`/api/v1/tasks`)

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/tasks` | Create a new task (attaches user & tenant from headers) |
| `GET` | `/api/v1/tasks` | List all tasks |
| `GET` | `/api/v1/tasks/{id}` | Get a task by ID |
| `PUT` | `/api/v1/tasks/{id}` | Update a task by ID |
| `DELETE` | `/api/v1/tasks/{id}` | Delete a task by ID |

#### 1. Create a Task (with Request Headers)
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "X-User-ID: alice-123" \
  -H "X-Tenant-ID: acme-corp" \
  -d '{"title": "Learn Go Clean Architecture", "description": "Pass context to business logic"}'
```
**Response (`201 Created`):**
```json
{
  "id": 1,
  "title": "Learn Go Clean Architecture",
  "description": "Pass context to business logic",
  "completed": false,
  "created_by": "alice-123",
  "tenant_id": "acme-corp",
  "created_at": "2026-11-22T01:35:00Z",
  "updated_at": "2026-11-22T01:35:00Z"
}
```

#### 2. Business Check on Headers (`X-Role: read-only`)
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "X-Role: read-only" \
  -d '{"title": "Forbidden task"}'
```
**Response (`403 Forbidden`):**
```json
{
  "error": "forbidden: insufficient permissions: read-only users cannot create tasks"
}
```
