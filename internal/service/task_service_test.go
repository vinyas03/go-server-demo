package service_test

import (
	"context"
	"errors"
	"testing"

	"go-learn/internal/model"
	"go-learn/internal/repository"
	"go-learn/internal/service"
)

func TestCreateTask(t *testing.T) {
	repo := repository.NewMemoryTaskRepository()
	svc := service.NewTaskService(repo)

	t.Run("successful creation with context headers and metadata", func(t *testing.T) {
		// Provide context with simulated request metadata (user ID, tenant ID, headers)
		ctx := model.WithRequestContext(context.Background(), model.RequestContext{
			UserID:   "usr_999",
			TenantID: "tenant_acme",
			Headers: map[string]string{
				"X-Role": "admin",
			},
		})

		input := model.CreateTaskInput{
			Title:       "Test task",
			Description: "Test description",
		}

		task, err := svc.CreateTask(ctx, input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if task.ID != 1 {
			t.Errorf("expected task ID 1, got %d", task.ID)
		}
		if task.Title != "Test task" {
			t.Errorf("expected title 'Test task', got '%s'", task.Title)
		}
		if task.CreatedBy != "usr_999" {
			t.Errorf("expected CreatedBy 'usr_999', got '%s'", task.CreatedBy)
		}
		if task.TenantID != "tenant_acme" {
			t.Errorf("expected TenantID 'tenant_acme', got '%s'", task.TenantID)
		}
	})

	t.Run("business check rejects read-only role from context", func(t *testing.T) {
		ctx := model.WithRequestContext(context.Background(), model.RequestContext{
			Headers: map[string]string{
				"X-Role": "read-only",
			},
		})

		_, err := svc.CreateTask(ctx, model.CreateTaskInput{
			Title: "Forbidden Task",
		})
		if err == nil {
			t.Fatalf("expected error for read-only role, got nil")
		}
		if !errors.Is(err, model.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("empty title returns validation error", func(t *testing.T) {
		input := model.CreateTaskInput{
			Title: "   ",
		}

		_, err := svc.CreateTask(context.Background(), input)
		if err == nil {
			t.Fatalf("expected error for empty title, got nil")
		}
		if !errors.Is(err, model.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})
}

func TestUpdateAndGetTask(t *testing.T) {
	repo := repository.NewMemoryTaskRepository()
	svc := service.NewTaskService(repo)
	ctx := context.Background()

	task, err := svc.CreateTask(ctx, model.CreateTaskInput{
		Title: "Initial title",
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	completed := true
	newTitle := "Updated title"
	updated, err := svc.UpdateTask(ctx, task.ID, model.UpdateTaskInput{
		Title:     &newTitle,
		Completed: &completed,
	})
	if err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	if updated.Title != newTitle || !updated.Completed {
		t.Errorf("unexpected updated task values: %+v", updated)
	}
}
