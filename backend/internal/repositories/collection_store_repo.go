package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// SavedCollectionRepository is the per-user saved-collections contract.
//
// As with the cart, ownership is structural: every statement is filtered by
// user_id, and deletes match on (id, user_id) so a shopper holding another
// account's collection id still gets a plain "not found".
type SavedCollectionRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]*models.SavedCollection, error)
	Create(ctx context.Context, item *models.SavedCollection) (*models.SavedCollection, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type savedCollectionRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewSavedCollectionRepository constructs a SavedCollectionRepository.
func NewSavedCollectionRepository(db *gorm.DB, log *zap.Logger) SavedCollectionRepository {
	return &savedCollectionRepository{db: db, log: log}
}

func (r *savedCollectionRepository) List(ctx context.Context, userID uuid.UUID) ([]*models.SavedCollection, error) {
	var items []*models.SavedCollection
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("repositories: listing saved collections: %w", err)
	}
	return items, nil
}

// Create saves a collection. A duplicate (same user, name and department) is
// reported as a conflict so the caller can tell the shopper it is already
// saved, rather than silently creating a second copy.
func (r *savedCollectionRepository) Create(ctx context.Context, item *models.SavedCollection) (*models.SavedCollection, error) {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, utils.NewConflict("Collection already saved")
		}
		if isForeignKeyViolation(err) {
			return nil, utils.NewNotFound("User")
		}
		return nil, fmt.Errorf("repositories: saving collection: %w", err)
	}
	return item, nil
}

func (r *savedCollectionRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.SavedCollection{})
	if res.Error != nil {
		return fmt.Errorf("repositories: deleting saved collection: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return utils.NewNotFound("Saved collection")
	}
	return nil
}

// ─── Collection cart ─────────────────────────────────────────────────────────

// CollectionCartRepository is the per-user collection-cart contract.
type CollectionCartRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]*models.CollectionCartItem, error)
	Create(ctx context.Context, item *models.CollectionCartItem) (*models.CollectionCartItem, error)
	UpdateCollection(ctx context.Context, userID, id uuid.UUID, payload models.JSONB) (*models.CollectionCartItem, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	Clear(ctx context.Context, userID uuid.UUID) error
}

type collectionCartRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewCollectionCartRepository constructs a CollectionCartRepository.
func NewCollectionCartRepository(db *gorm.DB, log *zap.Logger) CollectionCartRepository {
	return &collectionCartRepository{db: db, log: log}
}

func (r *collectionCartRepository) List(ctx context.Context, userID uuid.UUID) ([]*models.CollectionCartItem, error) {
	var items []*models.CollectionCartItem
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("repositories: listing collection cart: %w", err)
	}
	return items, nil
}

func (r *collectionCartRepository) Create(ctx context.Context, item *models.CollectionCartItem) (*models.CollectionCartItem, error) {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, utils.NewConflict("Collection already in cart")
		}
		if isForeignKeyViolation(err) {
			return nil, utils.NewNotFound("User")
		}
		return nil, fmt.Errorf("repositories: adding collection to cart: %w", err)
	}
	return item, nil
}

// UpdateCollection replaces the stored snapshot for one entry — used when a
// shopper removes a single slot from a collection already in their cart.
func (r *collectionCartRepository) UpdateCollection(ctx context.Context, userID, id uuid.UUID, payload models.JSONB) (*models.CollectionCartItem, error) {
	res := r.db.WithContext(ctx).Model(&models.CollectionCartItem{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("collection", payload)
	if res.Error != nil {
		return nil, fmt.Errorf("repositories: updating collection cart item: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, utils.NewNotFound("Collection cart item")
	}

	var stored models.CollectionCartItem
	if err := r.db.WithContext(ctx).
		First(&stored, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, fmt.Errorf("repositories: reading back collection cart item: %w", err)
	}
	return &stored, nil
}

func (r *collectionCartRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.CollectionCartItem{})
	if res.Error != nil {
		return fmt.Errorf("repositories: deleting collection cart item: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return utils.NewNotFound("Collection cart item")
	}
	return nil
}

func (r *collectionCartRepository) Clear(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.CollectionCartItem{}).Error; err != nil {
		return fmt.Errorf("repositories: clearing collection cart: %w", err)
	}
	return nil
}
