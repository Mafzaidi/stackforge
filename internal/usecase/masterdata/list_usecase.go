package masterdata

import (
	"context"
	"errors"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
)

type listUseCase struct {
	repo   repository.MasterDataRepository
	logger service.Logger
}

// NewListUseCase creates a new instance of ListUseCase.
func NewListUseCase(
	repo repository.MasterDataRepository,
	logger service.Logger,
) ListUseCase {
	return &listUseCase{repo: repo, logger: logger}
}

// Execute retrieves menu items from master data, filtered by user permissions.
func (uc *listUseCase) Execute(ctx context.Context, module, typ string) ([]entity.MasterData, error) {
	if module == "" {
		return nil, errors.New("module is required")
	}

	if typ == "" {
		return nil, errors.New("type is required")
	}

	active := true

	items, err := uc.repo.List(ctx, repository.ListFilter{
		Module:   &module,
		Type:     &typ,
		IsActive: &active,
	})
	if err != nil {
		uc.logger.Error("failed to fetch master data", service.Fields{
			"module": module,
			"type":   typ,
			"error":  err.Error(),
		})
		return nil, errors.New("failed to fetch master data")
	}

	var result []entity.MasterData
	for _, item := range items {

		result = append(result, entity.MasterData{
			ID:          item.ID,
			Module:      item.Module,
			Type:        item.Type,
			Name:        item.Name,
			Description: item.Description,
			Icon:        item.Icon,
			SortOrder:   item.SortOrder,
			IsActive:    item.IsActive,
			Metadata:    item.Metadata,
		})
	}

	return result, nil
}
