// Package personalization assembles a recommendation.PersonalizationContext for
// a shopper from whatever sources of preference and behaviour are available.
//
// It is the seam between "where personal data lives" and "how the ranker uses
// it". Today the sources are the onboarding profile and a mock past-orders
// table; tomorrow they can be a real Myntra account, wishlist, browsing history
// and recently-viewed products. Crucially, all of that variation is contained
// here: the builder's output is always a recommendation.PersonalizationContext,
// and the recommendation pipeline consumes that one type. Adding a new signal
// source is a change to this package and (at most) a new field on the context —
// never a change to the engine, the scorer, or the collection builder.
//
// Dependency direction is one-way: this package imports recommendation (for the
// context type) and repositories/models (for the data). recommendation imports
// nothing from here, so there is no cycle and the engine stays ignorant of
// where its personalization input came from.
package personalization

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
)

// maxPurchaseSignals bounds how many distinct purchased brands/categories feed
// the boost, so a shopper with a long history does not get an unbounded pile of
// weak signals — the most recent purchases are the most predictive anyway.
const maxPurchaseSignals = 8

// Builder loads a shopper's personalization context from persisted sources.
type Builder struct {
	profiles repositories.UserProfileRepository
	orders   repositories.PastOrderRepository
	log      *zap.Logger
}

// NewBuilder constructs a Builder.
func NewBuilder(
	profiles repositories.UserProfileRepository,
	orders repositories.PastOrderRepository,
	log *zap.Logger,
) *Builder {
	return &Builder{profiles: profiles, orders: orders, log: log}
}

// Build returns the personalization context for a user, or nil when there is
// nothing to personalize on (no profile and no orders). Returning nil rather
// than an empty context is deliberate: nil is the signal the engine uses to
// skip the boost entirely, so an un-onboarded shopper gets the exact,
// unmodified ranking.
//
// A data-access failure is returned as an error, but callers are expected to
// treat personalization as best-effort: the correct behaviour on failure is to
// proceed with a nil context (an un-personalized but correct result), never to
// fail the shopper's request. The assistant service does exactly that.
func (b *Builder) Build(ctx context.Context, userID uuid.UUID) (*recommendation.PersonalizationContext, error) {
	profile, err := b.profiles.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	orders, err := b.orders.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	if profile == nil && len(orders) == 0 {
		return nil, nil
	}

	pc := &recommendation.PersonalizationContext{}

	if profile != nil {
		pc.City = profile.City
		pc.PreferredBrands = nonEmpty(profile.BrandsList())
		pc.PreferredColors = nonEmpty(profile.ColorsList())
		pc.PreferredStyles = parseStyles(profile.StylesList())
		pc.PreferredCategories = parseCategories(profile.CategoriesList())
	}

	if len(orders) > 0 {
		pc.PurchasedBrands, pc.PurchasedCategories = purchaseSignals(orders)
	}

	if pc.IsEmpty() {
		return nil, nil
	}
	return pc, nil
}

// purchaseSignals extracts the distinct brands and categories a shopper has
// bought, most recent first, capped at maxPurchaseSignals each. Orders arrive
// newest-first from the repository, so a simple in-order dedupe preserves
// recency.
func purchaseSignals(orders []*models.PastOrder) ([]string, []models.CategoryName) {
	brandSeen := map[string]struct{}{}
	catSeen := map[string]struct{}{}
	brands := make([]string, 0, maxPurchaseSignals)
	cats := make([]models.CategoryName, 0, maxPurchaseSignals)

	for _, o := range orders {
		if o.Brand != "" && len(brands) < maxPurchaseSignals {
			if _, ok := brandSeen[o.Brand]; !ok {
				brandSeen[o.Brand] = struct{}{}
				brands = append(brands, o.Brand)
			}
		}
		if o.Category != "" && len(cats) < maxPurchaseSignals {
			if c, ok := models.ParseCategory(o.Category); ok {
				if _, seen := catSeen[string(c)]; !seen {
					catSeen[string(c)] = struct{}{}
					cats = append(cats, c)
				}
			}
		}
	}
	return brands, cats
}

// parseStyles maps stored preference strings onto the canonical StyleTag
// vocabulary, dropping anything unrecognised. Reusing models.ParseStyle means a
// stored "Boho" resolves to the same StyleTag the extractor and scorer use.
func parseStyles(raw []string) []models.StyleTag {
	out := make([]models.StyleTag, 0, len(raw))
	seen := map[models.StyleTag]struct{}{}
	for _, s := range raw {
		if st, ok := models.ParseStyle(s); ok {
			if _, dup := seen[st]; !dup {
				seen[st] = struct{}{}
				out = append(out, st)
			}
		}
	}
	return out
}

// parseCategories maps stored preference strings onto the canonical
// CategoryName vocabulary, dropping anything unrecognised.
func parseCategories(raw []string) []models.CategoryName {
	out := make([]models.CategoryName, 0, len(raw))
	seen := map[models.CategoryName]struct{}{}
	for _, c := range raw {
		if cat, ok := models.ParseCategory(c); ok {
			if _, dup := seen[cat]; !dup {
				seen[cat] = struct{}{}
				out = append(out, cat)
			}
		}
	}
	return out
}

// nonEmpty trims blanks out of a stored string list.
func nonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
