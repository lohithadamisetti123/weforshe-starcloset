package recommendation

import "math"

// Calibration maps a raw relevance score onto the 60–99 "AI match" percentage
// the UI renders.
//
// This is presentation, not ranking — it never reorders anything — but it is
// not cosmetic either. The number is a claim made to a shopper, so it has to be
// honest in two ways the raw score is not:
//
//  1. Raw scores are compressed. Even a perfect answer rarely exceeds ~0.85,
//     because quality and value are never saturated. Showing "83% match" for a
//     flawless pick reads as a hedge.
//  2. Raw scores ignore how much was asked. Scoring 0.8 against six stated
//     constraints is strong evidence of a good match; scoring 0.8 against
//     "show me something" means the engine mostly liked the star rating. The
//     first deserves a higher confidence claim than the second.
//
// So the raw score is rescaled across its realistic operating band, then capped
// by how much the shopper actually specified. The floor of 60 matches the
// frontend's existing expectation and reflects that a curated catalogue has no
// truly terrible picks.

const (
	// matchFloor and matchCeil bound the displayed percentage.
	matchFloor = 60.0
	matchCeil  = 99.0
	// scoreFloor and scoreCeil are the realistic operating band of the raw
	// score. Anything below the floor is a weak match, anything above the
	// ceiling is as good as the engine gets.
	scoreFloor = 0.22
	scoreCeil  = 0.86
	// gamma < 1 expands the top of the range, so the difference between a good
	// and an excellent match is visible rather than crushed into two points.
	gamma = 0.80
)

// confidenceCap is the highest match percentage claimable for a query with n
// stated constraints. Beyond four constraints there is enough signal to make
// the strongest claim the engine offers.
func confidenceCap(specified int) float64 {
	switch specified {
	case 0:
		return 88
	case 1:
		return 92
	case 2:
		return 95
	case 3:
		return 97
	default:
		return matchCeil
	}
}

// MatchPercent converts a raw score in [0,1] into the displayed 60–99 figure,
// capped by how much the intent specified.
func MatchPercent(score float64, intent *Intent) int {
	norm := clamp01((score - scoreFloor) / (scoreCeil - scoreFloor))
	expanded := math.Pow(norm, gamma)

	pct := matchFloor + (matchCeil-matchFloor)*expanded

	if cap := confidenceCap(intent.SpecifiedCount()); pct > cap {
		pct = cap
	}
	if pct < matchFloor {
		pct = matchFloor
	}
	return int(math.Round(pct))
}

// ApplyCalibration fills MatchPercent on every scored product in place.
func ApplyCalibration(products []ScoredProduct, intent *Intent) {
	for i := range products {
		products[i].MatchPercent = MatchPercent(products[i].Score, intent)
	}
	enforceMonotonicity(products)
}

// enforceMonotonicity guarantees that a lower-ranked product never displays a
// higher match percentage than the one above it.
//
// The diversity pass can legitimately promote a slightly less relevant product
// into slot 2, which would otherwise surface as a shortlist reading
// 94, 91, 92, 88 — and a shopper who notices that has caught the product
// contradicting itself. Ranking stays as MMR chose; only the displayed number
// is clamped down to stay consistent with the order shown.
func enforceMonotonicity(products []ScoredProduct) {
	for i := 1; i < len(products); i++ {
		if products[i].MatchPercent > products[i-1].MatchPercent {
			products[i].MatchPercent = products[i-1].MatchPercent
		}
	}
}
