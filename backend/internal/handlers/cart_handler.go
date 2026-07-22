package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/middleware"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
	"github.com/myntra/bharat-ai-backend/internal/validator"
)

// CartHandler serves a shopper's cart, saved collections and collection cart.
//
// Every route resolves the owner with middleware.MustUserID — the authenticated
// account (or guest session) on the request — and passes it down. No route reads
// an owner id from the path or body, so one shopper cannot address another's
// cart. Deletes are matched on (id, user_id) in the repository, so even a
// correctly-guessed id belonging to someone else returns a plain 404.
type CartHandler struct {
	cart *services.CartService
	val  *validator.Validator
}

// NewCartHandler constructs a CartHandler.
func NewCartHandler(cart *services.CartService, val *validator.Validator) *CartHandler {
	return &CartHandler{cart: cart, val: val}
}

// ─── Product cart ────────────────────────────────────────────────────────────

// List handles GET /api/v1/cart.
func (h *CartHandler) List(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	items, err := h.cart.ListCart(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, items)
}

// Add handles POST /api/v1/cart.
func (h *CartHandler) Add(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	var req dto.AddCartItemRequest
	if !bindJSON(c, h.val, &req) {
		return
	}
	item, err := h.cart.AddToCart(c.Request.Context(), userID, req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.Created(c, "Added to cart", item)
}

// Remove handles DELETE /api/v1/cart/:productId.
func (h *CartHandler) Remove(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	if err := h.cart.RemoveFromCart(c.Request.Context(), userID, c.Param("productId")); err != nil {
		utils.Fail(c, err)
		return
	}
	utils.NoContent(c)
}

// Clear handles DELETE /api/v1/cart.
func (h *CartHandler) Clear(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	if err := h.cart.ClearCart(c.Request.Context(), userID); err != nil {
		utils.Fail(c, err)
		return
	}
	utils.NoContent(c)
}

// ─── Saved collections ───────────────────────────────────────────────────────

// ListSaved handles GET /api/v1/saved-collections.
func (h *CartHandler) ListSaved(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	items, err := h.cart.ListSaved(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, items)
}

// Save handles POST /api/v1/saved-collections.
func (h *CartHandler) Save(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	var req dto.SaveCollectionRequest
	if !bindJSON(c, h.val, &req) {
		return
	}
	item, err := h.cart.SaveCollection(c.Request.Context(), userID, req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.Created(c, "Collection saved", item)
}

// DeleteSaved handles DELETE /api/v1/saved-collections/:id.
func (h *CartHandler) DeleteSaved(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	if err := h.cart.DeleteSaved(c.Request.Context(), userID, c.Param("id")); err != nil {
		utils.Fail(c, err)
		return
	}
	utils.NoContent(c)
}

// ─── Collection cart ─────────────────────────────────────────────────────────

// ListCollectionCart handles GET /api/v1/collection-cart.
func (h *CartHandler) ListCollectionCart(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	items, err := h.cart.ListCollectionCart(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, items)
}

// AddCollection handles POST /api/v1/collection-cart.
func (h *CartHandler) AddCollection(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	var req dto.AddCollectionCartRequest
	if !bindJSON(c, h.val, &req) {
		return
	}
	item, err := h.cart.AddCollectionToCart(c.Request.Context(), userID, req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.Created(c, "Collection added to cart", item)
}

// UpdateCollection handles PUT /api/v1/collection-cart/:id.
func (h *CartHandler) UpdateCollection(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	var req dto.UpdateCollectionCartRequest
	if !bindJSON(c, h.val, &req) {
		return
	}
	item, err := h.cart.UpdateCollectionCartItem(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, item)
}

// DeleteCollection handles DELETE /api/v1/collection-cart/:id.
func (h *CartHandler) DeleteCollection(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	if err := h.cart.DeleteCollectionCartItem(c.Request.Context(), userID, c.Param("id")); err != nil {
		utils.Fail(c, err)
		return
	}
	utils.NoContent(c)
}

// ClearCollectionCart handles DELETE /api/v1/collection-cart.
func (h *CartHandler) ClearCollectionCart(c *gin.Context) {
	userID, ok := middleware.MustUserID(c)
	if !ok {
		return
	}
	if err := h.cart.ClearCollectionCart(c.Request.Context(), userID); err != nil {
		utils.Fail(c, err)
		return
	}
	utils.NoContent(c)
}
