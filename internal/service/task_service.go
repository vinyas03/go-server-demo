package service

import (
	"context"
	"fmt"
	"strings"

	"go-learn/internal/model"
	"go-learn/internal/repository"

	"go.elastic.co/apm/v2"
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
	// Start a new APM span for tracing the CreateTask operation
	//here we are starting a new APM span for the CreateTask operation. This allows us to trace the execution of this method and capture performance metrics, errors, and other relevant information.
	//  The span is associated with the current context, which may contain request-scoped metadata such as user ID, tenant ID, and other headers.
	span, ctx := apm.StartSpan(ctx, "TaskService.CreateTask", "app.service")
	//this defer statement ensures that the span is ended when the function returns, regardless of whether it returns normally or due to an error. This is important for proper resource management and accurate tracing.
	defer span.End()

	// 1. Context Cancellation / Timeout check
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 2. Business Check on Headers / Context Values
	role := model.GetHeader(ctx, "X-Role")
	if strings.EqualFold(role, "read-only") {
		err := fmt.Errorf("%w: read-only users cannot create tasks", model.ErrForbidden)
		apm.CaptureError(ctx, err).Send()
		return nil, err
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
		apm.CaptureError(ctx, err).Send()
		return nil, err
	}

	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, id int64) (*model.Task, error) {
	span, ctx := apm.StartSpan(ctx, "TaskService.GetTask", "app.service")
	defer span.End()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid task id", model.ErrInvalidInput)
	}

	task, err := s.repo.GetByID(ctx, id)
	if err != nil && err != model.ErrNotFound {
		apm.CaptureError(ctx, err).Send()
	}
	return task, err
}

func (s *taskService) ListTasks(ctx context.Context) ([]*model.Task, error) {
	span, ctx := apm.StartSpan(ctx, "TaskService.ListTasks", "app.service")
	defer span.End()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return s.repo.GetAll(ctx)
}

func (s *taskService) UpdateTask(ctx context.Context, id int64, input model.UpdateTaskInput) (*model.Task, error) {
	span, ctx := apm.StartSpan(ctx, "TaskService.UpdateTask", "app.service")
	defer span.End()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Business Check: verify role permissions from context
	role := model.GetHeader(ctx, "X-Role")
	if strings.EqualFold(role, "read-only") {
		err := fmt.Errorf("%w: read-only users cannot modify tasks", model.ErrForbidden)
		apm.CaptureError(ctx, err).Send()
		return nil, err
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
		apm.CaptureError(ctx, err).Send()
		return nil, err
	}

	return task, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id int64) error {
	span, ctx := apm.StartSpan(ctx, "TaskService.DeleteTask", "app.service")
	defer span.End()

	if err := ctx.Err(); err != nil {
		return err
	}

	// Business Check: verify role permissions from context
	role := model.GetHeader(ctx, "X-Role")
	if strings.EqualFold(role, "read-only") {
		err := fmt.Errorf("%w: read-only users cannot delete tasks", model.ErrForbidden)
		apm.CaptureError(ctx, err).Send()
		return err
	}

	if id <= 0 {
		return fmt.Errorf("%w: invalid task id", model.ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}
