package repository

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

type TagFilter struct {
	UserID   *string
	Module   *string
	RefID    *string
	IsActive *bool
}

type TagRepository interface {
	GetByID(ctx context.Context, id string) (*entity.Tag, error)
	List(ctx context.Context, filter TagFilter) ([]*entity.Tag, error)
	Create(ctx context.Context, t *entity.Tag) error
	Update(ctx context.Context, t *entity.Tag) error
	BulkUpsert(ctx context.Context, tags []*entity.Tag) error
	Delete(ctx context.Context, id string) error
}
