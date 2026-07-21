package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/middleware"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
	"github.com/myntra/bharat-ai-backend/internal/validator"
)

// ProfileHandler serves a shopper's personalization profile and purchase
// history. Every route is identity-scoped: the user id comes from the auth
// middleware (a real account or a minted guest), never from the request body,
// so one shopper can neither read nor write another's profile.
type ProfileHandler struct {
	profiles *services.ProfileService
	val      *validator.Validator
}

// NewProfileHandler constructs a ProfileHandler.
func NewProfileHandler(profiles *services.ProfileService, val *validator.Validator) *ProfileHandler {
	return &ProfileHandler{profiles: profiles, val: val}
}

// Get handles GET /api/v1/profile/preferences.
//
//	@Summary	Get the current shopper's profile
//	@Tags		profile
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	utils.Response{data=dto.ProfileResponse}
//	@Router		/profile/preferences [get]
func (h *ProfileHandler) Get(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	profile, err := h.profiles.Get(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, profile)
}

// Update handles PUT /api/v1/profile/preferences — the onboarding submission
// and every later edit.
//
//	@Summary	Create or update the shopper's profile
//	@Tags		profile
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		body	body		dto.UpdateProfileRequest	true	"Profile"
//	@Success	200		{object}	utils.Response{data=dto.ProfileResponse}
//	@Failure	400		{object}	utils.ErrorResponse
//	@Router		/profile/preferences [put]
func (h *ProfileHandler) Update(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	var req dto.UpdateProfileRequest
	if !bindJSON(c, h.val, &req) {
		return
	}
	profile, err := h.profiles.Update(c.Request.Context(), userID, req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OKMessage(c, "Profile saved", profile)
}

// Status handles GET /api/v1/profile/status — the cheap check the client makes
// on load to decide whether to show onboarding.
//
//	@Summary	Onboarding status
//	@Tags		profile
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	utils.Response{data=dto.OnboardingStatusResponse}
//	@Router		/profile/status [get]
func (h *ProfileHandler) Status(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	status, err := h.profiles.Status(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, status)
}

// Orders handles GET /api/v1/profile/orders — the shopper's purchase history.
//
//	@Summary	List purchase history
//	@Tags		profile
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	utils.Response{data=[]dto.PastOrderResponse}
//	@Router		/profile/orders [get]
func (h *ProfileHandler) Orders(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	orders, err := h.profiles.Orders(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, orders)
}
