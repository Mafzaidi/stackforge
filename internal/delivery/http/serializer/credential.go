package serializer

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// CredentialResponse is the API representation of a credential.
// Sensitive fields (encrypted passwords, notes) are deliberately excluded.
type CredentialResponse struct {
	ID               string              `json:"id"`
	UserID           string              `json:"user_id"`
	VaultID          string              `json:"vault_id,omitempty"`
	CategoryID       string              `json:"category_id,omitempty"`
	Title            string              `json:"title"`
	SiteUrl          string              `json:"site_url,omitempty"`
	FaviconUrl       *string             `json:"favicon_url,omitempty"`
	IsFavorite       bool                `json:"is_favorite"`
	PasswordStrength int64               `json:"password_strength"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	Vault            *VaultResponse      `json:"vault,omitempty"`
	Category         *MasterDataResponse `json:"category,omitempty"`
	Tags             []Tag               `json:"tags,omitempty"`
}

// FromCredential converts a domain credential entity to an API response.
func FromCredential(e *entity.Credential) CredentialResponse {
	var vault *VaultResponse
	if e.Vault != nil {
		v := FromVault(e.Vault)
		vault = &v
	}

	var category *MasterDataResponse
	if e.Category != nil {
		c := FromMasterData(*e.Category)
		category = &c
	}

	return CredentialResponse{
		ID:               e.ID,
		UserID:           e.UserID,
		VaultID:          e.VaultID,
		CategoryID:       e.CategoryID,
		Title:            e.Title,
		SiteUrl:          e.SiteUrl,
		FaviconUrl:       e.FaviconUrl,
		IsFavorite:       e.IsFavorite,
		PasswordStrength: e.PasswordStrength,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
		Vault:            vault,
		Category:         category,
		Tags:             FromTagList(e.Tags),
	}
}

// FromCredentialList converts a slice of credential entities.
func FromCredentialList(entities []*entity.Credential) []CredentialResponse {
	result := make([]CredentialResponse, len(entities))
	for i, e := range entities {
		result[i] = FromCredential(e)
	}
	return result
}
