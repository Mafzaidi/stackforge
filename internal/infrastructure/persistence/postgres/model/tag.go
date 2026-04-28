package model

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

type Tag struct {
	ID        string    `gorm:"column:id;primaryKey"`
	UserID    string    `gorm:"column:user_id"`
	Name      string    `gorm:"column:name"`
	Color     *string   `gorm:"column:color"`
	Module    string    `gorm:"column:module"`
	RefID     string    `gorm:"column:ref_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
	IsActive  bool      `gorm:"column:is_active"`
}

func (Tag) TableName() string {
	return "stackforge_service.tags"
}

func TagFromEntity(e *entity.Tag) *Tag {
	return &Tag{
		ID:        e.ID,
		UserID:    e.UserID,
		Name:      e.Name,
		Color:     e.Color,
		Module:    e.Module,
		RefID:     e.RefID,
		CreatedAt: e.CreatedAt,
		IsActive:  e.IsActive,
	}
}

func (t *Tag) ToEntity() *entity.Tag {
	return &entity.Tag{
		ID:        t.ID,
		UserID:    t.UserID,
		Name:      t.Name,
		Color:     t.Color,
		Module:    t.Module,
		RefID:     t.RefID,
		CreatedAt: t.CreatedAt,
		IsActive:  t.IsActive,
	}
}
