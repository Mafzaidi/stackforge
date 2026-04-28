package credential

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
)

// Repository implements repository.CredentialRepository with dual-write behavior.
// It writes to PostgreSQL as the primary store and delegates denormalized document
// creation to DenormalizedWriter for the MongoDB secondary store.
// Secondary write failures are logged but do not fail the operation (fire-and-forget).
type Repository struct {
	primary            repository.CredentialRepository
	secondary          repository.CredentialRepository // Optional: for read operations if needed
	denormalizedWriter *DenormalizedWriter
	logger             service.Logger
}

// NewRepository creates a new dual-write credential Repository.
// The primary repository handles PostgreSQL persistence, while the denormalizedWriter
// handles MongoDB denormalized document writes as a fire-and-forget secondary.
func NewRepository(
	primary repository.CredentialRepository,
	secondary repository.CredentialRepository,
	denormalizedWriter *DenormalizedWriter,
	logger service.Logger,
) repository.CredentialRepository {
	return &Repository{primary: primary, secondary: secondary, denormalizedWriter: denormalizedWriter, logger: logger}
}

// GetByID retrieves a credential by its ID from the primary store.
func (r *Repository) GetByID(ctx context.Context, id string) (*entity.Credential, error) {
	return r.primary.GetByID(ctx, id)
}

// List retrieves all credentials from the primary store.
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*entity.Credential, error) {
	return r.secondary.List(ctx, limit, offset)
}

// ListByUserID retrieves credentials for a specific user with pagination from the primary store.
func (r *Repository) ListByUserID(ctx context.Context, userID string, page, limit int) (*repository.PaginatedCredentials, error) {
	return r.primary.ListByUserID(ctx, userID, page, limit)
}

// CountByUserID counts the total number of credentials for a specific user in the primary store.
func (r *Repository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	return r.primary.CountByUserID(ctx, userID)
}

// Create writes the credential to PostgreSQL (primary), then writes a denormalized
// document to MongoDB (secondary) via DenormalizedWriter. Tags are fetched internally
// by the writer from TagRepository. If the secondary write fails, the error is logged
// but the operation returns nil (fire-and-forget).
func (r *Repository) Create(ctx context.Context, c *entity.Credential) error {
	if err := r.primary.Create(ctx, c); err != nil {
		return err
	}
	if err := r.denormalizedWriter.CreateDenormalized(ctx, c); err != nil {
		r.logger.Error("dual-write: denormalized credential create failed", service.Fields{"error": err.Error(), "id": c.ID})
	}
	return nil
}

// Update writes changes to PostgreSQL (primary), then replaces the denormalized
// document in MongoDB (secondary). Secondary write failures are logged but do not
// fail the operation.
func (r *Repository) Update(ctx context.Context, c *entity.Credential) error {
	if err := r.primary.Update(ctx, c); err != nil {
		return err
	}
	if err := r.denormalizedWriter.UpdateDenormalized(ctx, c); err != nil {
		r.logger.Error("dual-write: denormalized credential update failed", service.Fields{"error": err.Error(), "id": c.ID})
	}
	return nil
}

// Delete removes the credential from PostgreSQL (primary), then removes the
// denormalized document from MongoDB (secondary). Secondary write failures are
// logged but do not fail the operation.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.primary.Delete(ctx, id); err != nil {
		return err
	}
	if err := r.denormalizedWriter.Delete(ctx, id); err != nil {
		r.logger.Error("dual-write: denormalized credential delete failed", service.Fields{"error": err.Error(), "id": id})
	}
	return nil
}
