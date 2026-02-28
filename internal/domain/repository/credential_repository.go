package repository

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// CredentialRepository defines the interface for credential persistence operations.
// This interface is implementation-agnostic and can be implemented by
// in-memory, PostgreSQL, or any other storage mechanism.
type CredentialRepository interface {
	// GetByID retrieves a credential by its ID.
	// Returns nil if the credential is not found.
	GetByID(ctx context.Context, id string) (*entity.Credential, error)

	// List retrieves all credentials.
	List(ctx context.Context) ([]*entity.Credential, error)

	// Create persists a new credential.
	Create(ctx context.Context, c *entity.Credential) error

	// Update persists changes to an existing credential.
	Update(ctx context.Context, c *entity.Credential) error

	// Delete removes a credential by its ID.
	Delete(ctx context.Context, id string) error
}
