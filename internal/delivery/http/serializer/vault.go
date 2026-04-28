package serializer

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// VaultResponse is the API representation of a vault.
type VaultResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Icon        *string   `json:"icon"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FromVault converts a domain vault entity to an API response.
func FromVault(e *entity.Vault) VaultResponse {
	return VaultResponse{
		ID:          e.ID,
		UserID:      e.UserID,
		Name:        e.Name,
		Description: stringPtrOrNil(e.Description),
		Icon:        stringPtrOrNil(e.Icon),
		IsDefault:   e.IsDefault,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// FromVaultList converts a slice of vault entities.
func FromVaultList(entities []*entity.Vault) []VaultResponse {
	result := make([]VaultResponse, len(entities))
	for i, e := range entities {
		result[i] = FromVault(e)
	}
	return result
}
