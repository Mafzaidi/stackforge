package userprofiles

import (
	"context"
	"errors"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
)

type getByUserIDUseCase struct {
	repo   repository.UserProfilesRepository
	logger service.Logger
}

func NewGetByUserIDUseCase(
	repo repository.UserProfilesRepository,
	logger service.Logger,
) GetByUserIDUseCase {
	return &getByUserIDUseCase{repo: repo, logger: logger}
}

func (uc *getByUserIDUseCase) Execute(ctx context.Context, userID string) (*entity.UserProfiles, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	profile, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("failed to fetch user profile", service.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.New("failed to fetch user profile")
	}
	return profile, nil
}
