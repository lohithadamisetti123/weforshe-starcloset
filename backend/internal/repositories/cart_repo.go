package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// CartRepository is the per-user product-cart data access contract.
//
// Every method takes a userID and every statement filters on it. That is the
// ownership guarantee: there is no method that can read or mutate a cart row
// without naming the owner, so a handler cannot accidentally expose one
// shopper's cart to another even if it were passed an id belonging to someone
// else.
type CartRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]*models.CartItem, error)
	Add(ctx context.Context, userID, productID uuid.UUID) (*models.CartItem, error)
	Remove(ctx context.Context, userID, productID uuid.UUID) error
	Clear(ctx context.Context, userID uuid.UUID) error
}

type cartRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewCartRepository constructs a CartRepository.
func NewCartRepository(db *gorm.DB, log *zap.Logger) CartRepository {
	return &cartRepository{db: db, log: log}
}

// List returns a user's cart, newest first, with products preloaded so the cart
// page renders in a constant number of queries rather than 2N+1.
func (r *cartRepository) List(ctx context.Context, userID uuid.UUID) ([]*models.CartItem, error) {
	var items []*models.CartItem
	if err := r.db.WithContext(ctx).Preload("Product.Category").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("repositories: listing cart: %w", err)
	}
	return items, nil
}

// Add puts a product in a user's cart, incrementing the quantity when it is
// already there.
//
// The (user_id, product_id) unique index arbitrates: a second add is an
// increment, not a duplicate row and not a 409. Doing it in one ON CONFLICT
// statement also removes the read-then-write race between two tabs.
func (r *cartRepository) Add(ctx context.Context, userID, productID uuid.UUID) (*models.CartItem, error) {
	item := &models.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
		Quantity:  1,
	}

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"quantity":   gorm.Expr("cart_items.quantity + 1"),
			"updated_at": time.Now(),
		}),
	}).Create(item).Error
	if err != nil {
		if isForeignKeyViolation(err) {
			// The FK does the existence check: no extra SELECT, and no race
			// between that SELECT and the insert.
			return nil, utils.NewNotFound("Product")
		}
		return nil, fmt.Errorf("repositories: adding to cart: %w", err)
	}

	// On conflict the generated ID is fiction — read back the row that exists.
	var stored models.CartItem
	if err := r.db.WithContext(ctx).Preload("Product.Category").
		First(&stored, "user_id = ? AND product_id = ?", userID, productID).Error; err != nil {
		return nil, fmt.Errorf("repositories: reading back cart item: %w", err)
	}
	return &stored, nil
}

// Remove deletes a product from a user's cart.
func (r *cartRepository) Remove(ctx context.Context, userID, productID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&models.CartItem{})
	if res.Error != nil {
		return fmt.Errorf("repositories: removing cart item: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		// Scoping the delete by user_id and reporting a plain "not found" for a
		// miss means one shopper cannot probe for, or delete, another's cart.
		return utils.NewNotFound("Cart item")
	}
	return nil
}

// Clear empties a user's cart.
func (r *cartRepository) Clear(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.CartItem{}).Error; err != nil {
		return fmt.Errorf("repositories: clearing cart: %w", err)
	}
	return nil
}
