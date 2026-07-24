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
)

// StarProgressRepository is the per-user, per-department unlock progress store.
//
// Like the cart, every method takes a userID and every statement filters on it,
// so a shopper can only ever read or advance their own progress. The
// (user_id, department) unique index makes each department's counter a single
// upsertable row — Increment is one ON CONFLICT statement, free of the
// read-then-write race two tabs would otherwise create.
type StarProgressRepository interface {
	// List returns every department row the user has (may be empty).
	List(ctx context.Context, userID uuid.UUID) ([]*models.StarProgress, error)
	// Increment advances one department's counter by delta (creating the row at
	// delta if absent) and returns the resulting value. delta is clamped so the
	// counter never drops below zero.
	Increment(ctx context.Context, userID uuid.UUID, department string, delta int) (*models.StarProgress, error)
}

type starProgressRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewStarProgressRepository constructs a StarProgressRepository.
func NewStarProgressRepository(db *gorm.DB, log *zap.Logger) StarProgressRepository {
	return &starProgressRepository{db: db, log: log}
}

// List returns the user's progress rows.
func (r *starProgressRepository) List(ctx context.Context, userID uuid.UUID) ([]*models.StarProgress, error) {
	var rows []*models.StarProgress
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("repositories: listing star progress: %w", err)
	}
	return rows, nil
}

// Increment advances one department's counter atomically.
func (r *starProgressRepository) Increment(ctx context.Context, userID uuid.UUID, department string, delta int) (*models.StarProgress, error) {
	seed := delta
	if seed < 0 {
		seed = 0
	}
	row := &models.StarProgress{
		ID:         uuid.New(),
		UserID:     userID,
		Department: department,
		Items:      seed,
	}

	// ON CONFLICT adds delta to the existing counter; GREATEST(...,0) enforces
	// the non-negative floor in the same statement so a decrement can never push
	// a department below zero.
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "department"}},
		DoUpdates: clause.Assignments(map[string]any{
			"items":      gorm.Expr("GREATEST(star_progress.items + ?, 0)", delta),
			"updated_at": time.Now(),
		}),
	}).Create(row).Error
	if err != nil {
		return nil, fmt.Errorf("repositories: incrementing star progress: %w", err)
	}

	// On conflict the generated row is fiction — read back the stored counter.
	var stored models.StarProgress
	if err := r.db.WithContext(ctx).
		First(&stored, "user_id = ? AND department = ?", userID, department).Error; err != nil {
		return nil, fmt.Errorf("repositories: reading back star progress: %w", err)
	}
	return &stored, nil
}
