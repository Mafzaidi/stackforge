package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
	"github.com/mafzaidi/stackforge/internal/delivery/http/serializer"
	"github.com/mafzaidi/stackforge/internal/domain/service"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
	"github.com/mafzaidi/stackforge/internal/usecase/menu"
)

// MenuHandler handles menu HTTP requests.
type MenuHandler struct {
	listUC menu.ListUseCase
	cfg    service.Config
}

// NewMenuHandler creates a new menu handler.
func NewMenuHandler(listUC menu.ListUseCase, cfg service.Config) *MenuHandler {
	return &MenuHandler{listUC: listUC, cfg: cfg}
}

// List handles GET /api/menus
func (h *MenuHandler) List(c *gin.Context) {
	claims, err := middleware.GetClaims(c)
	if err != nil {
		response.Unauthorized(c, "Missing authentication")
		return
	}

	items, err := h.listUC.Execute(c.Request.Context(), claims, h.cfg.GetAppCode())
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve menus")
		return
	}

	response.Success(c, "Menus retrieved successfully", serializer.FromMenuItemList(items))
}
