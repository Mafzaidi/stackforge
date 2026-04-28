package repository

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// PaginatedCredentials represents paginated credential results.
type PaginatedCredentials struct {
	Items        []*entity.Credential
	TotalItems   int64
	TotalPages   int
	CurrentPage  int
	ItemsPerPage int
}

// CredentialRepository defines the interface for credential persistence operations.
// This interface is implementation-agnostic and can be implemented by
// in-memory, PostgreSQL, MongoDB, or any other storage mechanism.
type CredentialRepository interface {
	// GetByID retrieves a credential by its ID.
	// Returns nil if the credential is not found.
	GetByID(ctx context.Context, id string) (*entity.Credential, error)

	// List retrieves all credentials.
	List(ctx context.Context, limit, offset int) ([]*entity.Credential, error)

	// ListByUserID retrieves credentials for a specific user with pagination.
	// Results are sorted by CreatedAt descending.
	ListByUserID(ctx context.Context, userID string, page, limit int) (*PaginatedCredentials, error)

	// CountByUserID counts the total number of credentials for a specific user.
	CountByUserID(ctx context.Context, userID string) (int64, error)

	// Create persists a new credential.
	Create(ctx context.Context, c *entity.Credential) error

	// Update persists changes to an existing credential.
	Update(ctx context.Context, c *entity.Credential) error

	// Delete removes a credential by its ID.
	Delete(ctx context.Context, id string) error
}
