package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
	"github.com/mafzaidi/stackforge/internal/usecase/credential"
)

// CredentialHandler handles credential HTTP requests
type CredentialHandler struct {
	listUC credential.ListUseCase
}

// NewCredentialHandler creates a new credential handler
func NewCredentialHandler(listUC credential.ListUseCase) *CredentialHandler {
	return &CredentialHandler{
		listUC: listUC,
	}
}

// List returns all credentials
// GET /api/credentials
func (h *CredentialHandler) List(c *gin.Context) {
	credentials, err := h.listUC.Execute(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Failed to list credentials")
		return
	}
	
	response.Success(c, gin.H{"items": credentials})
}
