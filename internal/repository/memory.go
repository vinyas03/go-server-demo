package repository

import (
	"context"
	"sync"
	"time"

	"go-learn/internal/model"
)

// MemoryTaskRepository provides a thread-safe in-memory implementation of TaskRepository.
type MemoryTaskRepository struct {
	mu     sync.RWMutex
	tasks  map[int64]*model.Task
	nextID int64
}

// NewMemoryTaskRepository initializes a new in-memory task repository.
func NewMemoryTaskRepository() *MemoryTaskRepository {
	return &MemoryTaskRepository{
		tasks:  make(map[int64]*model.Task),
		nextID: 1,
	}
}

func (r *MemoryTaskRepository) Create(ctx context.Context, task *model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.ID = r.nextID
	r.nextID++
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now

	// Clone to avoid external pointer mutation
	copied := *task
	r.tasks[task.ID] = &copied
	return nil
}

func (r *MemoryTaskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, model.ErrNotFound
	}

	copied := *task
	return &copied, nil
}

func (r *MemoryTaskRepository) GetAll(ctx context.Context) ([]*model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		copied := *task
		result = append(result, &copied)
	}
	return result, nil
}

func (r *MemoryTaskRepository) Update(ctx context.Context, task *model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return model.ErrNotFound
	}

	task.UpdatedAt = time.Now().UTC()
	copied := *task
	r.tasks[task.ID] = &copied
	return nil
}

func (r *MemoryTaskRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return model.ErrNotFound
	}

	delete(r.tasks, id)
	return nil
}
