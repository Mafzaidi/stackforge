package vault

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

type CreateUseCase interface {
	Execute(ctx context.Context, userID, name, description string) (*entity.Vault, error)
}

type ListUseCase interface {
	Execute(ctx context.Context, userID string) ([]*entity.Vault, error)
}
