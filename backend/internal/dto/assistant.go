package dto

import (
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
)

// ChatRequest is the body of POST /api/v1/assistant/chat.
type ChatRequest struct {
	// Prompt is the shopper's natural-language brief.
	Prompt string `json:"prompt" binding:"required" validate:"required,min=2,max=500" example:"I need a wedding outfit under ₹3000"`

	// Persist controls whether this turn is written to chat history. It is a
	// pointer so that omitting it means "yes" (the frontend sends no such
	// field) while an explicit false can still switch it off for previews.
	Persist *bool `json:"persist,omitempty"`
}

// ShouldPersist reports whether the turn should be recorded.
func (r ChatRequest) ShouldPersist() bool {
	return r.Persist == nil || *r.Persist
}

// AIRecommendationResponse is the assistant's reply.
//
// It mirrors this TypeScript interface exactly:
//
//	export interface AIRecommendationResponse {
//	  summary: string
//	  budget: string
//	  occasion: string
//	  style: string
//	  reason: string
//	  products: Product[]
//	}
//
// Two details are load-bearing and easy to get wrong:
//
//  1. The key is `products`, not `recommendedProducts`. The written brief says
//     the latter; the shipped frontend reads the former
//     (`ResponseCard.tsx` maps over `response.products`). The running code is
//     the real contract, so `products` wins.
//  2. `budget`, `occasion` and `style` are *strings*, not numbers or enums —
//     they are display labels like "₹3,000", "everyday" and "versatile", and
//     the mock falls back to exactly those words when nothing was extracted.
//     Sending a number for `budget` would render as a bare "3000" next to a
//     "Budget" chip that was designed around a formatted currency string.
type AIRecommendationResponse struct {
	Summary  string    `json:"summary"`
	Budget   string    `json:"budget"`
	Occasion string    `json:"occasion"`
	Style    string    `json:"style"`
	Reason   string    `json:"reason"`
	Products []Product `json:"products"`

	// --- Additive fields -----------------------------------------------------
	// Everything below is new. The frontend ignores unknown keys, so these cost
	// it nothing today and give it somewhere to grow into.

	// RecommendationID identifies this run, so the client can favourite from it
	// or deep-link to /recommendations/:id.
	RecommendationID string `json:"recommendationId,omitempty"`

	// Meta exposes what the engine understood and how it got there.
	Meta *AssistantMeta `json:"meta,omitempty"`
}

// AssistantMeta is the engine's self-report for one run.
//
// This is here because "why did it show me that?" is the first question anyone
// asks of a recommender, and the answer already exists inside the engine. It is
// also what makes the demo defensible: the extracted intent is visible rather
// than implied.
type AssistantMeta struct {
	// Source is which understanding path ran: "rules", "gemini" or "hybrid".
	Source string `json:"source"`
	// Confidence is the merged extractor confidence in [0,1].
	Confidence float64 `json:"confidence"`
	// Intent is the structured reading of the prompt.
	Intent IntentView `json:"intent"`
	// CandidatePool is how many products were scored.
	CandidatePool int `json:"candidatePool"`
	// Relaxed lists constraints that had to be widened to fill the shortlist.
	Relaxed []string `json:"relaxed,omitempty"`
	// LatencyMS is end-to-end engine latency.
	LatencyMS int `json:"latencyMs"`
	// Personalized reports whether a shopper personalization context was
	// applied to this run — true only when the shopper has a profile or order
	// history that produced boosts. The client uses it to show a "Personalized
	// for you" cue; it never affects ranking.
	Personalized bool `json:"personalized"`
}

// IntentView is the client-facing projection of a recommendation.Intent.
type IntentView struct {
	BudgetMin     *int     `json:"budgetMin,omitempty"`
	BudgetMax     *int     `json:"budgetMax,omitempty"`
	Occasion      string   `json:"occasion,omitempty"`
	Style         string   `json:"style,omitempty"`
	Gender        string   `json:"gender,omitempty"`
	Category      string   `json:"category,omitempty"`
	Colors        []string `json:"colors,omitempty"`
	Brands        []string `json:"brands,omitempty"`
	ExcludeColors []string `json:"excludeColors,omitempty"`
	ExcludeBrands []string `json:"excludeBrands,omitempty"`
	Keywords      []string `json:"keywords,omitempty"`
	Sort          string   `json:"sort,omitempty"`
}

// NewIntentView projects an engine intent onto the wire.
func NewIntentView(i *recommendation.Intent) IntentView {
	if i == nil {
		return IntentView{}
	}
	v := IntentView{
		Occasion:      i.OccasionOrEmpty(),
		Style:         i.StyleOrEmpty(),
		Gender:        i.GenderOrEmpty(),
		Category:      i.CategoryOrEmpty(),
		Colors:        i.Colors,
		Brands:        i.Brands,
		ExcludeColors: i.ExcludeColors,
		ExcludeBrands: i.ExcludeBrands,
		Keywords:      i.Keywords,
		Sort:          string(i.Sort),
	}
	if i.Budget != nil {
		if i.Budget.HasMin() {
			v.BudgetMin = i.Budget.Min
		}
		if i.Budget.HasMax() {
			v.BudgetMax = i.Budget.Max
		}
	}
	return v
}

// NewAIRecommendationResponse builds the assistant reply from an engine result.
func NewAIRecommendationResponse(r *recommendation.Result, recID string, latencyMS int) AIRecommendationResponse {
	if r == nil {
		return AIRecommendationResponse{Products: []Product{}}
	}
	return AIRecommendationResponse{
		Summary:          r.Summary,
		Budget:           r.BudgetLabel,
		Occasion:         r.OccasionLabel,
		Style:            r.StyleLabel,
		Reason:           r.Reason,
		Products:         NewScoredProducts(r.Products),
		RecommendationID: recID,
		Meta: &AssistantMeta{
			Source:        string(r.Intent.Source),
			Confidence:    round2(r.Intent.Confidence),
			Intent:        NewIntentView(r.Intent),
			CandidatePool: r.PoolSize,
			Relaxed:       r.Relaxed,
			LatencyMS:     latencyMS,
		},
	}
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
