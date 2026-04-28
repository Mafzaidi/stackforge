package model

import (
	"encoding/json"
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// MasterData is the GORM model for the stackforge_service.master_data table.
type MasterData struct {
	ID          string          `gorm:"column:id;primaryKey"`
	Module      string          `gorm:"column:module"`
	Type        string          `gorm:"column:type"`
	Name        string          `gorm:"column:name"`
	Description string          `gorm:"column:description"`
	Icon        string          `gorm:"column:icon"`
	Color       string          `gorm:"column:color"`
	SortOrder   string          `gorm:"column:sort_order"`
	IsActive    bool            `gorm:"column:is_active"`
	Metadata    json.RawMessage `gorm:"column:metadata;type:jsonb"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
}

// TableName returns the fully qualified table name with schema prefix.
func (MasterData) TableName() string {
	return "stackforge_service.master_data"
}

// MasterDataFromEntity converts a domain entity to a GORM model.
func MasterDataFromEntity(e *entity.MasterData) *MasterData {
	var metadata json.RawMessage
	if e.Metadata != nil {
		data, err := json.Marshal(e.Metadata)
		if err == nil {
			metadata = data
		}
	}

	return &MasterData{
		ID:          e.ID,
		Module:      e.Module,
		Type:        e.Type,
		Name:        e.Name,
		Description: e.Description,
		Icon:        e.Icon,
		Color:       e.Color,
		SortOrder:   e.SortOrder,
		IsActive:    e.IsActive,
		Metadata:    metadata,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// ToEntity converts the GORM model back to a domain entity.
func (m *MasterData) ToEntity() *entity.MasterData {
	var metadata map[string]interface{}
	if m.Metadata != nil {
		_ = json.Unmarshal(m.Metadata, &metadata)
	}

	return &entity.MasterData{
		ID:          m.ID,
		Module:      m.Module,
		Type:        m.Type,
		Name:        m.Name,
		Description: m.Description,
		Icon:        m.Icon,
		Color:       m.Color,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		Metadata:    metadata,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
