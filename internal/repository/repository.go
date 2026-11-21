package repository

import (
	"context"
	"go-learn/internal/model"
)

// TaskRepository defines the interface for task storage operations.
// This interface allows swapping in-memory store with PostgreSQL, MySQL, Mongo, etc.
type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id int64) (*model.Task, error)
	GetAll(ctx context.Context) ([]*model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id int64) error
}
