package serializer

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// TodoResponse is the API representation of a todo item.
type TodoResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FromTodo converts a domain todo entity to an API response.
func FromTodo(e *entity.Todo) TodoResponse {
	return TodoResponse{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		Done:        e.Done,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// FromTodoList converts a slice of todo entities.
func FromTodoList(entities []*entity.Todo) []TodoResponse {
	result := make([]TodoResponse, len(entities))
	for i, e := range entities {
		result[i] = FromTodo(e)
	}
	return result
}
