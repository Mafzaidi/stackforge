package memory

import (
	"context"
	"fmt"
	"math"
	"sort"
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

func (r *credentialRepository) List(ctx context.Context, limit, offset int) ([]*entity.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	credentials := make([]*entity.Credential, 0, len(r.data))
	for _, credential := range r.data {
		credentials = append(credentials, credential)
	}
	return credentials, nil
}

func (r *credentialRepository) ListByUserID(ctx context.Context, userID string, page, limit int) (*repository.PaginatedCredentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Filter by user_id
	var userCreds []*entity.Credential
	for _, cred := range r.data {
		if cred.UserID == userID {
			userCreds = append(userCreds, cred)
		}
	}

	// Sort by CreatedAt descending
	sort.Slice(userCreds, func(i, j int) bool {
		return userCreds[i].CreatedAt.After(userCreds[j].CreatedAt)
	})

	totalItems := int64(len(userCreds))
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	// Apply pagination
	offset := (page - 1) * limit
	end := offset + limit
	if offset > len(userCreds) {
		offset = len(userCreds)
	}
	if end > len(userCreds) {
		end = len(userCreds)
	}
	paged := userCreds[offset:end]

	return &repository.PaginatedCredentials{
		Items:        paged,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		CurrentPage:  page,
		ItemsPerPage: limit,
	}, nil
}

func (r *credentialRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, cred := range r.data {
		if cred.UserID == userID {
			count++
		}
	}
	return count, nil
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
