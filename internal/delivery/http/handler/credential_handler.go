package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
	"github.com/mafzaidi/stackforge/internal/delivery/http/serializer"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
	"github.com/mafzaidi/stackforge/internal/usecase/credential"
)

type CreateCredentialPayload struct {
	Title            string   `json:"title" binding:"required,max=255"`
	SiteUrl          string   `json:"site_url" binding:"required"`
	FaviconUrl       string   `json:"favicon_url" binding:"required"`
	Username         string   `json:"username" binding:"required,max=255"`
	Password         string   `json:"password" binding:"required,max=10000"`
	Notes            string   `json:"notes"`
	IsFavorite       bool     `json:"is_favorite" binding:"required"`
	PasswordStrength int32    `json:"password_strength" binding:"required"`
	VaultID          string   `json:"vault_id"`
	CategoryID       string   `json:"category_id"`
	Tags             []string `json:"tags"`
}

type credentialListQuery struct {
	Page  string `form:"page"`
	Limit string `form:"limit"`
}

type CredentialHandler struct {
	listUC   credential.ListUseCase
	createUC credential.CreateUseCase
}

// NewCredentialHandler creates a new credential handler.
func NewCredentialHandler(
	listUC credential.ListUseCase,
	createUC credential.CreateUseCase,
) *CredentialHandler {
	return &CredentialHandler{
		listUC:   listUC,
		createUC: createUC,
	}
}

func (h *CredentialHandler) Create(c *gin.Context) {
	claims, err := middleware.GetClaims(c)
	if err != nil {
		response.Unauthorized(c, "Missing or invalid authentication token")
		return
	}

	var req CreateCredentialPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cred, err := h.createUC.Execute(
		c.Request.Context(),
		claims.Subject,
		req.Title,
		req.SiteUrl,
		req.FaviconUrl,
		req.Username,
		req.Password,
		req.Notes,
		req.IsFavorite,
		req.PasswordStrength,
		req.VaultID,
		req.CategoryID,
		req.Tags,
	)
	if err != nil {
		switch err.Error() {
		case "invalid master password":
			response.Unauthorized(c, err.Error())
		case "master password not set up yet, please set up your master password first":
			response.BadRequest(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Created(c, "Credential created successfully", serializer.FromCredential(cred))
}

func (h *CredentialHandler) List(c *gin.Context) {
	var query credentialListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	limit := 10
	offset := 0
	if query.Limit != "" {
		if _, err := fmt.Sscanf(query.Limit, "%d", &limit); err != nil {
			response.BadRequest(c, "Invalid limit parameter")
			return
		}
	}
	if query.Page != "" {
		var page int
		if _, err := fmt.Sscanf(query.Page, "%d", &page); err != nil {
			response.BadRequest(c, "Invalid page parameter")
			return
		}
		offset = (page - 1) * limit
	}

	credentials, err := h.listUC.Execute(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalServerError(c, "Failed to list credentials")
		return
	}

	response.SuccessWithPagination(c, "Credentials retrieved successfully", serializer.FromCredentialList(credentials), response.Pagination{
		Page:       1,
		Limit:      10,
		TotalItems: len(credentials),
		TotalPages: 1,
	})
}
