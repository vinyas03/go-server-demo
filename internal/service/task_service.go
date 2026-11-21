package service

import (
	"context"
	"fmt"
	"strings"

	"go-learn/internal/model"
	"go-learn/internal/repository"
)

// TaskService defines the business operations for tasks.
type TaskService interface {
	CreateTask(ctx context.Context, input model.CreateTaskInput) (*model.Task, error)
	GetTask(ctx context.Context, id int64) (*model.Task, error)
	ListTasks(ctx context.Context) ([]*model.Task, error)
	UpdateTask(ctx context.Context, id int64, input model.UpdateTaskInput) (*model.Task, error)
	DeleteTask(ctx context.Context, id int64) error
}

type taskService struct {
	repo repository.TaskRepository
}

// NewTaskService creates a new TaskService instance injecting the repository dependency.
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{
		repo: repo,
	}
}

func (s *taskService) CreateTask(ctx context.Context, input model.CreateTaskInput) (*model.Task, error) {
	// 1. Context Cancellation / Timeout check - these checks ensure that if the context has been canceled or has timed out, we stop processing and return an error immediately.
	//when can context be canceled or timeout? Context can be canceled or timeout in scenarios such as:
	// 1. Client Disconnect: If the client disconnects before the request is completed, the context will be canceled.
	// 2. Request Timeout: If a timeout is set for the request and it exceeds that duration, the context will be canceled.
	// 3. Manual Cancellation: If some part of the application explicitly cancels the context (e.g., due to an error or a specific condition), it will be canceled.

	//context deadline uses:
	//say for example, we have a long-running operation that takes a significant amount of time to complete. If we set a context deadline for this operation,
	// it allows us to specify a maximum duration for the operation to complete.
	// If the operation exceeds this duration, the context will be canceled, and we can handle the cancellation gracefully (e.g., by returning an error or cleaning up resources).

	//context cancellation uses:
	//context cancellation is useful in scenarios where we want to stop processing a request or operation based on certain conditions.
	// For example, if a user cancels a request in a web application, we can cancel the context associated with that request,
	// which will propagate the cancellation to all goroutines and operations that are using that context.
	//how this is detected by the server ? it is detected by the server through the context's Done channel. When the context is canceled or times out, the Done channel is closed,
	//  and any goroutines that are waiting on that channel will be notified of the cancellation.
	//imagine a scenario where a user initiates a long-running operation, such as generating a report. If the user decides to cancel the operation before it completes,
	//then the http connection will be closed, and the server will detect this closure through the context's Done channel. The server can then stop processing the operation and return
	//an appropriate response to the user.

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 2. Business Check on Headers / Context Values
	// Example: Inspect client role or permission passed via context
	role := model.GetHeader(ctx, "X-Role")
	if strings.EqualFold(role, "read-only") {
		return nil, fmt.Errorf("%w: read-only users cannot create tasks", model.ErrForbidden)
	}

	// 3. Input validation
	trimmedTitle := strings.TrimSpace(input.Title)
	if trimmedTitle == "" {
		return nil, fmt.Errorf("%w: title cannot be empty", model.ErrInvalidInput)
	}

	if len(trimmedTitle) > 120 {
		return nil, fmt.Errorf("%w: title cannot exceed 120 characters", model.ErrInvalidInput)
	}

	// 4. Populate entity with context metadata (User ID, Tenant ID)
	task := &model.Task{
		Title:       trimmedTitle,
		Description: strings.TrimSpace(input.Description),
		Completed:   false,
		CreatedBy:   model.GetUserID(ctx),
		TenantID:    model.GetTenantID(ctx),
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, id int64) (*model.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid task id", model.ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *taskService) ListTasks(ctx context.Context) ([]*model.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return s.repo.GetAll(ctx)
}

func (s *taskService) UpdateTask(ctx context.Context, id int64, input model.UpdateTaskInput) (*model.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Business Check: verify role permissions from context
	role := model.GetHeader(ctx, "X-Role")
	if strings.EqualFold(role, "read-only") {
		return nil, fmt.Errorf("%w: read-only users cannot modify tasks", model.ErrForbidden)
	}

	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid task id", model.ErrInvalidInput)
	}

	// Fetch existing task to ensure it exists
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply partial updates with validation
	if input.Title != nil {
		trimmedTitle := strings.TrimSpace(*input.Title)
		if trimmedTitle == "" {
			return nil, fmt.Errorf("%w: title cannot be empty", model.ErrInvalidInput)
		}
		if len(trimmedTitle) > 120 {
			return nil, fmt.Errorf("%w: title cannot exceed 120 characters", model.ErrInvalidInput)
		}
		task.Title = trimmedTitle
	}

	if input.Description != nil {
		task.Description = strings.TrimSpace(*input.Description)
	}

	if input.Completed != nil {
		task.Completed = *input.Completed
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Business Check: verify role permissions from context
	role := model.GetHeader(ctx, "X-Role")
	if strings.EqualFold(role, "read-only") {
		return fmt.Errorf("%w: read-only users cannot delete tasks", model.ErrForbidden)
	}

	if id <= 0 {
		return fmt.Errorf("%w: invalid task id", model.ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}
