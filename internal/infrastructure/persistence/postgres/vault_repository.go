package postgres

import (
	"context"
	"errors"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/infrastructure/persistence/postgres/model"
	"gorm.io/gorm"
)

type vaultRepository struct {
	db *gorm.DB
}

func NewVaultRepository(db *gorm.DB) repository.VaultRepository {
	return &vaultRepository{
		db: db,
	}
}
func (r *vaultRepository) GetByID(ctx context.Context, id string) (*entity.Vault, error) {
	var m model.Vault
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return m.ToEntity(), nil
}

func (r *vaultRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*entity.Vault, error) {
	var m model.Vault
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("name = ?", name).
		First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return m.ToEntity(), nil
}

func (r *vaultRepository) List(ctx context.Context, userID string) ([]*entity.Vault, error) {
	var models []model.Vault
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	vaults := make([]*entity.Vault, len(models))
	for i, m := range models {
		vaults[i] = m.ToEntity()
	}
	return vaults, nil
}

func (r *vaultRepository) Create(ctx context.Context, v *entity.Vault) error {
	gormModel := model.VaultFromEntity(v)
	return r.db.WithContext(ctx).Create(gormModel).Error
}
