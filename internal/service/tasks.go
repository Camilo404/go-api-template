// Package service contains the business logic of the application.
// Handlers stay thin (parse, call service, render) while services
// own validation, orchestration, and store calls. This separation
// makes the logic easy to unit test without HTTP plumbing.
package service

import (
	"context"
	"strconv"

	"github.com/Camilo404/go-api-template/internal/models"
	"github.com/Camilo404/go-api-template/internal/store"
)

// TaskService operates on tasks. Replace "Task" with your entity name.
type TaskService struct {
	store store.TaskStorer
}

// NewTaskService wires a TaskService over the given store.
func NewTaskService(s store.TaskStorer) *TaskService {
	return &TaskService{store: s}
}

// List returns a page of tasks.
func (s *TaskService) List(ctx context.Context, page, perPage int) ([]models.Task, int, error) {
	return s.store.List(ctx, page, perPage)
}

// Get fetches a single task by its string id (e.g. URL path param).
func (s *TaskService) Get(ctx context.Context, idStr string) (models.Task, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return models.Task{}, models.ErrInvalidID
	}
	return s.store.GetByID(ctx, id)
}

// Create validates input and persists a new task.
func (s *TaskService) Create(ctx context.Context, in models.CreateTaskInput) (models.Task, error) {
	if err := in.Validate(); err != nil {
		return models.Task{}, err
	}
	return s.store.Create(ctx, in)
}

// Update validates input, parses the id, and applies the patch.
func (s *TaskService) Update(ctx context.Context, idStr string, in models.UpdateTaskInput) (models.Task, error) {
	if err := in.Validate(); err != nil {
		return models.Task{}, err
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return models.Task{}, models.ErrInvalidID
	}
	return s.store.Update(ctx, id, in)
}

// Delete removes a task by id.
func (s *TaskService) Delete(ctx context.Context, idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return models.ErrInvalidID
	}
	return s.store.Delete(ctx, id)
}
