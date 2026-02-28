package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
)

// credentialRepository implements repository.CredentialRepository using in-memory storage.
type credentialRepository struct {
	mu   sync.RWMutex
	data map[string]*entity.Credential
}

// NewCredentialRepository creates a new in-memory credential repository.
func NewCredentialRepository() repository.CredentialRepository {
	return &credentialRepository{
		data: make(map[string]*entity.Credential),
	}
}

func (r *credentialRepository) GetByID(ctx context.Context, id string) (*entity.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	credential, exists := r.data[id]
	if !exists {
		return nil, fmt.Errorf("credential not found: %s", id)
	}
	return credential, nil
}

func (r *credentialRepository) List(ctx context.Context) ([]*entity.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	credentials := make([]*entity.Credential, 0, len(r.data))
	for _, credential := range r.data {
		credentials = append(credentials, credential)
	}
	return credentials, nil
}

func (r *credentialRepository) Create(ctx context.Context, c *entity.Credential) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[c.ID]; exists {
		return fmt.Errorf("credential already exists: %s", c.ID)
	}
	r.data[c.ID] = c
	return nil
}

func (r *credentialRepository) Update(ctx context.Context, c *entity.Credential) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[c.ID]; !exists {
		return fmt.Errorf("credential not found: %s", c.ID)
	}
	r.data[c.ID] = c
	return nil
}

func (r *credentialRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return fmt.Errorf("credential not found: %s", id)
	}
	delete(r.data, id)
	return nil
}
