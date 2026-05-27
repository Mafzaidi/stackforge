package credential

import (
	"context"
	"errors"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
	"golang.org/x/crypto/bcrypt"
)

// DecryptedCredentialFields holds the decrypted sensitive fields of a credential.
type DecryptedCredentialFields struct {
	ID       string
	Username string
	Password string
	Notes    string
}

type viewUseCase struct {
	repo            repository.CredentialRepository
	userProfileRepo repository.UserProfilesRepository
	encryptor       Encryptor
	logger          service.Logger
}

// NewViewUseCase creates a new instance of ViewUseCase.
func NewViewUseCase(
	repo repository.CredentialRepository,
	userProfileRepo repository.UserProfilesRepository,
	encryptor Encryptor,
	logger service.Logger,
) ViewUseCase {
	return &viewUseCase{
		repo:            repo,
		userProfileRepo: userProfileRepo,
		encryptor:       encryptor,
		logger:          logger,
	}
}

// Execute verifies the master password, retrieves the credential, and decrypts the sensitive fields.
func (uc *viewUseCase) Execute(ctx context.Context, userID, credentialID, masterPassword string) (*DecryptedCredentialFields, *entity.Credential, error) {
	if credentialID == "" {
		return nil, nil, errors.New("credential ID is required")
	}
	if masterPassword == "" {
		return nil, nil, errors.New("master password is required")
	}

	// Get user profile to verify master password
	profile, err := uc.userProfileRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("user profile not found", service.Fields{
			"user_id": userID,
			"action":  "VIEW_CREDENTIAL",
			"error":   err.Error(),
		})
		return nil, nil, errors.New("user profile not found")
	}

	if profile.MasterPasswordHash == nil {
		return nil, nil, errors.New("master password not set up yet")
	}

	// Verify master password against stored bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(*profile.MasterPasswordHash), []byte(masterPassword)); err != nil {
		uc.logger.Error("invalid master password", service.Fields{
			"user_id": userID,
			"action":  "VIEW_CREDENTIAL",
		})
		return nil, nil, errors.New("invalid master password")
	}

	// Fetch the credential
	cred, err := uc.repo.GetByID(ctx, credentialID)
	if err != nil {
		uc.logger.Error("credential not found", service.Fields{
			"user_id":       userID,
			"credential_id": credentialID,
			"action":        "VIEW_CREDENTIAL",
			"error":         err.Error(),
		})
		return nil, nil, errors.New("credential not found")
	}

	// Verify ownership
	if cred.UserID != userID {
		return nil, nil, errors.New("unauthorized access to credential")
	}

	// Derive decryption key from stored master password hash
	key := deriveKeyFromHash(*profile.MasterPasswordHash)

	// Decrypt username
	decryptedUsername, err := uc.encryptor.DecryptWithKey(cred.UsernameEncrypted, key)
	if err != nil {
		uc.logger.Error("failed to decrypt username", service.Fields{
			"user_id":       userID,
			"credential_id": credentialID,
			"action":        "VIEW_CREDENTIAL",
			"error":         err.Error(),
		})
		return nil, nil, errors.New("failed to decrypt credential")
	}

	// Decrypt password
	decryptedPassword, err := uc.encryptor.DecryptWithKey(cred.PasswordEncrypted, key)
	if err != nil {
		uc.logger.Error("failed to decrypt password", service.Fields{
			"user_id":       userID,
			"credential_id": credentialID,
			"action":        "VIEW_CREDENTIAL",
			"error":         err.Error(),
		})
		return nil, nil, errors.New("failed to decrypt credential")
	}

	// Decrypt notes if present
	var decryptedNotes string
	if cred.NotesEncrypted != nil && *cred.NotesEncrypted != "" {
		decryptedNotes, err = uc.encryptor.DecryptWithKey(*cred.NotesEncrypted, key)
		if err != nil {
			uc.logger.Error("failed to decrypt notes", service.Fields{
				"user_id":       userID,
				"credential_id": credentialID,
				"action":        "VIEW_CREDENTIAL",
				"error":         err.Error(),
			})
			return nil, nil, errors.New("failed to decrypt credential")
		}
	}

	uc.logger.Info("credential decrypted successfully", service.Fields{
		"user_id":       userID,
		"credential_id": credentialID,
		"action":        "VIEW_CREDENTIAL",
	})

	decrypted := &DecryptedCredentialFields{
		ID:       cred.ID,
		Username: decryptedUsername,
		Password: decryptedPassword,
		Notes:    decryptedNotes,
	}

	return decrypted, cred, nil
}
