package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/myntra/bharat-ai-backend/internal/middleware"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// DashboardHandler serves the recommendations dashboard's summary figures.
type DashboardHandler struct {
	dashboard *services.DashboardService
}

// NewDashboardHandler constructs a DashboardHandler.
func NewDashboardHandler(dashboard *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboard: dashboard}
}

// Stats handles GET /api/v1/dashboard/stats.
//
//	@Summary		Get the shopper's recommendation dashboard figures
//	@Tags			dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.Response{data=dto.DashboardStats}
//	@Router			/dashboard/stats [get]
func (h *DashboardHandler) Stats(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	stats, err := h.dashboard.Stats(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, stats)
}
