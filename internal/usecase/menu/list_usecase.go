package menu

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

type listUseCase struct {
	repo repository.MasterDataRepository
}

// NewListUseCase creates a new instance of ListUseCase.
func NewListUseCase(repo repository.MasterDataRepository) ListUseCase {
	return &listUseCase{repo: repo}
}

// Execute retrieves menu items from master data, filtered by user permissions.
func (uc *listUseCase) Execute(ctx context.Context, claims *entity.Claims, appCode string) ([]entity.MenuItem, error) {
	module := "navigation"
	typ := "sidebar_menu"
	active := true

	items, err := uc.repo.List(ctx, repository.ListFilter{
		Module:   &module,
		Type:     &typ,
		IsActive: &active,
	})
	if err != nil {
		return nil, err
	}

	var result []entity.MenuItem
	for _, item := range items {
		requiredPerm, _ := item.Metadata["required_permission"].(string)

		if requiredPerm != "" && !claims.HasPermission(appCode, requiredPerm) {
			continue
		}

		url, _ := item.Metadata["url"].(string)

		result = append(result, entity.MenuItem{
			ID:        item.ID,
			Title:     item.Name,
			URL:       url,
			Icon:      item.Icon,
			SortOrder: item.SortOrder,
		})
	}

	return result, nil
}
