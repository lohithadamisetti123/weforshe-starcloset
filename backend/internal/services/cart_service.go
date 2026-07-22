package services

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// CartService owns a shopper's product cart, saved collections and collection
// cart — the three stores that used to live in browser localStorage under a
// single global key each, and therefore leaked between accounts sharing a
// browser.
//
// Every method here takes the authenticated user id from its caller and passes
// it straight to a repository that filters on it. The service never accepts an
// owner id from a request body, so there is no path by which a client can name
// a cart other than its own.
type CartService struct {
	cart        repositories.CartRepository
	saved       repositories.SavedCollectionRepository
	collections repositories.CollectionCartRepository
	log         *zap.Logger
}

// NewCartService constructs a CartService.
func NewCartService(
	cart repositories.CartRepository,
	saved repositories.SavedCollectionRepository,
	collections repositories.CollectionCartRepository,
	log *zap.Logger,
) *CartService {
	return &CartService{cart: cart, saved: saved, collections: collections, log: log}
}

// ─── Product cart ────────────────────────────────────────────────────────────

// ListCart returns the user's cart.
func (s *CartService) ListCart(ctx context.Context, userID uuid.UUID) ([]dto.CartItem, error) {
	items, err := s.cart.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewCartItems(items), nil
}

// AddToCart adds a product (or increments its quantity).
func (s *CartService) AddToCart(ctx context.Context, userID uuid.UUID, req dto.AddCartItemRequest) (dto.CartItem, error) {
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return dto.CartItem{}, utils.NewBadRequest("Invalid product id", err)
	}
	item, err := s.cart.Add(ctx, userID, productID)
	if err != nil {
		return dto.CartItem{}, err
	}
	return dto.NewCartItem(item), nil
}

// RemoveFromCart removes a product from the user's cart.
func (s *CartService) RemoveFromCart(ctx context.Context, userID uuid.UUID, productIDRaw string) error {
	productID, err := uuid.Parse(productIDRaw)
	if err != nil {
		return utils.NewBadRequest("Invalid product id", err)
	}
	return s.cart.Remove(ctx, userID, productID)
}

// ClearCart empties the user's cart.
func (s *CartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return s.cart.Clear(ctx, userID)
}

// ─── Saved collections ───────────────────────────────────────────────────────

// ListSaved returns the user's saved collections.
func (s *CartService) ListSaved(ctx context.Context, userID uuid.UUID) ([]dto.SavedCollectionResponse, error) {
	items, err := s.saved.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewSavedCollections(items), nil
}

// SaveCollection stores a collection snapshot for the user.
func (s *CartService) SaveCollection(ctx context.Context, userID uuid.UUID, req dto.SaveCollectionRequest) (dto.SavedCollectionResponse, error) {
	item := &models.SavedCollection{
		UserID:     userID,
		Name:       req.Name,
		Department: req.Department,
		Prompt:     req.Prompt,
		Collection: models.JSONB(req.Collection),
	}
	stored, err := s.saved.Create(ctx, item)
	if err != nil {
		return dto.SavedCollectionResponse{}, err
	}
	return dto.NewSavedCollection(stored), nil
}

// DeleteSaved removes one of the user's saved collections.
func (s *CartService) DeleteSaved(ctx context.Context, userID uuid.UUID, idRaw string) error {
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return utils.NewBadRequest("Invalid collection id", err)
	}
	return s.saved.Delete(ctx, userID, id)
}

// ─── Collection cart ─────────────────────────────────────────────────────────

// ListCollectionCart returns the user's collection cart.
func (s *CartService) ListCollectionCart(ctx context.Context, userID uuid.UUID) ([]dto.CollectionCartResponse, error) {
	items, err := s.collections.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewCollectionCartItems(items), nil
}

// AddCollectionToCart adds a whole collection to the user's cart.
func (s *CartService) AddCollectionToCart(ctx context.Context, userID uuid.UUID, req dto.AddCollectionCartRequest) (dto.CollectionCartResponse, error) {
	item := &models.CollectionCartItem{
		UserID:     userID,
		Name:       req.Name,
		Department: req.Department,
		Collection: models.JSONB(req.Collection),
	}
	stored, err := s.collections.Create(ctx, item)
	if err != nil {
		return dto.CollectionCartResponse{}, err
	}
	return dto.NewCollectionCartItem(stored), nil
}

// UpdateCollectionCartItem replaces one entry's snapshot (slot removal).
func (s *CartService) UpdateCollectionCartItem(ctx context.Context, userID uuid.UUID, idRaw string, req dto.UpdateCollectionCartRequest) (dto.CollectionCartResponse, error) {
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return dto.CollectionCartResponse{}, utils.NewBadRequest("Invalid collection cart id", err)
	}
	stored, err := s.collections.UpdateCollection(ctx, userID, id, models.JSONB(req.Collection))
	if err != nil {
		return dto.CollectionCartResponse{}, err
	}
	return dto.NewCollectionCartItem(stored), nil
}

// DeleteCollectionCartItem removes one entry from the user's collection cart.
func (s *CartService) DeleteCollectionCartItem(ctx context.Context, userID uuid.UUID, idRaw string) error {
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return utils.NewBadRequest("Invalid collection cart id", err)
	}
	return s.collections.Delete(ctx, userID, id)
}

// ClearCollectionCart empties the user's collection cart.
func (s *CartService) ClearCollectionCart(ctx context.Context, userID uuid.UUID) error {
	return s.collections.Clear(ctx, userID)
}
