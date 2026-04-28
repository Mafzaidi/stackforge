package postgres

import (
	"context"
	"errors"
	"math"

	"gorm.io/gorm"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/infrastructure/persistence/postgres/model"
)

// credentialRepository implements repository.CredentialRepository using PostgreSQL via GORM.
type credentialRepository struct {
	db *gorm.DB
}

// NewCredentialRepository creates a new PostgreSQL credential repository.
// Accepts a GORM database client from ConnectionManager.
// Requirements: 4.1, 4.5, 4.6, 4.7, 4.8, 4.9, 4.10
func NewCredentialRepository(db *gorm.DB) repository.CredentialRepository {
	return &credentialRepository{
		db: db,
	}
}

func (r *credentialRepository) GetByID(ctx context.Context, id string) (*entity.Credential, error) {
	var m model.Credential
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return m.ToEntity(), nil
}

func (r *credentialRepository) List(ctx context.Context, limit, offset int) ([]*entity.Credential, error) {
	var models []model.Credential
	result := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	credentials := make([]*entity.Credential, len(models))
	for i, m := range models {
		credentials[i] = m.ToEntity()
	}
	return credentials, nil
}

func (r *credentialRepository) ListByUserID(ctx context.Context, userID string, page, limit int) (*repository.PaginatedCredentials, error) {
	offset := (page - 1) * limit

	var models []model.Credential
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	items := make([]*entity.Credential, len(models))
	for i, m := range models {
		items[i] = m.ToEntity()
	}

	totalItems, err := r.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	return &repository.PaginatedCredentials{
		Items:        items,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		CurrentPage:  page,
		ItemsPerPage: limit,
	}, nil
}

func (r *credentialRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&model.Credential{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (r *credentialRepository) Create(ctx context.Context, c *entity.Credential) error {
	m := model.CredentialFromEntity(c)
	return r.db.WithContext(ctx).Create(m).Error
}
func (r *credentialRepository) Update(ctx context.Context, c *entity.Credential) error {
	m := model.CredentialFromEntity(c)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *credentialRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Credential{}).Error
}
