package recommendation

import (
	"fmt"
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// Explanation turns scores back into language.
//
// A shortlist without reasons is a slot machine. The engine already knows
// exactly why each product won — the Breakdown holds the per-signal
// contributions — so throwing that away at the response boundary and shipping
// six anonymous cards would waste the most useful thing the ranker produces.
//
// Everything here is generated from the Breakdown rather than from a second
// LLM call. That is deliberate: an explanation produced by a model that cannot
// see the scoring is a plausible story about the ranking, not the ranking's
// actual cause, and it will eventually contradict the numbers next to it. These
// sentences cannot drift from the scores because they are derived from them.

// Fallback labels. These match the shipped frontend's mock exactly, so a user
// with an unparseable prompt sees the copy the design was built around.
const (
	labelFlexibleBudget = "flexible budget"
	labelEverydayOcc    = "everyday"
	labelVersatileStyle = "versatile"
)

// BudgetLabel renders the shopper's price envelope for display.
func BudgetLabel(b *BudgetRange) string {
	switch {
	case b.IsZero():
		return labelFlexibleBudget
	case b.HasMax() && b.HasMin():
		return fmt.Sprintf("%s–%s", utils.FormatINR(*b.Min), utils.FormatINR(*b.Max))
	case b.HasMax():
		return utils.FormatINR(*b.Max)
	default:
		return "above " + utils.FormatINR(*b.Min)
	}
}

// OccasionLabel renders the occasion for display, falling back to "everyday".
func OccasionLabel(i *Intent) string {
	if i == nil || i.Occasion == nil {
		return labelEverydayOcc
	}
	return string(*i.Occasion)
}

// StyleLabel renders the style for display, falling back to "versatile".
func StyleLabel(i *Intent) string {
	if i == nil || i.Style == nil {
		return labelVersatileStyle
	}
	return string(*i.Style)
}

// ReasonsFor produces short, human justifications for a single product, most
// significant first, driven by which signals actually contributed.
func ReasonsFor(sp *ScoredProduct, intent *Intent) []string {
	if sp == nil || sp.Product == nil {
		return nil
	}
	p := sp.Product
	reasons := make([]string, 0, 4)

	for _, signal := range sp.Breakdown.TopSignals(4) {
		s := sp.Breakdown[signal]
		if r := reasonForSignal(signal, s, p, intent); r != "" {
			reasons = append(reasons, r)
		}
	}

	// A product that wins on nothing in particular still deserves an honest
	// line rather than silence.
	if len(reasons) == 0 {
		reasons = append(reasons, fmt.Sprintf("A well-rated %s pick from %s",
			strings.ToLower(categoryName(p)), p.Brand))
	}
	return reasons
}

func reasonForSignal(signal string, s SignalScore, p *models.Product, intent *Intent) string {
	switch signal {
	case SignalOccasion:
		if intent.Occasion == nil {
			return ""
		}
		if p.Occasion == *intent.Occasion {
			return fmt.Sprintf("Made for %s wear", utils.LowerFirst(string(p.Occasion)))
		}
		if s.Raw >= 0.45 {
			// The honest version of a near-match: say what it is and why it
			// still works, rather than implying an exact hit.
			return fmt.Sprintf("A %s piece that carries over well to %s",
				utils.LowerFirst(string(p.Occasion)), utils.LowerFirst(string(*intent.Occasion)))
		}
		return ""

	case SignalBudget:
		if intent.Budget.IsZero() {
			return ""
		}
		b := intent.Budget

		// Direction is checked from the actual price against the actual
		// envelope first, and only then from the score's magnitude. Checking
		// magnitude alone is what caused the original bug this replaced: the
		// decay curve for a ~5% ceiling overshoot (see budget.go,
		// overshootDecay) still scores above 0.7, which would otherwise land
		// in the "Comfortably inside budget" branch below for a price that is
		// not, in fact, inside the budget.
		over := b.HasMax() && p.Price > *b.Max
		under := b.HasMin() && !b.HasMax() && p.Price < *b.Min

		switch {
		case over:
			if s.Raw >= 0.35 {
				return fmt.Sprintf("Slightly over budget at %s, but worth a look",
					utils.FormatINR(p.Price))
			}
			return ""
		case under:
			// The mirror case for a floor-only ask ("above ₹2000", "premium").
			if s.Raw >= 0.35 {
				return fmt.Sprintf("Just under your %s budget at %s, but worth a look",
					BudgetLabel(b), utils.FormatINR(p.Price))
			}
			return ""
		case s.Raw >= 0.95 && b.HasMax():
			return fmt.Sprintf("Uses your %s budget well at %s",
				BudgetLabel(b), utils.FormatINR(p.Price))
		case s.Raw >= 0.7:
			return fmt.Sprintf("Comfortably inside budget at %s", utils.FormatINR(p.Price))
		case b.HasMax() && p.Price <= *b.Max:
			// Genuinely within the stated envelope — the low raw score here
			// reflects the engine's sweet-spot curve (a price well under the
			// ceiling scores below a price spent near it, see budget.go), not
			// any actual shortcoming a shopper would perceive.
			return fmt.Sprintf("Well within budget at %s", utils.FormatINR(p.Price))
		default:
			return ""
		}

	case SignalStyle:
		if intent.Style == nil {
			return ""
		}
		if p.Style == *intent.Style {
			return fmt.Sprintf("Squarely %s in style", utils.LowerFirst(string(p.Style)))
		}
		if s.Raw >= 0.45 {
			return fmt.Sprintf("%s styling, close to the %s look you asked for",
				string(p.Style), utils.LowerFirst(string(*intent.Style)))
		}
		return ""

	case SignalCategory:
		if s.Raw < 0.6 {
			return ""
		}
		if intent.Category != nil && p.Category != nil && p.Category.Name == *intent.Category {
			return fmt.Sprintf("Exactly the %s you asked for", utils.LowerFirst(categoryName(p)))
		}
		if intent.Occasion != nil {
			return fmt.Sprintf("%s are a natural fit for %s", categoryName(p),
				utils.LowerFirst(string(*intent.Occasion)))
		}
		return ""

	case SignalGender:
		if intent.Gender == nil || s.Raw < 0.8 {
			return ""
		}
		if p.Gender == models.GenderUnisex {
			return "Cut as unisex, so it works either way"
		}
		return ""

	case SignalColor:
		if s.Raw < 0.7 {
			return ""
		}
		return fmt.Sprintf("In %s, matching your colour brief", p.Color)

	case SignalBrand:
		if s.Raw < 0.8 {
			return ""
		}
		return fmt.Sprintf("From %s, the brand you named", p.Brand)

	case SignalText:
		if s.Raw < 0.35 {
			return ""
		}
		return "Matches the specifics of what you described"

	case SignalQuality:
		if p.Rating >= 4.5 && p.RatingCount >= 100 {
			return fmt.Sprintf("Rated %.1f by %d shoppers", p.Rating, p.RatingCount)
		}
		if p.Rating >= 4.3 {
			return fmt.Sprintf("Strongly rated at %.1f", p.Rating)
		}
		return ""

	case SignalValue:
		d := p.EffectiveDiscount()
		if d >= 30 {
			return fmt.Sprintf("%d%% off right now", d)
		}
		return ""
	}
	return ""
}

func categoryName(p *models.Product) string {
	if p.Category == nil {
		return "product"
	}
	return string(p.Category.Name)
}

// BuildSummary writes the assistant's opening line.
//
// It mirrors the tone the frontend was designed against while telling the
// shopper what the engine actually understood — which doubles as a cheap
// correction affordance. If the engine misread "sangeet" as Office, the summary
// says "office" and the shopper immediately knows to rephrase, instead of
// silently concluding the recommendations are bad.
func BuildSummary(intent *Intent, products []ScoredProduct) string {
	prompt := utils.Truncate(strings.TrimSpace(intent.Query), 90)

	if len(products) == 0 {
		return fmt.Sprintf("I couldn't find a good match for %q in the catalogue right now. "+
			"Try widening the budget or naming a different occasion.", prompt)
	}

	understood := describeUnderstanding(intent)
	if understood == "" {
		return fmt.Sprintf("Here's a curated edit based on %q — picked for fit, value, "+
			"and how well each piece works for the occasion.", prompt)
	}

	return fmt.Sprintf("Here's a curated edit for %s — %d pieces picked for fit, value, "+
		"and how well each one works for the occasion.", understood, len(products))
}

// describeUnderstanding renders the intent as a natural phrase such as
// "a wedding in ethnic style, for women, under ₹3,000".
func describeUnderstanding(intent *Intent) string {
	parts := make([]string, 0, 4)

	if intent.Occasion != nil {
		parts = append(parts, utils.LowerFirst(string(*intent.Occasion)))
	}
	if intent.Category != nil {
		parts = append(parts, utils.LowerFirst(string(*intent.Category)))
	}
	if len(parts) == 0 {
		return ""
	}

	phrase := "a " + strings.Join(parts, " ")
	if intent.Style != nil {
		phrase += fmt.Sprintf(" in %s style", utils.LowerFirst(string(*intent.Style)))
	}
	if intent.Gender != nil && *intent.Gender != models.GenderUnisex {
		phrase += fmt.Sprintf(", for %s", utils.LowerFirst(string(*intent.Gender)))
	}
	if intent.Budget.HasMax() {
		phrase += fmt.Sprintf(", under %s", utils.FormatINR(*intent.Budget.Max))
	} else if intent.Budget.HasMin() {
		phrase += fmt.Sprintf(", above %s", utils.FormatINR(*intent.Budget.Min))
	}
	return phrase
}

// BuildReason writes the paragraph explaining the shortlist as a whole. It is
// assembled from the signals that actually drove this ranking, so it changes
// shape with the query rather than reciting a fixed template.
func BuildReason(intent *Intent, products []ScoredProduct, relaxed []string) string {
	if len(products) == 0 {
		return "No products cleared the bar for this brief. The catalogue may not " +
			"carry this combination of occasion, style and price yet."
	}

	var b strings.Builder

	criteria := make([]string, 0, 5)
	if intent.Occasion != nil {
		criteria = append(criteria, fmt.Sprintf("a %s occasion", utils.LowerFirst(string(*intent.Occasion))))
	}
	if intent.Style != nil {
		criteria = append(criteria, fmt.Sprintf("a %s aesthetic", utils.LowerFirst(string(*intent.Style))))
	}
	if intent.Gender != nil {
		criteria = append(criteria, fmt.Sprintf("%s fits", utils.LowerFirst(string(*intent.Gender))))
	}
	if len(intent.Colors) > 0 {
		criteria = append(criteria, fmt.Sprintf("%s tones", utils.JoinWithAnd(intent.Colors)))
	}
	if len(intent.Brands) > 0 {
		criteria = append(criteria, utils.JoinWithAnd(intent.Brands))
	}

	if len(criteria) > 0 {
		b.WriteString("These were shortlisted for ")
		b.WriteString(utils.JoinWithAnd(criteria))
	} else {
		b.WriteString("With no strong brief to work from, these were shortlisted on " +
			"rating strength and value")
	}

	if intent.Budget.HasMax() {
		inBudget := 0
		for _, p := range products {
			if InBudget(intent.Budget, p.Product.Price) {
				inBudget++
			}
		}
		b.WriteString(fmt.Sprintf(", ranked by AI match across fit, rating and price. "+
			"%d of %d land within your %s", inBudget, len(products), BudgetLabel(intent.Budget)))
	} else {
		b.WriteString(", ranked by AI match across fit, rating and price")
	}
	b.WriteString(".")

	// Honesty about compromises. If the engine had to loosen a constraint the
	// shopper set, saying so is worth more than pretending it did not.
	if len(relaxed) > 0 {
		b.WriteString(fmt.Sprintf(" To fill the edit, I widened %s.", utils.JoinWithAnd(relaxed)))
	}

	// Name the spread so the shortlist reads as a considered set rather than a
	// list of unrelated items.
	if cats := distinctCategories(products); len(cats) > 1 {
		b.WriteString(fmt.Sprintf(" The mix spans %s so you have real options rather than "+
			"six variations of one idea.", utils.JoinWithAnd(cats)))
	}

	return b.String()
}

func distinctCategories(products []ScoredProduct) []string {
	seen := make(map[string]struct{}, len(products))
	out := make([]string, 0, 4)
	for _, p := range products {
		if p.Product == nil || p.Product.Category == nil {
			continue
		}
		name := strings.ToLower(string(p.Product.Category.Name))
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
		if len(out) == 3 {
			break
		}
	}
	return out
}
