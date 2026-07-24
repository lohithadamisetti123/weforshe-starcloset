package dto

import (
	"encoding/json"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// ─── Product cart ────────────────────────────────────────────────────────────

// AddCartItemRequest is the body of POST /api/v1/cart.
//
// Note there is no userId field, here or anywhere in this file: ownership comes
// from the authenticated identity on the request, never from the body. A client
// cannot ask to write into someone else's cart because there is nowhere to say
// whose cart it is.
type AddCartItemRequest struct {
	// Any RFC-4122 UUID version is accepted: the seed catalogue mints product IDs
	// deterministically (UUIDv5 from the slug) as well as randomly (UUIDv4), so a
	// uuid4-only rule would 422 over half the catalogue — including every product
	// the collection builder surfaces.
	ProductID string `json:"productId" binding:"required" validate:"required,uuid"`
}

// CartItem is the wire shape of one cart line. The product is embedded so the
// cart page renders without a second round-trip per line.
type CartItem struct {
	ID       string  `json:"id"`
	Product  Product `json:"product"`
	Quantity int     `json:"quantity"`
	AddedAt  string  `json:"addedAt"`
}

// NewCartItem maps a cart model.
func NewCartItem(c *models.CartItem) CartItem {
	if c == nil {
		return CartItem{}
	}
	return CartItem{
		ID:       c.ID.String(),
		Product:  NewProduct(c.Product),
		Quantity: c.Quantity,
		AddedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// NewCartItems maps a slice, always returning a non-nil slice.
func NewCartItems(items []*models.CartItem) []CartItem {
	out := make([]CartItem, 0, len(items))
	for _, c := range items {
		out = append(out, NewCartItem(c))
	}
	return out
}

// ─── Saved collections ───────────────────────────────────────────────────────

// SaveCollectionRequest is the body of POST /api/v1/saved-collections.
//
// The collection travels as raw JSON: it is a snapshot of exactly what the UI
// rendered, and re-validating its internal shape here would couple this
// endpoint to the ranker's output format for no benefit — the client stores it
// and the client renders it back.
type SaveCollectionRequest struct {
	Name       string          `json:"name" binding:"required" validate:"required,max=180"`
	Department string          `json:"department" validate:"max=48"`
	Prompt     string          `json:"prompt" validate:"max=2000"`
	Collection json.RawMessage `json:"collection" binding:"required"`
}

// SavedCollectionResponse is the wire shape of a saved collection. The field
// names mirror what the client already stores, so the UI is unchanged.
type SavedCollectionResponse struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Prompt     string          `json:"prompt"`
	Department string          `json:"department"`
	Collection json.RawMessage `json:"collection"`
	SavedAt    string          `json:"savedAt"`
}

// NewSavedCollection maps a saved-collection model.
func NewSavedCollection(s *models.SavedCollection) SavedCollectionResponse {
	if s == nil {
		return SavedCollectionResponse{}
	}
	return SavedCollectionResponse{
		ID:         s.ID.String(),
		Name:       s.Name,
		Prompt:     s.Prompt,
		Department: s.Department,
		Collection: json.RawMessage(s.Collection),
		SavedAt:    s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// NewSavedCollections maps a slice, always returning a non-nil slice.
func NewSavedCollections(items []*models.SavedCollection) []SavedCollectionResponse {
	out := make([]SavedCollectionResponse, 0, len(items))
	for _, s := range items {
		out = append(out, NewSavedCollection(s))
	}
	return out
}

// ─── Collection cart ─────────────────────────────────────────────────────────

// AddCollectionCartRequest is the body of POST /api/v1/collection-cart.
type AddCollectionCartRequest struct {
	Name       string          `json:"name" binding:"required" validate:"required,max=180"`
	Department string          `json:"department" validate:"max=48"`
	Collection json.RawMessage `json:"collection" binding:"required"`
}

// UpdateCollectionCartRequest is the body of PUT /api/v1/collection-cart/:id,
// used when a shopper removes one slot from a collection already in their cart.
type UpdateCollectionCartRequest struct {
	Collection json.RawMessage `json:"collection" binding:"required"`
}

// CollectionCartResponse is the wire shape of a collection-cart entry.
type CollectionCartResponse struct {
	ID         string          `json:"id"`
	Collection json.RawMessage `json:"collection"`
	AddedAt    string          `json:"addedAt"`
}

// NewCollectionCartItem maps a collection-cart model.
func NewCollectionCartItem(c *models.CollectionCartItem) CollectionCartResponse {
	if c == nil {
		return CollectionCartResponse{}
	}
	return CollectionCartResponse{
		ID:         c.ID.String(),
		Collection: json.RawMessage(c.Collection),
		AddedAt:    c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// NewCollectionCartItems maps a slice, always returning a non-nil slice.
func NewCollectionCartItems(items []*models.CollectionCartItem) []CollectionCartResponse {
	out := make([]CollectionCartResponse, 0, len(items))
	for _, c := range items {
		out = append(out, NewCollectionCartItem(c))
	}
	return out
}
