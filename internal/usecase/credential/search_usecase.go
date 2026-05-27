package credential

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

type searchUseCase struct {
	repo repository.CredentialRepository
}

// NewSearchUseCase creates a new instance of SearchUseCase.
func NewSearchUseCase(repo repository.CredentialRepository) SearchUseCase {
	return &searchUseCase{repo: repo}
}

// Execute searches credentials by keyword with pagination.
func (uc *searchUseCase) Execute(ctx context.Context, keyword string, limit, offset int) ([]*entity.Credential, int64, error) {
	return uc.repo.SearchByKeyword(ctx, keyword, limit, offset)
}
