package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/middleware"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
	"github.com/myntra/bharat-ai-backend/internal/validator"
)

// AssistantHandler exposes the natural-language shopping assistant.
//
// This is the product's USP, and the handler reflects that: everything it does
// is bind, delegate to AssistantService.Chat, respond. The five-stage pipeline
// that makes the answer good — intent extraction, retrieval, scoring,
// diversification, explanation — lives entirely in internal/recommendation and
// internal/ai, reachable and testable without an HTTP request in sight.
type AssistantHandler struct {
	assistant *services.AssistantService
	val       *validator.Validator
}

// NewAssistantHandler constructs an AssistantHandler.
func NewAssistantHandler(assistant *services.AssistantService, val *validator.Validator) *AssistantHandler {
	return &AssistantHandler{assistant: assistant, val: val}
}

// Chat handles POST /api/v1/assistant/chat.
//
//	@Summary		Ask the assistant for a ranked, explained shortlist
//	@Description	Runs both extractors (deterministic rules and Gemini, when configured), merges their reading of the prompt, and ranks the catalogue against it.
//	@Tags			assistant
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.ChatRequest	true	"The shopper's prompt"
//	@Success		200		{object}	utils.Response{data=dto.AIRecommendationResponse}
//	@Failure		422		{object}	utils.ErrorResponse
//	@Router			/assistant/chat [post]
func (h *AssistantHandler) Chat(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	var req dto.ChatRequest
	if !bindJSON(c, h.val, &req) {
		return
	}

	resp, err := h.assistant.Chat(c.Request.Context(), userID, req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, resp)
}

// Replay handles GET /api/v1/assistant/recommendations/:id.
//
// It re-renders a stored shortlist in the assistant's response shape, from the
// ranks that were persisted rather than a fresh engine run — see
// services.AssistantService.Replay for why that distinction matters.
//
//	@Summary		Re-fetch a past assistant answer by recommendation ID
//	@Tags			assistant
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Recommendation ID"
//	@Success		200	{object}	utils.Response{data=dto.AIRecommendationResponse}
//	@Failure		404	{object}	utils.ErrorResponse
//	@Router			/assistant/recommendations/{id} [get]
func (h *AssistantHandler) Replay(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Fail(c, utils.NewNotFound("Recommendation"))
		return
	}

	resp, err := h.assistant.Replay(c.Request.Context(), userID, id)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, resp)
}

// Collections handles POST /api/v1/assistant/collections.
//
// It returns the full set of themed, mutually-distinct collections for a
// department + prompt in one call — the variety-preserving replacement for the
// old pattern of three parallel /collection calls. See
// services.AssistantService.Collections.
//
//	@Summary		Build all themed collections for a department in one call
//	@Description	Returns three mutually-distinct collections (no product repeats across them) for the department and prompt.
//	@Tags			assistant
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.CollectionRequest	true	"Department and prompt"
//	@Success		200		{object}	utils.Response{data=dto.CollectionsResponse}
//	@Failure		422		{object}	utils.ErrorResponse
//	@Router			/assistant/collections [post]
func (h *AssistantHandler) Collections(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	var req dto.CollectionRequest
	if !bindJSON(c, h.val, &req) {
		return
	}

	resp, err := h.assistant.Collections(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, services.ErrUnsupportedDepartment) {
			utils.Fail(c, utils.NewValidation("This department isn't available yet", err.Error()))
			return
		}
		utils.Fail(c, err)
		return
	}
	utils.OK(c, resp)
}

// Departments handles GET /api/v1/assistant/departments.
//
// This is step one of the guided flow: the frontend renders one card per
// entry rather than hard-coding the list, and Available tells it whether to
// show the card as clickable or as "coming soon" — see
// recommendation.DepartmentSupport for why some entries aren't buildable yet.
//
//	@Summary		List the first-step department choices for the guided assistant
//	@Tags			assistant
//	@Produce		json
//	@Success		200	{object}	utils.Response{data=[]dto.DepartmentOption}
//	@Router			/assistant/departments [get]
func (h *AssistantHandler) Departments(c *gin.Context) {
	utils.OK(c, dto.NewDepartmentOptions())
}

// Collection handles POST /api/v1/assistant/collection.
//
// This is step two of the guided flow — a department plus a prompt produces
// a complete, budget-aware set of items (an outfit, not a single product).
// See services.AssistantService.Collection.
//
//	@Summary		Build a budget-aware collection for a department
//	@Description	Fills a per-department slot template (e.g. Outfit + Footwear + Accessories) from the stated budget, never mixing in products from another department.
//	@Tags			assistant
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.CollectionRequest	true	"Department and prompt"
//	@Success		200		{object}	utils.Response{data=dto.CollectionResponse}
//	@Failure		422		{object}	utils.ErrorResponse
//	@Router			/assistant/collection [post]
func (h *AssistantHandler) Collection(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	var req dto.CollectionRequest
	if !bindJSON(c, h.val, &req) {
		return
	}

	resp, err := h.assistant.Collection(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, services.ErrUnsupportedDepartment) {
			utils.Fail(c, utils.NewValidation("This department isn't available yet", err.Error()))
			return
		}
		utils.Fail(c, err)
		return
	}
	utils.OK(c, resp)
}
