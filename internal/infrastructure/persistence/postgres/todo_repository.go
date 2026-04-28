package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/infrastructure/persistence/postgres/model"
)

// todoRepository implements repository.TodoRepository using PostgreSQL via GORM.
type todoRepository struct {
	db *gorm.DB
}

// NewTodoRepository creates a new PostgreSQL todo repository.
// Accepts a GORM database client from ConnectionManager.
// Requirements: 4.2, 4.5, 4.6, 4.7, 4.8, 4.9
func NewTodoRepository(db *gorm.DB) repository.TodoRepository {
	return &todoRepository{
		db: db,
	}
}

func (r *todoRepository) GetByID(ctx context.Context, id string) (*entity.Todo, error) {
	var m model.Todo
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return m.ToEntity(), nil
}

func (r *todoRepository) List(ctx context.Context) ([]*entity.Todo, error) {
	var models []model.Todo
	result := r.db.WithContext(ctx).Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	todos := make([]*entity.Todo, len(models))
	for i, m := range models {
		todos[i] = m.ToEntity()
	}
	return todos, nil
}

func (r *todoRepository) Create(ctx context.Context, t *entity.Todo) error {
	m := model.TodoFromEntity(t)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *todoRepository) Update(ctx context.Context, t *entity.Todo) error {
	m := model.TodoFromEntity(t)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *todoRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Todo{}).Error
}
