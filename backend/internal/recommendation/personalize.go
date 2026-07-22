package recommendation

import (
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// This file adds *personalization* to the engine without changing how the
// engine ranks. It is deliberately additive and self-contained:
//
//   - The ten-signal scorer (scoring.go), the MMR pass, the relaxation ladder
//     and the collection builder are untouched. None of their logic is edited.
//   - Personalization enters through exactly one seam: the engine, after it has
//     scored the full candidate pool and before it sorts, calls
//     ApplyPersonalization. When no PersonalizationContext is attached to the
//     intent (the default for every existing test and code path), that call is
//     a no-op and the ranking is byte-identical to today.
//   - The boosts are ADDITIVE nudges on the base relevance score, and their
//     total is hard-capped (MaxTotalBoost). They can reorder products *within*
//     the set the engine already judged relevant; they can never override the
//     hard filters (gender, budget bounds live in retrieve() as WHERE clauses,
//     never here) and are capped small enough that they cannot out-rank a
//     materially better occasion/style/category fit.
//
// The data itself (which brands/colors/styles the shopper prefers, what they've
// bought, where they live) is assembled elsewhere — the personalization
// *builder* in internal/personalization loads it from the profile and order
// repositories. Swapping that source for real Myntra account data changes the
// builder, not this file and not the pipeline. That separation is the whole
// point: this module knows how to turn preferences into score nudges; it does
// not know or care where the preferences came from.

// BoostWeights are the maximum additive contribution each personalization
// signal can make to a product's [0,1] relevance score. They are intentionally
// small: personalization is a tie-breaker among already-relevant products, not
// a competing notion of relevance. A shopper who asked for a "wedding saree"
// must still be shown wedding sarees first, with their preferred brand
// surfacing among them — never a casual tee from that brand ahead of the saree.
type BoostWeights struct {
	PreferredBrand    float64
	PreferredColor    float64
	PreferredStyle    float64
	PreferredCategory float64
	PurchasedBrand    float64
	PurchasedCategory float64
	Location          float64
}

// DefaultBoostWeights are tuned so that a product hitting several preferences
// gains a visible-but-bounded edge. With the cap below, the most a fully
// on-preference product can gain is MaxTotalBoost — comfortably less than the
// occasion signal's weight, so occasion fit still wins.
var DefaultBoostWeights = BoostWeights{
	PreferredBrand:    0.05,
	PreferredColor:    0.03,
	PreferredStyle:    0.04,
	PreferredCategory: 0.04,
	PurchasedBrand:    0.05,
	PurchasedCategory: 0.03,
	Location:          0.03,
}

// MaxTotalBoost caps the sum of all personalization boosts applied to one
// product. This is the guarantee that personalization stays a nudge: even a
// product that matches every stated preference and the shopper's city cannot
// gain more than this, keeping it below the weight of the mandatory-intent
// signals it must never override.
const MaxTotalBoost = 0.15

// PersonalizationContext is the shopper-specific data the boosts read. It is
// plain, engine-facing data with no I/O and no knowledge of persistence, so the
// boost math is as pure and testable as the rest of the recommendation package.
//
// Every field is optional. A zero PersonalizationContext produces no boost,
// which is exactly what should happen for a shopper who has not onboarded.
type PersonalizationContext struct {
	// Stated preferences from onboarding.
	PreferredBrands     []string
	PreferredColors     []string
	PreferredStyles     []models.StyleTag
	PreferredCategories []models.CategoryName

	// Behavioural signals derived from purchase history. Kept separate from the
	// stated preferences because "bought before" and "said they like" are
	// different strengths of evidence and are weighted independently. Later,
	// wishlist / browsing / recently-viewed become additional fields here and
	// additional clauses in boost() — no pipeline change required.
	PurchasedBrands     []string
	PurchasedCategories []models.CategoryName

	// City drives the location signal (a soft climate/culture prior).
	City string

	// Weights allows an operator or test to override the defaults; the zero
	// value means "use DefaultBoostWeights".
	Weights *BoostWeights
}

// IsEmpty reports whether the context carries anything to personalize on.
func (pc *PersonalizationContext) IsEmpty() bool {
	if pc == nil {
		return true
	}
	return len(pc.PreferredBrands) == 0 && len(pc.PreferredColors) == 0 &&
		len(pc.PreferredStyles) == 0 && len(pc.PreferredCategories) == 0 &&
		len(pc.PurchasedBrands) == 0 && len(pc.PurchasedCategories) == 0 &&
		strings.TrimSpace(pc.City) == ""
}

func (pc *PersonalizationContext) weights() BoostWeights {
	if pc != nil && pc.Weights != nil {
		return *pc.Weights
	}
	return DefaultBoostWeights
}

// ApplyPersonalization adds each product's personalization boost to its base
// relevance score, in place. It is the single integration seam with the engine.
//
// It reads intent.Personalization; when that is nil or empty it returns without
// touching a single score, which is why attaching no context leaves the
// existing ranking exactly as it was. The boost is applied to Score (the
// pre-diversification relevance), so every downstream stage — the sort, the MMR
// pass, the match-percent calibration — reflects it naturally, with no change
// to any of those stages.
func ApplyPersonalization(scored []ScoredProduct, intent *Intent) {
	if intent == nil || intent.Personalization == nil {
		return
	}
	pc := intent.Personalization
	if pc.IsEmpty() {
		return
	}
	for i := range scored {
		if scored[i].Product == nil {
			continue
		}
		boost, _ := pc.boost(scored[i].Product)
		if boost <= 0 {
			continue
		}
		scored[i].Score = clamp01(scored[i].Score + boost)
	}
}

// boost returns the total additive boost for one product, capped at
// MaxTotalBoost, together with short human-readable reasons for the parts that
// fired (used by ReasonsFor for explainability). Returning the reasons here
// keeps the "why boosted" logic next to the "how much" logic so they cannot
// drift apart.
func (pc *PersonalizationContext) boost(p *models.Product) (float64, []string) {
	if pc == nil || p == nil {
		return 0, nil
	}
	w := pc.weights()

	type part struct {
		score  float64
		reason string
	}
	var parts []part

	// --- Preferred brand -----------------------------------------------------
	if m := BestBrandMatch(pc.PreferredBrands, p.Brand); m > 0 {
		parts = append(parts, part{m * w.PreferredBrand, "One of your favourite brands"})
	}

	// --- Preferred colour ----------------------------------------------------
	if m := BestColorMatch(pc.PreferredColors, p.Color); m > 0 {
		parts = append(parts, part{m * w.PreferredColor, "In a colour you like"})
	}

	// --- Preferred style -----------------------------------------------------
	// Reuses the engine's own StyleAffinity, so "likes Ethnic" also gives
	// partial credit to a Festive/Traditional piece, consistent with how the
	// scorer treats style everywhere else.
	if m := bestStyleAffinity(pc.PreferredStyles, p.Style); m > 0 {
		parts = append(parts, part{m * w.PreferredStyle, "Matches a style you prefer"})
	}

	// --- Preferred category --------------------------------------------------
	if p.Category != nil && containsCategory(pc.PreferredCategories, p.Category.Name) {
		parts = append(parts, part{w.PreferredCategory, "In a category you shop"})
	}

	// --- Previously purchased brand -----------------------------------------
	if m := BestBrandMatch(pc.PurchasedBrands, p.Brand); m > 0 {
		parts = append(parts, part{m * w.PurchasedBrand, "From a brand you've bought before"})
	}

	// --- Previously purchased category --------------------------------------
	if p.Category != nil && containsCategory(pc.PurchasedCategories, p.Category.Name) {
		parts = append(parts, part{w.PurchasedCategory, "In a category you've bought before"})
	}

	// --- Location ------------------------------------------------------------
	if ls, reason := locationBoost(pc.City, p, w.Location); ls > 0 {
		parts = append(parts, part{ls, reason})
	}

	var total float64
	reasons := make([]string, 0, len(parts))
	for _, pt := range parts {
		total += pt.score
		if pt.reason != "" {
			reasons = append(reasons, pt.reason)
		}
	}
	if total > MaxTotalBoost {
		total = MaxTotalBoost
	}
	return total, reasons
}

// ReasonsFor returns the personalization reasons that apply to a product, most
// impactful first, for a context. It is pure and read-only, intended to be
// called at the response boundary (by the DTO/service layer) to annotate why a
// product was surfaced for this shopper — without touching the engine's own
// explain layer.
func (pc *PersonalizationContext) ReasonsFor(p *models.Product) []string {
	if pc.IsEmpty() || p == nil {
		return nil
	}
	_, reasons := pc.boost(p)
	return reasons
}

// bestStyleAffinity returns the strongest StyleAffinity across the preferred
// styles.
func bestStyleAffinity(wanted []models.StyleTag, productStyle models.StyleTag) float64 {
	var best float64
	for _, w := range wanted {
		if m := StyleAffinity(w, productStyle); m > best {
			best = m
		}
	}
	return best
}

func containsCategory(list []models.CategoryName, c models.CategoryName) bool {
	for _, x := range list {
		if x == c {
			return true
		}
	}
	return false
}

// ─── Location signal ─────────────────────────────────────────────────────────
//
// Location is a soft climate/culture prior, never a hard rule and never a
// hardcoded recommendation. A shopper's city nudges products whose style,
// category or colour suits the local climate or wardrobe culture — Hyderabad
// and Chennai lean to lightweight/breathable fabrics, Delhi to layering and
// warmer pieces, Kolkata to ethnic/festive. It only ever adds a small boost;
// it never filters, and a product that matches nothing about the city simply
// gets no location boost.

type cityPrior struct {
	styles     []models.StyleTag
	categories []models.CategoryName
	// colorHint terms match loosely against the product's free-text colour.
	colorHints []string
}

// cityPriors is keyed by normalised city name. It is data, not logic, so adding
// a city is a one-line change and never a code path. Unknown cities fall
// through to no location boost, which is the correct, safe default.
var cityPriors = map[string]cityPrior{
	// Warm / humid — lightweight, breathable, casual leanings.
	"hyderabad": {styles: []models.StyleTag{models.StyleCasual}, colorHints: []string{"white", "beige", "sky", "light", "pastel", "linen"}},
	"chennai":   {styles: []models.StyleTag{models.StyleCasual}, colorHints: []string{"white", "beige", "cotton", "light", "pastel"}},
	"mumbai":    {styles: []models.StyleTag{models.StyleCasual, models.StyleWestern}, colorHints: []string{"white", "light", "pastel"}},
	"bengaluru": {styles: []models.StyleTag{models.StyleCasual, models.StyleWestern}},
	"bangalore": {styles: []models.StyleTag{models.StyleCasual, models.StyleWestern}},
	"pune":      {styles: []models.StyleTag{models.StyleCasual, models.StyleWestern}},
	"goa":       {styles: []models.StyleTag{models.StyleCasual, models.StyleWestern}, colorHints: []string{"white", "floral", "bright"}},

	// Cooler / layering leanings.
	"delhi":      {styles: []models.StyleTag{models.StyleFormal}, categories: []models.CategoryName{models.CategoryBlazers}, colorHints: []string{"black", "navy", "grey", "maroon", "dark"}},
	"new delhi":  {styles: []models.StyleTag{models.StyleFormal}, categories: []models.CategoryName{models.CategoryBlazers}, colorHints: []string{"black", "navy", "grey", "maroon", "dark"}},
	"gurgaon":    {styles: []models.StyleTag{models.StyleFormal}, categories: []models.CategoryName{models.CategoryBlazers}},
	"noida":      {styles: []models.StyleTag{models.StyleFormal}, categories: []models.CategoryName{models.CategoryBlazers}},
	"chandigarh": {styles: []models.StyleTag{models.StyleFormal}, categories: []models.CategoryName{models.CategoryBlazers}},
	"jaipur":     {styles: []models.StyleTag{models.StyleEthnic}, categories: []models.CategoryName{models.CategoryEthnicWear}},
	"lucknow":    {styles: []models.StyleTag{models.StyleEthnic}, categories: []models.CategoryName{models.CategoryKurtaSets}},

	// Strong ethnic / festive wardrobe culture.
	"kolkata":   {styles: []models.StyleTag{models.StyleEthnic}, categories: []models.CategoryName{models.CategorySarees, models.CategoryEthnicWear}},
	"ahmedabad": {styles: []models.StyleTag{models.StyleEthnic}, categories: []models.CategoryName{models.CategoryEthnicWear, models.CategoryKurtaSets}},
}

// locationBoost returns the location contribution and a reason for one product.
func locationBoost(city string, p *models.Product, weight float64) (float64, string) {
	key := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(city)), " "))
	if key == "" || weight <= 0 {
		return 0, ""
	}
	prior, ok := cityPriors[key]
	if !ok {
		return 0, ""
	}

	// The three sub-signals are OR'd and the boost is the weight scaled by how
	// many matched, so a product that fits the city on style *and* category
	// edges out one that fits on only one — without any single match ever
	// approaching the mandatory signals in size.
	matches := 0
	for _, s := range prior.styles {
		if StyleAffinity(s, p.Style) >= 0.6 {
			matches++
			break
		}
	}
	if p.Category != nil {
		for _, c := range prior.categories {
			if c == p.Category.Name {
				matches++
				break
			}
		}
	}
	lowerColor := strings.ToLower(p.Color)
	for _, hint := range prior.colorHints {
		if strings.Contains(lowerColor, hint) {
			matches++
			break
		}
	}

	if matches == 0 {
		return 0, ""
	}
	// Scale 1..3 matches across [0.5, 1.0] of the weight so even a single match
	// contributes, but a fuller fit contributes more.
	scale := 0.5 + 0.25*float64(minInt(matches, 3)-1)
	return weight * scale, "Suited to " + titleCity(key)
}

// titleCity upper-cases the first letter of each word in a normalised city key
// for display ("new delhi" → "New Delhi"). Kept local rather than using the
// deprecated strings.Title.
func titleCity(key string) string {
	words := strings.Fields(key)
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}
