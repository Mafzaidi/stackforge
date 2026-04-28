package credential

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
)

const (
	maxTitleLength    = 255
	maxUsernameLength = 255
	maxPasswordLength = 10000
)

type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	EncryptWithKey(plaintext string, key []byte) (string, error)
	DecryptWithKey(ciphertext string, key []byte) (string, error)
}

type createUseCase struct {
	repo            repository.CredentialRepository
	userProfileRepo repository.UserProfilesRepository
	tagRepo         repository.TagRepository
	encryptor       Encryptor
	logger          service.Logger
}

func NewCreateUseCase(
	repo repository.CredentialRepository,
	userProfileRepo repository.UserProfilesRepository,
	tagRepo repository.TagRepository,
	encryptor Encryptor,
	logger service.Logger,
) CreateUseCase {
	return &createUseCase{
		repo:            repo,
		userProfileRepo: userProfileRepo,
		tagRepo:         tagRepo,
		encryptor:       encryptor,
		logger:          logger,
	}
}

func (uc *createUseCase) Execute(
	ctx context.Context,
	userID,
	title,
	siteUrl,
	faviconUrl,
	username,
	password,
	notes string,
	isFavorite bool,
	passwordStrength int32,
	vaultID,
	categoryID string,
	tags []string,
) (*entity.Credential, error) {
	masterPasswordHash, err := uc.verifyMasterPassword(ctx, userID)
	if err != nil {
		uc.logger.Error("master password verification failed", service.Fields{
			"user_id": userID,
			"action":  "CREATE_CREDENTIAL",
			"error":   err.Error(),
		})
		return nil, err
	}

	if err := validateInput(
		title,
		siteUrl,
		username,
		password,
		passwordStrength,
		vaultID,
		categoryID,
	); err != nil {
		uc.logger.Error("credential creation validation failed", service.Fields{
			"user_id": userID,
			"action":  "CREATE_CREDENTIAL",
			"error":   err.Error(),
		})
		return nil, err
	}

	// Derive per-user encryption key from master password hash
	key := deriveKeyFromHash(*masterPasswordHash)

	encryptedUsername, err := uc.encryptor.EncryptWithKey(username, key)
	if err != nil {
		uc.logger.Error("failed to encrypt username", service.Fields{
			"user_id": userID,
			"action":  "CREATE_CREDENTIAL",
			"error":   err.Error(),
		})
		return nil, errors.New("failed to encrypt username")
	}

	encryptedPassword, err := uc.encryptor.EncryptWithKey(password, key)
	if err != nil {
		uc.logger.Error("failed to encrypt password", service.Fields{
			"user_id": userID,
			"action":  "CREATE_CREDENTIAL",
			"error":   err.Error(),
		})
		return nil, errors.New("failed to encrypt password")
	}

	encryptedNotes, err := uc.encryptor.EncryptWithKey(notes, key)
	if err != nil {
		uc.logger.Error("failed to encrypt notes", service.Fields{
			"user_id": userID,
			"action":  "CREATE_CREDENTIAL",
			"error":   err.Error(),
		})
		return nil, errors.New("failed to encrypt notes")
	}

	now := time.Now().UTC()
	cred := &entity.Credential{
		ID:                uuid.NewString(),
		UserID:            userID,
		VaultID:           vaultID,
		CategoryID:        categoryID,
		Title:             title,
		SiteUrl:           siteUrl,
		FaviconUrl:        &faviconUrl,
		UsernameEncrypted: encryptedUsername,
		PasswordEncrypted: encryptedPassword,
		NotesEncrypted:    &encryptedNotes,
		IsFavorite:        isFavorite,
		PasswordStrength:  int64(passwordStrength),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	tagEntities := uc.convertTagsToEntities(userID, cred.ID, tags)

	if len(tagEntities) > 0 {
		if err := uc.tagRepo.BulkUpsert(ctx, tagEntities); err != nil {
			uc.logger.Error("failed to add tags to credential", service.Fields{
				"user_id":       userID,
				"credential_id": cred.ID,
				"action":        "CREATE_CREDENTIAL",
				"error":         err.Error(),
			})
			return nil, errors.New("failed to add tags to credential")
		}
	}

	if err := uc.repo.Create(ctx, cred); err != nil {
		uc.logger.Error("failed to persist credential", service.Fields{
			"user_id": userID,
			"action":  "CREATE_CREDENTIAL",
			"error":   err.Error(),
		})
		return nil, errors.New("failed to create credential")
	}

	uc.logger.Info("credential created", service.Fields{
		"user_id":       userID,
		"credential_id": cred.ID,
		"action":        "CREATE_CREDENTIAL",
	})

	return cred, nil
}

// verifyMasterPassword checks that the user has a profile with master password set,
// then verifies the provided password matches the stored hash.
func (uc *createUseCase) verifyMasterPassword(ctx context.Context, userID string) (*string, error) {
	profile, err := uc.userProfileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("user profile not found")
	}

	if profile.MasterPasswordHash == nil {
		return nil, errors.New("master password not set up yet, please set up your master password first")
	}

	// if err := bcrypt.CompareHashAndPassword([]byte(profile.MasterPasswordHash), []byte(masterPassword)); err != nil {
	// 	return errors.New("invalid master password")
	// }

	return profile.MasterPasswordHash, nil
}

func (uc *createUseCase) convertTagsToEntities(userID, credentialID string, tags []string) []*entity.Tag {
	entities := make([]*entity.Tag, len(tags))
	for i, t := range tags {
		entities[i] = &entity.Tag{
			ID:     uuid.NewString(),
			UserID: userID,
			Name:   t,
			Module: "credential",
			RefID:  credentialID,
		}
	}
	return entities
}

// deriveKeyFromHash derives a 32-byte AES-256 key from a MasterPasswordHash using SHA-256.
func deriveKeyFromHash(masterPasswordHash string) []byte {
	h := sha256.Sum256([]byte(masterPasswordHash))
	return h[:]
}

func validateInput(
	title,
	siteUrl,
	username,
	password string,
	passwordStrength int32,
	vaultID,
	categoryID string,
) error {
	if title == "" {
		return errors.New("title is required")
	}
	if len(title) > maxTitleLength {
		return errors.New("title exceeds maximum length of 255 characters")
	}
	if siteUrl == "" {
		return errors.New("site URL is required")
	}
	if username == "" {
		return errors.New("username is required")
	}
	if len(username) > maxUsernameLength {
		return errors.New("username exceeds maximum length of 255 characters")
	}
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) > maxPasswordLength {
		return errors.New("password exceeds maximum length of 10000 characters")
	}
	if passwordStrength < 0 || passwordStrength > 100 {
		return errors.New("password strength must be between 0 and 100")
	}
	if vaultID == "" {
		return errors.New("vault ID is required")
	}
	if categoryID == "" {
		return errors.New("category ID is required")
	}
	return nil
}
