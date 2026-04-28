package menu

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// ListUseCase defines the interface for listing menu items filtered by user permissions.
type ListUseCase interface {
	Execute(ctx context.Context, claims *entity.Claims, appCode string) ([]entity.MenuItem, error)
}
