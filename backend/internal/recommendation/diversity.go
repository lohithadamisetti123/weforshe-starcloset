package recommendation

import "github.com/myntra/bharat-ai-backend/internal/models"

// Diversification via Maximal Marginal Relevance.
//
// This stage earns its place the first time you watch a pure relevance ranker
// answer "wedding outfit under ₹3000". The catalogue's best wedding pick is a
// Banarasi silk saree; the second, third and fourth best are also Banarasi silk
// sarees, because the features that made the first one win are shared. The
// shopper gets six near-identical products, one decision, and no sense that
// anyone curated anything.
//
// MMR fixes this by selecting greedily against relevance *minus* similarity to
// what has already been picked:
//
//	next = argmax_i [ λ·rel(i) − (1−λ)·max_{j ∈ selected} sim(i, j) ]
//
// The top pick is always the most relevant product — the shopper's headline
// answer is never sacrificed. Diversity pressure applies only from the second
// slot onward, which is exactly where a human stylist starts offering
// alternatives rather than repetitions.

// similarityWeights are the contributions to the pairwise similarity used by
// MMR. Category dominates because two kurta sets are substitutes for one
// another in a way that two maroon items simply are not.
const (
	simCategoryWeight = 0.55
	simBrandWeight    = 0.25
	simStyleWeight    = 0.12
	simColorWeight    = 0.08
	// simPriceBandWeight rewards spread across price points: showing a ₹600 and
	// a ₹2,900 option is more useful than two items eleven rupees apart.
	simPriceBandWeight = 0.10
)

// ProductSimilarity returns how substitutable two products are, in [0,1].
func ProductSimilarity(a, b *models.Product) float64 {
	if a == nil || b == nil {
		return 0
	}
	if a.ID == b.ID {
		return 1
	}

	var sim float64
	if a.CategoryID == b.CategoryID {
		sim += simCategoryWeight
	}
	if normalizeBrand(a.Brand) == normalizeBrand(b.Brand) {
		sim += simBrandWeight
	}
	if a.Style == b.Style {
		sim += simStyleWeight
	}
	if a.Color == b.Color {
		sim += simColorWeight
	}
	if samePriceBand(a.Price, b.Price) {
		sim += simPriceBandWeight
	}
	return clamp01(sim)
}

// samePriceBand reports whether two prices sit in the same coarse tier. The
// bands are the ones shoppers actually reason in, not equal-width buckets.
func samePriceBand(a, b int) bool {
	return priceBand(a) == priceBand(b)
}

func priceBand(p int) int {
	switch {
	case p < 1000:
		return 0
	case p < 2000:
		return 1
	case p < 3500:
		return 2
	case p < 6000:
		return 3
	default:
		return 4
	}
}

// Diversify selects topN products from scored using MMR with the given lambda.
//
// lambda == 1 is pure relevance (no diversification); lambda == 0 is pure
// novelty. The configured default of 0.72 keeps relevance firmly in charge
// while still breaking up runs of near-duplicates.
//
// scored is expected to be sorted by Score descending. The returned slice has
// Rank and FinalScore populated.
func Diversify(scored []ScoredProduct, topN int, lambda float64) []ScoredProduct {
	if topN <= 0 || len(scored) == 0 {
		return nil
	}
	if topN > len(scored) {
		topN = len(scored)
	}
	if lambda >= 1 {
		out := make([]ScoredProduct, topN)
		copy(out, scored[:topN])
		for i := range out {
			out[i].Rank = i + 1
			out[i].FinalScore = out[i].Score
		}
		return out
	}

	selected := make([]ScoredProduct, 0, topN)
	taken := make([]bool, len(scored))

	// Slot 1 is always the most relevant product, unmodified.
	selected = append(selected, scored[0])
	taken[0] = true
	selected[0].Rank = 1
	selected[0].FinalScore = scored[0].Score

	for len(selected) < topN {
		bestIdx := -1
		bestMMR := -1.0

		for i, cand := range scored {
			if taken[i] {
				continue
			}
			var maxSim float64
			for _, sel := range selected {
				if s := ProductSimilarity(cand.Product, sel.Product); s > maxSim {
					maxSim = s
				}
			}
			mmr := lambda*cand.Score - (1-lambda)*maxSim
			if mmr > bestMMR {
				bestMMR = mmr
				bestIdx = i
			}
		}

		if bestIdx < 0 {
			break
		}
		pick := scored[bestIdx]
		pick.Rank = len(selected) + 1
		// FinalScore records the post-diversification value that drove the
		// ordering, while Score keeps the untouched relevance. Both are
		// persisted so a surprising position can be traced to either the
		// ranker or the diversity pass.
		pick.FinalScore = bestMMR
		selected = append(selected, pick)
		taken[bestIdx] = true
	}

	return selected
}
