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

// HistoryHandler serves a shopper's past assistant conversations.
type HistoryHandler struct {
	history *services.HistoryService
	val     *validator.Validator
}

// NewHistoryHandler constructs a HistoryHandler.
func NewHistoryHandler(history *services.HistoryService, val *validator.Validator) *HistoryHandler {
	return &HistoryHandler{history: history, val: val}
}

// List handles GET /api/v1/history.
//
//	@Summary		List the shopper's past prompts
//	@Tags			history
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"
//	@Param			pageSize	query		int	false	"Page size"
//	@Success		200			{object}	utils.Response{data=[]dto.HistoryEntry,meta=utils.PageMeta}
//	@Router			/history [get]
func (h *HistoryHandler) List(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	var p dto.PaginationQuery
	if !bindQuery(c, h.val, &p) {
		return
	}

	entries, meta, err := h.history.List(c.Request.Context(), userID, p)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OKWithMeta(c, entries, meta)
}

// Create handles POST /api/v1/history.
//
// The assistant records its own turns automatically; this endpoint exists for
// a client that handled a prompt itself and wants it logged, and for seeding
// demo data — see dto.CreateHistoryRequest.
//
//	@Summary		Log a history entry
//	@Tags			history
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.CreateHistoryRequest	true	"Entry to record"
//	@Success		201		{object}	utils.Response{data=dto.HistoryEntry}
//	@Router			/history [post]
func (h *HistoryHandler) Create(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	var req dto.CreateHistoryRequest
	if !bindJSON(c, h.val, &req) {
		return
	}

	entry, err := h.history.Create(c.Request.Context(), userID, req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.Created(c, "History entry recorded", entry)
}

// UpdateStatus handles PATCH /api/v1/history/:id.
//
//	@Summary		Move a history entry between Completed, Saved and Purchased
//	@Tags			history
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"History entry ID"
//	@Param			body	body		dto.UpdateHistoryRequest	true	"New status"
//	@Success		200		{object}	utils.Response{data=dto.HistoryEntry}
//	@Failure		404		{object}	utils.ErrorResponse
//	@Router			/history/{id} [patch]
func (h *HistoryHandler) UpdateStatus(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Fail(c, utils.NewNotFound("History entry"))
		return
	}

	var req dto.UpdateHistoryRequest
	if !bindJSON(c, h.val, &req) {
		return
	}

	entry, err := h.history.UpdateStatus(c.Request.Context(), userID, id, req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, entry)
}

// Delete handles DELETE /api/v1/history/:id.
//
//	@Summary		Delete a history entry
//	@Tags			history
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"History entry ID"
//	@Success		204	"No Content"
//	@Failure		404	{object}	utils.ErrorResponse
//	@Router			/history/{id} [delete]
func (h *HistoryHandler) Delete(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Fail(c, utils.NewNotFound("History entry"))
		return
	}

	if err := h.history.Delete(c.Request.Context(), userID, id); err != nil {
		utils.Fail(c, err)
		return
	}
	utils.NoContent(c)
}
