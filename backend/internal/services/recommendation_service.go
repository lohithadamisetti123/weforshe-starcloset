package services

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// RecommendationService serves stored shortlists.
type RecommendationService struct {
	recs repositories.RecommendationRepository
	log  *zap.Logger
}

// NewRecommendationService constructs a RecommendationService.
func NewRecommendationService(recs repositories.RecommendationRepository, log *zap.Logger) *RecommendationService {
	return &RecommendationService{recs: recs, log: log}
}

// List returns a page of the shopper's past recommendations.
func (s *RecommendationService) List(ctx context.Context, userID uuid.UUID, p dto.PaginationQuery) ([]dto.RecommendationSummary, *utils.PageMeta, error) {
	p.Normalize()

	recs, total, err := s.recs.List(ctx, userID, p)
	if err != nil {
		return nil, nil, err
	}

	out := make([]dto.RecommendationSummary, 0, len(recs))
	for _, r := range recs {
		out = append(out, dto.NewRecommendationSummary(r))
	}

	meta := utils.NewPageMeta(p.Page, p.PageSize, total)
	return out, &meta, nil
}

// GetByID returns one stored shortlist, rendered as the assistant response the
// shopper originally received.
func (s *RecommendationService) GetByID(ctx context.Context, userID, id uuid.UUID) (*dto.AIRecommendationResponse, error) {
	rec, err := s.recs.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return dto.NewStoredRecommendation(rec), nil
}
