package userprofiles

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
	"golang.org/x/crypto/bcrypt"
)

type setupMasterPasswordUseCase struct {
	repo   repository.UserProfilesRepository
	logger service.Logger
}

// NewSetupMasterPasswordUseCase creates a new instance of SetupMasterPasswordUseCase.
func NewSetupMasterPasswordUseCase(
	repo repository.UserProfilesRepository,
	logger service.Logger,
) SetupMasterPasswordUseCase {
	return &setupMasterPasswordUseCase{repo: repo, logger: logger}
}

// Execute sets the master password on an existing user profile.
// Called when the user creates their first credential.
func (uc *setupMasterPasswordUseCase) Execute(ctx context.Context, userID string, masterPassword string) error {
	if userID == "" {
		return errors.New("user_id is required")
	}
	if masterPassword == "" {
		return errors.New("master_password is required")
	}

	profile, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("user profile not found", service.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
		return errors.New("user profile not found")
	}

	// Prevent overwriting an already-configured master password
	if profile.MasterPasswordHash != nil {
		return errors.New("master password already set")
	}

	salt, err := generateSalt()
	if err != nil {
		uc.logger.Error("failed to generate salt", service.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
		return errors.New("failed to generate password salt")
	}

	passwordHash, err := hashPassword(masterPassword)
	if err != nil {
		uc.logger.Error("failed to hash master password", service.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
		return errors.New("failed to process master password")
	}

	profile.MasterPasswordHash = &passwordHash
	profile.EncryptionKeyHash = deriveEncryptionKeyHash(masterPassword, salt)
	profile.PasswordSalt = &salt
	profile.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, profile); err != nil {
		uc.logger.Error("failed to update user profile", service.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
		return errors.New("failed to save master password")
	}

	uc.logger.Info("master password configured", service.Fields{
		"user_id":    userID,
		"profile_id": profile.ID,
	})

	return nil
}

func generateSalt() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func deriveEncryptionKeyHash(password, salt string) *string {
	h := sha256.Sum256([]byte("enc:" + password + salt))
	result := base64.StdEncoding.EncodeToString(h[:])
	return &result
}
