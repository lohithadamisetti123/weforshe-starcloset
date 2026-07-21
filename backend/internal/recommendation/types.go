// Package recommendation implements the ranking engine that turns a shopper's
// natural-language brief into an ordered, explainable shortlist.
//
// The engine is deliberately rule-based — no model training, no embeddings, no
// vector store. That is not a limitation to apologise for: with a catalogue of
// this size and a hard requirement that every pick be explainable to a
// merchandiser, a well-tuned scoring function beats a black box on both
// relevance and debuggability, and it responds in microseconds.
//
// The pipeline has five stages:
//
//  1. Understand  — an Intent is extracted from the prompt (internal/ai).
//  2. Retrieve    — a candidate pool is pulled with generous soft constraints.
//  3. Score       — each candidate is scored across ten independent signals
//     with query-adaptive weights (scoring.go).
//  4. Diversify   — a greedy MMR pass stops the shortlist collapsing into six
//     near-identical products (diversity.go).
//  5. Explain     — raw scores are calibrated to a match % and turned into
//     human sentences (calibrate.go, explain.go).
//
// Every stage is pure and independently testable; the engine itself only wires
// them together.
package recommendation

import (
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// BudgetRange is the price envelope the shopper asked for, in whole rupees.
// Min and Max are independently optional: "under ₹3000" sets only Max,
// "above ₹2000" only Min, "between ₹2000 and ₹5000" both.
type BudgetRange struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

// HasMax reports whether an upper bound was expressed.
func (b *BudgetRange) HasMax() bool { return b != nil && b.Max != nil && *b.Max > 0 }

// HasMin reports whether a lower bound was expressed.
func (b *BudgetRange) HasMin() bool { return b != nil && b.Min != nil && *b.Min > 0 }

// IsZero reports whether no bound at all was expressed.
func (b *BudgetRange) IsZero() bool { return !b.HasMax() && !b.HasMin() }

// Ceiling returns the effective upper bound, falling back to 0 when unbounded.
func (b *BudgetRange) Ceiling() int {
	if b.HasMax() {
		return *b.Max
	}
	return 0
}

// Floor returns the effective lower bound, falling back to 0.
func (b *BudgetRange) Floor() int {
	if b.HasMin() {
		return *b.Min
	}
	return 0
}

// SortHint is an explicit ordering the shopper asked for, overriding pure
// relevance ordering. "Show me the cheapest kurta" is a different request from
// "show me a kurta", and flattening the two loses real intent.
type SortHint string

// SortHint values.
const (
	SortRelevance SortHint = "relevance"
	SortPriceAsc  SortHint = "price_asc"
	SortPriceDesc SortHint = "price_desc"
	SortRating    SortHint = "rating"
	SortDiscount  SortHint = "discount"
	SortNewest    SortHint = "newest"
)

// IntentSource records which understanding path produced an Intent. It is
// persisted with every recommendation so that a relevance regression can be
// attributed to the LLM or to the rules without guesswork.
type IntentSource string

// IntentSource values.
const (
	SourceRules  IntentSource = "rules"  // deterministic extractor only
	SourceGemini IntentSource = "gemini" // LLM only
	SourceHybrid IntentSource = "hybrid" // both, merged field-by-field
)

// Intent is the structured reading of a shopper's prompt. Every field is
// optional: an empty Intent is a valid "show me something good" request, and
// the engine must still return a sensible shortlist for it.
//
// Pointer fields distinguish "not mentioned" from "mentioned as the zero
// value", which matters — Budget == nil means unconstrained, whereas
// Budget.Max == 0 would be nonsense the ranker must never act on.
type Intent struct {
	// Query is the shopper's raw prompt, retained for text relevance scoring.
	Query string `json:"query"`

	Budget   *BudgetRange         `json:"budget,omitempty"`
	Occasion *models.Occasion     `json:"occasion,omitempty"`
	Style    *models.StyleTag     `json:"style,omitempty"`
	Gender   *models.Gender       `json:"gender,omitempty"`
	Category *models.CategoryName `json:"category,omitempty"`

	// CategoryIn is a hard allow-list, set only by the collection builder
	// (see collection.go) when filling a specific slot ("must be Footwear").
	// It is never set by the AI extractors and is empty for every ordinary
	// search or chat request — Category above remains the only category
	// signal those paths ever see.
	CategoryIn []models.CategoryName `json:"categoryIn,omitempty"`

	// StrictBudget tells retrieve() not to drop the price ceiling during
	// relaxation. Set by BuildCollection so slot picks stay within the
	// remaining budget rather than returning the best-rated product
	// regardless of cost.
	StrictBudget bool `json:"-"`

	// Colors and Brands are slices because "black or navy" and "Nike or Puma"
	// are ordinary asks. Any one match is a full match.
	Colors []string `json:"colors,omitempty"`
	Brands []string `json:"brands,omitempty"`

	// Exclusions carry negations ("not ethnic", "no black", "anything but Zara").
	// These are applied as hard filters rather than score penalties: a shopper
	// who says "no black" and is shown black has been ignored, not merely
	// under-served.
	ExcludeColors []string          `json:"excludeColors,omitempty"`
	ExcludeStyles []models.StyleTag `json:"excludeStyles,omitempty"`
	ExcludeBrands []string          `json:"excludeBrands,omitempty"`
	// ExcludeProducts prevents specific product IDs from being returned.
	// Used by BuildThemeCollections to ensure each themed collection draws
	// from a distinct set of products, never repeating a product across themes.
	ExcludeProducts []string `json:"-"`

	// Keywords are the content-bearing terms left after stop-word removal,
	// used by the BM25 text signal.
	Keywords []string `json:"keywords,omitempty"`

	// TypeTerms are canonical product-type keys for a requested type (watch,
	// necklace, sneakers, saree…). They narrow the pool within a broad
	// category — "watch" within Accessories, "heels" within Footwear — where
	// the category alone can't distinguish sub-types. Set by the rule
	// extractor and resolved against the type vocabulary; see producttype.go.
	TypeTerms []string `json:"typeTerms,omitempty"`

	Sort SortHint `json:"sort,omitempty"`

	// Confidence is the extractor's self-assessed reliability in [0,1]. It
	// damps the semantic signals: a hesitant reading should not out-shout a
	// product's own quality.
	Confidence float64 `json:"confidence"`

	// Source records the extraction path taken.
	Source IntentSource `json:"source"`

	// Notes carries extractor commentary for observability. Never user-facing.
	Notes []string `json:"notes,omitempty"`

	// Personalization, when set, carries the shopper-specific preference and
	// history data used by the additive boost layer (see personalize.go). It is
	// intentionally NOT part of intent extraction — the AI extractors never set
	// it — and NOT serialized: it is an engine input, not part of the prompt's
	// meaning. When nil (every existing search, chat and test path), the engine
	// behaves exactly as before. Because it rides on the Intent, it propagates
	// unchanged into every collection slot (which shallow-copy the intent), so
	// personalization reaches both direct chat and collection building through
	// this single field.
	Personalization *PersonalizationContext `json:"-"`
}

// NewIntent returns an Intent seeded with the raw query and neutral defaults.
func NewIntent(query string) *Intent {
	return &Intent{
		Query:      strings.TrimSpace(query),
		Sort:       SortRelevance,
		Confidence: 0.5,
		Source:     SourceRules,
	}
}

// HasSemanticSignal reports whether the shopper expressed anything the ranker
// can act on beyond raw text. When false the engine switches to a
// quality-and-value-led "best of the catalogue" strategy rather than pretending
// to have understood a brief that was never given.
func (i *Intent) HasSemanticSignal() bool {
	if i == nil {
		return false
	}
	return i.Occasion != nil || i.Style != nil || i.Gender != nil ||
		i.Category != nil || len(i.Colors) > 0 || len(i.Brands) > 0 ||
		!i.Budget.IsZero()
}

// SpecifiedCount returns how many distinct constraints the shopper expressed.
// It feeds the confidence calibration: a six-constraint brief that scores well
// is a far stronger match than a one-word prompt that scores the same.
func (i *Intent) SpecifiedCount() int {
	if i == nil {
		return 0
	}
	n := 0
	if i.Occasion != nil {
		n++
	}
	if i.Style != nil {
		n++
	}
	if i.Gender != nil {
		n++
	}
	if i.Category != nil {
		n++
	}
	if !i.Budget.IsZero() {
		n++
	}
	if len(i.Colors) > 0 {
		n++
	}
	if len(i.Brands) > 0 {
		n++
	}
	return n
}

// OccasionOrEmpty returns the occasion as a plain string, or "".
func (i *Intent) OccasionOrEmpty() string {
	if i == nil || i.Occasion == nil {
		return ""
	}
	return string(*i.Occasion)
}

// StyleOrEmpty returns the style as a plain string, or "".
func (i *Intent) StyleOrEmpty() string {
	if i == nil || i.Style == nil {
		return ""
	}
	return string(*i.Style)
}

// GenderOrEmpty returns the gender as a plain string, or "".
func (i *Intent) GenderOrEmpty() string {
	if i == nil || i.Gender == nil {
		return ""
	}
	return string(*i.Gender)
}

// CategoryOrEmpty returns the category as a plain string, or "".
func (i *Intent) CategoryOrEmpty() string {
	if i == nil || i.Category == nil {
		return ""
	}
	return string(*i.Category)
}

// Signal names. These are the keys of a Breakdown and appear verbatim in the
// persisted jsonb, so they are constants rather than inline literals.
const (
	SignalOccasion = "occasion"
	SignalStyle    = "style"
	SignalGender   = "gender"
	SignalCategory = "category"
	SignalBudget   = "budget"
	SignalColor    = "color"
	SignalBrand    = "brand"
	SignalText     = "text"
	SignalQuality  = "quality"
	SignalValue    = "value"
)

// SignalOrder is the canonical ordering used when explaining a score, roughly
// by how much a shopper cares about each dimension.
var SignalOrder = []string{
	SignalOccasion, SignalBudget, SignalStyle, SignalCategory, SignalGender,
	SignalColor, SignalBrand, SignalText, SignalQuality, SignalValue,
}

// SignalScore is one dimension's contribution to a product's total.
type SignalScore struct {
	// Raw is the signal's own 0..1 assessment.
	Raw float64 `json:"raw"`
	// Weight is the query-adaptive weight this signal carried.
	Weight float64 `json:"weight"`
	// Contribution is Raw*Weight — what actually moved the total.
	Contribution float64 `json:"contribution"`
	// Active reports whether the query gave the engine anything to judge this
	// signal against. Inactive signals are excluded from explanations.
	Active bool `json:"active"`
}

// Breakdown maps signal name to its score. Persisted per recommended product,
// which is what makes "why was this ranked first?" a query rather than an
// investigation.
type Breakdown map[string]SignalScore

// Total returns the sum of all contributions — the product's relevance in [0,1].
func (b Breakdown) Total() float64 {
	var t float64
	for _, s := range b {
		t += s.Contribution
	}
	return t
}

// TopSignals returns up to n active signal names ordered by contribution,
// highest first. Ties break on SignalOrder so output is deterministic.
func (b Breakdown) TopSignals(n int) []string {
	type kv struct {
		k string
		v float64
	}
	ranked := make([]kv, 0, len(b))
	for _, name := range SignalOrder {
		s, ok := b[name]
		if !ok || !s.Active || s.Contribution <= 0 {
			continue
		}
		ranked = append(ranked, kv{name, s.Contribution})
	}
	// Insertion sort: the slice is at most ten elements and stability against
	// SignalOrder is required, which sort.Slice would not give for free.
	for i := 1; i < len(ranked); i++ {
		for j := i; j > 0 && ranked[j].v > ranked[j-1].v; j-- {
			ranked[j], ranked[j-1] = ranked[j-1], ranked[j]
		}
	}
	if n > len(ranked) {
		n = len(ranked)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, ranked[i].k)
	}
	return out
}

// ScoredProduct is a catalogue product with everything the engine concluded
// about it for one particular query.
type ScoredProduct struct {
	Product *models.Product

	// Score is the pre-diversification relevance in [0,1].
	Score float64
	// FinalScore is the score after the MMR diversity pass. Ordering uses this.
	FinalScore float64
	// MatchPercent is the calibrated 60..99 figure the UI shows.
	MatchPercent int
	// Rank is the 1-based final position.
	Rank int

	Breakdown Breakdown
	// Reasons are short human-readable justifications, most important first.
	Reasons []string
}

// Result is a complete engine run.
type Result struct {
	Intent   *Intent
	Products []ScoredProduct

	// Summary and Reason are the assistant-voice strings the frontend renders
	// verbatim.
	Summary string
	Reason  string

	// BudgetLabel, OccasionLabel and StyleLabel are display strings, already
	// formatted for the UI (e.g. "₹3,000", "everyday", "versatile").
	BudgetLabel   string
	OccasionLabel string
	StyleLabel    string

	// PoolSize is how many candidates were scored, and Relaxed reports whether
	// a constraint had to be loosened to fill the shortlist. Both are surfaced
	// as response metadata so the client can be honest about near-misses.
	PoolSize int
	Relaxed  []string
}
