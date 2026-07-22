package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserProfile holds a shopper's onboarding answers and stated preferences.
//
// It is one row per user (unique index on user_id) rather than columns on the
// users table because a profile is optional, edited independently, and expected
// to grow — keeping it separate means the hot users table stays narrow and a
// profile write never contends with an auth read.
//
// The preference lists are stored as JSONB (see jsonb.go) rather than as
// normalised join tables: they are always read back whole to build a scoring
// context, never queried by element, and their vocabulary is expected to churn.
// This is the same trade the project already makes for the AI intent and the
// per-product score breakdowns.
type UserProfile struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`

	Name   string `gorm:"type:varchar(120);not null;default:''" json:"name"`
	Gender Gender `gorm:"type:varchar(16)" json:"gender,omitempty"`
	City   string `gorm:"type:varchar(120);not null;default:''" json:"city"`

	// BudgetMin and BudgetMax are the shopper's stated everyday price envelope,
	// in whole rupees. Pointers so "not stated" is distinct from ₹0. These feed
	// personalization only as context; the per-request budget from a prompt
	// always takes precedence and remains the hard filter.
	BudgetMin *int `gorm:"column:budget_min" json:"budgetMin,omitempty"`
	BudgetMax *int `gorm:"column:budget_max" json:"budgetMax,omitempty"`

	// The four preference lists. Stored as jsonb string arrays.
	FavoriteStyles     JSONB `gorm:"type:jsonb;not null;default:'[]'" json:"favoriteStyles"`
	FavoriteColors     JSONB `gorm:"type:jsonb;not null;default:'[]'" json:"favoriteColors"`
	FavoriteBrands     JSONB `gorm:"type:jsonb;not null;default:'[]'" json:"favoriteBrands"`
	FavoriteCategories JSONB `gorm:"type:jsonb;not null;default:'[]'" json:"favoriteCategories"`
	// FavoriteOccasions seeds the personalized feed and reinforces the engine's
	// occasion signal; Sizes are stored for display and future filtering (the
	// current catalogue carries no size dimension, so sizes do not yet affect
	// ranking).
	FavoriteOccasions JSONB `gorm:"type:jsonb;not null;default:'[]'" json:"favoriteOccasions"`
	Sizes             JSONB `gorm:"type:jsonb;not null;default:'[]'" json:"sizes"`

	// Onboarded marks that the shopper completed the onboarding flow at least
	// once, so the client shows it exactly once and offers edit thereafter.
	Onboarded bool `gorm:"not null;default:false" json:"onboarded"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName pins the table name.
func (UserProfile) TableName() string { return "user_profiles" }

// BeforeSave assigns an ID when absent and normalises empty JSONB arrays so a
// row never stores SQL NULL where the app expects `[]`.
func (p *UserProfile) BeforeSave(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.FavoriteStyles = ensureJSONArray(p.FavoriteStyles)
	p.FavoriteColors = ensureJSONArray(p.FavoriteColors)
	p.FavoriteBrands = ensureJSONArray(p.FavoriteBrands)
	p.FavoriteCategories = ensureJSONArray(p.FavoriteCategories)
	p.FavoriteOccasions = ensureJSONArray(p.FavoriteOccasions)
	p.Sizes = ensureJSONArray(p.Sizes)
	return nil
}

// StylesList / ColorsList / BrandsList / CategoriesList decode the stored JSONB
// arrays into plain string slices. A decode failure yields an empty slice
// rather than an error: a malformed preference blob should cost personalization
// signal, never fail the request that reads it.
func (p *UserProfile) StylesList() []string     { return decodeStringList(p.FavoriteStyles) }
func (p *UserProfile) ColorsList() []string     { return decodeStringList(p.FavoriteColors) }
func (p *UserProfile) BrandsList() []string     { return decodeStringList(p.FavoriteBrands) }
func (p *UserProfile) CategoriesList() []string { return decodeStringList(p.FavoriteCategories) }
func (p *UserProfile) OccasionsList() []string  { return decodeStringList(p.FavoriteOccasions) }
func (p *UserProfile) SizesList() []string      { return decodeStringList(p.Sizes) }

// ensureJSONArray returns a valid `[]` JSONB when the value is empty/invalid.
func ensureJSONArray(j JSONB) JSONB {
	if len(j) == 0 {
		return JSONB("[]")
	}
	return j
}

// decodeStringList unmarshals a JSONB array into []string, tolerating nulls.
func decodeStringList(j JSONB) []string {
	var out []string
	if err := j.Decode(&out); err != nil {
		return nil
	}
	return out
}
