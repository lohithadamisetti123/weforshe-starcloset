package services

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/personas"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// StarClosetService powers the V2 department journey: the department rosters
// (resolved against the persona store) and each shopper's independent
// per-department unlock progress. It touches neither the recommendation engine
// nor the cart — it only reads persona config and reads/advances progress rows.
type StarClosetService struct {
	store    personas.Store
	progress repositories.StarProgressRepository
	log      *zap.Logger
}

// NewStarClosetService constructs a StarClosetService.
func NewStarClosetService(store personas.Store, progress repositories.StarProgressRepository, log *zap.Logger) *StarClosetService {
	return &StarClosetService{store: store, progress: progress, log: log}
}

// Departments returns the three department rosters with their stylist cards.
func (s *StarClosetService) Departments() []dto.StarDepartmentResponse {
	return dto.NewStarDepartments(s.store)
}

// Progress returns a shopper's complete per-department unlock map.
func (s *StarClosetService) Progress(ctx context.Context, userID uuid.UUID) (map[string]int, error) {
	rows, err := s.progress.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewProgressMap(rows), nil
}

// Increment advances one department's counter and returns the updated map.
//
// The department is validated against the known roster keys so a caller cannot
// invent a department, and delta defaults to 1 — one cart-add is one step. Only
// the named department moves; the others are never read or written, which is
// what keeps the three journeys independent.
func (s *StarClosetService) Increment(ctx context.Context, userID uuid.UUID, department string, delta int) (map[string]int, error) {
	canon := personas.NormalizeDepartment(department)
	if canon == "" {
		return nil, utils.NewValidation("Unknown department", department)
	}
	if delta <= 0 {
		delta = 1
	}
	if _, err := s.progress.Increment(ctx, userID, canon, delta); err != nil {
		return nil, err
	}
	rows, err := s.progress.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	s.log.Info("star progress advanced",
		zap.String("userId", userID.String()),
		zap.String("department", canon),
		zap.Int("delta", delta))
	return dto.NewProgressMap(rows), nil
}
