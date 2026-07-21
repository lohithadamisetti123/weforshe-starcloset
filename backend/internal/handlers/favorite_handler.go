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

// FavoriteHandler serves a shopper's saved products.
type FavoriteHandler struct {
	favorites *services.FavoriteService
	val       *validator.Validator
}

// NewFavoriteHandler constructs a FavoriteHandler.
func NewFavoriteHandler(favorites *services.FavoriteService, val *validator.Validator) *FavoriteHandler {
	return &FavoriteHandler{favorites: favorites, val: val}
}

// Add handles POST /api/v1/favorites.
//
//	@Summary		Save a product
//	@Tags			favorites
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.AddFavoriteRequest	true	"Product to save"
//	@Success		201		{object}	utils.Response{data=dto.Favorite}
//	@Failure		404		{object}	utils.ErrorResponse
//	@Router			/favorites [post]
func (h *FavoriteHandler) Add(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	var req dto.AddFavoriteRequest
	if !bindJSON(c, h.val, &req) {
		return
	}

	fav, err := h.favorites.Add(c.Request.Context(), userID, req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.Created(c, "Saved", fav)
}

// Remove handles DELETE /api/v1/favorites/:id.
//
// The path parameter is the *product* ID, matching what a product card in the
// client already has in hand — see services.FavoriteService.Remove.
//
//	@Summary		Remove a saved product
//	@Tags			favorites
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Product ID"
//	@Success		204	"No Content"
//	@Failure		404	{object}	utils.ErrorResponse
//	@Router			/favorites/{id} [delete]
func (h *FavoriteHandler) Remove(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Fail(c, utils.NewNotFound("Favorite"))
		return
	}

	if err := h.favorites.Remove(c.Request.Context(), userID, productID); err != nil {
		utils.Fail(c, err)
		return
	}
	utils.NoContent(c)
}

// List handles GET /api/v1/favorites.
//
//	@Summary		List saved products
//	@Tags			favorites
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.Response{data=[]dto.Favorite}
//	@Router			/favorites [get]
func (h *FavoriteHandler) List(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}

	favs, err := h.favorites.List(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, favs)
}
