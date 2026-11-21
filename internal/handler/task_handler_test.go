package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-learn/internal/handler"
	"go-learn/internal/middleware"
	"go-learn/internal/model"
	"go-learn/internal/repository"
	"go-learn/internal/service"

	"go.elastic.co/apm/module/apmhttp/v2"
)

func setupTestServer() http.Handler {
	repo := repository.NewMemoryTaskRepository()
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Wrap with RequestContextMiddleware and CubeAPM tracer
	wrapped := middleware.Chain(
		mux,
		middleware.Recovery,
		middleware.RequestContextMiddleware,
	)

	return apmhttp.Wrap(wrapped)
}

func TestTaskHandler_CreateWithHeaders(t *testing.T) {
	server := setupTestServer()

	// 1. POST request with X-User-ID and X-Tenant-ID headers
	payload := []byte(`{"title":"Task with Context","description":"Headers in business layer"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "alice-123")
	req.Header.Set("X-Tenant-ID", "org-alpha")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var created model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if created.CreatedBy != "alice-123" {
		t.Errorf("expected CreatedBy 'alice-123', got '%s'", created.CreatedBy)
	}
	if created.TenantID != "org-alpha" {
		t.Errorf("expected TenantID 'org-alpha', got '%s'", created.TenantID)
	}

	// 2. POST request with X-Role: read-only header (business check rejection)
	reqReadOnly := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(payload))
	reqReadOnly.Header.Set("Content-Type", "application/json")
	reqReadOnly.Header.Set("X-Role", "read-only")
	wReadOnly := httptest.NewRecorder()

	server.ServeHTTP(wReadOnly, reqReadOnly)

	if wReadOnly.Code != http.StatusForbidden {
		t.Errorf("expected status 403 Forbidden for read-only role, got %d", wReadOnly.Code)
	}
}

func TestTaskHandler_CreateAndGet(t *testing.T) {
	server := setupTestServer()

	// 1. Create a task via POST
	payload := []byte(`{"title":"Learn Clean Go Architecture","description":"Practice net/http"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var created model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if created.ID != 1 {
		t.Errorf("expected task id 1, got %d", created.ID)
	}

	// 2. Fetch the created task via GET
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/1", nil)
	getW := httptest.NewRecorder()

	server.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d. Body: %s", getW.Code, getW.Body.String())
	}

	// 3. Fetch non-existing task
	notReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/999", nil)
	notW := httptest.NewRecorder()

	server.ServeHTTP(notW, notReq)

	if notW.Code != http.StatusNotFound {
		t.Errorf("expected status 404 Not Found, got %d", notW.Code)
	}
}
