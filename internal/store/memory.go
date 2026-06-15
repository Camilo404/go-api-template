package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Camilo404/go-api-template/internal/models"
)

// MemoryTaskStore is a goroutine-safe in-memory implementation of
// TaskStorer. Use it for tests, demos, and local development. Replace
// it with a real backend in production.
type MemoryTaskStore struct {
	mu     sync.RWMutex
	tasks  map[int]models.Task
	nextID int
}

// NewMemoryTaskStore returns an empty store ready for use.
func NewMemoryTaskStore() *MemoryTaskStore {
	return &MemoryTaskStore{
		tasks:  make(map[int]models.Task),
		nextID: 1,
	}
}

// List returns a page of tasks ordered by ID ascending along with the
// total count. Returns an empty slice (not nil) when the page is empty
// so JSON encoders emit `[]` instead of `null`.
func (s *MemoryTaskStore) List(_ context.Context, page, perPage int) ([]models.Task, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]models.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		all = append(all, t)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	total := len(all)
	start := (page - 1) * perPage
	if start >= total {
		return []models.Task{}, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

// GetByID fetches a single task. Returns models.ErrNotFound if absent.
func (s *MemoryTaskStore) GetByID(_ context.Context, id int) (models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return models.Task{}, models.ErrNotFound
	}
	return t, nil
}

// Create assigns a new ID and timestamps, then stores the task.
func (s *MemoryTaskStore) Create(_ context.Context, in models.CreateTaskInput) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	t := models.Task{
		ID:          s.nextID,
		Title:       in.Title,
		Description: in.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.tasks[t.ID] = t
	s.nextID++
	return t, nil
}

// Update applies the provided fields and refreshes UpdatedAt. Nil
// fields in `in` are left untouched.
func (s *MemoryTaskStore) Update(_ context.Context, id int, in models.UpdateTaskInput) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return models.Task{}, models.ErrNotFound
	}
	if in.Title != nil {
		t.Title = *in.Title
	}
	if in.Description != nil {
		t.Description = *in.Description
	}
	if in.Completed != nil {
		t.Completed = *in.Completed
	}
	t.UpdatedAt = time.Now().UTC()
	s.tasks[id] = t
	return t, nil
}

// Delete removes a task. Returns models.ErrNotFound if absent.
func (s *MemoryTaskStore) Delete(_ context.Context, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return models.ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}
