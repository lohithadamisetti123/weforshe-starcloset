package dto

import (
	"github.com/myntra/bharat-ai-backend/internal/models"
)

// AddFavoriteRequest is the body of POST /api/v1/favorites.
type AddFavoriteRequest struct {
	ProductID string `json:"productId" binding:"required" validate:"required,uuid"`
}

// Favorite is the wire shape of a saved product.
//
// The product is embedded rather than referenced by ID. The favourites page
// renders product cards, so returning bare IDs would force the client into N+1
// round-trips to draw a single grid.
type Favorite struct {
	ID        string  `json:"id"`
	ProductID string  `json:"productId"`
	Product   Product `json:"product"`
	CreatedAt string  `json:"createdAt"`
}

// NewFavorite maps a favourite model.
func NewFavorite(f *models.FavoriteProduct) Favorite {
	if f == nil {
		return Favorite{}
	}
	return Favorite{
		ID:        f.ID.String(),
		ProductID: f.ProductID.String(),
		Product:   NewProduct(f.Product),
		CreatedAt: f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// NewFavorites maps a slice, always returning a non-nil slice.
func NewFavorites(fs []*models.FavoriteProduct) []Favorite {
	out := make([]Favorite, 0, len(fs))
	for _, f := range fs {
		out = append(out, NewFavorite(f))
	}
	return out
}
