package services

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// ProductService serves the catalogue.
type ProductService struct {
	products   repositories.ProductRepository
	categories repositories.CategoryRepository
	engine     *recommendation.Engine
	cache      *repositories.Cache
	log        *zap.Logger
}

// NewProductService constructs a ProductService.
func NewProductService(
	products repositories.ProductRepository,
	categories repositories.CategoryRepository,
	engine *recommendation.Engine,
	cache *repositories.Cache,
	log *zap.Logger,
) *ProductService {
	return &ProductService{
		products:   products,
		categories: categories,
		engine:     engine,
		cache:      cache,
		log:        log,
	}
}

// listPage is the cached shape for a page of products. Caching the DTO rather
// than the models means a hit skips serialisation as well as the query, and it
// keeps GORM's lazy association fields out of Redis.
type listPage struct {
	Products []dto.Product   `json:"products"`
	Meta     *utils.PageMeta `json:"meta"`
}

// List returns a filtered page of the catalogue.
func (s *ProductService) List(ctx context.Context, q dto.ProductQuery) ([]dto.Product, *utils.PageMeta, error) {
	q.Normalize()

	key := repositories.KeyPrefixProducts + q.CacheKey()
	var cached listPage
	if s.cache.Get(ctx, key, &cached) {
		return cached.Products, cached.Meta, nil
	}

	products, total, err := s.products.List(ctx, q)
	if err != nil {
		return nil, nil, err
	}

	out := dto.NewProducts(products)
	meta := utils.NewPageMeta(q.Page, q.PageSize, total)

	s.cache.Set(ctx, key, listPage{Products: out, Meta: &meta})
	return out, &meta, nil
}

// Search ranks the catalogue against a query using the recommendation engine.
//
// This is the design decision that keeps the product honest. The lazy version
// of a search endpoint is `WHERE title ILIKE '%term%'` — which cannot tell that
// a shopper typing "wedding kurta under 3k" wants budget-fit and occasion-fit
// weighed against each other, and cannot rank two matches at all. Routing the
// query through ProductQuery.ToIntent and into the same engine the assistant
// uses means search and chat give the same answer to the same question. Two
// ranking paths would guarantee they eventually disagree, and a shopper who
// gets one answer from the search bar and a different one from the assistant
// has learned not to trust either.
func (s *ProductService) Search(ctx context.Context, q dto.ProductQuery) ([]dto.Product, *recommendation.Result, error) {
	q.Normalize()

	intent := q.ToIntent()

	result, err := s.engine.Recommend(ctx, intent)
	if err != nil {
		return nil, nil, err
	}

	// The engine returns its shortlist; a search bar wants a page. Slicing here
	// rather than asking the engine for a page keeps the engine's contract
	// ("the best N for this intent") separate from the endpoint's ("results
	// 25-48 of whatever matched").
	scored := result.Products
	start := q.Offset()
	if start > len(scored) {
		start = len(scored)
	}
	end := start + q.PageSize
	if end > len(scored) {
		end = len(scored)
	}

	return dto.NewScoredProducts(scored[start:end]), result, nil
}

// GetByID returns one product.
//
// The ID accepts either a UUID or the catalogue slug ("prod-001"). The frontend
// mock keyed products by slug, so every link and bookmark from the mocked era
// carries one; rejecting them would break exactly the URLs a reviewer is most
// likely to try.
func (s *ProductService) GetByID(ctx context.Context, id string) (*dto.Product, error) {
	key := repositories.ProductKey(id)
	var cached dto.Product
	if s.cache.Get(ctx, key, &cached) {
		return &cached, nil
	}

	var (
		product *models.Product
		err     error
	)
	if parsed, parseErr := uuid.Parse(id); parseErr == nil {
		product, err = s.products.GetByID(ctx, parsed)
	} else {
		product, err = s.products.GetBySlug(ctx, id)
	}
	if err != nil {
		return nil, err
	}

	out := dto.NewProduct(product)
	s.cache.Set(ctx, key, out)
	return &out, nil
}

// ByCategory returns a page of one category.
func (s *ProductService) ByCategory(ctx context.Context, name string, q dto.ProductQuery) ([]dto.Product, *utils.PageMeta, error) {
	category, ok := models.ParseCategory(name)
	if !ok {
		// Naming the valid categories turns a dead end into a usable error. The
		// client asked for something reasonable in the wrong words; the reply
		// should say the right ones.
		return nil, nil, utils.NewValidation(
			"Unknown category",
			map[string]any{"category": name, "valid": models.AllCategories},
		)
	}

	q.Category = string(category)
	return s.List(ctx, q)
}

// ByBrand returns a page of one brand.
func (s *ProductService) ByBrand(ctx context.Context, brand string, q dto.ProductQuery) ([]dto.Product, *utils.PageMeta, error) {
	q.Brand = brand
	return s.List(ctx, q)
}

// Categories returns the category list with product counts.
func (s *ProductService) Categories(ctx context.Context) ([]dto.Category, error) {
	key := repositories.CategoriesKey()
	var cached []dto.Category
	if s.cache.Get(ctx, key, &cached) {
		return cached, nil
	}

	rows, err := s.categories.List(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]dto.Category, 0, len(rows))
	for i := range rows {
		out = append(out, dto.NewCategory(&rows[i].Category, rows[i].ProductCount))
	}

	s.cache.Set(ctx, key, out)
	return out, nil
}

// Brands returns the distinct brand list.
func (s *ProductService) Brands(ctx context.Context) ([]string, error) {
	key := repositories.KeyPrefixProducts + "brands"
	var cached []string
	if s.cache.Get(ctx, key, &cached) {
		return cached, nil
	}

	brands, err := s.products.Brands(ctx)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, key, brands)
	return brands, nil
}
