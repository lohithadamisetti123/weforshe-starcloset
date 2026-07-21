package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Recommendation is one full run of the ranking engine: the prompt that caused
// it, the intent that was extracted from it, and the ranked slice of products
// that came out.
//
// Persisting the intent and the per-item score breakdown is what makes the
// engine debuggable in production: when someone reports "this pick made no
// sense", the row explains precisely which signals fired.
type Recommendation struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_rec_user_created,priority:1" json:"userId"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`

	Prompt  string `gorm:"type:text;not null" json:"prompt"`
	Summary string `gorm:"type:text;not null" json:"summary"`
	Reason  string `gorm:"type:text;not null" json:"reason"`

	BudgetLabel string `gorm:"type:varchar(48);not null;default:''" json:"budgetLabel"`
	Occasion    string `gorm:"type:varchar(32);not null;default:''" json:"occasion"`
	Style       string `gorm:"type:varchar(32);not null;default:''" json:"style"`

	// Source records which NLU path produced the intent: "gemini",
	// "rules" (Gemini unavailable) or "hybrid" (both, merged).
	Source string `gorm:"type:varchar(16);not null;default:'rules'" json:"source"`

	// Intent is the serialised recommendation.Intent that drove the ranking.
	Intent JSONB `gorm:"type:jsonb" json:"intent"`

	// LatencyMS is the end-to-end engine latency, kept for the dashboard and
	// for spotting Gemini slowdowns without an APM stack.
	LatencyMS int `gorm:"not null;default:0" json:"latencyMs"`

	Items []RecommendationItem `gorm:"foreignKey:RecommendationID;constraint:OnDelete:CASCADE" json:"items,omitempty"`

	CreatedAt time.Time      `gorm:"index:idx_rec_user_created,priority:2,sort:desc" json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName pins the table name.
func (Recommendation) TableName() string { return "recommendations" }

// BeforeCreate assigns an ID when absent.
func (r *Recommendation) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// RecommendationItem is a single ranked product within a Recommendation.
type RecommendationItem struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RecommendationID uuid.UUID `gorm:"type:uuid;not null;index" json:"recommendationId"`

	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"productId"`
	Product   *Product  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`

	// Rank is 1-based position as shown to the user.
	Rank int `gorm:"not null" json:"rank"`
	// Score is the raw 0..1 relevance score before presentation calibration.
	Score float64 `gorm:"not null" json:"score"`
	// MatchPercent is the calibrated 60..99 figure the UI renders.
	MatchPercent int `gorm:"not null" json:"matchPercent"`

	// Breakdown is the serialised per-signal contribution map.
	Breakdown JSONB `gorm:"type:jsonb" json:"breakdown"`
	// Reasons is the serialised []string of human-readable justifications.
	Reasons JSONB `gorm:"type:jsonb" json:"reasons"`

	CreatedAt time.Time `json:"createdAt"`
}

// TableName pins the table name.
func (RecommendationItem) TableName() string { return "recommendation_items" }

// BeforeCreate assigns an ID when absent.
func (r *RecommendationItem) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
