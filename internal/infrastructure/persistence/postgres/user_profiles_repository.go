package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/infrastructure/persistence/postgres/model"
)

// userProfilesRepository implements repository.UserProfilesRepository using PostgreSQL via GORM.
type userProfilesRepository struct {
	db *gorm.DB
}

// NewUserProfilesRepository creates a new PostgreSQL user profiles repository.
// Accepts a GORM database client from ConnectionManager.
// Requirements: 4.3, 4.5, 4.6, 4.7, 4.8, 4.9
func NewUserProfilesRepository(db *gorm.DB) repository.UserProfilesRepository {
	return &userProfilesRepository{
		db: db,
	}
}

func (r *userProfilesRepository) GetByID(ctx context.Context, id string) (*entity.UserProfiles, error) {
	var m model.UserProfiles
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return m.ToEntity(), nil
}

func (r *userProfilesRepository) GetByUserID(ctx context.Context, userID string) (*entity.UserProfiles, error) {
	var m model.UserProfiles
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return m.ToEntity(), nil
}

func (r *userProfilesRepository) List(ctx context.Context) ([]*entity.UserProfiles, error) {
	var models []model.UserProfiles
	result := r.db.WithContext(ctx).Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	profiles := make([]*entity.UserProfiles, len(models))
	for i, m := range models {
		profiles[i] = m.ToEntity()
	}
	return profiles, nil
}

func (r *userProfilesRepository) Create(ctx context.Context, p *entity.UserProfiles) error {
	m := model.UserProfilesFromEntity(p)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *userProfilesRepository) Update(ctx context.Context, p *entity.UserProfiles) error {
	m := model.UserProfilesFromEntity(p)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *userProfilesRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.UserProfiles{}).Error
}
