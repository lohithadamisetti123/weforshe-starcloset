package recommendation

import "math"

// Quality and value are the two signals that never switch off. Even a prompt
// with no extractable intent at all ("show me something nice") must produce a
// defensible shortlist, and when the shopper has told us nothing these carry
// the whole ranking.

const (
	// ratingPriorWeight (m) is the number of pseudo-ratings pulled toward the
	// catalogue mean. It is the strength of the prior: a product with 40
	// ratings is shrunk noticeably, one with 1,200 barely at all.
	ratingPriorWeight = 60.0
	// ratingFloor and ratingCeil bound the 1–5 star scale for normalisation.
	ratingFloor = 1.0
	ratingCeil  = 5.0
	// discountSaturation is the discount percentage at which the value signal
	// saturates. Beyond it, more discount says more about the markup than about
	// the deal.
	discountSaturation = 55.0
)

// DefaultMeanRating is the fallback C term when no catalogue statistics are
// available — an empty pool, or a repository that could not compute the mean.
// A curated fashion catalogue clusters around four stars, so this is a
// defensible prior rather than a magic number, and it keeps the Bayesian
// average well-defined instead of collapsing every quality score to zero.
const DefaultMeanRating = 4.0

// QualityPrior converts a raw rating into a trustworthy 0..1 score using a
// Bayesian average.
//
// The naive alternative — ranking on the raw star rating — reliably promotes a
// 5.0-from-3-ratings novelty above a 4.6-from-2,000-ratings staple. That is the
// wrong answer for a shopper, so ratings are shrunk toward the catalogue mean
// in proportion to how little evidence backs them:
//
//	adjusted = (v/(v+m))·R + (m/(v+m))·C
//
// where R is the product's rating, v its rating count, C the catalogue mean and
// m the prior weight. This is the same estimator behind IMDb's Top 250, and it
// is the right shape for the problem: confidence should scale with evidence.
func QualityPrior(rating float64, ratingCount int, catalogueMean float64) float64 {
	if catalogueMean <= 0 {
		catalogueMean = 4.0
	}
	v := float64(ratingCount)
	if v < 0 {
		v = 0
	}
	adjusted := (v/(v+ratingPriorWeight))*rating +
		(ratingPriorWeight/(v+ratingPriorWeight))*catalogueMean

	return clamp01((adjusted - ratingFloor) / (ratingCeil - ratingFloor))
}

// ValueSignal scores the deal on offer in [0,1]. It saturates so that a 70%-off
// clearance item cannot buy its way to the top of a well-matched shortlist —
// discount is a tiebreaker between comparable products, not a substitute for
// relevance.
func ValueSignal(discountPercent int) float64 {
	if discountPercent <= 0 {
		return 0
	}
	d := float64(discountPercent)
	if d > discountSaturation {
		d = discountSaturation
	}
	// sqrt shape: early discount is felt more than the last few points.
	return clamp01(math.Sqrt(d / discountSaturation))
}

// CatalogueMeanRating computes the mean rating across a candidate pool, the C
// term of the Bayesian average.
func CatalogueMeanRating(ratings []float64) float64 {
	if len(ratings) == 0 {
		return DefaultMeanRating
	}
	var sum float64
	for _, r := range ratings {
		sum += r
	}
	return sum / float64(len(ratings))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
