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
	// Business validation
	trimmedTitle := strings.TrimSpace(input.Title)
	if trimmedTitle == "" {
		return nil, fmt.Errorf("%w: title cannot be empty", model.ErrInvalidInput)
	}

	if len(trimmedTitle) > 120 {
		return nil, fmt.Errorf("%w: title cannot exceed 120 characters", model.ErrInvalidInput)
	}

	task := &model.Task{
		Title:       trimmedTitle,
		Description: strings.TrimSpace(input.Description),
		Completed:   false,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, id int64) (*model.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid task id", model.ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *taskService) ListTasks(ctx context.Context) ([]*model.Task, error) {
	return s.repo.GetAll(ctx)
}

func (s *taskService) UpdateTask(ctx context.Context, id int64, input model.UpdateTaskInput) (*model.Task, error) {
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
	if id <= 0 {
		return fmt.Errorf("%w: invalid task id", model.ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}
