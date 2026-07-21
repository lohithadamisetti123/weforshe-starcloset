package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
)

// DashboardService computes the figures on the recommendations dashboard.
//
// The frontend currently derives these client-side from its fixture:
// averageAIMatch is literally `78 + avgRating * 4`, and averageBudget is the
// mean price of whatever happens to be on screen. Both are plausible-looking
// numbers with no referent. Serving them from here replaces invention with
// measurement — averageAIMatch becomes the mean of what the engine actually
// scored, and recommendationsToday becomes a count of things that happened.
type DashboardService struct {
	history  repositories.HistoryRepository
	recs     repositories.RecommendationRepository
	products repositories.ProductRepository
	cache    *repositories.Cache
	log      *zap.Logger
}

// NewDashboardService constructs a DashboardService.
func NewDashboardService(
	history repositories.HistoryRepository,
	recs repositories.RecommendationRepository,
	products repositories.ProductRepository,
	cache *repositories.Cache,
	log *zap.Logger,
) *DashboardService {
	return &DashboardService{
		history:  history,
		recs:     recs,
		products: products,
		cache:    cache,
		log:      log,
	}
}

// Stats returns the dashboard figures for a shopper.
//
// Nothing here is cached across users — the numbers are per-user by definition
// and the assistant invalidates the key on every write. The cost is a handful
// of indexed aggregates on a page that is not in any hot path.
func (s *DashboardService) Stats(ctx context.Context, userID uuid.UUID) (*dto.DashboardStats, error) {
	stats := &dto.DashboardStats{
		// The frontend renders these unconditionally, so they must be present
		// and sensible for a brand-new shopper with no history at all. A dash
		// beats an empty string; zero beats null.
		MostSelectedCategory: "—",
	}

	// "Today" is the shopper's day, not the server's. UTC midnight is 5:30am in
	// Bengaluru, which would show an Indian shopper their morning's searches
	// under yesterday. IST is hard-coded rather than negotiated per-request
	// because this product is for one market and pretending otherwise would be
	// a lie told in configuration.
	since := startOfDayIST(time.Now())

	hStats, err := s.history.Stats(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	stats.RecommendationsToday = int(hStats.TotalToday)
	stats.TotalRecommendations = hStats.TotalAllTime

	total, err := s.products.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalProducts = total

	avgMatch, err := s.recs.AverageMatch(ctx, userID)
	if err != nil {
		return nil, err
	}
	stats.AverageAIMatch = int(avgMatch + 0.5)

	tallies, err := s.recs.TopCategories(ctx, userID, 6)
	if err != nil {
		return nil, err
	}
	if len(tallies) > 0 {
		stats.MostSelectedCategory = tallies[0].Category
		stats.CategoryBreakdown = make([]dto.Slice, 0, len(tallies))
		for _, t := range tallies {
			stats.CategoryBreakdown = append(stats.CategoryBreakdown, dto.Slice{
				Name:  t.Category,
				Value: int(t.Total),
			})
		}
	}

	// averageBudget reports what the shopper's recommended products actually
	// cost, which is the number they can act on. Averaging their *stated*
	// budgets would answer a different and less useful question — and most
	// prompts state no budget at all, so the mean would be over a self-selected
	// minority.
	avgPrice, err := s.averageRecommendedPrice(ctx, userID)
	if err != nil {
		return nil, err
	}
	stats.AverageBudget = avgPrice

	return stats, nil
}

// averageRecommendedPrice is the mean price of everything this shopper has been
// shown.
func (s *DashboardService) averageRecommendedPrice(ctx context.Context, userID uuid.UUID) (int, error) {
	recs, _, err := s.recs.List(ctx, userID, dto.PaginationQuery{Page: 1, PageSize: 50})
	if err != nil {
		return 0, err
	}

	var (
		sum   int
		count int
	)
	for _, r := range recs {
		for i := range r.Items {
			if r.Items[i].Product == nil {
				continue
			}
			sum += r.Items[i].Product.Price
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	return sum / count, nil
}

// istOffset is India Standard Time, UTC+5:30. Constructed with FixedZone rather
// than LoadLocation("Asia/Kolkata") so the binary does not depend on the tzdata
// files being present in a scratch container — India has no DST, so the fixed
// offset is exactly correct rather than merely convenient.
var istOffset = time.FixedZone("IST", 5*60*60+30*60)

// startOfDayIST returns midnight IST for the day containing t.
func startOfDayIST(t time.Time) time.Time {
	local := t.In(istOffset)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, istOffset).UTC()
}
