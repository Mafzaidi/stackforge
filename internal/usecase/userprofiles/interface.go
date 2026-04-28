package userprofiles

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// CreateUseCase creates an empty user profile on first login (no master password needed).
type CreateUseCase interface {
	Execute(ctx context.Context, userID string) (*entity.UserProfiles, error)
}

type GetByUserIDUseCase interface {
	Execute(ctx context.Context, userID string) (*entity.UserProfiles, error)
}

// SetupMasterPasswordUseCase sets the master password on an existing profile
// when the user creates their first credential.
type SetupMasterPasswordUseCase interface {
	Execute(ctx context.Context, userID, masterPassword string) error
}
