package dualwrite

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
)

type userProfilesRepository struct {
	primary   repository.UserProfilesRepository
	secondary repository.UserProfilesRepository
	logger    service.Logger
}

func NewUserProfilesRepository(primary, secondary repository.UserProfilesRepository, logger service.Logger) repository.UserProfilesRepository {
	return &userProfilesRepository{primary: primary, secondary: secondary, logger: logger}
}

func (r *userProfilesRepository) GetByID(ctx context.Context, id string) (*entity.UserProfiles, error) {
	return r.primary.GetByID(ctx, id)
}

func (r *userProfilesRepository) GetByUserID(ctx context.Context, userID string) (*entity.UserProfiles, error) {
	return r.primary.GetByUserID(ctx, userID)
}

func (r *userProfilesRepository) List(ctx context.Context) ([]*entity.UserProfiles, error) {
	return r.primary.List(ctx)
}

func (r *userProfilesRepository) Create(ctx context.Context, p *entity.UserProfiles) error {
	if err := r.primary.Create(ctx, p); err != nil {
		return err
	}
	if err := r.secondary.Create(ctx, p); err != nil {
		r.logger.Error("dual-write: secondary user_profiles create failed", service.Fields{"error": err.Error(), "id": p.ID})
	}
	return nil
}

func (r *userProfilesRepository) Update(ctx context.Context, p *entity.UserProfiles) error {
	if err := r.primary.Update(ctx, p); err != nil {
		return err
	}
	if err := r.secondary.Update(ctx, p); err != nil {
		r.logger.Error("dual-write: secondary user_profiles update failed", service.Fields{"error": err.Error(), "id": p.ID})
	}
	return nil
}

func (r *userProfilesRepository) Delete(ctx context.Context, id string) error {
	if err := r.primary.Delete(ctx, id); err != nil {
		return err
	}
	if err := r.secondary.Delete(ctx, id); err != nil {
		r.logger.Error("dual-write: secondary user_profiles delete failed", service.Fields{"error": err.Error(), "id": id})
	}
	return nil
}
