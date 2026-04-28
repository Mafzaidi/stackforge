package masterdata

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// ListUseCase defines the interface for listing menu items filtered by user permissions.
type ListUseCase interface {
	Execute(ctx context.Context, module, typ string) ([]entity.MasterData, error)
}
