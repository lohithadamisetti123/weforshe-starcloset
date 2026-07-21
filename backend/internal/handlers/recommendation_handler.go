package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/middleware"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
	"github.com/myntra/bharat-ai-backend/internal/validator"
)

// RecommendationHandler serves stored engine runs.
type RecommendationHandler struct {
	recs *services.RecommendationService
	val  *validator.Validator
}

// NewRecommendationHandler constructs a RecommendationHandler.
func NewRecommendationHandler(recs *services.RecommendationService, val *validator.Validator) *RecommendationHandler {
	return &RecommendationHandler{recs: recs, val: val}
}

// List handles GET /api/v1/recommendations.
//
//	@Summary		List the shopper's past recommendation runs
//	@Tags			recommendations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"
//	@Param			pageSize	query		int	false	"Page size"
//	@Success		200			{object}	utils.Response{data=[]dto.RecommendationSummary,meta=utils.PageMeta}
//	@Router			/recommendations [get]
func (h *RecommendationHandler) List(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	var p dto.PaginationQuery
	if !bindQuery(c, h.val, &p) {
		return
	}

	recs, meta, err := h.recs.List(c.Request.Context(), userID, p)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OKWithMeta(c, recs, meta)
}

// GetByID handles GET /api/v1/recommendations/:id.
//
//	@Summary		Get one stored recommendation run
//	@Tags			recommendations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Recommendation ID"
//	@Success		200	{object}	utils.Response{data=dto.AIRecommendationResponse}
//	@Failure		404	{object}	utils.ErrorResponse
//	@Router			/recommendations/{id} [get]
func (h *RecommendationHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Fail(c, utils.NewNotFound("Recommendation"))
		return
	}

	rec, err := h.recs.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, rec)
}
