package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
	"github.com/mafzaidi/stackforge/internal/delivery/http/serializer"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
	"github.com/mafzaidi/stackforge/internal/usecase/vault"
)

type CreateVaultPayload struct {
	Name        string `json:"name" binding:"required,max=255"`
	Description string `json:"description" binding:"required,max=255"`
}

type VaultHandler struct {
	createUC vault.CreateUseCase
	listUC   vault.ListUseCase
}

func NewVaultHandler(
	createUC vault.CreateUseCase,
	listUC vault.ListUseCase,
) *VaultHandler {
	return &VaultHandler{
		createUC: createUC,
		listUC:   listUC,
	}
}

func (h *VaultHandler) Create(c *gin.Context) {
	claims, err := middleware.GetClaims(c)
	if err != nil {
		response.Unauthorized(c, "Missing or invalid authentication token")
		return
	}

	var req CreateVaultPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	v, err := h.createUC.Execute(
		c.Request.Context(),
		claims.Subject,
		req.Name,
		req.Description,
	)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Vault created successfully", serializer.FromVault(v))
}

func (h *VaultHandler) List(c *gin.Context) {
	claims, err := middleware.GetClaims(c)
	if err != nil {
		response.Unauthorized(c, "Missing or invalid authentication token")
		return
	}

	vaults, err := h.listUC.Execute(c.Request.Context(), claims.Subject)
	if err != nil {
		response.InternalServerError(c, "Failed to list vaults")
		return
	}

	response.Success(c, "Vaults retrieved successfully", serializer.FromVaultList(vaults))
}
