package vault

import (
	"context"
	"errors"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
)

type listUseCase struct {
	repo   repository.VaultRepository
	logger service.Logger
}

func NewListUseCase(
	repo repository.VaultRepository,
	logger service.Logger,
) ListUseCase {
	return &listUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (uc *listUseCase) Execute(ctx context.Context, userID string) ([]*entity.Vault, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	items, err := uc.repo.List(ctx, userID)

	if err != nil {
		uc.logger.Error("failed to fetch vault", service.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.New("failed to fetch vault")
	}

	return items, nil
}
