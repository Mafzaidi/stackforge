package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
	"github.com/mafzaidi/stackforge/internal/delivery/http/serializer"
	"github.com/mafzaidi/stackforge/internal/domain/service"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
	"github.com/mafzaidi/stackforge/internal/usecase/masterdata"
)

type masterDatalistQuery struct {
	Module string `form:"module"`
	Type   string `form:"type"`
}

type MasterDataHandler struct {
	listUC masterdata.ListUseCase
	cfg    service.Config
}

func NewMasterDataHandler(listUC masterdata.ListUseCase, cfg service.Config) *MasterDataHandler {
	return &MasterDataHandler{listUC: listUC, cfg: cfg}
}

// List handles GET /api/master-data
func (h *MasterDataHandler) List(c *gin.Context) {
	var query masterDatalistQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Bad request")
		return
	}
	_, err := middleware.GetClaims(c)
	if err != nil {
		response.Unauthorized(c, "Missing or invalid authentication token")
		return
	}

	md, err := h.listUC.Execute(c.Request.Context(), query.Module, query.Type)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve master data")
		return
	}

	response.Success(c, "Master data retrieved successfully", serializer.FromMasterDataList(md))
}
