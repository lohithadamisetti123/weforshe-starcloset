package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/middleware"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
	"github.com/myntra/bharat-ai-backend/internal/validator"
)

// AuthHandler exposes registration, login and token refresh.
type AuthHandler struct {
	auth *services.AuthService
	val  *validator.Validator
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(auth *services.AuthService, val *validator.Validator) *AuthHandler {
	return &AuthHandler{auth: auth, val: val}
}

// Register handles POST /api/v1/auth/register.
//
//	@Summary		Register a new account
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RegisterRequest	true	"Registration details"
//	@Success		201		{object}	utils.Response{data=dto.AuthResponse}
//	@Failure		409		{object}	utils.ErrorResponse
//	@Failure		422		{object}	utils.ErrorResponse
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !bindJSON(c, h.val, &req) {
		return
	}

	resp, err := h.auth.Register(c.Request.Context(), req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.Created(c, "Account created", resp)
}

// Login handles POST /api/v1/auth/login.
//
//	@Summary		Authenticate with email and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.LoginRequest	true	"Credentials"
//	@Success		200		{object}	utils.Response{data=dto.AuthResponse}
//	@Failure		401		{object}	utils.ErrorResponse
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, h.val, &req) {
		return
	}

	resp, err := h.auth.Login(c.Request.Context(), req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, resp)
}

// Refresh handles POST /api/v1/auth/refresh.
//
//	@Summary		Exchange a refresh token for a new pair
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RefreshRequest	true	"Refresh token"
//	@Success		200		{object}	utils.Response{data=dto.AuthResponse}
//	@Failure		401		{object}	utils.ErrorResponse
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if !bindJSON(c, h.val, &req) {
		return
	}

	resp, err := h.auth.Refresh(c.Request.Context(), req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, resp)
}

// Profile handles GET /api/v1/profile.
//
//	@Summary		Get the authenticated account
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.Response{data=dto.User}
//	@Failure		401	{object}	utils.ErrorResponse
//	@Router			/profile [get]
func (h *AuthHandler) Profile(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	user, err := h.auth.Profile(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, user)
}
