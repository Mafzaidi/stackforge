package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

type userProfilesRepository struct {
	mu   sync.RWMutex
	data map[string]*entity.UserProfiles // keyed by ID
}

// NewUserProfilesRepository creates a new in-memory user profiles repository.
func NewUserProfilesRepository() repository.UserProfilesRepository {
	return &userProfilesRepository{
		data: make(map[string]*entity.UserProfiles),
	}
}

func (r *userProfilesRepository) GetByID(ctx context.Context, id string) (*entity.UserProfiles, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.data[id]
	if !exists {
		return nil, fmt.Errorf("user profile not found: %s", id)
	}
	return p, nil
}

func (r *userProfilesRepository) GetByUserID(ctx context.Context, userID string) (*entity.UserProfiles, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.data {
		if p.UserID == userID {
			return p, nil
		}
	}
	return nil, fmt.Errorf("user profile not found for user: %s", userID)
}

func (r *userProfilesRepository) List(ctx context.Context) ([]*entity.UserProfiles, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	profiles := make([]*entity.UserProfiles, 0, len(r.data))
	for _, p := range r.data {
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (r *userProfilesRepository) Create(ctx context.Context, p *entity.UserProfiles) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[p.ID]; exists {
		return fmt.Errorf("user profile already exists: %s", p.ID)
	}
	r.data[p.ID] = p
	return nil
}

func (r *userProfilesRepository) Update(ctx context.Context, p *entity.UserProfiles) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[p.ID]; !exists {
		return fmt.Errorf("user profile not found: %s", p.ID)
	}
	r.data[p.ID] = p
	return nil
}

func (r *userProfilesRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return fmt.Errorf("user profile not found: %s", id)
	}
	delete(r.data, id)
	return nil
}
