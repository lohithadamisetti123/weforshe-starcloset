package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// ProductRepository is the catalogue's data access contract.
//
// It deliberately includes Candidates, the method the recommendation engine
// needs, so the repository satisfies recommendation.CandidateSource without an
// adapter. The engine still depends only on that one-method interface; this is
// simply the production implementation of it.
type ProductRepository interface {
	recommendation.CandidateSource

	List(ctx context.Context, q dto.ProductQuery) ([]*models.Product, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Product, error)
	Count(ctx context.Context) (int64, error)
	Brands(ctx context.Context) ([]string, error)
	// Colors returns the distinct catalogue color list, used alongside Brands
	// to build the ai.Vocabulary the rule extractor matches against at boot.
	Colors(ctx context.Context) ([]string, error)
	MeanRating(ctx context.Context) (float64, error)
}

// productRepository is the GORM-backed implementation.
type productRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewProductRepository constructs a ProductRepository.
func NewProductRepository(db *gorm.DB, log *zap.Logger) ProductRepository {
	return &productRepository{db: db, log: log}
}

// Candidates returns the pool the engine ranks.
//
// This is the hot path for every assistant request, and its job is narrow on
// purpose: apply only the constraints that genuinely disqualify a product, then
// get out of the way. The temptation is to push occasion and style into the
// WHERE clause too — it would be faster and it would be wrong. A shopper asking
// for "something for a sangeet" who gets zero rows because nothing is tagged
// Festival has been failed by SQL; the same shopper handed a pool including
// Wedding and Party pieces gets a graded answer from the scorer.
//
// The pool is capped (REC_CANDIDATE_POOL, default 200) rather than unbounded.
// With a catalogue this size the cap never binds, but it means the engine's
// cost is bounded by configuration instead of by however large the catalogue
// grows — the failure mode of an unbounded pool is a slow, silent regression.
func (r *productRepository) Candidates(ctx context.Context, f recommendation.CandidateFilter) ([]*models.Product, error) {
	q := r.db.WithContext(ctx).Model(&models.Product{}).Preload("Category")

	if f.Gender != nil {
		// Unisex products always qualify: excluding them from a "men's shirt"
		// query would drop the entire unisex rail for no reason a shopper would
		// endorse.
		q = q.Where("gender IN ?", []models.Gender{*f.Gender, models.GenderUnisex})
	}
	if len(f.Categories) > 0 {
		q = q.Where("category_id IN (SELECT id FROM categories WHERE name IN ?)", f.Categories)
	}
	if f.MinPrice != nil {
		q = q.Where("price >= ?", *f.MinPrice)
	}
	if f.MaxPrice != nil {
		q = q.Where("price <= ?", *f.MaxPrice)
	}
	if len(f.ExcludeStyles) > 0 {
		q = q.Where("style NOT IN ?", f.ExcludeStyles)
	}
	for _, c := range f.ExcludeColors {
		if c = strings.TrimSpace(c); c != "" {
			// Colour is a free-text field ("Bottle Green", "Off White"), so an
			// exclusion has to match loosely or "no green" misses "Bottle
			// Green" — exactly the shade the shopper meant.
			q = q.Where("LOWER(color) NOT LIKE ?", "%"+strings.ToLower(c)+"%")
		}
	}
	for _, b := range f.ExcludeBrands {
		if b = strings.TrimSpace(b); b != "" {
			q = q.Where("LOWER(brand) <> ?", strings.ToLower(b))
		}
	}
	// ExcludeProducts blocks specific product UUIDs.
	// We use individual != clauses rather than NOT IN (?) because GORM's
	// PostgreSQL dialect does not reliably expand []string into a uuid NOT IN
	// list — the cast id::text = ANY(?) form also has driver quoting issues.
	// Individual clauses are safe and the list is always short (≤ 12 items).
	for _, pid := range f.ExcludeProducts {
		if pid = strings.TrimSpace(pid); pid != "" {
			q = q.Where("id::text != ?", pid)
		}
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 200
	}

	// Ordering by rating before the cap means that if the cap ever does bind,
	// the products it discards are the ones least likely to have been shown.
	// An arbitrary cut would make results depend on physical row order.
	var products []*models.Product
	if err := q.Order("rating DESC, rating_count DESC, id ASC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("repositories: fetching candidates: %w", err)
	}
	return products, nil
}

// List returns a filtered, paginated page of products with the total count.
//
// Unlike Candidates, this backs a browse UI where filters are explicit clicks,
// so here the filters *are* hard: a shopper who ticks "Sarees" and sees a shirt
// has been ignored. The distinction between the two methods is the distinction
// between interpreting a wish and honouring an instruction.
func (r *productRepository) List(ctx context.Context, q dto.ProductQuery) ([]*models.Product, int64, error) {
	base := r.db.WithContext(ctx).Model(&models.Product{})
	base = r.applyFilters(base, q)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repositories: counting products: %w", err)
	}
	if total == 0 {
		return []*models.Product{}, 0, nil
	}

	// The count and the page must be computed from the same filter set but not
	// the same *gorm.DB: GORM mutates the statement in place, so reusing the
	// handle after Count would carry the count's SELECT into the page query.
	page := r.db.WithContext(ctx).Model(&models.Product{}).Preload("Category")
	page = r.applyFilters(page, q)
	page = r.applyOrder(page, q)

	var products []*models.Product
	if err := page.Limit(q.PageSize).Offset(q.Offset()).Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("repositories: listing products: %w", err)
	}
	return products, total, nil
}

// applyFilters translates a ProductQuery into WHERE clauses.
func (r *productRepository) applyFilters(q *gorm.DB, f dto.ProductQuery) *gorm.DB {
	if f.Q != "" {
		term := "%" + strings.ToLower(strings.TrimSpace(f.Q)) + "%"
		// search_doc is denormalised at write time precisely so this stays one
		// indexed predicate instead of an OR across four columns.
		q = q.Where("search_doc LIKE ?", term)
	}
	if f.Category != "" {
		if name, ok := models.ParseCategory(f.Category); ok {
			q = q.Where("category_id IN (?)",
				r.db.Model(&models.Category{}).Select("id").Where("name = ?", name))
		} else {
			// An unrecognised category yields nothing rather than everything.
			// Silently ignoring the filter would show a shopper the full
			// catalogue and let them believe it was all "Sneakers".
			q = q.Where("1 = 0")
		}
	}
	if f.Brand != "" {
		q = q.Where("LOWER(brand) = ?", strings.ToLower(strings.TrimSpace(f.Brand)))
	}
	if f.Gender != "" {
		if g, ok := models.ParseGender(f.Gender); ok {
			q = q.Where("gender IN ?", []models.Gender{g, models.GenderUnisex})
		}
	}
	if f.Occasion != "" {
		if o, ok := models.ParseOccasion(f.Occasion); ok {
			q = q.Where("occasion = ?", o)
		} else {
			q = q.Where("1 = 0")
		}
	}
	if f.Style != "" {
		if s, ok := models.ParseStyle(f.Style); ok {
			q = q.Where("style = ?", s)
		} else {
			q = q.Where("1 = 0")
		}
	}
	if f.Color != "" {
		q = q.Where("LOWER(color) LIKE ?", "%"+strings.ToLower(strings.TrimSpace(f.Color))+"%")
	}
	if f.MinPrice != nil {
		q = q.Where("price >= ?", *f.MinPrice)
	}
	if f.MaxPrice != nil {
		q = q.Where("price <= ?", *f.MaxPrice)
	}
	if f.MinRating != nil {
		q = q.Where("rating >= ?", *f.MinRating)
	}
	return q
}

// applyOrder translates the sort hint into an ORDER BY.
//
// Every ordering ends in `id ASC`. Without a unique tiebreak, two products with
// the same price have no defined order, and Postgres is free to return them
// differently between calls — which shows up as a product appearing on both
// page 1 and page 2 while another never appears at all. Deterministic
// pagination requires a total order.
func (r *productRepository) applyOrder(q *gorm.DB, f dto.ProductQuery) *gorm.DB {
	switch f.SortHint() {
	case recommendation.SortPriceAsc:
		return q.Order("price ASC, id ASC")
	case recommendation.SortPriceDesc:
		return q.Order("price DESC, id ASC")
	case recommendation.SortRating:
		return q.Order("rating DESC, rating_count DESC, id ASC")
	case recommendation.SortDiscount:
		return q.Order("discount DESC, price ASC, id ASC")
	case recommendation.SortNewest:
		return q.Order("created_at DESC, id ASC")
	default:
		// "Relevance" without the engine is meaningless, so the browse endpoint
		// falls back to the catalogue's own quality signal. The search endpoint
		// routes through the engine instead and never reaches this branch.
		return q.Order("rating DESC, rating_count DESC, id ASC")
	}
}

// GetByID fetches one product.
func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	var p models.Product
	err := r.db.WithContext(ctx).Preload("Category").First(&p, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFound("Product")
		}
		return nil, fmt.Errorf("repositories: fetching product %s: %w", id, err)
	}
	return &p, nil
}

// GetBySlug fetches a product by its catalogue slug ("prod-001").
//
// The frontend's mock data keys on these slugs, so any bookmark or hard-coded
// link from the mocked era resolves against the real API instead of 404ing.
func (r *productRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var p models.Product
	err := r.db.WithContext(ctx).Preload("Category").First(&p, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFound("Product")
		}
		return nil, fmt.Errorf("repositories: fetching product %q: %w", slug, err)
	}
	return &p, nil
}

// GetByIDs fetches products in bulk, in the order requested.
//
// The ordering matters: callers pass IDs that are already ranked — a stored
// recommendation's shortlist, a favourites list — and returning them in
// database order would silently discard that ranking. This is also the
// N+1 fix for favourites and recommendation replay: one query, not one per row.
func (r *productRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Product, error) {
	if len(ids) == 0 {
		return []*models.Product{}, nil
	}

	var found []*models.Product
	if err := r.db.WithContext(ctx).Preload("Category").
		Where("id IN ?", ids).Find(&found).Error; err != nil {
		return nil, fmt.Errorf("repositories: fetching products by ids: %w", err)
	}

	byID := make(map[uuid.UUID]*models.Product, len(found))
	for _, p := range found {
		byID[p.ID] = p
	}

	ordered := make([]*models.Product, 0, len(ids))
	for _, id := range ids {
		// A missing ID is skipped, not an error: a product deleted after a
		// recommendation was stored should shorten the replayed grid, not break
		// the history page that references it.
		if p, ok := byID[id]; ok {
			ordered = append(ordered, p)
		}
	}
	return ordered, nil
}

// Count returns the number of live products.
func (r *productRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("repositories: counting products: %w", err)
	}
	return n, nil
}

// Brands returns the distinct brand list, alphabetically.
func (r *productRepository) Brands(ctx context.Context) ([]string, error) {
	var brands []string
	if err := r.db.WithContext(ctx).Model(&models.Product{}).
		Distinct().Order("brand ASC").Pluck("brand", &brands).Error; err != nil {
		return nil, fmt.Errorf("repositories: listing brands: %w", err)
	}
	return brands, nil
}

// Colors returns the distinct catalogue color list, alphabetically.
//
// This feeds the rule extractor's vocabulary alongside Brands: shoppers echo
// the catalogue's own color names ("Bottle Green", "Rust") as often as they use
// generic family words ("green"), and the extractor cannot match a term it was
// never given.
func (r *productRepository) Colors(ctx context.Context) ([]string, error) {
	var colors []string
	if err := r.db.WithContext(ctx).Model(&models.Product{}).
		Distinct().Order("color ASC").Pluck("color", &colors).Error; err != nil {
		return nil, fmt.Errorf("repositories: listing colors: %w", err)
	}
	return colors, nil
}

// MeanRating returns the catalogue's mean rating.
//
// This feeds the Bayesian quality prior, which shrinks each product's rating
// toward the catalogue mean in proportion to how few ratings it has. Computing
// it from the real catalogue rather than hard-coding 4.0 is what keeps a
// 5.0-from-two-reviews product from outranking a 4.4-from-3,000 one.
func (r *productRepository) MeanRating(ctx context.Context) (float64, error) {
	var mean *float64
	if err := r.db.WithContext(ctx).Model(&models.Product{}).
		Select("AVG(rating)").Scan(&mean).Error; err != nil {
		return 0, fmt.Errorf("repositories: computing mean rating: %w", err)
	}
	if mean == nil {
		// An empty catalogue has no mean. Returning the neutral default keeps
		// the prior well-defined instead of collapsing every score to zero.
		return recommendation.DefaultMeanRating, nil
	}
	return *mean, nil
}
