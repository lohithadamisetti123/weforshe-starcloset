package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// PastOrderRepository is the data-access contract for purchase history.
//
// The interface is deliberately narrow and source-agnostic: List/Count/CreateMany
// say nothing about the orders being mock. Replacing the mock generator with a
// real Myntra order import means calling CreateMany with real rows (or reading
// from a different table behind the same interface) — nothing downstream, and
// certainly not the recommendation pipeline, needs to change.
type PastOrderRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]*models.PastOrder, error)
	Count(ctx context.Context, userID uuid.UUID) (int64, error)
	CreateMany(ctx context.Context, orders []*models.PastOrder) error
}

type pastOrderRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewPastOrderRepository constructs a PastOrderRepository.
func NewPastOrderRepository(db *gorm.DB, log *zap.Logger) PastOrderRepository {
	return &pastOrderRepository{db: db, log: log}
}

// List returns a user's orders, newest first.
func (r *pastOrderRepository) List(ctx context.Context, userID uuid.UUID) ([]*models.PastOrder, error) {
	var orders []*models.PastOrder
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("ordered_at DESC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("repositories: listing past orders: %w", err)
	}
	return orders, nil
}

// Count returns how many orders a user has. Used to decide whether the mock
// history needs seeding, so it does not get re-seeded on every profile edit.
func (r *pastOrderRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&models.PastOrder{}).
		Where("user_id = ?", userID).
		Count(&n).Error; err != nil {
		return 0, fmt.Errorf("repositories: counting past orders: %w", err)
	}
	return n, nil
}

// CreateMany inserts a batch of orders in one statement.
func (r *pastOrderRepository) CreateMany(ctx context.Context, orders []*models.PastOrder) error {
	if len(orders) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&orders).Error; err != nil {
		return fmt.Errorf("repositories: creating past orders: %w", err)
	}
	return nil
}
