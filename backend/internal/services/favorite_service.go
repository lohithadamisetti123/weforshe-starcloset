package services

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// FavoriteService serves saved products.
type FavoriteService struct {
	favorites repositories.FavoriteRepository
	log       *zap.Logger
}

// NewFavoriteService constructs a FavoriteService.
func NewFavoriteService(favorites repositories.FavoriteRepository, log *zap.Logger) *FavoriteService {
	return &FavoriteService{favorites: favorites, log: log}
}

// Add saves a product for a shopper.
func (s *FavoriteService) Add(ctx context.Context, userID uuid.UUID, req dto.AddFavoriteRequest) (*dto.Favorite, error) {
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		// Validation has already checked the uuid4 tag, so reaching here means
		// the tag and this parse disagree — treat it as not-found rather than
		// as a 500 the shopper can do nothing about.
		return nil, utils.NewNotFound("Product")
	}

	fav, err := s.favorites.Add(ctx, userID, productID)
	if err != nil {
		return nil, err
	}

	out := dto.NewFavorite(fav)
	return &out, nil
}

// Remove deletes a saved product. The path parameter is the *product* ID, not
// the favourite row's ID: the client is holding a product card and knows which
// product it wants unsaved, not which join row represents that fact.
func (s *FavoriteService) Remove(ctx context.Context, userID, productID uuid.UUID) error {
	return s.favorites.Remove(ctx, userID, productID)
}

// List returns a shopper's saved products, newest first.
func (s *FavoriteService) List(ctx context.Context, userID uuid.UUID) ([]dto.Favorite, error) {
	favs, err := s.favorites.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewFavorites(favs), nil
}
