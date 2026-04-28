package model

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// Todo is the GORM model for the stackforge_service.todos table.
type Todo struct {
	ID          string    `gorm:"column:id;primaryKey"`
	Title       string    `gorm:"column:title"`
	Description string    `gorm:"column:description"`
	Done        bool      `gorm:"column:done"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName returns the fully qualified table name with schema prefix.
func (Todo) TableName() string {
	return "stackforge_service.todos"
}

// TodoFromEntity converts a domain entity to a GORM model.
func TodoFromEntity(e *entity.Todo) *Todo {
	return &Todo{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		Done:        e.Done,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// ToEntity converts the GORM model back to a domain entity.
func (m *Todo) ToEntity() *entity.Todo {
	return &entity.Todo{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Done:        m.Done,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
