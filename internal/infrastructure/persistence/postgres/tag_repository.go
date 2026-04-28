package postgres

import (
	"context"
	"errors"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/infrastructure/persistence/postgres/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) repository.TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) GetByID(ctx context.Context, id string) (*entity.Tag, error) {
	var t model.Tag
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&t)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return t.ToEntity(), nil
}

func (r *tagRepository) List(ctx context.Context, filter repository.TagFilter) ([]*entity.Tag, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", filter.UserID)

	if filter.Module != nil {
		query = query.Where("module = ?", *filter.Module)
	}
	if filter.RefID != nil {
		query = query.Where("ref_id = ?", *filter.RefID)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	query = query.Order("name ASC")

	var models []model.Tag
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	tags := make([]*entity.Tag, len(models))
	for i, m := range models {
		tags[i] = m.ToEntity()
	}
	return tags, nil
}

func (r *tagRepository) Create(ctx context.Context, t *entity.Tag) error {
	gormModel := model.TagFromEntity(t)
	return r.db.WithContext(ctx).Create(gormModel).Error
}

func (r *tagRepository) Update(ctx context.Context, t *entity.Tag) error {
	gormModel := model.TagFromEntity(t)
	return r.db.WithContext(ctx).Save(gormModel).Error
}

func (r *tagRepository) BulkUpsert(ctx context.Context, tags []*entity.Tag) error {
	gormModels := make([]model.Tag, len(tags))
	for i, t := range tags {
		gormModels[i] = *model.TagFromEntity(t)
	}
	return r.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			UpdateAll: true,
		},
	).Create(&gormModels).Error
}

func (r *tagRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Tag{}).Error
}
