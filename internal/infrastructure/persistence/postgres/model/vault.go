package model

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

type Vault struct {
	ID          string    `gorm:"column:id;primaryKey"`
	UserID      string    `gorm:"column:user_id"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	Icon        string    `gorm:"column:icon"`
	IsDefault   bool      `gorm:"column:is_default"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Vault) TableName() string {
	return "stackforge_service.vaults"
}

func VaultFromEntity(e *entity.Vault) *Vault {
	return &Vault{
		ID:          e.ID,
		UserID:      e.UserID,
		Name:        e.Name,
		Description: e.Description,
		Icon:        e.Icon,
		IsDefault:   e.IsDefault,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func (m *Vault) ToEntity() *entity.Vault {
	return &entity.Vault{
		ID:          m.ID,
		UserID:      m.UserID,
		Name:        m.Name,
		Description: m.Description,
		Icon:        m.Icon,
		IsDefault:   m.IsDefault,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
