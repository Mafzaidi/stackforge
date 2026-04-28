package credential

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

type CreateUseCase interface {
	Execute(
		ctx context.Context,
		userID,
		title,
		siteUrl,
		faviconUrl,
		username,
		password,
		notes string,
		isFavorite bool,
		passwordStrength int32,
		vaultID,
		categoryID string,
		tags []string,
	) (*entity.Credential, error)
}

type ListUseCase interface {
	Execute(ctx context.Context, limit, offset int) ([]*entity.Credential, error)
}

type ListByUserUseCase interface {
	Execute(ctx context.Context, userID string, page, limit int) (*repository.PaginatedCredentials, error)
}
