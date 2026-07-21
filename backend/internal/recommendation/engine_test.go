package recommendation

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/myntra/bharat-ai-backend/internal/config"
	"github.com/myntra/bharat-ai-backend/internal/models"
)

// --- Fixture catalogue -------------------------------------------------------
//
// A small, hand-curated catalogue covering enough of the vocabulary space
// (occasion x gender x style x price x color x brand) to exercise every stage
// of the pipeline without a database. Real relevance regressions have been
// caught by exactly this kind of test in the commit history this file is
// meant to guard.

func product(title, brand, category string, price, mrp int, rating float64, ratingCount int,
	gender models.Gender, occasion models.Occasion, style models.StyleTag, color string) *models.Product {
	p := &models.Product{
		ID:          uuid.New(),
		Slug:        strings.ToLower(strings.ReplaceAll(title, " ", "-")),
		Title:       title,
		Brand:       brand,
		Description: title + " by " + brand,
		Image:       "https://picsum.photos/seed/" + title + "/600/800",
		Category:    &models.Category{ID: uuid.New(), Name: models.CategoryName(category)},
		Price:       price,
		MRP:         mrp,
		Rating:      rating,
		RatingCount: ratingCount,
		Gender:      gender,
		Occasion:    occasion,
		Style:       style,
		Color:       color,
	}
	_ = p.BeforeSave(nil)
	return p
}

func fixtureCatalogue() []*models.Product {
	return []*models.Product{
		product("Embroidered Anarkali Set", "Biba", "Kurta Sets", 2899, 3999, 4.5, 900,
			models.GenderWomen, models.OccasionWedding, models.StyleEthnic, "Maroon"),
		product("Banarasi Silk Saree", "FabIndia", "Sarees", 4200, 5500, 4.7, 1400,
			models.GenderWomen, models.OccasionWedding, models.StyleEthnic, "Red"),
		product("Sequin Party Gown", "Forever 21", "Dresses", 2600, 3200, 4.1, 300,
			models.GenderWomen, models.OccasionParty, models.StyleWestern, "Black"),
		product("Cotton Casual Kurti", "W", "Tops", 799, 1199, 4.0, 2200,
			models.GenderWomen, models.OccasionCasual, models.StyleEthnic, "Blue"),
		product("Formal Office Shirt", "Van Heusen", "Shirts", 1499, 1999, 4.3, 1100,
			models.GenderMen, models.OccasionOffice, models.StyleFormal, "White"),
		product("Slim Fit Chinos", "Levi's", "Jeans", 1799, 2499, 4.2, 800,
			models.GenderMen, models.OccasionOffice, models.StyleFormal, "Beige"),
		product("Streetwear Hoodie", "Nike", "Tops", 2999, 3999, 4.4, 1600,
			models.GenderUnisex, models.OccasionCasual, models.StyleStreetwear, "Grey"),
		product("Wedding Sherwani", "Manyavar", "Ethnic Wear", 6500, 8999, 4.6, 500,
			models.GenderMen, models.OccasionWedding, models.StyleEthnic, "Gold"),
		product("Minimal White Sneakers", "Puma", "Footwear", 2499, 3199, 4.3, 2000,
			models.GenderUnisex, models.OccasionCasual, models.StyleMinimal, "White"),
		product("Festive Silk Kurta", "Fabindia", "Kurta Sets", 1899, 2499, 4.1, 650,
			models.GenderMen, models.OccasionFestival, models.StyleEthnic, "Mustard"),
		product("Date Night Slip Dress", "Zara", "Dresses", 3400, 4200, 4.0, 250,
			models.GenderWomen, models.OccasionDateNight, models.StyleWestern, "Black"),
		product("Travel Denim Jacket", "Roadster", "Jeans", 1999, 2699, 3.9, 700,
			models.GenderUnisex, models.OccasionTravel, models.StyleStreetwear, "Blue"),
		product("Bohemian Maxi Dress", "Rare Rabbit", "Dresses", 3100, 3900, 4.2, 320,
			models.GenderWomen, models.OccasionFestival, models.StyleBohemian, "Rust"),
		product("Indo-Western Jacket Set", "Manyavar", "Ethnic Wear", 5200, 6800, 4.4, 410,
			models.GenderMen, models.OccasionWedding, models.StyleIndoWestern, "Navy"),
		product("Budget Cotton Saree", "W", "Sarees", 899, 1299, 3.6, 180,
			models.GenderWomen, models.OccasionCasual, models.StyleEthnic, "Green"),
	}
}

// stubSource is an in-memory CandidateSource that applies the same hard
// filters a real repository would (see internal/repositories/product_repo.go,
// Candidates), so the engine's retrieval and relaxation ladder are exercised
// exactly as they run in production.
type stubSource struct {
	products []*models.Product
}

func (s *stubSource) Candidates(_ context.Context, f CandidateFilter) ([]*models.Product, error) {
	out := make([]*models.Product, 0, len(s.products))
	for _, p := range s.products {
		if f.Gender != nil && p.Gender != *f.Gender && p.Gender != models.GenderUnisex {
			continue
		}
		if f.MinPrice != nil && p.Price < *f.MinPrice {
			continue
		}
		if f.MaxPrice != nil && p.Price > *f.MaxPrice {
			continue
		}
		if containsColor(f.ExcludeColors, p.Color) {
			continue
		}
		if containsStyle(f.ExcludeStyles, p.Style) {
			continue
		}
		if containsBrand(f.ExcludeBrands, p.Brand) {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func containsColor(list []string, color string) bool {
	for _, c := range list {
		if strings.Contains(strings.ToLower(color), strings.ToLower(c)) {
			return true
		}
	}
	return false
}

func containsStyle(list []models.StyleTag, style models.StyleTag) bool {
	for _, s := range list {
		if s == style {
			return true
		}
	}
	return false
}

func containsBrand(list []string, brand string) bool {
	for _, b := range list {
		if strings.EqualFold(b, brand) {
			return true
		}
	}
	return false
}

func testConfig() config.RecommendationConfig {
	cfg := config.RecommendationConfig{
		Weights: config.Weights{
			Occasion: 0.24, Style: 0.15, Gender: 0.10, Category: 0.09,
			Budget: 0.18, Color: 0.05, Brand: 0.04, Text: 0.09,
			Quality: 0.09, Value: 0.05,
		},
		DiversityLambda: 0.72,
		TopN:            6,
		CandidatePool:   200,
	}
	cfg.Weights = cfg.Weights.Normalized()
	return cfg
}

func newTestEngine(products []*models.Product) *Engine {
	return NewEngine(&stubSource{products: products}, testConfig())
}

func genderPtr(g models.Gender) *models.Gender               { return &g }
func occasionPtr(o models.Occasion) *models.Occasion         { return &o }
func stylePtr(s models.StyleTag) *models.StyleTag            { return &s }
func categoryPtr(c models.CategoryName) *models.CategoryName { return &c }
func intPtr(n int) *int                                      { return &n }

// --- Tests -------------------------------------------------------------------

func TestEngine_RanksByOccasionAndGender(t *testing.T) {
	engine := newTestEngine(fixtureCatalogue())
	intent := NewIntent("wedding outfit")
	intent.Gender = genderPtr(models.GenderWomen)
	intent.Occasion = occasionPtr(models.OccasionWedding)

	result, err := engine.Recommend(context.Background(), intent)
	if err != nil {
		t.Fatalf("Recommend returned error: %v", err)
	}
	if len(result.Products) == 0 {
		t.Fatal("expected a non-empty shortlist")
	}

	top := result.Products[0]
	if top.Product.Gender != models.GenderWomen && top.Product.Gender != models.GenderUnisex {
		t.Errorf("top result gender = %s, want Women or Unisex", top.Product.Gender)
	}
	if top.Product.Occasion != models.OccasionWedding {
		t.Errorf("top result occasion = %s, want Wedding (got %q)", top.Product.Occasion, top.Product.Title)
	}
}

func TestEngine_NeverEmptyUnderImpossibleConstraints(t *testing.T) {
	engine := newTestEngine(fixtureCatalogue())

	// A combination nothing in the fixture catalogue satisfies simultaneously:
	// a men's, festival-occasion, sub-₹500 item. The engine's contract is to
	// relax constraints until it can answer rather than return nothing.
	intent := NewIntent("cheap festival wear")
	intent.Gender = genderPtr(models.GenderMen)
	intent.Occasion = occasionPtr(models.OccasionFestival)
	intent.Budget = &BudgetRange{Max: intPtr(500)}

	result, err := engine.Recommend(context.Background(), intent)
	if err != nil {
		t.Fatalf("Recommend returned error: %v", err)
	}
	if len(result.Products) == 0 {
		t.Fatal("engine returned an empty shortlist despite a non-empty catalogue")
	}
	if len(result.Relaxed) == 0 {
		t.Error("expected at least one relaxation to be recorded")
	}
}

func TestEngine_HardExclusionsAreNeverReturned(t *testing.T) {
	engine := newTestEngine(fixtureCatalogue())
	intent := NewIntent("party dress, not black")
	intent.Gender = genderPtr(models.GenderWomen)
	intent.Occasion = occasionPtr(models.OccasionParty)
	intent.ExcludeColors = []string{"black"}

	result, err := engine.Recommend(context.Background(), intent)
	if err != nil {
		t.Fatalf("Recommend returned error: %v", err)
	}
	for _, sp := range result.Products {
		if strings.Contains(strings.ToLower(sp.Product.Color), "black") {
			t.Errorf("excluded color %q returned as %q", sp.Product.Color, sp.Product.Title)
		}
	}
}

func TestEngine_MatchPercentIsBoundedAndMonotonic(t *testing.T) {
	engine := newTestEngine(fixtureCatalogue())
	intent := NewIntent("office wear")
	intent.Gender = genderPtr(models.GenderMen)
	intent.Occasion = occasionPtr(models.OccasionOffice)

	result, err := engine.Recommend(context.Background(), intent)
	if err != nil {
		t.Fatalf("Recommend returned error: %v", err)
	}

	prev := 101
	for _, sp := range result.Products {
		if sp.MatchPercent < 60 || sp.MatchPercent > 99 {
			t.Errorf("%q: matchPercent = %d, want within [60,99]", sp.Product.Title, sp.MatchPercent)
		}
		if sp.MatchPercent > prev {
			t.Errorf("%q: matchPercent %d exceeds previous rank's %d — not monotonic", sp.Product.Title, sp.MatchPercent, prev)
		}
		prev = sp.MatchPercent
	}
}

func TestEngine_VagueQueryFallsBackToQualityLed(t *testing.T) {
	// "show me something nice" — no semantic signal at all. The engine must
	// still return a full, sensible shortlist rather than an error or an
	// arbitrary catalogue slice.
	engine := newTestEngine(fixtureCatalogue())
	intent := NewIntent("show me something nice")

	result, err := engine.Recommend(context.Background(), intent)
	if err != nil {
		t.Fatalf("Recommend returned error: %v", err)
	}
	if len(result.Products) == 0 {
		t.Fatal("expected a non-empty shortlist for a vague prompt")
	}
	if result.Summary == "" {
		t.Error("expected a non-empty summary")
	}
}

func TestEngine_ExplicitSortOverridesDiversification(t *testing.T) {
	engine := newTestEngine(fixtureCatalogue())
	intent := NewIntent("cheapest first")
	intent.Sort = SortPriceAsc

	result, err := engine.Recommend(context.Background(), intent)
	if err != nil {
		t.Fatalf("Recommend returned error: %v", err)
	}
	for i := 1; i < len(result.Products); i++ {
		if result.Products[i].Product.Price < result.Products[i-1].Product.Price {
			t.Errorf("price_asc sort violated at rank %d: %d then %d",
				i, result.Products[i-1].Product.Price, result.Products[i].Product.Price)
		}
	}
}

func TestEngine_RanksAreSequentialFromOne(t *testing.T) {
	engine := newTestEngine(fixtureCatalogue())
	result, err := engine.Recommend(context.Background(), NewIntent("casual wear"))
	if err != nil {
		t.Fatalf("Recommend returned error: %v", err)
	}
	for i, sp := range result.Products {
		if sp.Rank != i+1 {
			t.Errorf("product %d: rank = %d, want %d", i, sp.Rank, i+1)
		}
	}
}

func TestEngine_CategoryFilterNarrowsResults(t *testing.T) {
	engine := newTestEngine(fixtureCatalogue())
	intent := NewIntent("sarees")
	intent.Category = categoryPtr(models.CategorySarees)
	intent.Gender = genderPtr(models.GenderWomen)

	result, err := engine.Recommend(context.Background(), intent)
	if err != nil {
		t.Fatalf("Recommend returned error: %v", err)
	}
	if len(result.Products) == 0 {
		t.Fatal("expected non-empty results for sarees")
	}

	// Category is a soft signal (see engine.go's CandidateFilter comment), so
	// this checks the category nudges a saree into the shortlist at all,
	// rather than asserting it claims rank 1 — that would pin the test to the
	// exact signal weights rather than to the behaviour worth guaranteeing.
	found := false
	for _, sp := range result.Products {
		if sp.Product.Category != nil && sp.Product.Category.Name == models.CategorySarees {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected at least one Saree in the shortlist when Category=Sarees was requested")
	}
}

func TestEngine_NilIntentDoesNotPanic(t *testing.T) {
	engine := newTestEngine(fixtureCatalogue())
	result, err := engine.Recommend(context.Background(), nil)
	if err != nil {
		t.Fatalf("Recommend returned error for nil intent: %v", err)
	}
	if len(result.Products) == 0 {
		t.Fatal("expected a non-empty shortlist for a nil intent")
	}
}

func TestEngine_EmptyCatalogueReturnsEmptyNotError(t *testing.T) {
	engine := newTestEngine(nil)
	result, err := engine.Recommend(context.Background(), NewIntent("anything"))
	if err != nil {
		t.Fatalf("Recommend returned error for empty catalogue: %v", err)
	}
	if len(result.Products) != 0 {
		t.Errorf("expected zero products from an empty catalogue, got %d", len(result.Products))
	}
	if result.Summary == "" {
		t.Error("expected a summary even for an empty result")
	}
}
