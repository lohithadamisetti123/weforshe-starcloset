package recommendation

import (
	"github.com/myntra/bharat-ai-backend/internal/config"
	"github.com/myntra/bharat-ai-backend/internal/models"
)

// Scorer turns (Intent, Product) into an explainable relevance score.
//
// The central design decision here is query-adaptive weighting.
//
// A fixed weight vector is what most rule-based rankers do, and it is subtly
// broken. If colour carries 5% of the score and the shopper never mentioned a
// colour, then under fixed weights every product forfeits that 5% — the signal
// contributes nothing but still dilutes the total, and worse, the totals are no
// longer comparable between a richly-specified query and a vague one.
//
// Instead, a signal is *active* only when the query gives the engine something
// to judge it against. Inactive weight is redistributed across the signals that
// are active. The consequences fall out naturally and are exactly what you want:
//
//   - "wedding saree under ₹3000, red" → occasion, budget, category and colour
//     carry nearly the whole score. Quality is a tiebreaker.
//   - "show me something nice" → nothing semantic is active, so quality and
//     value absorb the entire weight and the engine returns the genuine best of
//     the catalogue instead of an arbitrary six.
//
// The score therefore always means the same thing — "how well does this answer
// the brief that was actually given" — regardless of how much brief there was.
type Scorer struct {
	weights config.Weights
}

// NewScorer returns a Scorer using the given base weights.
func NewScorer(w config.Weights) *Scorer {
	return &Scorer{weights: w.Normalized()}
}

// scoringContext carries the per-request state shared across every product so
// it is computed once rather than per candidate.
type scoringContext struct {
	intent        *Intent
	textIndex     *TextIndex
	queryTerms    []string
	catalogueMean float64
	weights       config.Weights
	active        map[string]bool
}

// newScoringContext resolves which signals are active for this query and
// redistributes the inactive weight.
func (s *Scorer) newScoringContext(intent *Intent, pool []*models.Product) *scoringContext {
	ratings := make([]float64, 0, len(pool))
	for _, p := range pool {
		if p != nil {
			ratings = append(ratings, p.Rating)
		}
	}

	active := map[string]bool{
		SignalOccasion: intent.Occasion != nil,
		SignalStyle:    intent.Style != nil,
		SignalGender:   intent.Gender != nil,
		// Category is judged whenever the shopper named one *or* named an
		// occasion — an occasion implies a category prior, which is a real
		// judgement about fit and deserves to be scored.
		SignalCategory: intent.Category != nil || intent.Occasion != nil,
		SignalBudget:   !intent.Budget.IsZero(),
		SignalColor:    len(intent.Colors) > 0,
		SignalBrand:    len(intent.Brands) > 0,
		SignalText:     len(intent.Keywords) > 0,
		// Quality and value are always in play: they are properties of the
		// product, not answers to a question the shopper asked.
		SignalQuality: true,
		SignalValue:   true,
	}

	return &scoringContext{
		intent:        intent,
		textIndex:     NewTextIndex(pool),
		queryTerms:    intent.Keywords,
		catalogueMean: CatalogueMeanRating(ratings),
		weights:       redistribute(s.weights, active),
		active:        active,
	}
}

// redistribute scales the weights of active signals so they sum to 1.0,
// absorbing whatever the inactive signals would have carried.
//
// The proportions among active signals are preserved, so operator tuning of the
// base weights keeps its meaning: doubling REC_W_BUDGET still doubles budget's
// influence relative to occasion, whichever signals happen to fire.
func redistribute(base config.Weights, active map[string]bool) config.Weights {
	raw := map[string]float64{
		SignalOccasion: base.Occasion,
		SignalStyle:    base.Style,
		SignalGender:   base.Gender,
		SignalCategory: base.Category,
		SignalBudget:   base.Budget,
		SignalColor:    base.Color,
		SignalBrand:    base.Brand,
		SignalText:     base.Text,
		SignalQuality:  base.Quality,
		SignalValue:    base.Value,
	}

	var activeSum float64
	for name, w := range raw {
		if active[name] {
			activeSum += w
		}
	}
	if activeSum <= 0 {
		// Degenerate configuration: every active signal has zero weight. Fall
		// back to quality-only rather than dividing by zero or returning an
		// all-zero ranking.
		return config.Weights{Quality: 1.0}
	}

	scale := 1.0 / activeSum
	out := config.Weights{}
	if active[SignalOccasion] {
		out.Occasion = raw[SignalOccasion] * scale
	}
	if active[SignalStyle] {
		out.Style = raw[SignalStyle] * scale
	}
	if active[SignalGender] {
		out.Gender = raw[SignalGender] * scale
	}
	if active[SignalCategory] {
		out.Category = raw[SignalCategory] * scale
	}
	if active[SignalBudget] {
		out.Budget = raw[SignalBudget] * scale
	}
	if active[SignalColor] {
		out.Color = raw[SignalColor] * scale
	}
	if active[SignalBrand] {
		out.Brand = raw[SignalBrand] * scale
	}
	if active[SignalText] {
		out.Text = raw[SignalText] * scale
	}
	if active[SignalQuality] {
		out.Quality = raw[SignalQuality] * scale
	}
	if active[SignalValue] {
		out.Value = raw[SignalValue] * scale
	}
	return out
}

// ScoreAll scores every product in the pool against the intent.
func (s *Scorer) ScoreAll(intent *Intent, pool []*models.Product) []ScoredProduct {
	ctx := s.newScoringContext(intent, pool)

	out := make([]ScoredProduct, 0, len(pool))
	for _, p := range pool {
		if p == nil {
			continue
		}
		out = append(out, s.scoreOne(ctx, p))
	}
	return out
}

// scoreOne evaluates a single product against the prepared context.
func (s *Scorer) scoreOne(ctx *scoringContext, p *models.Product) ScoredProduct {
	in := ctx.intent
	w := ctx.weights
	bd := make(Breakdown, len(SignalOrder))

	set := func(name string, raw, weight float64) {
		a := ctx.active[name]
		if !a {
			bd[name] = SignalScore{Raw: raw, Weight: 0, Contribution: 0, Active: false}
			return
		}
		bd[name] = SignalScore{
			Raw:          raw,
			Weight:       weight,
			Contribution: raw * weight,
			Active:       true,
		}
	}

	// --- Occasion -----------------------------------------------------------
	var occasionRaw float64
	if in.Occasion != nil {
		occasionRaw = OccasionAffinity(*in.Occasion, p.Occasion)
	}
	set(SignalOccasion, occasionRaw, w.Occasion)

	// --- Style --------------------------------------------------------------
	var styleRaw float64
	if in.Style != nil {
		styleRaw = StyleAffinity(*in.Style, p.Style)
	}
	set(SignalStyle, styleRaw, w.Style)

	// --- Gender -------------------------------------------------------------
	var genderRaw float64
	if in.Gender != nil {
		genderRaw = GenderAffinity(*in.Gender, p.Gender)
	}
	set(SignalGender, genderRaw, w.Gender)

	// --- Category -----------------------------------------------------------
	// An explicit category ask is authoritative; otherwise fall back to the
	// occasion's category prior.
	var categoryRaw float64
	switch {
	case in.Category != nil && p.Category != nil:
		if p.Category.Name == *in.Category {
			categoryRaw = 1.0
		} else if in.Occasion != nil {
			// The shopper named a category and we do not have it — grade the
			// alternative on how well it suits the occasion instead of
			// zeroing it, so a near-miss still beats an irrelevance.
			categoryRaw = 0.35 * CategoryPrior(*in.Occasion, p.Category.Name)
		}
	case in.Occasion != nil && p.Category != nil:
		categoryRaw = CategoryPrior(*in.Occasion, p.Category.Name)
	}
	set(SignalCategory, categoryRaw, w.Category)

	// --- Budget -------------------------------------------------------------
	budgetRaw := BudgetFit(in.Budget, p.Price)
	set(SignalBudget, budgetRaw, w.Budget)

	// --- Colour -------------------------------------------------------------
	colorRaw := BestColorMatch(in.Colors, p.Color)
	set(SignalColor, colorRaw, w.Color)

	// --- Brand --------------------------------------------------------------
	brandRaw := BestBrandMatch(in.Brands, p.Brand)
	set(SignalBrand, brandRaw, w.Brand)

	// --- Free text ----------------------------------------------------------
	textRaw := ctx.textIndex.Score(p.ID.String(), ctx.queryTerms)
	set(SignalText, textRaw, w.Text)

	// --- Quality ------------------------------------------------------------
	qualityRaw := QualityPrior(p.Rating, p.RatingCount, ctx.catalogueMean)
	set(SignalQuality, qualityRaw, w.Quality)

	// --- Value --------------------------------------------------------------
	valueRaw := ValueSignal(p.EffectiveDiscount())
	set(SignalValue, valueRaw, w.Value)

	total := bd.Total()

	// Damp the whole score by extractor confidence when the reading was shaky,
	// but never below a floor: a hesitant Intent should soften the engine's
	// certainty, not erase a genuinely good product's score.
	if in.Confidence > 0 && in.Confidence < 1 && in.HasSemanticSignal() {
		damp := 0.85 + 0.15*in.Confidence
		total *= damp
	}

	return ScoredProduct{
		Product:   p,
		Score:     clamp01(total),
		Breakdown: bd,
	}
}
