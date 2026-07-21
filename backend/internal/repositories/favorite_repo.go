package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// FavoriteRepository is the saved-products data access contract.
type FavoriteRepository interface {
	Add(ctx context.Context, userID, productID uuid.UUID) (*models.FavoriteProduct, error)
	Remove(ctx context.Context, userID, productID uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID) ([]*models.FavoriteProduct, error)
	ProductIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	Exists(ctx context.Context, userID, productID uuid.UUID) (bool, error)
}

type favoriteRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewFavoriteRepository constructs a FavoriteRepository.
func NewFavoriteRepository(db *gorm.DB, log *zap.Logger) FavoriteRepository {
	return &favoriteRepository{db: db, log: log}
}

// Add saves a product for a user.
//
// Favouriting is idempotent: tapping a filled heart twice is a double-tap, not
// a request for two rows, and a 409 on the second tap would be a rude answer to
// a question the shopper did not ask. The unique index does the arbitration and
// OnConflict-DoNothing makes the second call a successful no-op.
func (r *favoriteRepository) Add(ctx context.Context, userID, productID uuid.UUID) (*models.FavoriteProduct, error) {
	fav := &models.FavoriteProduct{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
	}

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
		DoNothing: true,
	}).Create(fav).Error
	if err != nil {
		if isForeignKeyViolation(err) {
			// The FK is doing existence-checking for us: no extra SELECT to
			// confirm the product is real, and no race between that SELECT and
			// the insert.
			return nil, utils.NewNotFound("Product")
		}
		return nil, fmt.Errorf("repositories: adding favorite: %w", err)
	}

	// On a conflict nothing was written, so the generated ID is fiction. Read
	// back the row that actually exists.
	var stored models.FavoriteProduct
	if err := r.db.WithContext(ctx).Preload("Product.Category").
		First(&stored, "user_id = ? AND product_id = ?", userID, productID).Error; err != nil {
		return nil, fmt.Errorf("repositories: reading back favorite: %w", err)
	}
	return &stored, nil
}

// Remove deletes a saved product.
func (r *favoriteRepository) Remove(ctx context.Context, userID, productID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&models.FavoriteProduct{})
	if res.Error != nil {
		return fmt.Errorf("repositories: removing favorite: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		// Scoping the delete by user_id and reporting "not found" for a miss
		// means one shopper cannot probe for, or delete, another's favourites.
		return utils.NewNotFound("Favorite")
	}
	return nil
}

// List returns a user's saved products, newest first.
func (r *favoriteRepository) List(ctx context.Context, userID uuid.UUID) ([]*models.FavoriteProduct, error) {
	var favs []*models.FavoriteProduct
	// Preloading the product and its category is what keeps the favourites page
	// at three queries instead of 2N+1.
	if err := r.db.WithContext(ctx).Preload("Product.Category").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&favs).Error; err != nil {
		return nil, fmt.Errorf("repositories: listing favorites: %w", err)
	}
	return favs, nil
}

// ProductIDs returns just the IDs a user has saved, for cheap membership marking.
func (r *favoriteRepository) ProductIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	if err := r.db.WithContext(ctx).Model(&models.FavoriteProduct{}).
		Where("user_id = ?", userID).
		Pluck("product_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("repositories: listing favorite ids: %w", err)
	}
	return ids, nil
}

// Exists reports whether a user has saved a product.
func (r *favoriteRepository) Exists(ctx context.Context, userID, productID uuid.UUID) (bool, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&models.FavoriteProduct{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&n).Error; err != nil {
		return false, fmt.Errorf("repositories: checking favorite: %w", err)
	}
	return n > 0, nil
}
