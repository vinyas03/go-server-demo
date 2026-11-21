package model

import "time"

// Task represents the core domain entity.
type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedBy   string    `json:"created_by,omitempty"`
	TenantID    string    `json:"tenant_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTaskInput represents the required fields to create a new task.
type CreateTaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTaskInput represents the fields that can be updated on an existing task.
type UpdateTaskInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}
