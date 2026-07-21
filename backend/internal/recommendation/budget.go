package recommendation

import "math"

// Budget is the signal shoppers feel most sharply. A ₹3,000 brief answered with
// a ₹9,000 saree is not a recommendation, it is an advertisement.
//
// The curve below encodes three claims that a naive `price <= budget` filter
// gets wrong:
//
//  1. Spending the budget is the goal, not avoiding it. Someone who says
//     "under ₹3000" usually wants the best thing near ₹3,000, not the cheapest
//     thing in the catalogue. The peak therefore sits in the 70–100% band, not
//     at zero.
//  2. A conspicuously cheap item is a weak answer, not a great one. It scores
//     acceptably — it is still in budget — but it should not outrank a
//     well-matched item that uses the budget.
//  3. Going over budget is a soft failure that decays fast, not a hard one.
//     ₹3,150 against a ₹3,000 brief is worth showing; ₹6,000 is not.

const (
	// budgetSweetSpotLow is where the "using your budget well" plateau begins,
	// as a fraction of the ceiling.
	budgetSweetSpotLow = 0.70
	// overshootDecay controls how fast score falls once price exceeds the
	// ceiling. At 6.0 a 10% overshoot keeps ~55% of the signal and a 40%
	// overshoot keeps ~9%.
	overshootDecay = 6.0
	// cheapFloor is the score a nearly-free item gets against a stated budget.
	cheapFloor = 0.55
	// hardCeilingMultiplier bounds what the retrieval layer will even consider
	// against a stated ceiling.
	hardCeilingMultiplier = 1.6
)

// BudgetFit scores a price in [0,1] against the shopper's stated envelope.
// An absent or empty envelope returns a neutral 0.5 and the caller will have
// deactivated the signal anyway, so it never distorts the total.
func BudgetFit(b *BudgetRange, price int) float64 {
	if b.IsZero() || price <= 0 {
		return 0.5
	}

	switch {
	case b.HasMax() && b.HasMin():
		return rangeFit(*b.Min, *b.Max, price)
	case b.HasMax():
		return ceilingFit(*b.Max, price)
	default:
		return floorFit(*b.Min, price)
	}
}

// ceilingFit scores against an upper bound ("under ₹3000").
func ceilingFit(maxPrice, price int) float64 {
	if maxPrice <= 0 {
		return 0.5
	}
	u := float64(price) / float64(maxPrice)

	switch {
	case u > 1.0:
		// Smooth exponential decay past the ceiling. Never returns exactly 0,
		// so a slight overshoot can still win on an otherwise perfect match.
		return math.Exp(-overshootDecay * (u - 1.0))
	case u >= budgetSweetSpotLow:
		// Full marks across the top 30% of the budget.
		return 1.0
	default:
		// Linear ramp from cheapFloor at ₹0 up to 1.0 at the sweet spot.
		return cheapFloor + (1.0-cheapFloor)*(u/budgetSweetSpotLow)
	}
}

// floorFit scores against a lower bound ("above ₹2000", "premium").
func floorFit(minPrice, price int) float64 {
	if minPrice <= 0 {
		return 0.5
	}
	u := float64(price) / float64(minPrice)
	if u >= 1.0 {
		// Comfortably above the floor is a full match; there is no upper bound
		// to punish against.
		return 1.0
	}
	// Undershooting a floor is a real miss, but a near-miss is still shown.
	return math.Max(0, 1.0-2.5*(1.0-u))
}

// rangeFit scores against a two-sided envelope ("between ₹2000 and ₹5000").
func rangeFit(minPrice, maxPrice, price int) float64 {
	if minPrice > maxPrice {
		minPrice, maxPrice = maxPrice, minPrice
	}
	if maxPrice <= 0 {
		return 0.5
	}
	switch {
	case price >= minPrice && price <= maxPrice:
		return 1.0
	case price > maxPrice:
		return math.Exp(-overshootDecay * (float64(price)/float64(maxPrice) - 1.0))
	default:
		if minPrice <= 0 {
			return 1.0
		}
		return math.Max(0, 1.0-2.5*(1.0-float64(price)/float64(minPrice)))
	}
}

// WithinHardCeiling reports whether a price is close enough to a stated ceiling
// to be worth showing at all.
//
// This is the one place the engine is willing to discard a product outright on
// price. It is applied conditionally by the retrieval layer — only when doing
// so still leaves a full shortlist — because returning six mildly over-budget
// options beats returning two perfect ones and four blanks.
func WithinHardCeiling(b *BudgetRange, price int) bool {
	if !b.HasMax() {
		return true
	}
	return float64(price) <= float64(*b.Max)*hardCeilingMultiplier
}

// InBudget reports strict compliance with the stated envelope, used for the
// "N of these are within budget" response metadata rather than for ranking.
func InBudget(b *BudgetRange, price int) bool {
	if b.IsZero() {
		return true
	}
	if b.HasMax() && price > *b.Max {
		return false
	}
	if b.HasMin() && price < *b.Min {
		return false
	}
	return true
}
