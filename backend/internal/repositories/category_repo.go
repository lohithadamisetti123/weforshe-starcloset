package repositories

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// CategoryWithCount pairs a category with how many live products sit in it.
type CategoryWithCount struct {
	models.Category
	ProductCount int64 `gorm:"column:product_count"`
}

// CategoryRepository is the category data access contract.
type CategoryRepository interface {
	List(ctx context.Context) ([]CategoryWithCount, error)
	GetByName(ctx context.Context, name models.CategoryName) (*models.Category, error)
}

type categoryRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewCategoryRepository constructs a CategoryRepository.
func NewCategoryRepository(db *gorm.DB, log *zap.Logger) CategoryRepository {
	return &categoryRepository{db: db, log: log}
}

// List returns every category with its product count.
//
// The count comes from a LEFT JOIN aggregate rather than a per-category COUNT
// in a loop. With ten categories the N+1 version would be survivable and still
// wrong — it is the pattern, not this instance, that eventually hurts.
func (r *categoryRepository) List(ctx context.Context) ([]CategoryWithCount, error) {
	var rows []CategoryWithCount

	err := r.db.WithContext(ctx).
		Model(&models.Category{}).
		Select("categories.*, COUNT(products.id) AS product_count").
		// The deleted_at predicate belongs in the JOIN condition, not in a
		// WHERE: in a WHERE it would filter away the categories that have no
		// live products at all, so an empty category would vanish from the
		// navigation rather than show a zero.
		Joins("LEFT JOIN products ON products.category_id = categories.id AND products.deleted_at IS NULL").
		Group("categories.id").
		Order("categories.name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repositories: listing categories: %w", err)
	}
	return rows, nil
}

// GetByName fetches one category.
func (r *categoryRepository) GetByName(ctx context.Context, name models.CategoryName) (*models.Category, error) {
	var c models.Category
	if err := r.db.WithContext(ctx).First(&c, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFound("Category")
		}
		return nil, fmt.Errorf("repositories: fetching category %q: %w", name, err)
	}
	return &c, nil
}
