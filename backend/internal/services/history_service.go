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

// HistoryService serves the shopper's past conversations.
type HistoryService struct {
	history repositories.HistoryRepository
	cache   *repositories.Cache
	log     *zap.Logger
}

// NewHistoryService constructs a HistoryService.
func NewHistoryService(history repositories.HistoryRepository, cache *repositories.Cache, log *zap.Logger) *HistoryService {
	return &HistoryService{history: history, cache: cache, log: log}
}

// List returns a page of history entries.
//
// History is not cached. It is per-user, it changes on every assistant call,
// and a stale history page is a shopper watching their own action fail to
// appear — the one place where a 10-minute TTL would be read as a bug rather
// than as speed.
func (s *HistoryService) List(ctx context.Context, userID uuid.UUID, p dto.PaginationQuery) ([]dto.HistoryEntry, *utils.PageMeta, error) {
	p.Normalize()

	entries, total, err := s.history.List(ctx, userID, p)
	if err != nil {
		return nil, nil, err
	}

	meta := utils.NewPageMeta(p.Page, p.PageSize, total)
	return dto.NewHistoryEntries(entries), &meta, nil
}

// Create records an entry the client asked us to store.
//
// The assistant writes its own history, so this exists for the client to log a
// prompt it handled itself — and, more usefully, for the Postman collection and
// tests to seed history without driving the whole AI pipeline.
func (s *HistoryService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateHistoryRequest) (*dto.HistoryEntry, error) {
	status := models.HistoryStatusCompleted
	if req.Status != "" {
		parsed, ok := models.ParseHistoryStatus(req.Status)
		if !ok {
			return nil, utils.NewValidation("Unknown status", map[string]any{
				"status": req.Status,
				"valid":  models.AllHistoryStatuses,
			})
		}
		status = parsed
	}

	entry := &models.ChatHistory{
		ID:                  uuid.New(),
		UserID:              userID,
		Prompt:              req.Prompt,
		BudgetLabel:         req.Budget,
		Occasion:            req.Occasion,
		Style:               req.Style,
		ProductsRecommended: req.ProductsRecommended,
		Status:              status,
	}

	if err := s.history.Create(ctx, entry); err != nil {
		return nil, err
	}

	s.cache.Delete(ctx, repositories.StatsKey())

	out := dto.NewHistoryEntry(entry)
	return &out, nil
}

// UpdateStatus moves an entry between Completed, Saved and Purchased.
func (s *HistoryService) UpdateStatus(ctx context.Context, userID, id uuid.UUID, req dto.UpdateHistoryRequest) (*dto.HistoryEntry, error) {
	status, ok := models.ParseHistoryStatus(req.Status)
	if !ok {
		return nil, utils.NewValidation("Unknown status", map[string]any{
			"status": req.Status,
			"valid":  models.AllHistoryStatuses,
		})
	}

	entry, err := s.history.UpdateStatus(ctx, userID, id, status)
	if err != nil {
		return nil, err
	}

	out := dto.NewHistoryEntry(entry)
	return &out, nil
}

// Delete removes an entry.
func (s *HistoryService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	if err := s.history.Delete(ctx, userID, id); err != nil {
		return err
	}
	s.cache.Delete(ctx, repositories.StatsKey())
	return nil
}
