package repository

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

type VaultRepository interface {
	GetByID(ctx context.Context, id string) (*entity.Vault, error)
	GetByUserIDAndName(ctx context.Context, userID, name string) (*entity.Vault, error)
	List(ctx context.Context, userID string) ([]*entity.Vault, error)
	Create(ctx context.Context, v *entity.Vault) error
}
