package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/middleware"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
	"github.com/myntra/bharat-ai-backend/internal/validator"
)

// StarClosetHandler exposes the V2 department journey: the public department
// rosters and the authenticated per-department unlock progress.
type StarClosetHandler struct {
	svc *services.StarClosetService
	val *validator.Validator
}

// NewStarClosetHandler constructs a StarClosetHandler.
func NewStarClosetHandler(svc *services.StarClosetService, val *validator.Validator) *StarClosetHandler {
	return &StarClosetHandler{svc: svc, val: val}
}

// Departments handles GET /api/v1/starcloset/departments.
//
//	@Summary	List the StarCloset shopping departments and their stylist rosters
//	@Tags		starcloset
//	@Produce	json
//	@Success	200	{object}	utils.Response{data=[]dto.StarDepartmentResponse}
//	@Router		/starcloset/departments [get]
func (h *StarClosetHandler) Departments(c *gin.Context) {
	utils.OK(c, h.svc.Departments())
}

// GetProgress handles GET /api/v1/starcloset/progress.
//
//	@Summary	Fetch the shopper's per-department unlock progress
//	@Tags		starcloset
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	utils.Response{data=dto.ProgressResponse}
//	@Router		/starcloset/progress [get]
func (h *StarClosetHandler) GetProgress(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	m, err := h.svc.Progress(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, dto.ProgressResponse{Progress: m})
}

// Increment handles POST /api/v1/starcloset/progress.
//
// One cart-add advances one department. The body names the department and an
// optional delta (default 1). Only the currently-shopped department moves.
//
//	@Summary	Advance one department's unlock progress
//	@Tags		starcloset
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		body	body		dto.IncrementProgressRequest	true	"Department and delta"
//	@Success	200		{object}	utils.Response{data=dto.ProgressResponse}
//	@Failure	422		{object}	utils.ErrorResponse
//	@Router		/starcloset/progress [post]
func (h *StarClosetHandler) Increment(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	var req dto.IncrementProgressRequest
	if !bindJSON(c, h.val, &req) {
		return
	}
	m, err := h.svc.Increment(c.Request.Context(), userID, req.Department, req.Delta)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, dto.ProgressResponse{Progress: m})
}
