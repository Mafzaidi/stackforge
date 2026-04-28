package repository

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

type ListFilter struct {
	Module   *string
	Type     *string
	Name     *string
	IsActive *bool
}

// MasterDataRepository defines the interface for master data persistence.
type MasterDataRepository interface {
	GetByID(ctx context.Context, id string) (*entity.MasterData, error)
	List(ctx context.Context, filter ListFilter) ([]*entity.MasterData, error)
	Create(ctx context.Context, m *entity.MasterData) error
	Update(ctx context.Context, m *entity.MasterData) error
	Delete(ctx context.Context, id string) error
}
