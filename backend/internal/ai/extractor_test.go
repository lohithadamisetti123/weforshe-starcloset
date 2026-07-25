package ai

import (
	"testing"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
)

// testVocab mirrors a slice of the real seeded catalogue's brand and color
// vocabulary — enough for the matcher tests below without a database.
func testVocab() Vocabulary {
	return Vocabulary{
		Brands: []string{"H&M", "Levi's", "Nike", "Puma", "W", "AND", "Biba", "Zara", "FabIndia"},
		Colors: []string{"Maroon", "Bottle Green", "Off White", "Mustard", "Rust"},
	}
}

// --- Hinglish occasion lexicon -------------------------------------------
//
// This is the feature the product's name ("Bharat AI") is actually about: a
// shopper typing in the way they'd type to a friend, not in the site's own
// vocabulary, must still be understood.

func TestRuleExtractor_HinglishOccasions(t *testing.T) {
	cases := []struct {
		prompt string
		want   models.Occasion
	}{
		// "shaadi" has no ceremony-specific occasion of its own, so it resolves
		// to the wedding bucket. The named ceremonies below each have a
		// first-class occasion in models.Occasion, and the extractor is
		// expected to preserve that precision rather than flatten it: the
		// engine's occasionAffinity matrix is what maps e.g. Sangeet back onto
		// Wedding-tagged catalogue products (at 0.85), so keeping the specific
		// reading gives strictly better ranking than the coarse bucket would.
		{"need something for my cousin's shaadi", models.OccasionWedding},
		{"outfit for sangeet night", models.OccasionSangeet},
		{"what to wear for mehendi function", models.OccasionMehendi},
		{"haldi ceremony outfit please", models.OccasionHaldi},
		{"going to a diwali party", models.OccasionDiwali},
		{"navratri garba outfit", models.OccasionNavratri},
		{"eid outfit for the family", models.OccasionFestival},
		{"holi celebration clothes", models.OccasionFestival},
	}

	extractor := NewRuleExtractor(testVocab())
	for _, tc := range cases {
		t.Run(tc.prompt, func(t *testing.T) {
			intent := extractor.Extract(tc.prompt)
			if intent.Occasion == nil {
				t.Fatalf("Extract(%q): occasion = nil, want %s", tc.prompt, tc.want)
			}
			if *intent.Occasion != tc.want {
				t.Errorf("Extract(%q): occasion = %s, want %s", tc.prompt, *intent.Occasion, tc.want)
			}
		})
	}
}

// --- Budget parsing --------------------------------------------------------

func TestRuleExtractor_BudgetParsing(t *testing.T) {
	cases := []struct {
		prompt  string
		wantMin *int
		wantMax *int
	}{
		{"kurta under 2000", nil, intPtr(2000)},
		{"something under 2k", nil, intPtr(2000)},
		{"budget of 1.5k to 3k", intPtr(1500), intPtr(3000)},
		{"between 2000 and 5000", intPtr(2000), intPtr(5000)},
		{"above 3000", intPtr(3000), nil},
		{"cheap and simple", nil, intPtr(1500)},
		{"something premium", intPtr(3000), nil},
	}

	for _, tc := range cases {
		t.Run(tc.prompt, func(t *testing.T) {
			budget := ExtractBudget(tc.prompt)
			if budget == nil {
				budget = &recommendation.BudgetRange{}
			}
			gotMin := budget.Min
			gotMax := budget.Max

			if (tc.wantMin == nil) != (gotMin == nil) {
				t.Errorf("Extract(%q): min = %v, want %v", tc.prompt, ptrStr(gotMin), ptrStr(tc.wantMin))
			} else if tc.wantMin != nil && *gotMin != *tc.wantMin {
				t.Errorf("Extract(%q): min = %d, want %d", tc.prompt, *gotMin, *tc.wantMin)
			}

			if (tc.wantMax == nil) != (gotMax == nil) {
				t.Errorf("Extract(%q): max = %v, want %v", tc.prompt, ptrStr(gotMax), ptrStr(tc.wantMax))
			} else if tc.wantMax != nil && *gotMax != *tc.wantMax {
				t.Errorf("Extract(%q): max = %d, want %d", tc.prompt, *gotMax, *tc.wantMax)
			}
		})
	}
}

// TestRuleExtractor_BudgetRejectsUnrelatedNumbers guards against the classic
// false-positive: a number in the prompt that has nothing to do with money
// ("25 year old", "size 8") must not be read as a budget.
func TestRuleExtractor_BudgetRejectsUnrelatedNumbers(t *testing.T) {
	budget := ExtractBudget("outfit for my 25 year old cousin")
	// No numeric or qualitative budget language at all should yield a nil
	// range, not a zero-value one — see ExtractBudget's final fallback.
	if budget != nil && (budget.Min != nil || budget.Max != nil) {
		t.Errorf("expected no budget extracted from an age mention, got min=%v max=%v",
			ptrStr(budget.Min), ptrStr(budget.Max))
	}
}

// --- Negation / exclusion ----------------------------------------------------

func TestRuleExtractor_NegationExcludesRatherThanWants(t *testing.T) {
	extractor := NewRuleExtractor(testVocab())
	intent := extractor.Extract("party dress but not black, no maroon either")

	if len(intent.ExcludeColors) == 0 {
		t.Fatal("expected excluded colors from a negated prompt")
	}
	foundBlack, foundMaroon := false, false
	for _, c := range intent.ExcludeColors {
		if c == "black" {
			foundBlack = true
		}
		if c == "maroon" {
			foundMaroon = true
		}
	}
	if !foundBlack {
		t.Error("expected 'black' in excluded colors")
	}
	if !foundMaroon {
		t.Error("expected 'maroon' in excluded colors")
	}
	for _, c := range intent.Colors {
		if c == "black" || c == "maroon" {
			t.Errorf("negated color %q leaked into wanted Colors", c)
		}
	}
}

// --- Brand matching (short-name safety) --------------------------------------

func TestRuleExtractor_ShortBrandNameRequiresWholeWord(t *testing.T) {
	extractor := NewRuleExtractor(testVocab())

	// "W" is a real brand in the catalogue and also a letter that appears in
	// almost every sentence. It must only match as a standalone word.
	noise := extractor.Extract("wedding outfit with a nice fit")
	for _, b := range noise.Brands {
		if b == "W" {
			t.Errorf("brand 'W' matched inside unrelated text: %v", noise.Brands)
		}
	}

	explicit := extractor.Extract("something from W please")
	found := false
	for _, b := range explicit.Brands {
		if b == "W" {
			found = true
		}
	}
	if !found {
		t.Error("expected brand 'W' to match when mentioned as a standalone word")
	}
}

// TestRuleExtractor_RangeConnectorWordDoesNotMatchAsBrand guards against a bug
// found during manual testing: "AND" is both a real catalogue brand and the
// English conjunction used in range phrasing ("between 1500 and 3000"). Whole-
// word matching alone cannot distinguish the two — "and" genuinely is a whole
// word in a range phrase, not just a substring collision — so the fix has to
// remove the range phrase before brand matching ever sees it. See
// maskBudgetSpan.
func TestRuleExtractor_RangeConnectorWordDoesNotMatchAsBrand(t *testing.T) {
	extractor := NewRuleExtractor(testVocab())

	cases := []string{
		"formal shirt for office, between 1500 and 3000",
		"kurta between 2000 and 5000",
		"something nice, 1.5k and 3k",
	}
	for _, prompt := range cases {
		t.Run(prompt, func(t *testing.T) {
			intent := extractor.Extract(prompt)
			for _, b := range intent.Brands {
				if b == "AND" {
					t.Errorf("Extract(%q): brand 'AND' false-matched from range syntax, got brands=%v",
						prompt, intent.Brands)
				}
			}
			// The range itself must still parse correctly — the fix must not
			// have broken budget extraction while fixing brand matching.
			if intent.Budget == nil || !intent.Budget.HasMin() || !intent.Budget.HasMax() {
				t.Errorf("Extract(%q): expected a parsed budget range, got %+v", prompt, intent.Budget)
			}
		})
	}

	// A genuine standalone mention of the brand, well away from any numeric
	// range, must still work — the fix must not overcorrect into never
	// matching "AND" at all.
	explicit := extractor.Extract("something from AND please")
	found := false
	for _, b := range explicit.Brands {
		if b == "AND" {
			found = true
		}
	}
	if !found {
		t.Error("expected brand 'AND' to match when mentioned as a standalone word outside a range")
	}
}

// --- Vague prompts yield low confidence, not false precision -----------------

func TestRuleExtractor_VaguePromptHasLowConfidence(t *testing.T) {
	extractor := NewRuleExtractor(testVocab())
	intent := extractor.Extract("show me something nice")

	if intent.HasSemanticSignal() {
		t.Error("expected no semantic signal from a fully vague prompt")
	}
	if intent.Confidence > 0.4 {
		t.Errorf("confidence = %.2f for a vague prompt, want a low, honest confidence (<=0.4)", intent.Confidence)
	}
}

func ptrStr(p *int) string {
	if p == nil {
		return "<nil>"
	}
	return itoaHelper(*p)
}

func itoaHelper(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
