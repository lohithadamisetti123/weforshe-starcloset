package dto

import (
	"time"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// DashboardStats is the wire shape of the recommendations dashboard header.
//
// It mirrors this TypeScript interface exactly:
//
//	export interface DashboardStats {
//	  recommendationsToday: number
//	  averageBudget: number
//	  mostSelectedCategory: string
//	  averageAIMatch: number
//	}
//
// The frontend currently computes these client-side from its product fixture
// (RecommendationsPage.tsx). Serving them from the API turns invented numbers
// into real ones — averageAIMatch in particular becomes the mean of what the
// engine actually scored rather than `78 + avgRating * 4`.
type DashboardStats struct {
	RecommendationsToday int     `json:"recommendationsToday"`
	AverageBudget        int     `json:"averageBudget"`
	MostSelectedCategory string  `json:"mostSelectedCategory"`
	AverageAIMatch       int     `json:"averageAIMatch"`
	TotalProducts        int64   `json:"totalProducts"`
	TotalRecommendations int64   `json:"totalRecommendations"`
	CategoryBreakdown    []Slice `json:"categoryBreakdown,omitempty"`
	BudgetBreakdown      []Slice `json:"budgetBreakdown,omitempty"`
}

// Slice is one labelled bucket, shaped for the frontend's recharts components.
type Slice struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// RecommendationSummary is the wire shape of a stored recommendation run.
type RecommendationSummary struct {
	ID        string    `json:"id"`
	Prompt    string    `json:"prompt"`
	Summary   string    `json:"summary"`
	Reason    string    `json:"reason"`
	Budget    string    `json:"budget"`
	Occasion  string    `json:"occasion"`
	Style     string    `json:"style"`
	Source    string    `json:"source"`
	CreatedAt string    `json:"createdAt"`
	Products  []Product `json:"products"`
}

// NewRecommendationSummary projects a stored recommendation onto the wire.
//
// The products carry the match percentages and reasons recorded at the time the
// shortlist was produced, not fresh ones. A stored recommendation is a record
// of what a shopper was told; recomputing it against today's catalogue would
// make their history quietly disagree with their memory.
func NewRecommendationSummary(r *models.Recommendation) RecommendationSummary {
	products := make([]Product, 0, len(r.Items))
	for i := range r.Items {
		item := &r.Items[i]
		if item.Product == nil {
			// Deleted since. Shorten the grid rather than emit a null the
			// frontend's .map() would fault on.
			continue
		}
		p := NewProduct(item.Product)
		match := item.MatchPercent
		p.MatchPercent = &match
		p.MatchReasons = decodeStringList(item.Reasons)
		products = append(products, p)
	}

	return RecommendationSummary{
		ID:        r.ID.String(),
		Prompt:    r.Prompt,
		Summary:   r.Summary,
		Reason:    r.Reason,
		Budget:    r.BudgetLabel,
		Occasion:  r.Occasion,
		Style:     r.Style,
		Source:    r.Source,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		Products:  products,
	}
}

// NewStoredRecommendation renders a stored recommendation in the assistant's
// response shape, so a replayed shortlist and a fresh one are the same object
// to the client.
func NewStoredRecommendation(r *models.Recommendation) *AIRecommendationResponse {
	summary := NewRecommendationSummary(r)
	return &AIRecommendationResponse{
		Summary:          summary.Summary,
		Budget:           summary.Budget,
		Occasion:         summary.Occasion,
		Style:            summary.Style,
		Reason:           summary.Reason,
		Products:         summary.Products,
		RecommendationID: summary.ID,
		Meta: &AssistantMeta{
			Source:    r.Source,
			LatencyMS: r.LatencyMS,
		},
	}
}

// decodeStringList unpacks a JSONB []string, tolerating absence and drift.
func decodeStringList(raw models.JSONB) []string {
	if len(raw) == 0 {
		return nil
	}
	var out []string
	if err := raw.Decode(&out); err != nil {
		return nil
	}
	return out
}
