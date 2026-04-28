package serializer

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// MasterDataResponse is the API representation of a master data record.
type MasterDataResponse struct {
	ID          string                 `json:"id"`
	Module      string                 `json:"module"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description *string                `json:"description"`
	Icon        *string                `json:"icon"`
	Color       *string                `json:"color"`
	SortOrder   *string                `json:"sort_order"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// FromMasterData converts a domain master data entity to an API response.
func FromMasterData(e entity.MasterData) MasterDataResponse {
	return MasterDataResponse{
		ID:          e.ID,
		Module:      e.Module,
		Type:        e.Type,
		Name:        e.Name,
		Description: stringPtrOrNil(e.Description),
		Icon:        stringPtrOrNil(e.Icon),
		Color:       stringPtrOrNil(e.Color),
		SortOrder:   stringPtrOrNil(e.SortOrder),
		IsActive:    e.IsActive,
		Metadata:    e.Metadata,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// FromMasterDataList converts a slice of master data entities.
func FromMasterDataList(entities []entity.MasterData) []MasterDataResponse {
	result := make([]MasterDataResponse, len(entities))
	for i, e := range entities {
		result[i] = FromMasterData(e)
	}
	return result
}
