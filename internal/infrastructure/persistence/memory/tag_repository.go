package memory

import (
	"context"
	"sync"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

type tagRepository struct {
	mu    sync.RWMutex
	items map[string]*entity.Tag
}

func NewTagRepository() repository.TagRepository {
	return &tagRepository{items: make(map[string]*entity.Tag)}
}

func (r *tagRepository) GetByID(ctx context.Context, id string) (*entity.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if tag, exists := r.items[id]; exists {
		return tag, nil
	}
	return nil, nil
}

func (r *tagRepository) List(ctx context.Context, filter repository.TagFilter) ([]*entity.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tags := make([]*entity.Tag, 0)
	for _, tag := range r.items {
		if filter.UserID != nil && tag.UserID != *filter.UserID {
			continue
		}
		if filter.Module != nil && tag.Module != *filter.Module {
			continue
		}
		if filter.RefID != nil && tag.RefID != *filter.RefID {
			continue
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *tagRepository) Create(ctx context.Context, t *entity.Tag) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[t.ID] = t
	return nil
}

func (r *tagRepository) Update(ctx context.Context, t *entity.Tag) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[t.ID]; !exists {
		return nil // or return an error if you want to indicate that the tag doesn't exist
	}
	r.items[t.ID] = t
	return nil
}

func (r *tagRepository) BulkUpsert(ctx context.Context, tags []*entity.Tag) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range tags {
		r.items[t.ID] = t
	}
	return nil
}

func (r *tagRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}
