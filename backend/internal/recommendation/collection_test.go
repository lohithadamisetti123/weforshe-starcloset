package recommendation

import (
	"context"
	"testing"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// collectionFixture is a small, purpose-built catalogue covering every slot
// a collection template can ask for (outfit, top, bottom, footwear,
// accessories, for both genders) — the shared engine_test.go fixture doesn't
// carry any Accessories products, which would make every collection test
// exercise the "slot skipped" path whether that's what's being tested or not.
func collectionFixture() []*models.Product {
	return []*models.Product{
		// Women, ethnic / wedding
		product("Embroidered Anarkali Set", "Biba", "Kurta Sets", 2199, 2999, 4.5, 900,
			models.GenderWomen, models.OccasionWedding, models.StyleEthnic, "Maroon"),
		product("Banarasi Silk Saree", "FabIndia", "Sarees", 2799, 3500, 4.7, 1400,
			models.GenderWomen, models.OccasionWedding, models.StyleEthnic, "Red"),
		product("Jhumka Earrings", "Rubans", "Accessories", 899, 1200, 4.6, 500,
			models.GenderWomen, models.OccasionWedding, models.StyleEthnic, "Gold"),
		product("Embellished Juttis", "Nike", "Footwear", 1199, 1600, 4.4, 700,
			models.GenderWomen, models.OccasionWedding, models.StyleEthnic, "Rose Gold"),

		// Women, western / office
		product("Formal Blouse Top", "Zara", "Tops", 1499, 1999, 3.6, 150,
			models.GenderWomen, models.OccasionOffice, models.StyleFormal, "White"),
		product("Tailored Trousers", "AND", "Jeans", 1799, 2299, 4.1, 350,
			models.GenderWomen, models.OccasionOffice, models.StyleFormal, "Black"),
		product("Office Pumps", "Metro", "Footwear", 1999, 2599, 4.0, 300,
			models.GenderWomen, models.OccasionOffice, models.StyleFormal, "Nude"),
		product("Minimal Watch", "Titan", "Accessories", 2999, 3699, 4.5, 600,
			models.GenderWomen, models.OccasionOffice, models.StyleFormal, "Silver"),
		// Deliberately the strongest candidate in its price/quality bracket
		// among the Top-slot pool (Tops ∪ Dresses) — see
		// TestBuildCollection_DressDropsRedundantBottomSlot, which needs the
		// Top slot to resolve to this one-piece Dress deterministically
		// rather than depending on close-call scoring between it and the
		// Blouse Top above.
		product("Wrap Dress", "H&M", "Dresses", 1999, 2599, 4.9, 900,
			models.GenderWomen, models.OccasionOffice, models.StyleFormal, "Navy"),

		// Men, western / office
		product("Formal Office Shirt", "Van Heusen", "Shirts", 1499, 1999, 4.3, 1100,
			models.GenderMen, models.OccasionOffice, models.StyleFormal, "White"),
		product("Slim Fit Chinos", "Levi's", "Jeans", 1799, 2499, 4.2, 800,
			models.GenderMen, models.OccasionOffice, models.StyleFormal, "Beige"),
		product("Oxford Shoes", "Aldo", "Footwear", 3499, 4299, 4.4, 500,
			models.GenderMen, models.OccasionOffice, models.StyleFormal, "Brown"),
		product("Leather Belt", "AND", "Accessories", 1199, 1599, 4.1, 250,
			models.GenderMen, models.OccasionOffice, models.StyleFormal, "Black"),

		// Men, ethnic / wedding
		product("Wedding Sherwani", "Manyavar", "Ethnic Wear", 6500, 8999, 4.6, 500,
			models.GenderMen, models.OccasionWedding, models.StyleEthnic, "Gold"),
		product("Mojari Footwear", "Manyavar", "Footwear", 1899, 2399, 4.3, 200,
			models.GenderMen, models.OccasionWedding, models.StyleEthnic, "Gold"),
	}
}

func collectionTestEngine() *Engine {
	return newTestEngine(collectionFixture())
}

func TestBuildCollection_WomenWeddingStaysWithinBudget(t *testing.T) {
	engine := collectionTestEngine()
	intent := NewIntent("wedding collection under 5000")
	intent.Occasion = occasionPtr(models.OccasionWedding)
	intent.Budget = &BudgetRange{Max: intPtr(5000)}

	c, err := engine.BuildCollection(context.Background(), DepartmentWomen, intent)
	if err != nil {
		t.Fatalf("BuildCollection returned error: %v", err)
	}
	if len(c.Items) == 0 {
		t.Fatal("expected a non-empty collection")
	}
	if c.TotalPrice > 5000 {
		t.Errorf("total price %d exceeds stated budget of 5000", c.TotalPrice)
	}
	for _, it := range c.Items {
		if it.Product.Product.Gender != models.GenderWomen && it.Product.Product.Gender != models.GenderUnisex {
			t.Errorf("slot %s: gender = %s, want Women or Unisex", it.Slot, it.Product.Product.Gender)
		}
	}
	// The wedding fixture set has an outfit, footwear and accessory option
	// within budget, so nothing should have needed to be skipped here.
	if len(c.SkippedSlots) != 0 {
		t.Errorf("expected no skipped slots, got %v", c.SkippedSlots)
	}
}

func TestBuildCollection_NeverMixesDepartments(t *testing.T) {
	engine := collectionTestEngine()
	intent := NewIntent("office collection")
	intent.Occasion = occasionPtr(models.OccasionOffice)

	c, err := engine.BuildCollection(context.Background(), DepartmentMen, intent)
	if err != nil {
		t.Fatalf("BuildCollection returned error: %v", err)
	}
	for _, it := range c.Items {
		if it.Product.Product.Gender == models.GenderWomen {
			t.Errorf("Men department collection included a Women's product: %s", it.Product.Product.Title)
		}
	}
}

func TestBuildCollection_MenOfficeProducesShirtTrouserFootwearAccessories(t *testing.T) {
	engine := collectionTestEngine()
	intent := NewIntent("formal office collection")
	intent.Occasion = occasionPtr(models.OccasionOffice)
	intent.Style = stylePtr(models.StyleFormal)

	c, err := engine.BuildCollection(context.Background(), DepartmentMen, intent)
	if err != nil {
		t.Fatalf("BuildCollection returned error: %v", err)
	}
	gotSlots := make(map[string]bool)
	for _, it := range c.Items {
		gotSlots[it.Slot] = true
	}
	for _, want := range []string{"Shirt", "Trouser", "Footwear", "Accessories"} {
		if !gotSlots[want] {
			t.Errorf("expected slot %q to be filled, got slots: %v", want, gotSlots)
		}
	}
}

func TestBuildCollection_DressDropsRedundantBottomSlot(t *testing.T) {
	engine := collectionTestEngine()
	// Western-leaning (no wedding/festival occasion, no ethnic style), so the
	// Women template's Top slot pool is {Tops, Dresses} — and the fixture's
	// Wrap Dress is constructed to be the clear quality-signal winner in that
	// pool (see collectionFixture), so this doesn't depend on a close scoring
	// call to land a specific way.
	intent := NewIntent("office collection")
	intent.Occasion = occasionPtr(models.OccasionOffice)
	intent.Style = stylePtr(models.StyleFormal)

	c, err := engine.BuildCollection(context.Background(), DepartmentWomen, intent)
	if err != nil {
		t.Fatalf("BuildCollection returned error: %v", err)
	}
	var topItem *CollectionItem
	for i := range c.Items {
		if c.Items[i].Slot == "Top" {
			topItem = &c.Items[i]
		}
	}
	if topItem == nil {
		t.Fatal("expected a Top slot to be filled")
	}
	if topItem.Product.Product.Title != "Wrap Dress" {
		t.Fatalf("expected the Top slot to resolve to the Wrap Dress, got %q — "+
			"either scoring shifted or the fixture needs adjusting", topItem.Product.Product.Title)
	}
	for _, it := range c.Items {
		if it.Slot == "Bottom" {
			t.Error("Bottom slot present alongside a one-piece Dress in the Top slot — should have been dropped")
		}
	}
}

func TestBuildCollection_UnsupportedDepartmentReturnsError(t *testing.T) {
	engine := collectionTestEngine()
	intent := NewIntent("cute dress for my daughter")

	_, err := engine.BuildCollection(context.Background(), DepartmentKidsGirls, intent)
	if err == nil {
		t.Fatal("expected an error for an unsupported department, got nil")
	}
}

func TestBuildCollection_FootwearDepartmentIsSingleSlot(t *testing.T) {
	engine := collectionTestEngine()
	intent := NewIntent("formal shoes")

	c, err := engine.BuildCollection(context.Background(), DepartmentFootwear, intent)
	if err != nil {
		t.Fatalf("BuildCollection returned error: %v", err)
	}
	if len(c.Items) != 1 || c.Items[0].Slot != "Footwear" {
		t.Errorf("expected exactly one Footwear slot, got %+v", c.Items)
	}
	if c.Items[0].Product.Product.Category == nil || c.Items[0].Product.Product.Category.Name != models.CategoryFootwear {
		t.Errorf("expected a Footwear product, got category %v", c.Items[0].Product.Product.Category)
	}
}

func TestBuildCollection_TightBudgetSkipsSlotsRatherThanExceedingIt(t *testing.T) {
	engine := collectionTestEngine()
	intent := NewIntent("wedding collection under 3000")
	intent.Occasion = occasionPtr(models.OccasionWedding)
	// Tight enough that the outfit alone (cheapest wedding outfit in the
	// fixture is 2899) consumes nearly the whole budget, leaving no room for
	// footwear or accessories.
	intent.Budget = &BudgetRange{Max: intPtr(3000)}

	c, err := engine.BuildCollection(context.Background(), DepartmentWomen, intent)
	if err != nil {
		t.Fatalf("BuildCollection returned error: %v", err)
	}
	if c.TotalPrice > 3000 {
		t.Errorf("total price %d exceeds stated budget of 3000", c.TotalPrice)
	}
	if len(c.SkippedSlots) == 0 {
		t.Error("expected at least one slot to be skipped under a tight budget")
	}
}
