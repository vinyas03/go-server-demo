package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-learn/internal/model"
	"go-learn/internal/service"
)

// TaskHandler handles incoming HTTP requests for task resources.
type TaskHandler struct {
	service service.TaskService
}

// NewTaskHandler constructs a TaskHandler with injected business service.
func NewTaskHandler(service service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

// RegisterRoutes registers endpoints using standard net/http ServeMux routing.
func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tasks", h.Create)
	mux.HandleFunc("GET /api/v1/tasks", h.List)
	mux.HandleFunc("GET /api/v1/tasks/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/tasks/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/tasks/{id}", h.Delete)
}

// Create handles POST /api/v1/tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input model.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	task, err := h.service.CreateTask(r.Context(), input)
	if err != nil {
		RespondWithError(w, err)
		return
	}

	JSON(w, http.StatusCreated, task)
}

// List handles GET /api/v1/tasks
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.ListTasks(r.Context())
	if err != nil {
		RespondWithError(w, err)
		return
	}

	JSON(w, http.StatusOK, tasks)
}

// GetByID handles GET /api/v1/tasks/{id}
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid task id: must be an integer")
		return
	}

	task, err := h.service.GetTask(r.Context(), id)
	if err != nil {
		RespondWithError(w, err)
		return
	}

	JSON(w, http.StatusOK, task)
}

// Update handles PUT /api/v1/tasks/{id}
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid task id: must be an integer")
		return
	}

	var input model.UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	updated, err := h.service.UpdateTask(r.Context(), id, input)
	if err != nil {
		RespondWithError(w, err)
		return
	}

	JSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /api/v1/tasks/{id}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid task id: must be an integer")
		return
	}

	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		RespondWithError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
