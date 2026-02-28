package memory

import (
	"context"
	"sync"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

// todoRepository is an in-memory implementation of repository.TodoRepository (for development/demo).
type todoRepository struct {
	mu    sync.RWMutex
	items map[string]*entity.Todo
}

// NewTodoRepository creates a new in-memory todo repository.
func NewTodoRepository() repository.TodoRepository {
	return &todoRepository{items: make(map[string]*entity.Todo)}
}

func (r *todoRepository) GetByID(ctx context.Context, id string) (*entity.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.items[id]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (r *todoRepository) List(ctx context.Context) ([]*entity.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entity.Todo, 0, len(r.items))
	for _, t := range r.items {
		out = append(out, t)
	}
	return out, nil
}

func (r *todoRepository) Create(ctx context.Context, t *entity.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[t.ID] = t
	return nil
}

func (r *todoRepository) Update(ctx context.Context, t *entity.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[t.ID] = t
	return nil
}

func (r *todoRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}
