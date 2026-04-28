package memory

import (
	"context"
	"sync"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

// masterDataRepository is an in-memory implementation of repository.MasterDataRepository (for development/demo).
type masterDataRepository struct {
	mu    sync.RWMutex
	items map[string]*entity.MasterData
}

// NewMasterDataRepository creates a new in-memory master data repository.
func NewMasterDataRepository() repository.MasterDataRepository {
	return &masterDataRepository{items: make(map[string]*entity.MasterData)}
}

func (r *masterDataRepository) GetByID(ctx context.Context, id string) (*entity.MasterData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.items[id]
	if !ok {
		return nil, nil
	}
	return m, nil
}

func (r *masterDataRepository) List(ctx context.Context, filter repository.ListFilter) ([]*entity.MasterData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*entity.MasterData
	for _, m := range r.items {
		if filter.Module != nil && m.Module != *filter.Module {
			continue
		}
		if filter.Type != nil && m.Type != *filter.Type {
			continue
		}
		if filter.Name != nil && m.Name != *filter.Name {
			continue
		}
		if filter.IsActive != nil && m.IsActive != *filter.IsActive {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *masterDataRepository) Create(ctx context.Context, m *entity.MasterData) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[m.ID] = m
	return nil
}

func (r *masterDataRepository) Update(ctx context.Context, m *entity.MasterData) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[m.ID] = m
	return nil
}

func (r *masterDataRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}
