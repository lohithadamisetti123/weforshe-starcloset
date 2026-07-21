package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// HistoryRepository is the chat history data access contract.
type HistoryRepository interface {
	Create(ctx context.Context, entry *models.ChatHistory) error
	List(ctx context.Context, userID uuid.UUID, p dto.PaginationQuery) ([]*models.ChatHistory, int64, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*models.ChatHistory, error)
	UpdateStatus(ctx context.Context, userID, id uuid.UUID, status models.HistoryStatus) (*models.ChatHistory, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	Stats(ctx context.Context, userID uuid.UUID, since time.Time) (*HistoryStats, error)
}

// HistoryStats is the aggregate behind the dashboard cards.
type HistoryStats struct {
	TotalToday   int64
	TotalAllTime int64
	AvgProducts  float64
}

type historyRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewHistoryRepository constructs a HistoryRepository.
func NewHistoryRepository(db *gorm.DB, log *zap.Logger) HistoryRepository {
	return &historyRepository{db: db, log: log}
}

// Create inserts a history entry.
func (r *historyRepository) Create(ctx context.Context, entry *models.ChatHistory) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("repositories: creating history entry: %w", err)
	}
	return nil
}

// List returns a user's history, newest first.
//
// Every query in this repository is scoped by user_id in the same statement as
// the id lookup, never as a check afterwards. That is what makes an
// authorisation bug here structurally difficult rather than merely unlikely:
// there is no code path that reads a row before deciding whether it may.
func (r *historyRepository) List(ctx context.Context, userID uuid.UUID, p dto.PaginationQuery) ([]*models.ChatHistory, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.ChatHistory{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repositories: counting history: %w", err)
	}
	if total == 0 {
		return []*models.ChatHistory{}, 0, nil
	}

	var entries []*models.ChatHistory
	// idx_history_user_created covers exactly this (user_id, created_at DESC)
	// access pattern, so the history page stays an index scan as it grows.
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id ASC").
		Limit(p.PageSize).Offset(p.Offset()).
		Find(&entries).Error; err != nil {
		return nil, 0, fmt.Errorf("repositories: listing history: %w", err)
	}
	return entries, total, nil
}

// GetByID fetches one entry belonging to a user.
func (r *historyRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*models.ChatHistory, error) {
	var entry models.ChatHistory
	err := r.db.WithContext(ctx).First(&entry, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// "Not found" rather than "forbidden" for someone else's entry: a
			// 403 would confirm the ID exists, which is information the caller
			// has not earned.
			return nil, utils.NewNotFound("History entry")
		}
		return nil, fmt.Errorf("repositories: fetching history entry: %w", err)
	}
	return &entry, nil
}

// UpdateStatus moves an entry between Completed, Saved and Purchased.
func (r *historyRepository) UpdateStatus(ctx context.Context, userID, id uuid.UUID, status models.HistoryStatus) (*models.ChatHistory, error) {
	res := r.db.WithContext(ctx).Model(&models.ChatHistory{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", status)
	if res.Error != nil {
		return nil, fmt.Errorf("repositories: updating history status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, utils.NewNotFound("History entry")
	}
	return r.GetByID(ctx, userID, id)
}

// Delete soft-deletes an entry.
func (r *historyRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.ChatHistory{})
	if res.Error != nil {
		return fmt.Errorf("repositories: deleting history entry: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return utils.NewNotFound("History entry")
	}
	return nil
}

// Stats aggregates a user's history for the dashboard.
func (r *historyRepository) Stats(ctx context.Context, userID uuid.UUID, since time.Time) (*HistoryStats, error) {
	stats := &HistoryStats{}

	if err := r.db.WithContext(ctx).Model(&models.ChatHistory{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&stats.TotalToday).Error; err != nil {
		return nil, fmt.Errorf("repositories: counting today's history: %w", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.ChatHistory{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalAllTime).Error; err != nil {
		return nil, fmt.Errorf("repositories: counting history: %w", err)
	}

	// AVG over zero rows is NULL, not 0, so the destination is a pointer.
	// Scanning NULL into a float64 is an error, and the dashboard of a
	// brand-new user is precisely the zero-row case.
	var avg *float64
	if err := r.db.WithContext(ctx).Model(&models.ChatHistory{}).
		Where("user_id = ?", userID).
		Select("AVG(products_recommended)").Scan(&avg).Error; err != nil {
		return nil, fmt.Errorf("repositories: averaging history: %w", err)
	}
	if avg != nil {
		stats.AvgProducts = *avg
	}

	return stats, nil
}
