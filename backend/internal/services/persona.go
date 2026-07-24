package services

import (
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/assistant"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/personas"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
)

// applyPersona folds a stylist persona's Style DNA into the intent, right after
// intent extraction and personalization, and just before the engine runs.
//
// It is the single seam between "which celebrity is styling this" and "how the
// rule-based engine ranks". The persona contributes exactly two things:
//
//  1. Additive, capped preference boosts — the persona's preferred styles,
//     brands and colours are merged into intent.Personalization, the same
//     PersonalizationContext the shopper's own profile feeds. Because it is one
//     context with one MaxTotalBoost cap, a persona can nudge the ranking toward
//     its aesthetic but can never override occasion/style/category fit or the
//     hard filters. No new engine stage, no new cap.
//
//  2. Soft signal defaults — when the shopper named no style (or, for direct
//     chat, no gender), the persona's styleOverride / defaultGender fill the
//     gap. An explicit shopper signal always wins; the persona only speaks into
//     silence.
//
// The engine stays entirely rule-based: no persona and no LLM ever selects a
// product. Passing a nil persona is a no-op, so every non-StarCloset path is
// byte-identical to before.
func applyPersona(intent *recommendation.Intent, p *personas.Persona) {
	if intent == nil || p == nil {
		return
	}

	// --- Style DNA → personalization boosts ---------------------------------
	pc := intent.Personalization
	if pc == nil {
		pc = &recommendation.PersonalizationContext{}
	}
	pc.PreferredBrands = mergeStrings(pc.PreferredBrands, p.PreferredBrands)
	pc.PreferredColors = mergeStrings(pc.PreferredColors, p.PreferredColors)
	pc.PreferredStyles = mergeStyles(pc.PreferredStyles, p.PreferredStyles)
	intent.Personalization = pc

	// --- Soft style default --------------------------------------------------
	if intent.Style == nil {
		if st, ok := models.ParseStyle(p.StyleOverride); ok {
			s := st
			intent.Style = &s
		}
	}

	// --- Soft gender default (direct chat only; collections use department) --
	if intent.Gender == nil {
		if g, ok := models.ParseGender(p.DefaultGender); ok {
			gg := g
			intent.Gender = &gg
		}
	}
}

// mergeStrings unions two string slices, case-insensitively de-duplicated,
// preserving order (existing first, then new).
func mergeStrings(existing, add []string) []string {
	seen := make(map[string]struct{}, len(existing)+len(add))
	out := make([]string, 0, len(existing)+len(add))
	for _, s := range append(append([]string{}, existing...), add...) {
		t := strings.TrimSpace(s)
		if t == "" {
			continue
		}
		key := strings.ToLower(t)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	return out
}

// personaVoiceInput assembles the collection context the persona-voice module
// narrates. Occasion and budget labels reuse the engine's own formatters, and
// are left blank (so the voice falls back gracefully) when the shopper implied
// neither — a persona should never claim a budget the shopper never set.
func personaVoiceInput(prompt, dept string, intent *recommendation.Intent, count int, highlight string, pieces []string) assistant.VoiceInput {
	in := assistant.VoiceInput{
		Prompt:     prompt,
		Department: dept,
		ItemCount:  count,
		Highlight:  highlight,
		Pieces:     pieces,
	}
	if intent != nil {
		if intent.Occasion != nil {
			in.OccasionLabel = recommendation.OccasionLabel(intent)
		}
		if intent.Budget != nil && !intent.Budget.IsZero() {
			in.BudgetLabel = recommendation.BudgetLabel(intent.Budget)
		}
	}
	return in
}

// collectionsVoiceStats summarises a set of built collections for the voice: the
// total piece count, the headline product title, and short "Slot: Brand"
// descriptions drawn from the first non-empty collection.
func collectionsVoiceStats(cs []*recommendation.Collection) (count int, highlight string, pieces []string) {
	for _, c := range cs {
		if c == nil {
			continue
		}
		count += len(c.Items)
	}
	for _, c := range cs {
		if c == nil || len(c.Items) == 0 {
			continue
		}
		for _, it := range c.Items {
			if it.Product.Product == nil {
				continue
			}
			if highlight == "" {
				highlight = it.Product.Product.Title
			}
			pieces = append(pieces, it.Slot+": "+it.Product.Product.Brand)
		}
		break
	}
	return count, highlight, pieces
}

// productsVoiceStats summarises a flat ranked product list for the voice.
func productsVoiceStats(ps []recommendation.ScoredProduct) (count int, highlight string) {
	count = len(ps)
	if count > 0 && ps[0].Product != nil {
		highlight = ps[0].Product.Title
	}
	return count, highlight
}

// mergeStyles parses the persona's free-text styles against the canonical
// vocabulary and unions them into the existing style list. Unparseable entries
// are dropped rather than guessed.
func mergeStyles(existing []models.StyleTag, add []string) []models.StyleTag {
	seen := make(map[models.StyleTag]struct{}, len(existing))
	out := make([]models.StyleTag, 0, len(existing)+len(add))
	for _, s := range existing {
		if _, dup := seen[s]; !dup {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	for _, s := range add {
		st, ok := models.ParseStyle(s)
		if !ok {
			continue
		}
		if _, dup := seen[st]; dup {
			continue
		}
		seen[st] = struct{}{}
		out = append(out, st)
	}
	return out
}
