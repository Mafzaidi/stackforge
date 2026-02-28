package credential

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

type listUseCase struct {
	repo repository.CredentialRepository
}

// NewListUseCase creates a new instance of ListUseCase.
func NewListUseCase(repo repository.CredentialRepository) ListUseCase {
	return &listUseCase{repo: repo}
}

// Execute retrieves all credentials from the repository.
func (uc *listUseCase) Execute(ctx context.Context) ([]*entity.Credential, error) {
	return uc.repo.List(ctx)
}
