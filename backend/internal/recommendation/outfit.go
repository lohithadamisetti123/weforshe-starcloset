package recommendation

import (
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// ─── Outfit compatibility ────────────────────────────────────────────────────
//
// The difference between "six well-ranked products" and "an outfit" is that an
// outfit's pieces are chosen with reference to each other. This file scores a
// candidate against the pieces already chosen, so the second, third and fourth
// slots are filled by what goes *with* the anchor rather than by what would have
// won an independent popularity contest.
//
// Everything here is a bounded nudge on top of the existing relevance score, not
// a filter. Style is subjective and the catalogue's colour strings are free
// text; a hard rule built on either would empty slots and make the collection
// worse. The cap keeps a merely-compatible item from beating a much better one.

// maxCompatBonus bounds the total styling adjustment.
const maxCompatBonus = 0.25

type colorFamily int

const (
	colorUnknown    colorFamily = iota
	colorNeutral                // goes with everything: black, white, grey, navy, beige
	colorMetallic               // gold/silver — neutral for accessories
	colorWarm                   // red, rust, mustard, coral
	colorCool                   // blue, green, teal, olive
	colorPinkPurple             // pink, lavender, wine
)

// colorFamilyTokens is an ORDERED list, not a map, and the order is load-bearing.
// Map iteration in Go is randomised, so classifying "navy blue" against a map
// would land on neutral or cool depending on the run — making the same request
// return different outfits. Neutrals are tested first because a colour naming a
// neutral is a neutral ("navy blue" is navy).
var colorFamilyTokens = []struct {
	family colorFamily
	terms  []string
}{
	{colorNeutral, []string{
		"black", "white", "grey", "gray", "charcoal", "ivory", "cream", "beige",
		"tan", "brown", "khaki", "navy", "denim", "nude", "offwhite", "stone", "camel",
	}},
	{colorMetallic, []string{"gold", "golden", "silver", "bronze", "copper", "rosegold", "metallic", "gunmetal"}},
	{colorWarm, []string{"red", "maroon", "rust", "orange", "coral", "peach", "mustard", "yellow", "amber", "terracotta"}},
	{colorCool, []string{"blue", "teal", "green", "olive", "mint", "turquoise", "aqua", "sky", "emerald", "sage"}},
	{colorPinkPurple, []string{"pink", "magenta", "fuchsia", "lavender", "purple", "mauve", "wine", "plum", "lilac", "burgundy"}},
}

// familyOfColor classifies a free-text colour string.
func familyOfColor(color string) colorFamily {
	if strings.TrimSpace(color) == "" {
		return colorUnknown
	}
	tokens := tokenize(color)
	for _, entry := range colorFamilyTokens {
		for _, term := range entry.terms {
			if anyTokenMatches(tokens, term) {
				return entry.family
			}
		}
	}
	return colorUnknown
}

// colorHarmony scores how well two colours sit together, 0..1.
//
// The rules are the ones a stylist would state plainly: neutrals and metallics
// go with anything; tonal (same family) pairings are safe; two saturated
// accents from opposing families are the only combination actively discouraged.
func colorHarmony(a, b string) float64 {
	fa, fb := familyOfColor(a), familyOfColor(b)
	if fa == colorUnknown || fb == colorUnknown {
		return 0.6 // no information — stay neutral rather than guess
	}
	if fa == colorNeutral || fb == colorNeutral || fa == colorMetallic || fb == colorMetallic {
		return 1.0
	}
	if fa == fb {
		return 0.85
	}
	return 0.45
}

// compatibilityBonus scores a candidate against the outfit assembled so far.
//
// anchor is the first apparel piece chosen (the item the rest of the look is
// built around); picked is everything chosen so far, used for brand diversity.
func compatibilityBonus(
	cand *models.Product,
	spec SlotSpec,
	tmpl OccasionTemplate,
	anchor *models.Product,
	picked []CollectionItem,
) float64 {
	if cand == nil {
		return 0
	}

	var bonus float64

	// ── Formality coherence ──────────────────────────────────────────────
	// Distance from the slot's target dressiness. A sneaker in an office look
	// and an oxford in a beach look are both wrong, and this is what says so.
	target := spec.Formality
	if target == 0 {
		target = tmpl.Formality
	}
	gap := formalityOf(cand) - target
	if gap < 0 {
		gap = -gap
	}
	bonus += (0.5 - gap) * 0.16 // +0.08 at a perfect match, -0.08 at the extreme

	// ── Colour harmony with the anchor ───────────────────────────────────
	if anchor != nil && anchor.ID != cand.ID {
		bonus += (colorHarmony(anchor.Color, cand.Color) - 0.6) * 0.15
	}

	// ── Style coherence ──────────────────────────────────────────────────
	// Reuses the engine's existing style affinity so "goes together" means the
	// same thing here as everywhere else in the scorer.
	if anchor != nil {
		bonus += (StyleAffinity(anchor.Style, cand.Style) - 0.5) * 0.10
	}
	for _, s := range tmpl.Styles {
		if StyleAffinity(s, cand.Style) >= 0.9 {
			bonus += 0.04
			break
		}
	}

	// ── Footwear / accessory rules ───────────────────────────────────────
	// PreferTypes is how a template says "loafers or heels for the office"
	// without banning everything else and risking an empty slot.
	if len(spec.PreferTypes) > 0 && productMatchesAnyType(cand, spec.PreferTypes) {
		bonus += 0.10
	}

	// ── Occasion reinforcement ───────────────────────────────────────────
	if tmpl.Key != "" && cand.Occasion != "" {
		if k, ok := occasionTemplateKey(cand.Occasion); ok && k == tmpl.Key {
			bonus += 0.06
		}
	}

	// ── Brand diversity within the outfit ────────────────────────────────
	// A head-to-toe single-brand look reads like a catalogue page, not a
	// stylist's pick.
	candBrand := normBrand(cand.Brand)
	for _, it := range picked {
		if it.Product.Product != nil && normBrand(it.Product.Product.Brand) == candBrand {
			bonus -= 0.07
			break
		}
	}

	if bonus > maxCompatBonus {
		bonus = maxCompatBonus
	}
	if bonus < -maxCompatBonus {
		bonus = -maxCompatBonus
	}
	return bonus
}
