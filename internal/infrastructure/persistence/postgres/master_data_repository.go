package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/infrastructure/persistence/postgres/model"
)

type masterDataRepository struct {
	db *gorm.DB
}

func NewMasterDataRepository(db *gorm.DB) repository.MasterDataRepository {
	return &masterDataRepository{db: db}
}

func (r *masterDataRepository) GetByID(ctx context.Context, id string) (*entity.MasterData, error) {
	var m model.MasterData
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return m.ToEntity(), nil
}

func (r *masterDataRepository) List(ctx context.Context, filter repository.ListFilter) ([]*entity.MasterData, error) {
	query := r.db.WithContext(ctx)

	if filter.Module != nil {
		query = query.Where("module = ?", *filter.Module)
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Name != nil {
		query = query.Where("name = ?", *filter.Name)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	query = query.Order("sort_order ASC, name ASC")

	var models []model.MasterData
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	md := make([]*entity.MasterData, len(models))
	for i, m := range models {
		md[i] = m.ToEntity()
	}
	return md, nil
}

func (r *masterDataRepository) Create(ctx context.Context, m *entity.MasterData) error {
	gormModel := model.MasterDataFromEntity(m)
	return r.db.WithContext(ctx).Create(gormModel).Error
}

func (r *masterDataRepository) Update(ctx context.Context, m *entity.MasterData) error {
	gormModel := model.MasterDataFromEntity(m)
	return r.db.WithContext(ctx).Save(gormModel).Error
}

func (r *masterDataRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.MasterData{}).Error
}
