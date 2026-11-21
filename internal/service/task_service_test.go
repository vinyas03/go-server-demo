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
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
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
		if task.Completed {
			t.Errorf("expected task to be incomplete by default")
		}
	})

	t.Run("empty title returns validation error", func(t *testing.T) {
		input := model.CreateTaskInput{
			Title: "   ",
		}

		_, err := svc.CreateTask(ctx, input)
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
