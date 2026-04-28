package userprofiles

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
)

type createUseCase struct {
	repo   repository.UserProfilesRepository
	logger service.Logger
}

// NewCreateUseCase creates a new instance of CreateUseCase.
func NewCreateUseCase(
	repo repository.UserProfilesRepository,
	logger service.Logger,
) CreateUseCase {
	return &createUseCase{repo: repo, logger: logger}
}

// Execute creates an empty user profile without a master password.
// Called automatically on first login when no profile exists yet.
func (uc *createUseCase) Execute(ctx context.Context, userID string) (*entity.UserProfiles, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	now := time.Now().UTC()
	profile := &entity.UserProfiles{
		ID:        uuid.NewString(),
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
		// MasterPasswordHash, EncryptionKeyHash, PasswordSalt are empty
		// until user sets up master password via SetupMasterPasswordUseCase
	}

	if err := uc.repo.Create(ctx, profile); err != nil {
		uc.logger.Error("failed to persist user profile", service.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.New("failed to create user profile")
	}

	uc.logger.Info("user profile created", service.Fields{
		"user_id":    userID,
		"profile_id": profile.ID,
	})

	return profile, nil
}
