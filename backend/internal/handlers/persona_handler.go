package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
	"github.com/myntra/bharat-ai-backend/internal/validator"
)

// PersonaHandler exposes the StarCloset stylist personas: the shopper roster and
// the partner onboarding portal. It is bind → delegate → respond, like every
// other handler; all persona logic lives in services.PersonaService.
type PersonaHandler struct {
	personas *services.PersonaService
	val      *validator.Validator
}

// NewPersonaHandler constructs a PersonaHandler.
func NewPersonaHandler(personas *services.PersonaService, val *validator.Validator) *PersonaHandler {
	return &PersonaHandler{personas: personas, val: val}
}

// ListPublished handles GET /api/v1/assistant/personas.
//
//	@Summary	List the published stylist personas a shopper can choose from
//	@Tags		personas
//	@Produce	json
//	@Success	200	{object}	utils.Response{data=[]dto.StylistPersona}
//	@Router		/assistant/personas [get]
func (h *PersonaHandler) ListPublished(c *gin.Context) {
	utils.OK(c, h.personas.ListPublished())
}

// Schema handles GET /api/v1/partner/personas/schema.
//
//	@Summary	Fetch the persona onboarding form schema
//	@Tags		partner
//	@Produce	json
//	@Success	200	{object}	utils.Response{data=dto.PersonaSchemaResponse}
//	@Router		/partner/personas/schema [get]
func (h *PersonaHandler) Schema(c *gin.Context) {
	utils.OK(c, h.personas.Schema())
}

// ListAll handles GET /api/v1/partner/personas.
//
//	@Summary	List every persona (drafts included) for the partner portal
//	@Tags		partner
//	@Produce	json
//	@Success	200	{object}	utils.Response{data=[]dto.PersonaDetail}
//	@Router		/partner/personas [get]
func (h *PersonaHandler) ListAll(c *gin.Context) {
	utils.OK(c, h.personas.ListAll())
}

// Get handles GET /api/v1/partner/personas/:slug.
//
//	@Summary	Fetch one persona in full detail
//	@Tags		partner
//	@Produce	json
//	@Param		slug	path		string	true	"Persona slug"
//	@Success	200		{object}	utils.Response{data=dto.PersonaDetail}
//	@Failure	404		{object}	utils.ErrorResponse
//	@Router		/partner/personas/{slug} [get]
func (h *PersonaHandler) Get(c *gin.Context) {
	detail, err := h.personas.Get(c.Param("slug"))
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, detail)
}

// Create handles POST /api/v1/partner/personas.
//
//	@Summary	Create a new draft persona
//	@Tags		partner
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.OnboardPersonaPayload	true	"Persona"
//	@Success	201		{object}	utils.Response{data=dto.PersonaDetail}
//	@Failure	409		{object}	utils.ErrorResponse
//	@Router		/partner/personas [post]
func (h *PersonaHandler) Create(c *gin.Context) {
	var req dto.OnboardPersonaPayload
	if !bindJSON(c, h.val, &req) {
		return
	}
	detail, err := h.personas.Create(req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.Created(c, "Stylist created", detail)
}

// Update handles PUT /api/v1/partner/personas/:slug.
//
//	@Summary	Update an existing persona
//	@Tags		partner
//	@Accept		json
//	@Produce	json
//	@Param		slug	path		string						true	"Persona slug"
//	@Param		body	body		dto.OnboardPersonaPayload	true	"Persona"
//	@Success	200		{object}	utils.Response{data=dto.PersonaDetail}
//	@Failure	404		{object}	utils.ErrorResponse
//	@Router		/partner/personas/{slug} [put]
func (h *PersonaHandler) Update(c *gin.Context) {
	var req dto.OnboardPersonaPayload
	if !bindJSON(c, h.val, &req) {
		return
	}
	detail, err := h.personas.Update(c.Param("slug"), req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, detail)
}

// Delete handles DELETE /api/v1/partner/personas/:slug.
//
//	@Summary	Delete a persona
//	@Tags		partner
//	@Produce	json
//	@Param		slug	path	string	true	"Persona slug"
//	@Success	204
//	@Failure	404	{object}	utils.ErrorResponse
//	@Router		/partner/personas/{slug} [delete]
func (h *PersonaHandler) Delete(c *gin.Context) {
	if err := h.personas.Delete(c.Param("slug")); err != nil {
		utils.Fail(c, err)
		return
	}
	utils.NoContent(c)
}

// Publish handles POST /api/v1/partner/personas/:slug/publish.
//
//	@Summary	Publish a persona so shoppers can choose it
//	@Tags		partner
//	@Produce	json
//	@Param		slug	path		string	true	"Persona slug"
//	@Success	200		{object}	utils.Response{data=dto.PersonaDetail}
//	@Failure	404		{object}	utils.ErrorResponse
//	@Router		/partner/personas/{slug}/publish [post]
func (h *PersonaHandler) Publish(c *gin.Context) {
	detail, err := h.personas.Publish(c.Param("slug"))
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, detail)
}
