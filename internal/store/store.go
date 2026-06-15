// Package store defines the persistence interface and provides an
// in-memory implementation suitable for tests and local development.
//
// To plug in a real database, implement TaskStorer (or a new interface
// for your entity) and wire it in cmd/api/main.go.
package store

import (
	"context"

	"github.com/Camilo404/go-api-template/internal/models"
)

// TaskStorer is the contract handlers/services depend on. Keeping it
// narrow (one method per use case) makes mocking trivial.
type TaskStorer interface {
	List(ctx context.Context, page, perPage int) ([]models.Task, int, error)
	GetByID(ctx context.Context, id int) (models.Task, error)
	Create(ctx context.Context, in models.CreateTaskInput) (models.Task, error)
	Update(ctx context.Context, id int, in models.UpdateTaskInput) (models.Task, error)
	Delete(ctx context.Context, id int) error
}
