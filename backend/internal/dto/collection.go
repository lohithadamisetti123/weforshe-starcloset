package dto

import (
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
)

// CollectionRequest is the body of POST /assistant/collection.
//
// Department drives everything downstream — which gender/category the
// slot-filling is scoped to — and is carried explicitly on every request
// rather than remembered server-side. The frontend's conversation state
// holds "the shopper picked Women", not this API: a stateless request here
// is what lets "show another one under ₹6000" work by simply resending the
// same department with a new prompt, with no session concept required.
type CollectionRequest struct {
	Department string `json:"department" binding:"required" validate:"required"`
	Prompt     string `json:"prompt" binding:"required" validate:"required,min=2,max=500"`

	// ExcludeProductIds lets the client ask for collections drawn from products
	// it hasn't already been shown — this is what makes "More Collections"
	// return genuinely new sets rather than repeating the deterministic result.
	// Optional and defaulted empty, so the singular endpoint and first-page
	// requests are unaffected.
	ExcludeProductIds []string `json:"excludeProductIds,omitempty" validate:"omitempty,max=200"`

	// PersonaSlug names the StarCloset stylist styling this request. Optional:
	// an empty slug (or an unknown one) runs the engine exactly as before, with
	// no persona bias and no stylist reply. When set, the persona's Style DNA is
	// merged into the ranking context and its voice produces the stylistReply.
	PersonaSlug string `json:"personaSlug,omitempty" validate:"omitempty,max=80"`
}

// DepartmentOption is one first-step choice, with whether it currently has
// catalogue support — see recommendation.DepartmentSupport. The frontend
// renders every entry but can grey out or omit the unsupported ones without
// hard-coding the list itself.
type DepartmentOption struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// NewDepartmentOptions lists every department the product spec calls for,
// each flagged with whether it currently has real inventory behind it.
func NewDepartmentOptions() []DepartmentOption {
	depts := recommendation.AllDepartments
	out := make([]DepartmentOption, len(depts))
	for i, d := range depts {
		out[i] = DepartmentOption{Name: string(d), Available: recommendation.DepartmentSupport(d)}
	}
	return out
}

// CollectionItem is one filled slot in the response.
type CollectionItem struct {
	Slot    string  `json:"slot"`
	Product Product `json:"product"`
}

// CollectionResponse mirrors recommendation.Collection, translated into wire
// types the same way every other assistant response is — DTOs are the only
// thing ever serialized, never the internal engine types directly.
type CollectionResponse struct {
	Title        string           `json:"title"`
	Items        []CollectionItem `json:"items"`
	TotalPrice   int              `json:"totalPrice"`
	BudgetLimit  *int             `json:"budgetLimit,omitempty"`
	SkippedSlots []string         `json:"skippedSlots,omitempty"`
	Explanation  []string         `json:"explanation"`
	Department   string           `json:"department"`
	Meta         *AssistantMeta   `json:"meta,omitempty"`

	// StylistReply is the persona-voice narration for this collection, when a
	// stylist persona styled the request. Empty when no persona was used.
	StylistReply string `json:"stylistReply,omitempty"`
}

// CollectionsResponse carries the full set of themed collections built for one
// department + prompt in a single request.
//
// This is the contract the guided flow actually needs. The older singular
// endpoint returned one collection per request and relied on the frontend
// firing three parallel calls to get variety — but each of those calls was
// stateless, so the engine's cross-collection de-duplication (which guarantees
// no product repeats across the three themes) was computed and then thrown
// away, leaving the three responses free to overlap. Returning all themes from
// one call is what makes "three genuinely different collections" true rather
// than aspirational.
type CollectionsResponse struct {
	Department  string               `json:"department"`
	Collections []CollectionResponse `json:"collections"`
	Meta        *AssistantMeta       `json:"meta,omitempty"`

	// StylistReply is the persona-voice narration for this set of collections,
	// generated deterministically from the chosen stylist's templates. Empty
	// when no persona styled the request.
	StylistReply string `json:"stylistReply,omitempty"`
}

// NewCollectionsResponse builds the plural wire response from the engine's
// de-duplicated set of themed collections.
func NewCollectionsResponse(department string, cs []*recommendation.Collection, meta *AssistantMeta) CollectionsResponse {
	out := make([]CollectionResponse, 0, len(cs))
	for _, c := range cs {
		if c == nil {
			continue
		}
		// Meta is attached once at the top level, not repeated per collection.
		out = append(out, NewCollectionResponse(department, c, nil))
	}
	return CollectionsResponse{
		Department:  department,
		Collections: out,
		Meta:        meta,
	}
}

// NewCollectionResponse builds the wire response from the engine's result.
func NewCollectionResponse(department string, c *recommendation.Collection, meta *AssistantMeta) CollectionResponse {
	items := make([]CollectionItem, len(c.Items))
	for i, it := range c.Items {
		items[i] = CollectionItem{Slot: it.Slot, Product: NewScoredProduct(it.Product)}
	}
	return CollectionResponse{
		Title:        c.Title,
		Items:        items,
		TotalPrice:   c.TotalPrice,
		BudgetLimit:  c.BudgetLimit,
		SkippedSlots: c.SkippedSlots,
		Explanation:  c.Explanation,
		Department:   department,
		Meta:         meta,
	}
}
