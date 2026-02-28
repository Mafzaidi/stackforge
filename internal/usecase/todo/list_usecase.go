package todo

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

type listUseCase struct {
	repo repository.TodoRepository
}

// NewListUseCase creates a new instance of ListUseCase.
func NewListUseCase(repo repository.TodoRepository) ListUseCase {
	return &listUseCase{repo: repo}
}

// Execute retrieves all todos from the repository.
func (uc *listUseCase) Execute(ctx context.Context) ([]*entity.Todo, error) {
	return uc.repo.List(ctx)
}
