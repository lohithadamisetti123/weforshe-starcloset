package recommendation

import (
	"context"
	"sort"
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/config"
	"github.com/myntra/bharat-ai-backend/internal/models"
)

// CandidateSource supplies the products the engine ranks.
//
// The engine depends on this narrow interface rather than on a GORM repository
// so that the entire ranking pipeline is testable with a slice of structs and
// no database. That is not a hypothetical benefit: the engine's unit tests are
// the ones that actually protect relevance, and they must be fast enough to run
// on every save.
type CandidateSource interface {
	// Candidates returns the pool to rank, already narrowed by any hard filters
	// the query supports. Implementations must return a generous pool — the
	// engine, not the datastore, decides what is relevant.
	Candidates(ctx context.Context, f CandidateFilter) ([]*models.Product, error)
}

// CandidateFilter is the set of hard constraints pushed down to the datastore.
//
// Only genuinely disqualifying constraints belong here. Everything graded
// belongs in the scorer: pushing a soft preference into SQL turns a
// "prefers red" into a "must be red" and produces empty grids, which is the
// single most common way a recommender fails its user.
type CandidateFilter struct {
	Gender   *models.Gender
	MaxPrice *int
	MinPrice *int
	// Categories is a hard allow-list, unlike Intent.Category (a soft scoring
	// signal used everywhere else). It exists for collection slot-filling
	// (see collection.go): when building a "Footwear" slot, the candidate
	// pool must actually be footwear, not merely biased toward it.
	Categories      []models.CategoryName
	ExcludeColors   []string
	ExcludeStyles   []models.StyleTag
	ExcludeBrands   []string
	ExcludeProducts []string
	Limit           int
	// StrictBudget disables the price-drop relaxation step so collection
	// slots never return a product that exceeds the remaining budget.
	StrictBudget bool
}

// Engine ranks catalogue products against a shopper's intent.
type Engine struct {
	source CandidateSource
	scorer *Scorer
	cfg    config.RecommendationConfig
}

// NewEngine wires an engine from its dependencies.
func NewEngine(source CandidateSource, cfg config.RecommendationConfig) *Engine {
	return &Engine{
		source: source,
		scorer: NewScorer(cfg.Weights),
		cfg:    cfg,
	}
}

// Recommend runs the full pipeline for an intent.
//
// The method's contract is that it returns a usable shortlist for *any* intent.
// An empty result is only ever acceptable when the catalogue itself is empty;
// anything else is a bug, because a shopper who gets six imperfect options can
// refine, while a shopper who gets zero just leaves.
func (e *Engine) Recommend(ctx context.Context, intent *Intent) (*Result, error) {
	if intent == nil {
		intent = NewIntent("")
	}
	if len(intent.Keywords) == 0 && intent.Query != "" {
		intent.Keywords = ExtractKeywords(intent.Query)
	}

	pool, relaxed, err := e.retrieve(ctx, intent)
	if err != nil {
		return nil, err
	}

	// Narrow to the requested product type (watch, sneakers, saree…) before
	// scoring, so a "watch" query ranks watches rather than the best-rated
	// accessory of any kind. Soft: if nothing matches, the pool is unchanged.
	pool = FilterByType(pool, intent)

	result := &Result{
		Intent:        intent,
		PoolSize:      len(pool),
		Relaxed:       relaxed,
		BudgetLabel:   BudgetLabel(intent.Budget),
		OccasionLabel: OccasionLabel(intent),
		StyleLabel:    StyleLabel(intent),
	}

	if len(pool) == 0 {
		result.Summary = BuildSummary(intent, nil)
		result.Reason = BuildReason(intent, nil, relaxed)
		result.Products = []ScoredProduct{}
		return result, nil
	}

	// Stage 3 — score.
	scored := e.scorer.ScoreAll(intent, pool)

	// Stage 3b — personalize (additive). This is the single seam through which
	// per-shopper preferences and history nudge the ranking. It is a no-op when
	// no PersonalizationContext is attached to the intent, so every existing
	// path (and every test) is unaffected; the scoring in ScoreAll above is
	// never modified. The boost is applied here — before the sort and MMR — so
	// those stages, and the match-percent calibration, all reflect it without
	// any change to their own logic. See personalize.go.
	ApplyPersonalization(scored, intent)

	sortByScore(scored)

	// Stage 4 — diversify. An explicit sort request means the shopper wants a
	// list in that order, so MMR is skipped: reordering under someone who asked
	// for "cheapest first" would be a straightforward bug.
	var selected []ScoredProduct
	if intent.Sort != "" && intent.Sort != SortRelevance {
		selected = e.applySortHint(scored, intent.Sort)
	} else {
		selected = Diversify(scored, e.cfg.TopN, e.cfg.DiversityLambda)
	}

	// Stage 5 — calibrate and explain.
	ApplyCalibration(selected, intent)
	for i := range selected {
		selected[i].Reasons = ReasonsFor(&selected[i], intent)
	}

	result.Products = selected
	result.Summary = BuildSummary(intent, selected)
	result.Reason = BuildReason(intent, selected, relaxed)
	return result, nil
}

// retrieve fetches the candidate pool, progressively relaxing hard constraints
// until the shortlist can be filled.
//
// The relaxation ladder is ordered by how much a shopper minds losing each
// constraint. Gender is last because it is the one people notice and object to
// most; a hard price ceiling goes first because "under ₹3000" is nearly always
// a preference rather than a wallet limit, and a ₹3,100 perfect match is a
// better answer than a ₹900 mediocre one.
//
// Every relaxation is recorded and surfaced in the response, so the engine
// never quietly ignores what it was told.
func (e *Engine) retrieve(ctx context.Context, intent *Intent) ([]*models.Product, []string, error) {
	limit := e.cfg.CandidatePool
	if limit <= 0 {
		limit = 200
	}

	base := CandidateFilter{
		ExcludeColors:   intent.ExcludeColors,
		ExcludeStyles:   intent.ExcludeStyles,
		ExcludeBrands:   intent.ExcludeBrands,
		ExcludeProducts: intent.ExcludeProducts,
		Categories:      intent.CategoryIn,
		StrictBudget:    intent.StrictBudget,
		Limit:           limit,
	}
	if intent.Gender != nil {
		base.Gender = intent.Gender
	}
	if intent.Budget.HasMax() {
		var hard int
		if intent.StrictBudget {
			// Collection slot-filling: use the exact remaining budget as the
			// ceiling so the pool never contains items the honesty check would
			// skip anyway. This avoids the scorer picking the best-rated item
			// only for it to be blocked, leaving the slot empty.
			hard = *intent.Budget.Max
		} else {
			// Normal search: retrieve slightly beyond the stated ceiling so the
			// scorer gets to decide on near-misses rather than the WHERE clause.
			hard = int(float64(*intent.Budget.Max) * hardCeilingMultiplier)
		}
		base.MaxPrice = &hard
	}
	if intent.Budget.HasMin() {
		base.MinPrice = intent.Budget.Min
	}

	relaxed := make([]string, 0, 3)

	pool, err := e.retrievePool(ctx, base, intent)
	if err != nil {
		return nil, nil, err
	}
	if len(pool) >= e.cfg.TopN {
		return pool, relaxed, nil
	}

	// Ladder step 1 — drop the price bounds (unless StrictBudget is set,
	// which collection slot-filling uses to guarantee the pick stays within
	// the remaining budget).
	if !base.StrictBudget && (base.MaxPrice != nil || base.MinPrice != nil) {
		f := base
		f.MaxPrice, f.MinPrice = nil, nil
		if pool, err = e.source.Candidates(ctx, f); err != nil {
			return nil, nil, err
		}
		relaxed = append(relaxed, "the price range")
		if len(pool) >= e.cfg.TopN {
			return pool, relaxed, nil
		}
		base = f
	}

	// Ladder step 2 — drop the exclusions.
	if len(base.ExcludeColors) > 0 || len(base.ExcludeStyles) > 0 || len(base.ExcludeBrands) > 0 {
		f := base
		f.ExcludeColors, f.ExcludeStyles, f.ExcludeBrands = nil, nil, nil
		if pool, err = e.source.Candidates(ctx, f); err != nil {
			return nil, nil, err
		}
		relaxed = append(relaxed, "some of your exclusions")
		if len(pool) >= e.cfg.TopN {
			return pool, relaxed, nil
		}
		base = f
	}

	// Gender is intentionally NOT part of the relaxation ladder. Showing a
	// shopper the wrong gender ("men's shoes in a women's collection") is worse
	// than returning fewer results, so the gender filter is never dropped —
	// even as a last resort. If the gendered pool is thin, we return what we
	// have and let the slot be skipped upstream.
	return pool, relaxed, nil
}

// retrievePool fetches the primary candidate pool, guaranteeing that a named
// category is actually represented in it.
//
// Why this exists: CandidateSource orders by rating and hard-caps at
// cfg.CandidatePool. For a small catalogue that cap never binds, so the general
// pool contains every category and the scorer can surface the right one. As the
// catalogue grows past the cap, that stops being true — a "red saree" query
// pulls the top-N highest-rated products of the shopper's gender, and if sarees
// rate lower on average than, say, dresses, the pool can arrive with few sarees
// or none. The scorer then ranks whatever it was handed and cannot recover a
// product it never received. The symptom is "accurate for a while, then subtly
// wrong as we add inventory" — the single most common way this class of ranker
// degrades at scale.
//
// The fix keeps category SOFT (no hard WHERE that would break cross-category
// answers and empty grids) while making retrieval category-AWARE: when the
// shopper named a category, we additionally pull a category-scoped slice and
// union it into the general pool. The scorer still decides final order; it is
// simply guaranteed to see the relevant category. This only applies to ordinary
// search — collection slot-filling already hard-scopes via CandidateFilter
// .Categories (CategoryIn), so base.Categories being set means "leave it alone".
func (e *Engine) retrievePool(ctx context.Context, base CandidateFilter, intent *Intent) ([]*models.Product, error) {
	pool, err := e.source.Candidates(ctx, base)
	if err != nil {
		return nil, err
	}

	// Only blend for ordinary search with a named category. Collection
	// slot-filling (base.Categories set) is already category-scoped.
	if intent == nil || intent.Category == nil || len(base.Categories) > 0 {
		return pool, nil
	}

	catFilter := base
	catFilter.Categories = []models.CategoryName{*intent.Category}
	// A generous slice: enough that the scorer has real choice within the
	// category, capped so a huge single-category catalogue can't dominate the
	// blended pool.
	catFilter.Limit = e.cfg.TopN * 4
	if catFilter.Limit < 40 {
		catFilter.Limit = 40
	}
	catPool, err := e.source.Candidates(ctx, catFilter)
	if err != nil {
		// A failed enrichment fetch must not fail the whole request — the
		// general pool is already a valid answer.
		return pool, nil
	}

	return mergeUniqueProducts(catPool, pool), nil
}

// mergeUniqueProducts concatenates two candidate slices, dropping duplicates by
// product ID. The first slice's ordering is preserved and takes precedence, so
// callers put the pool they most want represented first.
func mergeUniqueProducts(first, second []*models.Product) []*models.Product {
	seen := make(map[string]struct{}, len(first)+len(second))
	out := make([]*models.Product, 0, len(first)+len(second))
	for _, list := range [][]*models.Product{first, second} {
		for _, p := range list {
			if p == nil {
				continue
			}
			id := p.ID.String()
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, p)
		}
	}
	return out
}

// applySortHint honours an explicit ordering request, returning the top N.
func (e *Engine) applySortHint(scored []ScoredProduct, hint SortHint) []ScoredProduct {
	// Relevance still selects *which* products qualify — a shopper asking for
	// "the cheapest wedding saree" wants cheap sarees, not the cheapest thing
	// in the catalogue. Only the presentation order of that qualified set is
	// then re-sorted.
	poolSize := e.cfg.TopN * 3
	if poolSize > len(scored) {
		poolSize = len(scored)
	}
	pool := make([]ScoredProduct, poolSize)
	copy(pool, scored[:poolSize])

	switch hint {
	case SortPriceAsc:
		sort.SliceStable(pool, func(i, j int) bool {
			return pool[i].Product.Price < pool[j].Product.Price
		})
	case SortPriceDesc:
		sort.SliceStable(pool, func(i, j int) bool {
			return pool[i].Product.Price > pool[j].Product.Price
		})
	case SortRating:
		sort.SliceStable(pool, func(i, j int) bool {
			return pool[i].Product.Rating > pool[j].Product.Rating
		})
	case SortDiscount:
		sort.SliceStable(pool, func(i, j int) bool {
			return pool[i].Product.EffectiveDiscount() > pool[j].Product.EffectiveDiscount()
		})
	case SortNewest:
		sort.SliceStable(pool, func(i, j int) bool {
			return pool[i].Product.CreatedAt.After(pool[j].Product.CreatedAt)
		})
	}

	n := e.cfg.TopN
	if n > len(pool) {
		n = len(pool)
	}
	out := pool[:n]
	for i := range out {
		out[i].Rank = i + 1
		out[i].FinalScore = out[i].Score
	}
	return out
}

// sortByScore orders scored products by relevance descending.
//
// Ties break on rating, then rating count, then product ID. The ID tiebreak is
// what makes the whole ranking deterministic: without it, two products with
// identical scores would swap positions between requests depending on map
// iteration order, and identical prompts would return different shortlists —
// which makes both caching and debugging impossible.
func sortByScore(scored []ScoredProduct) {
	sort.SliceStable(scored, func(i, j int) bool {
		a, b := scored[i], scored[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.Product.Rating != b.Product.Rating {
			return a.Product.Rating > b.Product.Rating
		}
		if a.Product.RatingCount != b.Product.RatingCount {
			return a.Product.RatingCount > b.Product.RatingCount
		}
		return strings.Compare(a.Product.ID.String(), b.Product.ID.String()) < 0
	})
}
