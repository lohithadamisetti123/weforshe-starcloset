// Package dto defines the wire contract.
//
// These types exist so that the database schema and the public API can change
// independently. That separation is doing real work here rather than being
// ceremony: the frontend was built first, against a mock, and its TypeScript
// types are the specification this backend has to satisfy. Serialising GORM
// models directly would leak `slug`, `searchDoc`, `deletedAt` and
// `categoryId` into responses the frontend never asked for, and would couple
// every future column rename to a frontend release.
//
// The rule for this package: if a field's JSON tag does not appear in
// `frontend/src/types/index.ts`, it does not belong in the response body.
package dto

import (
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
)

// Product is the wire shape of a catalogue product.
//
// It mirrors this TypeScript interface exactly:
//
//	export interface Product {
//	  id: string
//	  title: string
//	  brand: string
//	  category: Category
//	  price: number
//	  mrp: number
//	  discount: number
//	  rating: number
//	  ratingCount: number
//	  image: string
//	  gender: Gender
//	  occasion: Occasion
//	  style: StyleTag
//	  color: string
//	  description: string
//	  matchPercent?: number
//	}
//
// Field order below follows the interface so the two can be diffed by eye.
type Product struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Brand       string  `json:"brand"`
	Category    string  `json:"category"`
	Price       int     `json:"price"`
	MRP         int     `json:"mrp"`
	Discount    int     `json:"discount"`
	Rating      float64 `json:"rating"`
	RatingCount int     `json:"ratingCount"`
	Image       string  `json:"image"`
	Gender      string  `json:"gender"`
	Occasion    string  `json:"occasion"`
	Style       string  `json:"style"`
	Color       string  `json:"color"`
	Description string  `json:"description"`

	// MatchPercent is optional in the TypeScript interface and must be omitted
	// rather than sent as 0 when absent: `matchPercent?: number` means the
	// frontend checks for undefined, and a literal 0 would render as "0% match"
	// on every plain catalogue card.
	MatchPercent *int `json:"matchPercent,omitempty"`

	// MatchReasons is an additive field the mock never had. Additive is safe —
	// TypeScript structural typing ignores unknown properties — and it exposes
	// the engine's per-product justification for the frontend to adopt when
	// ready, without requiring any change today.
	MatchReasons []string `json:"matchReasons,omitempty"`
}

// NewProduct maps a catalogue model to its wire shape.
func NewProduct(p *models.Product) Product {
	if p == nil {
		return Product{}
	}

	category := ""
	if p.Category != nil {
		category = string(p.Category.Name)
	}

	return Product{
		ID:          p.ID.String(),
		Title:       p.Title,
		Brand:       p.Brand,
		Category:    category,
		Price:       p.Price,
		MRP:         p.MRP,
		Discount:    p.EffectiveDiscount(),
		Rating:      p.Rating,
		RatingCount: p.RatingCount,
		Image:       p.Image,
		Gender:      string(p.Gender),
		Occasion:    string(p.Occasion),
		Style:       string(p.Style),
		Color:       p.Color,
		Description: p.Description,
	}
}

// NewProducts maps a slice of models.
func NewProducts(ps []*models.Product) []Product {
	// A non-nil empty slice, so an empty result serialises as `[]` and not
	// `null`. The frontend calls .map() on this without a guard; `null` would
	// be a runtime crash rather than an empty grid.
	out := make([]Product, 0, len(ps))
	for _, p := range ps {
		out = append(out, NewProduct(p))
	}
	return out
}

// NewScoredProduct maps a ranked product, attaching the engine's match score
// and its reasons.
func NewScoredProduct(sp recommendation.ScoredProduct) Product {
	p := NewProduct(sp.Product)
	pct := sp.MatchPercent
	p.MatchPercent = &pct
	p.MatchReasons = sp.Reasons
	return p
}

// NewScoredProducts maps a ranked slice.
func NewScoredProducts(sps []recommendation.ScoredProduct) []Product {
	out := make([]Product, 0, len(sps))
	for _, sp := range sps {
		out = append(out, NewScoredProduct(sp))
	}
	return out
}

// Category is the wire shape of a garment family.
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	// ProductCount is populated by the listing endpoint, which the frontend's
	// category chart can consume directly.
	ProductCount int64 `json:"productCount"`
}

// NewCategory maps a category model.
func NewCategory(c *models.Category, count int64) Category {
	if c == nil {
		return Category{}
	}
	return Category{
		ID:           c.ID.String(),
		Name:         string(c.Name),
		Slug:         c.Slug,
		Description:  c.Description,
		ProductCount: count,
	}
}
