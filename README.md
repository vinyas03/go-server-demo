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
│   │   ├── task_service.go      # Core business rules, validation, orchestrations
│   │   └── task_service_test.go # Business logic unit tests (independent of HTTP)
│   ├── repository/              # Persistence / Data Access Layer
│   │   ├── repository.go        # Storage interface (swappable with SQL, Mongo, etc.)
│   │   └── memory.go            # Thread-safe in-memory store (sync.RWMutex)
│   ├── model/                   # Domain Models & Entities
│   │   ├── task.go              # Task struct, DTOs (CreateTaskInput, UpdateTaskInput)
│   │   └── errors.go            # Domain error sentinels (ErrNotFound, ErrInvalidInput)
│   └── middleware/              # Cross-cutting HTTP middlewares
│       └── middleware.go        # Request logging, panic recovery, and CORS
├── go.mod
└── README.md
```

### Layer Responsibilities & Data Flow

```
[ HTTP Request ]
       │
       ▼
[ Middleware Chain ] ───────── (Logging, Panic Recovery, CORS)
       │
       ▼
[ Controller / Handler ] ───── Parses HTTP parameters/body, maps errors to HTTP codes
       │ (Calls Service)
       ▼
[ Business Logic / Service ] ─ Validates business rules (No net/http dependencies)
       │ (Calls Repository)
       ▼
[ Data Access / Repository ] ─ Persists & retrieves data (Interface-driven)
```

---

## 🚀 Getting Started

### 1. Run the API Server

```bash
go run ./cmd/api
```
The server will start on port `8080` (or the port specified by the `PORT` environment variable):
```
Server starting on http://localhost:8080
```

### 2. Run the Unit & Handler Tests

```bash
go test -v ./...
```

---

## 📡 API Endpoints

### Health Check
- **`GET /health`**
  ```bash
  curl http://localhost:8080/health
  ```
  **Response:**
  ```json
  {
    "status": "healthy",
    "time": "2026-11-22T00:50:00Z"
  }
  ```

---

### Task Management (`/api/v1/tasks`)

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/tasks` | Create a new task |
| `GET` | `/api/v1/tasks` | List all tasks |
| `GET` | `/api/v1/tasks/{id}` | Get a task by ID |
| `PUT` | `/api/v1/tasks/{id}` | Update a task by ID |
| `DELETE` | `/api/v1/tasks/{id}` | Delete a task by ID |

#### 1. Create a Task
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn Go Clean Architecture", "description": "Build layered net/http API"}'
```
**Response (`201 Created`):**
```json
{
  "id": 1,
  "title": "Learn Go Clean Architecture",
  "description": "Build layered net/http API",
  "completed": false,
  "created_at": "2026-11-22T00:50:00Z",
  "updated_at": "2026-11-22T00:50:00Z"
}
```

#### 2. List Tasks
```bash
curl http://localhost:8080/api/v1/tasks
```

#### 3. Get Task by ID
```bash
curl http://localhost:8080/api/v1/tasks/1
```

#### 4. Update a Task
```bash
curl -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'
```

#### 5. Delete a Task
```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/1
```
**Response (`204 No Content`)**

---

## 🛡️ Key Architectural Highlights

1. **Standard Library Only (`net/http`)**:
   Uses Go 1.22+ enhanced ServeMux routing (`POST /api/v1/tasks`, `GET /api/v1/tasks/{id}` with `r.PathValue("id")`). No heavy third-party framework dependencies needed.
2. **Separation of Concerns**:
   - `handler` knows nothing about database/storage.
   - `service` knows nothing about HTTP (`net/http`), JSON, or status codes.
   - `repository` is an interface, making switching from in-memory to PostgreSQL or MongoDB seamless.
3. **Robust Production Readiness**:
   - Panic recovery middleware prevents server crashes on runtime panics.
   - Structured logging middleware tracks request latency and HTTP status.
   - Server timeouts (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`) defend against Slowloris attacks.
   - Graceful shutdown handles `SIGINT`/`SIGTERM` allowing ongoing requests to complete.
