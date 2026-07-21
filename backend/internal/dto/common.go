package dto

import (
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
)

// ProductQuery is the query string of the product listing and search endpoints.
//
// Every field is optional. The zero value is a valid "browse everything"
// request, which is what the catalogue grid issues on first paint.
type ProductQuery struct {
	// Q is a free-text search term.
	Q string `form:"q" validate:"max=200"`

	Category string `form:"category" validate:"max=48"`
	Brand    string `form:"brand" validate:"max=80"`
	Gender   string `form:"gender" validate:"omitempty,oneof=Women Men Unisex"`
	// The quoted member is not decoration: "Date Night" contains a space, and
	// validator/v10 splits oneof on whitespace, so unquoted it would silently
	// become two bogus members and reject a real occasion.
	Occasion string `form:"occasion" validate:"omitempty,oneof=Wedding Office College Party Travel Festival Casual 'Date Night'" example:"Wedding"`
	Style    string `form:"style" validate:"max=24"`
	Color    string `form:"color" validate:"max=48"`

	MinPrice  *int     `form:"minPrice" validate:"omitempty,gte=0,lte=10000000"`
	MaxPrice  *int     `form:"maxPrice" validate:"omitempty,gte=0,lte=10000000"`
	MinRating *float64 `form:"minRating" validate:"omitempty,gte=0,lte=5"`

	// Sort accepts relevance | price_asc | price_desc | rating | discount | newest.
	Sort string `form:"sort" validate:"omitempty,oneof=relevance price_asc price_desc rating discount newest"`

	Page     int `form:"page" validate:"omitempty,gte=1"`
	PageSize int `form:"pageSize" validate:"omitempty,gte=1,lte=100"`
}

// Pagination defaults. PageSize is capped so a client cannot ask for the whole
// catalogue in one response and turn a browse into a denial of service.
const (
	DefaultPage     = 1
	DefaultPageSize = 24
	MaxPageSize     = 100
)

// Normalize applies defaults and clamps out-of-range values.
//
// It corrects rather than rejects: a client sending page=0 has made a trivial
// mistake and wants page 1, and answering that with a 422 is pedantry. Genuinely
// invalid values (a non-numeric page) are still caught by binding.
func (q *ProductQuery) Normalize() {
	if q.Page < 1 {
		q.Page = DefaultPage
	}
	if q.PageSize < 1 {
		q.PageSize = DefaultPageSize
	}
	if q.PageSize > MaxPageSize {
		q.PageSize = MaxPageSize
	}
	if q.MinPrice != nil && q.MaxPrice != nil && *q.MinPrice > *q.MaxPrice {
		q.MinPrice, q.MaxPrice = q.MaxPrice, q.MinPrice
	}
	q.Q = strings.TrimSpace(q.Q)
	if q.Sort == "" {
		if q.Q != "" {
			q.Sort = string(recommendation.SortRelevance)
		} else {
			// Relevance is meaningless without a query, so an unfiltered browse
			// defaults to the products people actually like.
			q.Sort = string(recommendation.SortRating)
		}
	}
}

// Offset returns the SQL offset for the current page.
func (q ProductQuery) Offset() int { return (q.Page - 1) * q.PageSize }

// SortHint resolves Sort onto the engine's type.
func (q ProductQuery) SortHint() recommendation.SortHint {
	switch q.Sort {
	case "price_asc":
		return recommendation.SortPriceAsc
	case "price_desc":
		return recommendation.SortPriceDesc
	case "rating":
		return recommendation.SortRating
	case "discount":
		return recommendation.SortDiscount
	case "newest":
		return recommendation.SortNewest
	default:
		return recommendation.SortRelevance
	}
}

// CacheKey builds a deterministic cache key for this query.
//
// Determinism is the whole point: a map-ordered or pointer-formatted key would
// produce a different string for identical queries, and the cache would sit at
// a 0% hit rate while looking perfectly healthy.
func (q ProductQuery) CacheKey() string {
	var b strings.Builder
	b.WriteString("products:v1")
	writeKV(&b, "q", strings.ToLower(q.Q))
	writeKV(&b, "cat", q.Category)
	writeKV(&b, "brand", q.Brand)
	writeKV(&b, "gender", q.Gender)
	writeKV(&b, "occ", q.Occasion)
	writeKV(&b, "style", q.Style)
	writeKV(&b, "color", q.Color)
	writeIntPtr(&b, "min", q.MinPrice)
	writeIntPtr(&b, "max", q.MaxPrice)
	if q.MinRating != nil {
		writeKV(&b, "rating", formatFloat(*q.MinRating))
	}
	writeKV(&b, "sort", q.Sort)
	writeKV(&b, "page", itoa(q.Page))
	writeKV(&b, "size", itoa(q.PageSize))
	return b.String()
}

// ToIntent converts a structured filter query into an engine Intent, so that
// the search endpoint ranks with the same logic as the assistant rather than
// with a second, subtly different implementation.
func (q ProductQuery) ToIntent() *recommendation.Intent {
	intent := recommendation.NewIntent(q.Q)
	intent.Sort = q.SortHint()
	// Filters are stated facts, not an inference from prose, so there is
	// nothing to be uncertain about.
	intent.Confidence = 1.0

	if g, ok := models.ParseGender(q.Gender); ok && q.Gender != "" {
		intent.Gender = &g
	}
	if o, ok := models.ParseOccasion(q.Occasion); ok && q.Occasion != "" {
		intent.Occasion = &o
	}
	if s, ok := models.ParseStyle(q.Style); ok && q.Style != "" {
		intent.Style = &s
	}
	if c, ok := models.ParseCategory(q.Category); ok && q.Category != "" {
		intent.Category = &c
	}
	if q.Color != "" {
		intent.Colors = []string{q.Color}
	}
	if q.Brand != "" {
		intent.Brands = []string{q.Brand}
	}
	if q.MinPrice != nil || q.MaxPrice != nil {
		intent.Budget = &recommendation.BudgetRange{Min: q.MinPrice, Max: q.MaxPrice}
	}
	if q.Q != "" {
		intent.Keywords = recommendation.ExtractKeywords(q.Q)
	}
	return intent
}

// SearchMeta accompanies the /products/search response.
//
// It carries the same pagination shape as a plain listing plus the engine's
// self-report, because search results come from the ranking engine (see
// ToIntent below) and a shopper comparing the search bar to the assistant
// should see the same kind of "why" for both.
type SearchMeta struct {
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
	Total      int      `json:"total"`
	TotalPages int      `json:"totalPages"`
	HasNext    bool     `json:"hasNext"`
	HasPrev    bool     `json:"hasPrev"`
	Source     string   `json:"source"`
	Confidence float64  `json:"confidence"`
	Relaxed    []string `json:"relaxed,omitempty"`
}

// NewSearchMeta builds the search response's metadata from an engine result.
func NewSearchMeta(q ProductQuery, r *recommendation.Result) SearchMeta {
	total := 0
	if r != nil {
		total = len(r.Products)
	}
	totalPages := 0
	if q.PageSize > 0 {
		totalPages = (total + q.PageSize - 1) / q.PageSize
	}
	m := SearchMeta{
		Page:       q.Page,
		PageSize:   q.PageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    q.Page < totalPages,
		HasPrev:    q.Page > 1,
	}
	if r != nil && r.Intent != nil {
		m.Source = string(r.Intent.Source)
		m.Confidence = round2(r.Intent.Confidence)
		m.Relaxed = r.Relaxed
	}
	return m
}

// PaginationQuery is the query string of simple paginated collections.
type PaginationQuery struct {
	Page     int `form:"page" validate:"omitempty,gte=1"`
	PageSize int `form:"pageSize" validate:"omitempty,gte=1,lte=100"`
}

// Normalize applies defaults and clamps.
func (p *PaginationQuery) Normalize() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}
}

// Offset returns the SQL offset.
func (p PaginationQuery) Offset() int { return (p.Page - 1) * p.PageSize }

func writeKV(b *strings.Builder, k, v string) {
	if v == "" {
		return
	}
	b.WriteByte(':')
	b.WriteString(k)
	b.WriteByte('=')
	b.WriteString(v)
}

func writeIntPtr(b *strings.Builder, k string, v *int) {
	if v == nil {
		return
	}
	writeKV(b, k, itoa(*v))
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func formatFloat(f float64) string {
	whole := int(f)
	frac := int((f-float64(whole))*10 + 0.5)
	return itoa(whole) + "." + itoa(frac)
}
