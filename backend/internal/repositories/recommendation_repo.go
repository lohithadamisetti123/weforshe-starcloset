package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// RecommendationRepository is the data access contract for stored shortlists.
type RecommendationRepository interface {
	Create(ctx context.Context, rec *models.Recommendation) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*models.Recommendation, error)
	List(ctx context.Context, userID uuid.UUID, p dto.PaginationQuery) ([]*models.Recommendation, int64, error)
	TopCategories(ctx context.Context, userID uuid.UUID, limit int) ([]CategoryTally, error)
	AverageMatch(ctx context.Context, userID uuid.UUID) (float64, error)
}

// CategoryTally counts how often a category has been recommended.
type CategoryTally struct {
	Category string `gorm:"column:category"`
	Total    int64  `gorm:"column:total"`
}

type recommendationRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewRecommendationRepository constructs a RecommendationRepository.
func NewRecommendationRepository(db *gorm.DB, log *zap.Logger) RecommendationRepository {
	return &recommendationRepository{db: db, log: log}
}

// Create persists a recommendation together with its ranked items.
//
// The parent and its items go in one transaction. A recommendation row with no
// items is not a partial success — it is a shortlist that recommends nothing,
// which the history page would happily render as an empty grid with a summary
// describing six products.
//
// GORM's association upsert would do this in one call, but writing the items
// explicitly keeps the ordering and the ID assignment visible rather than
// depending on reflection over a slice field.
func (r *recommendationRepository) Create(ctx context.Context, rec *models.Recommendation) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		items := rec.Items
		rec.Items = nil

		if err := tx.Create(rec).Error; err != nil {
			return fmt.Errorf("creating recommendation: %w", err)
		}

		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].ID = uuid.New()
			items[i].RecommendationID = rec.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return fmt.Errorf("creating recommendation items: %w", err)
		}

		rec.Items = items
		return nil
	})
	if err != nil {
		return fmt.Errorf("repositories: %w", err)
	}
	return nil
}

// GetByID fetches a stored recommendation with its products, in rank order.
//
// The nested Preload carries an explicit ORDER BY rank. Without it GORM would
// load the items in whatever order Postgres returns them, and the replayed
// shortlist would silently disagree with the one the shopper originally saw —
// same six products, different best answer.
func (r *recommendationRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*models.Recommendation, error) {
	var rec models.Recommendation
	err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("recommendation_items.rank ASC")
		}).
		Preload("Items.Product").
		Preload("Items.Product.Category").
		First(&rec, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFound("Recommendation")
		}
		return nil, fmt.Errorf("repositories: fetching recommendation: %w", err)
	}
	return &rec, nil
}

// List returns a user's recommendations, newest first.
func (r *recommendationRepository) List(ctx context.Context, userID uuid.UUID, p dto.PaginationQuery) ([]*models.Recommendation, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.Recommendation{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repositories: counting recommendations: %w", err)
	}
	if total == 0 {
		return []*models.Recommendation{}, 0, nil
	}

	var recs []*models.Recommendation
	if err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("recommendation_items.rank ASC")
		}).
		Preload("Items.Product").
		Preload("Items.Product.Category").
		Where("user_id = ?", userID).
		Order("created_at DESC, id ASC").
		Limit(p.PageSize).Offset(p.Offset()).
		Find(&recs).Error; err != nil {
		return nil, 0, fmt.Errorf("repositories: listing recommendations: %w", err)
	}
	return recs, total, nil
}

// TopCategories reports which categories a user gets recommended most.
//
// This backs the dashboard's "most selected category" card. It joins through to
// the live product rows rather than denormalising a category string onto the
// item, so a re-categorised product is reflected without a backfill.
func (r *recommendationRepository) TopCategories(ctx context.Context, userID uuid.UUID, limit int) ([]CategoryTally, error) {
	if limit <= 0 {
		limit = 5
	}

	var rows []CategoryTally
	err := r.db.WithContext(ctx).
		Model(&models.RecommendationItem{}).
		Select("categories.name AS category, COUNT(*) AS total").
		Joins("JOIN recommendations ON recommendations.id = recommendation_items.recommendation_id").
		Joins("JOIN products ON products.id = recommendation_items.product_id").
		Joins("JOIN categories ON categories.id = products.category_id").
		Where("recommendations.user_id = ? AND recommendations.deleted_at IS NULL", userID).
		Group("categories.name").
		// The count is the ranking; the name is only a tiebreak, so that two
		// categories on the same count don't swap places between refreshes.
		Order("total DESC, categories.name ASC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("repositories: tallying categories: %w", err)
	}
	return rows, nil
}

// AverageMatch returns the mean match percentage across a user's recommendations.
func (r *recommendationRepository) AverageMatch(ctx context.Context, userID uuid.UUID) (float64, error) {
	var avg *float64
	err := r.db.WithContext(ctx).
		Model(&models.RecommendationItem{}).
		Select("AVG(recommendation_items.match_percent)").
		Joins("JOIN recommendations ON recommendations.id = recommendation_items.recommendation_id").
		Where("recommendations.user_id = ? AND recommendations.deleted_at IS NULL", userID).
		Scan(&avg).Error
	if err != nil {
		return 0, fmt.Errorf("repositories: averaging match percent: %w", err)
	}
	if avg == nil {
		return 0, nil
	}
	return *avg, nil
}
