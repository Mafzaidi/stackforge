package dualwrite

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
)

type todoRepository struct {
	primary   repository.TodoRepository
	secondary repository.TodoRepository
	logger    service.Logger
}

func NewTodoRepository(primary, secondary repository.TodoRepository, logger service.Logger) repository.TodoRepository {
	return &todoRepository{primary: primary, secondary: secondary, logger: logger}
}

func (r *todoRepository) GetByID(ctx context.Context, id string) (*entity.Todo, error) {
	return r.primary.GetByID(ctx, id)
}

func (r *todoRepository) List(ctx context.Context) ([]*entity.Todo, error) {
	return r.primary.List(ctx)
}

func (r *todoRepository) Create(ctx context.Context, t *entity.Todo) error {
	if err := r.primary.Create(ctx, t); err != nil {
		return err
	}
	if err := r.secondary.Create(ctx, t); err != nil {
		r.logger.Error("dual-write: secondary todo create failed", service.Fields{"error": err.Error(), "id": t.ID})
	}
	return nil
}

func (r *todoRepository) Update(ctx context.Context, t *entity.Todo) error {
	if err := r.primary.Update(ctx, t); err != nil {
		return err
	}
	if err := r.secondary.Update(ctx, t); err != nil {
		r.logger.Error("dual-write: secondary todo update failed", service.Fields{"error": err.Error(), "id": t.ID})
	}
	return nil
}

func (r *todoRepository) Delete(ctx context.Context, id string) error {
	if err := r.primary.Delete(ctx, id); err != nil {
		return err
	}
	if err := r.secondary.Delete(ctx, id); err != nil {
		r.logger.Error("dual-write: secondary todo delete failed", service.Fields{"error": err.Error(), "id": id})
	}
	return nil
}
