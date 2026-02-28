package todo

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// ListUseCase defines the interface for listing todos.
type ListUseCase interface {
	Execute(ctx context.Context) ([]*entity.Todo, error)
}
