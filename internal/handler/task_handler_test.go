package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-learn/internal/handler"
	"go-learn/internal/model"
	"go-learn/internal/repository"
	"go-learn/internal/service"
)

func setupTestRouter() *http.ServeMux {
	repo := repository.NewMemoryTaskRepository()
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func TestTaskHandler_CreateAndGet(t *testing.T) {
	router := setupTestRouter()

	// 1. Create a task via POST
	payload := []byte(`{"title":"Learn Clean Go Architecture","description":"Practice net/http"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

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

	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d. Body: %s", getW.Code, getW.Body.String())
	}

	// 3. Fetch non-existing task
	notReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/999", nil)
	notW := httptest.NewRecorder()

	router.ServeHTTP(notW, notReq)

	if notW.Code != http.StatusNotFound {
		t.Errorf("expected status 404 Not Found, got %d", notW.Code)
	}
}
