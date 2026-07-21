package dto

import (
	"github.com/myntra/bharat-ai-backend/internal/models"
)

// HistoryEntry is the wire shape of one past prompt.
//
// It mirrors this TypeScript interface exactly:
//
//	export interface HistoryEntry {
//	  id: string
//	  prompt: string
//	  date: string
//	  budget: string
//	  occasion: string
//	  productsRecommended: number
//	  status: 'Completed' | 'Saved' | 'Purchased'
//	}
//
// `date` is a plain YYYY-MM-DD string rather than an RFC3339 timestamp. That is
// not laziness: HistoryTable.tsx renders the value directly and the shipped
// fixture uses '2026-07-12', so emitting a full timestamp would put
// "2026-07-12T09:31:44Z" into a table column sized for ten characters.
type HistoryEntry struct {
	ID                  string `json:"id"`
	Prompt              string `json:"prompt"`
	Date                string `json:"date"`
	Budget              string `json:"budget"`
	Occasion            string `json:"occasion"`
	ProductsRecommended int    `json:"productsRecommended"`
	Status              string `json:"status"`

	// Additive: lets the client re-open a past run's full recommendation.
	RecommendationID string `json:"recommendationId,omitempty"`
	// Additive: the exact instant, for clients that want real sorting.
	CreatedAt string `json:"createdAt,omitempty"`
}

// dateLayout is the YYYY-MM-DD form the frontend expects.
const dateLayout = "2006-01-02"

// NewHistoryEntry maps a history model.
func NewHistoryEntry(h *models.ChatHistory) HistoryEntry {
	if h == nil {
		return HistoryEntry{}
	}
	e := HistoryEntry{
		ID:                  h.ID.String(),
		Prompt:              h.Prompt,
		Date:                h.CreatedAt.Format(dateLayout),
		Budget:              h.BudgetLabel,
		Occasion:            h.Occasion,
		ProductsRecommended: h.ProductsRecommended,
		Status:              string(h.Status),
		CreatedAt:           h.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if h.RecommendationID != nil {
		e.RecommendationID = h.RecommendationID.String()
	}
	return e
}

// NewHistoryEntries maps a slice, always returning a non-nil slice.
func NewHistoryEntries(hs []*models.ChatHistory) []HistoryEntry {
	out := make([]HistoryEntry, 0, len(hs))
	for _, h := range hs {
		out = append(out, NewHistoryEntry(h))
	}
	return out
}

// CreateHistoryRequest is the body of POST /api/v1/history.
//
// This endpoint exists because the brief asks for it, but note that the
// assistant already records history automatically — a client should not
// normally need to. It is kept for clients that want to log a prompt they
// handled themselves, and for seeding demo data.
type CreateHistoryRequest struct {
	Prompt              string `json:"prompt" binding:"required" validate:"required,min=2,max=500"`
	Budget              string `json:"budget" validate:"max=48"`
	Occasion            string `json:"occasion" validate:"max=32"`
	Style               string `json:"style" validate:"max=32"`
	ProductsRecommended int    `json:"productsRecommended" validate:"gte=0,lte=100"`
	Status              string `json:"status" validate:"omitempty,oneof=Completed Saved Purchased"`
}

// UpdateHistoryRequest is the body of PATCH /api/v1/history/:id.
type UpdateHistoryRequest struct {
	Status string `json:"status" binding:"required" validate:"required,oneof=Completed Saved Purchased"`
}
