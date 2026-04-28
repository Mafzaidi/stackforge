package vault

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
)

const (
	maxNameLength = 255
	maxDescLength = 255
)

type createUseCase struct {
	repo   repository.VaultRepository
	logger service.Logger
}

func NewCreateUseCase(
	repo repository.VaultRepository,
	logger service.Logger,
) CreateUseCase {
	return &createUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (uc *createUseCase) Execute(ctx context.Context, userID, name, description string) (*entity.Vault, error) {
	if err := validateInput(name, description); err != nil {
		uc.logger.Error("vault creation validation failed", service.Fields{
			"user_id": userID,
			"action":  "CREATE_VAULT",
			"error":   err.Error(),
		})
		return nil, err
	}

	existingVault, err := uc.repo.GetByUserIDAndName(ctx, userID, name)
	if err != nil {
		uc.logger.Error("failed to check existing vault", service.Fields{
			"user_id": userID,
			"action":  "CREATE_VAULT",
			"error":   err.Error(),
		})
		return nil, errors.New("failed to create vault")
	}
	if existingVault != nil {
		return existingVault, errors.New("vault already exists")
	}

	now := time.Now().UTC()
	vault := &entity.Vault{
		ID:          uuid.NewString(),
		UserID:      userID,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.repo.Create(ctx, vault); err != nil {
		uc.logger.Error("failed to persist vault", service.Fields{
			"user_id": userID,
			"action":  "CREATE_VAULT",
			"error":   err.Error(),
		})
		return nil, errors.New("failed to create vault")
	}
	return vault, nil
}

func validateInput(name, description string) error {
	if name == "" {
		return errors.New("name is required")
	}
	if len(name) > maxNameLength {
		return errors.New("name exceeds maximum length of 255 characters")
	}
	if len(description) > maxDescLength {
		return errors.New("description exceeds maximum length of 255 characters")
	}
	return nil
}
