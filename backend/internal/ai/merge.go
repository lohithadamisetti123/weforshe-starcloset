package ai

import (
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
)

// Merging the two extractors is where the assistant's understanding is actually
// decided, so it is worth being explicit about the principle.
//
// The naive options are both bad:
//
//   - "Gemini wins, rules are the fallback" throws away the one component that
//     is provably correct. LLMs are genuinely unreliable at arithmetic and
//     verbatim extraction: they round ₹2,499 to ₹2,500, read "under 3000" as a
//     range of 2000–3000, and occasionally drop a digit. The regex does not.
//   - "Rules win, Gemini is a bonus" throws away the only component that can
//     read "something for the office party that isn't too stuffy".
//
// So arbitration is per field, decided by which extractor has the stronger
// claim on that particular kind of information:
//
//	Budget     -> rules win outright when they found a number in the text.
//	             A number present in the prompt is a fact, not an inference.
//	Occasion,  -> Gemini wins when confident, because these need world
//	Style,        knowledge ("interview" implies Office and Formal; "beach
//	Gender,       wedding in Goa" is Wedding, not Travel) that a keyword table
//	Category      cannot represent. The rules win when Gemini abstained.
//	Exclusions -> union of both. A missed exclusion actively contradicts the
//	             shopper, so a false positive here costs less than a miss.
//
// The result is strictly better than either alone, and it degrades cleanly: no
// key, no quota, or an outage leaves the rules running at full strength.

// Merge combines a rule-derived intent with a Gemini-derived one.
//
// rules must be non-nil. llm may be nil, which is the normal offline path and
// simply returns the rules unchanged.
func Merge(rules, llm *recommendation.Intent) *recommendation.Intent {
	if rules == nil && llm == nil {
		return recommendation.NewIntent("")
	}
	if llm == nil {
		return rules
	}
	if rules == nil {
		return llm
	}

	out := recommendation.NewIntent(rules.Query)
	out.Source = recommendation.SourceHybrid
	out.Keywords = rules.Keywords
	// Product-type terms come from the deterministic extractor only.
	out.TypeTerms = rules.TypeTerms

	// --- Budget: rules are authoritative when they parsed a number ----------
	switch {
	case rules.Budget != nil && !rules.Budget.IsZero():
		out.Budget = rules.Budget
		if llm.Budget != nil && !llm.Budget.IsZero() && !sameBudget(rules.Budget, llm.Budget) {
			// Worth recording: a persistent disagreement here is the signal
			// that either a regex or the prompt needs work.
			out.Notes = append(out.Notes, "budget: rules and gemini disagreed; used rules")
		}
	case llm.Budget != nil && !llm.Budget.IsZero():
		// The rules found nothing, so this came from implication rather than a
		// literal number ("something affordable"). Take it, but note it.
		out.Budget = llm.Budget
		out.Notes = append(out.Notes, "budget: inferred by gemini")
	}

	// --- Semantic fields: Gemini leads, rules fill the gaps ------------------
	confident := llm.Confidence >= 0.5

	out.Occasion = rules.Occasion
	if llm.Occasion != nil && (confident || rules.Occasion == nil) {
		out.Occasion = llm.Occasion
	}

	out.Style = rules.Style
	if llm.Style != nil && (confident || rules.Style == nil) {
		out.Style = llm.Style
	}

	out.Gender = rules.Gender
	if llm.Gender != nil && (confident || rules.Gender == nil) {
		out.Gender = llm.Gender
	}

	out.Category = rules.Category
	if llm.Category != nil && (confident || rules.Category == nil) {
		out.Category = llm.Category
	}

	// --- Colours and brands: prefer the richer answer ------------------------
	// Both extractors read these from the surface text, so neither has a
	// principled edge. The longer list is taken on the grounds that a missed
	// colour is a real loss while a spurious one costs at most 5% of the score.
	out.Colors = longer(rules.Colors, llm.Colors)
	out.Brands = longer(rules.Brands, llm.Brands)

	// --- Exclusions: union ---------------------------------------------------
	out.ExcludeColors = unionStrings(rules.ExcludeColors, llm.ExcludeColors)
	out.ExcludeBrands = unionStrings(rules.ExcludeBrands, llm.ExcludeBrands)
	out.ExcludeStyles = unionStyles(rules.ExcludeStyles, llm.ExcludeStyles)

	// A field cannot be both wanted and excluded; the exclusion is the more
	// deliberate statement and wins.
	out.Colors = removeExcluded(out.Colors, out.ExcludeColors)
	out.Brands = removeExcluded(out.Brands, out.ExcludeBrands)
	if out.Style != nil {
		for _, ex := range out.ExcludeStyles {
			if ex == *out.Style {
				out.Style = nil
				break
			}
		}
	}

	// --- Sort ----------------------------------------------------------------
	out.Sort = recommendation.SortRelevance
	switch {
	case rules.Sort != "" && rules.Sort != recommendation.SortRelevance:
		// "cheapest" is a literal keyword; the rules see it perfectly well.
		out.Sort = rules.Sort
	case llm.Sort != "" && llm.Sort != recommendation.SortRelevance:
		out.Sort = llm.Sort
	}

	// --- Confidence ----------------------------------------------------------
	out.Confidence = combinedConfidence(rules, llm, out)

	out.Notes = append(out.Notes, rules.Notes...)
	out.Notes = append(out.Notes, llm.Notes...)

	return out
}

// combinedConfidence blends the two self-assessments and then adjusts for
// whether they actually agreed.
//
// Agreement between two independent extractors is real evidence and should
// raise confidence above either one alone; disagreement means at least one is
// wrong, and the honest response is to lower confidence rather than to pick a
// winner and pretend. The scorer damps its output by this value, so getting it
// right directly controls how hard the engine leans on a shaky reading.
func combinedConfidence(rules, llm, merged *recommendation.Intent) float64 {
	// Gemini carries more weight because it is the better extractor when it
	// works; the rules anchor it because they cannot hallucinate.
	base := 0.6*llm.Confidence + 0.4*rules.Confidence

	agreements, comparisons := 0, 0
	compare := func(a, b bool, equal bool) {
		if a && b {
			comparisons++
			if equal {
				agreements++
			}
		}
	}

	compare(rules.Occasion != nil, llm.Occasion != nil,
		rules.Occasion != nil && llm.Occasion != nil && *rules.Occasion == *llm.Occasion)
	compare(rules.Style != nil, llm.Style != nil,
		rules.Style != nil && llm.Style != nil && *rules.Style == *llm.Style)
	compare(rules.Gender != nil, llm.Gender != nil,
		rules.Gender != nil && llm.Gender != nil && *rules.Gender == *llm.Gender)
	compare(rules.Category != nil, llm.Category != nil,
		rules.Category != nil && llm.Category != nil && *rules.Category == *llm.Category)

	if comparisons > 0 {
		ratio := float64(agreements) / float64(comparisons)
		// ±0.12 at the extremes: enough to matter, not enough to swamp the
		// underlying self-assessments.
		base += 0.12 * (2*ratio - 1)
	}

	// An intent with nothing in it cannot be confidently anything.
	if !merged.HasSemanticSignal() {
		base = minFloat(base, 0.4)
	}

	if base < 0.15 {
		base = 0.15
	}
	if base > 0.97 {
		base = 0.97
	}
	return base
}

func sameBudget(a, b *recommendation.BudgetRange) bool {
	return a.Floor() == b.Floor() && a.Ceiling() == b.Ceiling()
}

func longer(a, b []string) []string {
	if len(b) > len(a) {
		return b
	}
	if len(a) == 0 {
		return nil
	}
	return a
}

func unionStrings(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, list := range [][]string{a, b} {
		for _, s := range list {
			key := normalizeToken(s)
			if key == "" {
				continue
			}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func unionStyles(a, b []models.StyleTag) []models.StyleTag {
	seen := make(map[models.StyleTag]struct{}, len(a)+len(b))
	out := make([]models.StyleTag, 0, len(a)+len(b))
	for _, list := range [][]models.StyleTag{a, b} {
		for _, s := range list {
			if _, dup := seen[s]; dup {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func removeExcluded(wanted, excluded []string) []string {
	if len(wanted) == 0 || len(excluded) == 0 {
		return wanted
	}
	block := make(map[string]struct{}, len(excluded))
	for _, e := range excluded {
		block[normalizeToken(e)] = struct{}{}
	}
	out := make([]string, 0, len(wanted))
	for _, w := range wanted {
		if _, bad := block[normalizeToken(w)]; bad {
			continue
		}
		out = append(out, w)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeToken(s string) string {
	var b []rune
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b = append(b, r+32)
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b = append(b, r)
		}
	}
	return string(b)
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
