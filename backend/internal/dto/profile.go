package dto

import (
	"github.com/myntra/bharat-ai-backend/internal/models"
)

// UpdateProfileRequest is the body of PUT /api/v1/profile/preferences — the
// onboarding submission and every later profile edit share this one shape.
//
// Every field is optional so a partial edit (e.g. changing only the budget) is
// a valid request; the service treats the payload as the new complete profile,
// so the client always sends the full current state. The list caps keep a
// single shopper from storing an unbounded preference blob.
type UpdateProfileRequest struct {
	Name   string `json:"name" validate:"max=120"`
	Gender string `json:"gender" validate:"omitempty,oneof=Women Men Girls Boys Unisex"`
	City   string `json:"city" validate:"max=120"`

	BudgetMin *int `json:"budgetMin" validate:"omitempty,gte=0,lte=10000000"`
	BudgetMax *int `json:"budgetMax" validate:"omitempty,gte=0,lte=10000000"`

	FavoriteStyles     []string `json:"favoriteStyles" validate:"omitempty,max=20,dive,max=48"`
	FavoriteColors     []string `json:"favoriteColors" validate:"omitempty,max=20,dive,max=48"`
	FavoriteBrands     []string `json:"favoriteBrands" validate:"omitempty,max=20,dive,max=80"`
	FavoriteCategories []string `json:"favoriteCategories" validate:"omitempty,max=20,dive,max=48"`
	FavoriteOccasions  []string `json:"favoriteOccasions" validate:"omitempty,max=20,dive,max=48"`
	Sizes              []string `json:"sizes" validate:"omitempty,max=20,dive,max=24"`

	// Onboarded lets the client mark onboarding complete on first submit. It is
	// a pointer so an ordinary edit (which omits it) never accidentally flips a
	// completed profile back to not-onboarded.
	Onboarded *bool `json:"onboarded"`
}

// ProfileResponse is the wire shape of a shopper's profile.
type ProfileResponse struct {
	Name               string   `json:"name"`
	Gender             string   `json:"gender"`
	City               string   `json:"city"`
	BudgetMin          *int     `json:"budgetMin,omitempty"`
	BudgetMax          *int     `json:"budgetMax,omitempty"`
	FavoriteStyles     []string `json:"favoriteStyles"`
	FavoriteColors     []string `json:"favoriteColors"`
	FavoriteBrands     []string `json:"favoriteBrands"`
	FavoriteCategories []string `json:"favoriteCategories"`
	FavoriteOccasions  []string `json:"favoriteOccasions"`
	Sizes              []string `json:"sizes"`
	Onboarded          bool     `json:"onboarded"`
}

// NewProfileResponse maps a profile model, tolerating a nil profile (a shopper
// who has not onboarded) by returning a zero-value, not-onboarded response.
func NewProfileResponse(p *models.UserProfile) ProfileResponse {
	if p == nil {
		return ProfileResponse{
			FavoriteStyles:     []string{},
			FavoriteColors:     []string{},
			FavoriteBrands:     []string{},
			FavoriteCategories: []string{},
			FavoriteOccasions:  []string{},
			Sizes:              []string{},
			Onboarded:          false,
		}
	}
	return ProfileResponse{
		Name:               p.Name,
		Gender:             string(p.Gender),
		City:               p.City,
		BudgetMin:          p.BudgetMin,
		BudgetMax:          p.BudgetMax,
		FavoriteStyles:     orEmpty(p.StylesList()),
		FavoriteColors:     orEmpty(p.ColorsList()),
		FavoriteBrands:     orEmpty(p.BrandsList()),
		FavoriteCategories: orEmpty(p.CategoriesList()),
		FavoriteOccasions:  orEmpty(p.OccasionsList()),
		Sizes:              orEmpty(p.SizesList()),
		Onboarded:          p.Onboarded,
	}
}

// OnboardingStatusResponse is the tiny payload of GET /api/v1/profile/status,
// which the client checks on load to decide whether to show onboarding.
type OnboardingStatusResponse struct {
	Onboarded bool `json:"onboarded"`
}

// PastOrderResponse is the wire shape of one purchase-history line.
type PastOrderResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Brand     string `json:"brand"`
	Category  string `json:"category"`
	Image     string `json:"image"`
	Price     int    `json:"price"`
	OrderedAt string `json:"orderedAt"`
}

// NewPastOrder maps a past-order model.
func NewPastOrder(o *models.PastOrder) PastOrderResponse {
	if o == nil {
		return PastOrderResponse{}
	}
	return PastOrderResponse{
		ID:        o.ID.String(),
		Title:     o.Title,
		Brand:     o.Brand,
		Category:  o.Category,
		Image:     o.Image,
		Price:     o.Price,
		OrderedAt: o.OrderedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// NewPastOrders maps a slice, always returning a non-nil slice so the client
// can map over it without a null guard.
func NewPastOrders(orders []*models.PastOrder) []PastOrderResponse {
	out := make([]PastOrderResponse, 0, len(orders))
	for _, o := range orders {
		out = append(out, NewPastOrder(o))
	}
	return out
}

// orEmpty guarantees a non-nil slice for JSON (`[]`, never `null`).
func orEmpty(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
