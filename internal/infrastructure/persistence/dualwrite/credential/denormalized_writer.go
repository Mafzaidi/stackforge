package credential

import (
	"context"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	"github.com/mafzaidi/stackforge/internal/domain/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const credentialCollection = "credentials"

// DenormalizedWriter orchestrates fetching related data and writing
// denormalized credential documents to MongoDB. It embeds vault, category,
// tags, and user profile data alongside credential fields for efficient reads.
type DenormalizedWriter struct {
	mongoColl       *mongo.Collection
	vaultRepo       repository.VaultRepository
	masterDataRepo  repository.MasterDataRepository
	tagRepo         repository.TagRepository
	userProfileRepo repository.UserProfilesRepository
	logger          service.Logger
}

// NewDenormalizedWriter creates a new DenormalizedWriter with all required dependencies.
// The tagRepo is used to fetch tags internally during denormalized document creation,
// eliminating the need for callers to pass tags explicitly.
func NewDenormalizedWriter(
	mongoClient *mongo.Client,
	dbName string,
	vaultRepo repository.VaultRepository,
	masterDataRepo repository.MasterDataRepository,
	tagRepo repository.TagRepository,
	userProfileRepo repository.UserProfilesRepository,
	logger service.Logger,
) *DenormalizedWriter {
	return &DenormalizedWriter{
		mongoColl:       mongoClient.Database(dbName).Collection(credentialCollection),
		vaultRepo:       vaultRepo,
		masterDataRepo:  masterDataRepo,
		tagRepo:         tagRepo,
		userProfileRepo: userProfileRepo,
		logger:          logger,
	}
}

// CreateDenormalized fetches all related data (vault, category, tags, user profile)
// and writes a denormalized credential document to MongoDB.
// Tags are fetched internally from TagRepository using the credential's UserID and ID.
func (w *DenormalizedWriter) CreateDenormalized(ctx context.Context, cred *entity.Credential) error {
	vault, category, userProfile := w.fetchRelatedData(ctx, cred)
	tags := w.fetchTags(ctx, cred.UserID, cred.ID)

	doc := w.assembleDocument(cred, vault, category, tags, userProfile)

	_, err := w.mongoColl.InsertOne(ctx, doc)
	if err != nil {
		w.logger.Error("denormalized-writer: MongoDB insert failed", service.Fields{
			"error":         err.Error(),
			"credential_id": cred.ID,
			"user_id":       cred.UserID,
		})
		return err
	}
	return nil
}

// UpdateDenormalized fetches the latest related data (including tags) and replaces
// the denormalized document in MongoDB for the given credential.
func (w *DenormalizedWriter) UpdateDenormalized(ctx context.Context, cred *entity.Credential) error {
	vault, category, userProfile := w.fetchRelatedData(ctx, cred)
	tags := w.fetchTags(ctx, cred.UserID, cred.ID)

	doc := w.assembleDocument(cred, vault, category, tags, userProfile)

	filter := bson.M{"credential_id": cred.ID}
	_, err := w.mongoColl.ReplaceOne(ctx, filter, doc)
	if err != nil {
		w.logger.Error("denormalized-writer: MongoDB update failed", service.Fields{
			"error":         err.Error(),
			"credential_id": cred.ID,
			"user_id":       cred.UserID,
		})
		return err
	}
	return nil
}

// Delete removes the denormalized credential document from MongoDB.
func (w *DenormalizedWriter) Delete(ctx context.Context, credentialID string) error {
	filter := bson.M{"credential_id": credentialID}
	_, err := w.mongoColl.DeleteOne(ctx, filter)
	if err != nil {
		w.logger.Error("denormalized-writer: MongoDB delete failed", service.Fields{
			"error":         err.Error(),
			"credential_id": credentialID,
		})
		return err
	}
	return nil
}

// fetchRelatedData retrieves vault, category, and user profile for a credential.
// If any lookup fails or returns nil, the corresponding embed will be nil in the document.
func (w *DenormalizedWriter) fetchRelatedData(
	ctx context.Context,
	cred *entity.Credential,
) (*entity.Vault, *entity.MasterData, *entity.UserProfiles) {
	var vault *entity.Vault
	var category *entity.MasterData
	var userProfile *entity.UserProfiles

	// Fetch vault
	v, err := w.vaultRepo.GetByID(ctx, cred.VaultID)
	if err != nil {
		w.logger.Warn("denormalized-writer: failed to fetch vault", service.Fields{
			"error":    err.Error(),
			"vault_id": cred.VaultID,
		})
	} else {
		vault = v
	}

	// Fetch category
	c, err := w.masterDataRepo.GetByID(ctx, cred.CategoryID)
	if err != nil {
		w.logger.Warn("denormalized-writer: failed to fetch category", service.Fields{
			"error":       err.Error(),
			"category_id": cred.CategoryID,
		})
	} else {
		category = c
	}

	// Fetch user profile
	p, err := w.userProfileRepo.GetByUserID(ctx, cred.UserID)
	if err != nil {
		w.logger.Warn("denormalized-writer: failed to fetch user profile", service.Fields{
			"error":   err.Error(),
			"user_id": cred.UserID,
		})
	} else {
		userProfile = p
	}

	return vault, category, userProfile
}

// fetchTags retrieves tags associated with a credential from TagRepository.
// It filters by UserID, Module="credential", and RefID=credentialID.
// Returns an empty slice if no tags are found or if the fetch fails.
func (w *DenormalizedWriter) fetchTags(ctx context.Context, userID, credentialID string) []*entity.Tag {
	module := "credential"
	filter := repository.TagFilter{
		UserID: &userID,
		Module: &module,
		RefID:  &credentialID,
	}

	tags, err := w.tagRepo.List(ctx, filter)
	if err != nil {
		w.logger.Warn("denormalized-writer: failed to fetch tags", service.Fields{
			"error":         err.Error(),
			"user_id":       userID,
			"credential_id": credentialID,
		})
		return []*entity.Tag{}
	}
	return tags
}

// assembleDocument builds the denormalized MongoDB document from a credential
// and its related data (vault, category, tags, user profile).
func (w *DenormalizedWriter) assembleDocument(
	cred *entity.Credential,
	vault *entity.Vault,
	category *entity.MasterData,
	tags []*entity.Tag,
	userProfile *entity.UserProfiles,
) *denormalizedCredentialDocument {
	return &denormalizedCredentialDocument{
		CredentialID:      cred.ID,
		UserID:            cred.UserID,
		VaultID:           cred.VaultID,
		CategoryID:        cred.CategoryID,
		Title:             cred.Title,
		SiteUrl:           cred.SiteUrl,
		FaviconUrl:        cred.FaviconUrl,
		UsernameEncrypted: cred.UsernameEncrypted,
		PasswordEncrypted: cred.PasswordEncrypted,
		NotesEncrypted:    cred.NotesEncrypted,
		IsFavorite:        cred.IsFavorite,
		PasswordStrength:  cred.PasswordStrength,
		LastUsedAt:        cred.LastUsedAt,
		PasswordChangedAt: cred.PasswordChangedAt,
		ExpiresAt:         cred.ExpiresAt,
		AutoLogin:         cred.AutoLogin,
		CreatedAt:         cred.CreatedAt,
		UpdatedAt:         cred.UpdatedAt,
		Vault:             toVaultEmbed(vault),
		Category:          toCategoryEmbed(category),
		Tags:              toTagEmbeds(tags),
		UserProfile:       toUserProfileEmbed(userProfile),
	}
}
