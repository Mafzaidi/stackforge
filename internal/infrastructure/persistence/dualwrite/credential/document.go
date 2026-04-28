package credential

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// denormalizedCredentialDocument is the MongoDB document structure
// for a credential with embedded related data.
type denormalizedCredentialDocument struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	CredentialID      string             `bson:"credential_id"`
	UserID            string             `bson:"user_id"`
	VaultID           string             `bson:"vault_id"`
	CategoryID        string             `bson:"category_id"`
	Title             string             `bson:"title"`
	SiteUrl           string             `bson:"site_url"`
	FaviconUrl        *string            `bson:"favicon_url,omitempty"`
	UsernameEncrypted string             `bson:"username_encrypted"`
	PasswordEncrypted string             `bson:"password_encrypted"`
	NotesEncrypted    *string            `bson:"notes_encrypted,omitempty"`
	IsFavorite        bool               `bson:"is_favorite"`
	PasswordStrength  int64              `bson:"password_strength"`
	LastUsedAt        *time.Time         `bson:"last_used_at,omitempty"`
	PasswordChangedAt *time.Time         `bson:"password_changed_at,omitempty"`
	ExpiresAt         *time.Time         `bson:"expires_at,omitempty"`
	AutoLogin         bool               `bson:"auto_login"`
	CreatedAt         time.Time          `bson:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at"`

	// Embedded documents
	Vault       *vaultEmbed       `bson:"vault,omitempty"`
	Category    *categoryEmbed    `bson:"category,omitempty"`
	Tags        []tagEmbed        `bson:"tags"`
	UserProfile *userProfileEmbed `bson:"user_profile,omitempty"`
}

type vaultEmbed struct {
	ID          string    `bson:"id"`
	UserID      string    `bson:"user_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Icon        string    `bson:"icon"`
	IsDefault   bool      `bson:"is_default"`
	CreatedAt   time.Time `bson:"created_at"`
}

type categoryEmbed struct {
	ID          string                 `bson:"id"`
	Module      string                 `bson:"module"`
	Type        string                 `bson:"type"`
	Name        string                 `bson:"name"`
	Description string                 `bson:"description"`
	Icon        string                 `bson:"icon"`
	Color       string                 `bson:"color"`
	SortOrder   string                 `bson:"sort_order"`
	IsActive    bool                   `bson:"is_active"`
	Metadata    map[string]interface{} `bson:"metadata,omitempty"`
	CreatedAt   time.Time              `bson:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at"`
}

type tagEmbed struct {
	ID        string    `bson:"id"`
	UserID    string    `bson:"user_id"`
	Name      string    `bson:"name"`
	Color     *string   `bson:"color,omitempty"`
	Module    string    `bson:"module"`
	RefID     string    `bson:"ref_id"`
	CreatedAt time.Time `bson:"created_at"`
}

type userProfileEmbed struct {
	ID                 string    `bson:"id"`
	UserID             string    `bson:"user_id"`
	EncryptionKeyHash  *string   `bson:"encryption_key_hash,omitempty"`
	MasterPasswordHash *string   `bson:"master_password_hash,omitempty"`
	PasswordSalt       *string   `bson:"password_salt,omitempty"`
	CreatedAt          time.Time `bson:"created_at"`
	UpdatedAt          time.Time `bson:"updated_at"`
}

// toVaultEmbed converts a Vault entity to its embedded document representation.
func toVaultEmbed(v *entity.Vault) *vaultEmbed {
	if v == nil {
		return nil
	}
	return &vaultEmbed{
		ID:          v.ID,
		UserID:      v.UserID,
		Name:        v.Name,
		Description: v.Description,
		Icon:        v.Icon,
		IsDefault:   v.IsDefault,
		CreatedAt:   v.CreatedAt,
	}
}

// toCategoryEmbed converts a MasterData entity to its embedded document representation.
func toCategoryEmbed(m *entity.MasterData) *categoryEmbed {
	if m == nil {
		return nil
	}
	return &categoryEmbed{
		ID:          m.ID,
		Module:      m.Module,
		Type:        m.Type,
		Name:        m.Name,
		Description: m.Description,
		Icon:        m.Icon,
		Color:       m.Color,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		Metadata:    m.Metadata,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// toTagEmbeds converts a slice of Tag entities to their embedded document representations.
// Returns an empty slice (not nil) when the input is empty or nil.
func toTagEmbeds(tags []*entity.Tag) []tagEmbed {
	if len(tags) == 0 {
		return []tagEmbed{}
	}
	embeds := make([]tagEmbed, len(tags))
	for i, t := range tags {
		embeds[i] = tagEmbed{
			ID:        t.ID,
			UserID:    t.UserID,
			Name:      t.Name,
			Color:     t.Color,
			Module:    t.Module,
			RefID:     t.RefID,
			CreatedAt: t.CreatedAt,
		}
	}
	return embeds
}

// toUserProfileEmbed converts a UserProfiles entity to its embedded document representation.
func toUserProfileEmbed(p *entity.UserProfiles) *userProfileEmbed {
	if p == nil {
		return nil
	}
	return &userProfileEmbed{
		ID:                 p.ID,
		UserID:             p.UserID,
		EncryptionKeyHash:  p.EncryptionKeyHash,
		MasterPasswordHash: p.MasterPasswordHash,
		PasswordSalt:       p.PasswordSalt,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}
