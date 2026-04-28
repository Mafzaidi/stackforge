package memory

import (
	"context"
	"sync"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

type vaultRepository struct {
	mu    sync.RWMutex
	items map[string]*entity.Vault
}

func NewVaultRepository() repository.VaultRepository {
	return &vaultRepository{items: make(map[string]*entity.Vault)}
}
func (r *vaultRepository) GetByID(ctx context.Context, id string) (*entity.Vault, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if vault, ok := r.items[id]; ok {
		return vault, nil
	}
	return nil, nil
}

func (r *vaultRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*entity.Vault, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, vault := range r.items {
		if vault.UserID == userID && vault.Name == name {
			return vault, nil
		}
	}
	return nil, nil
}

func (r *vaultRepository) List(ctx context.Context, userID string) ([]*entity.Vault, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	vaults := make([]*entity.Vault, 0)
	for _, vault := range r.items {
		if vault.UserID == userID {
			vaults = append(vaults, vault)
		}
	}
	return vaults, nil
}

func (r *vaultRepository) Create(ctx context.Context, v *entity.Vault) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[v.ID] = v
	return nil
}
