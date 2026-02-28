package repository

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// TodoRepository defines the interface for todo persistence operations.
// This interface is implementation-agnostic and can be implemented by
// in-memory, PostgreSQL, or any other storage mechanism.
type TodoRepository interface {
	// GetByID retrieves a todo by its ID.
	// Returns nil if the todo is not found.
	GetByID(ctx context.Context, id string) (*entity.Todo, error)

	// List retrieves all todos.
	List(ctx context.Context) ([]*entity.Todo, error)

	// Create persists a new todo.
	Create(ctx context.Context, t *entity.Todo) error

	// Update persists changes to an existing todo.
	Update(ctx context.Context, t *entity.Todo) error

	// Delete removes a todo by its ID.
	Delete(ctx context.Context, id string) error
}
