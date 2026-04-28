package repository

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

type UserProfilesRepository interface {
	GetByID(ctx context.Context, id string) (*entity.UserProfiles, error)

	GetByUserID(ctx context.Context, userID string) (*entity.UserProfiles, error)

	List(ctx context.Context) ([]*entity.UserProfiles, error)

	Create(ctx context.Context, t *entity.UserProfiles) error

	Update(ctx context.Context, t *entity.UserProfiles) error

	Delete(ctx context.Context, id string) error
}
