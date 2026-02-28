package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
	"github.com/mafzaidi/stackforge/internal/usecase/todo"
)

// TodoHandler handles todo HTTP requests
type TodoHandler struct {
	listUC todo.ListUseCase
}

// NewTodoHandler creates a new todo handler
func NewTodoHandler(listUC todo.ListUseCase) *TodoHandler {
	return &TodoHandler{
		listUC: listUC,
	}
}

// List returns all todos
// GET /api/todos
func (h *TodoHandler) List(c *gin.Context) {
	todos, err := h.listUC.Execute(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Failed to list todos")
		return
	}
	
	response.Success(c, gin.H{"items": todos})
}
