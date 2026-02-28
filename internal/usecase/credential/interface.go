package credential

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// ListUseCase defines the interface for listing credentials.
type ListUseCase interface {
	Execute(ctx context.Context) ([]*entity.Credential, error)
}
