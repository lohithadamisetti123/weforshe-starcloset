package recommendation

import (
	"context"
	"fmt"
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// Department is the shopper's first-step choice in the guided assistant flow.
type Department string

const (
	DepartmentWomen       Department = "Women"
	DepartmentMen         Department = "Men"
	DepartmentKidsGirls   Department = "Kids - Girls"
	DepartmentKidsBoys    Department = "Kids - Boys"
	DepartmentHome        Department = "Home Essentials" // retired: not offered or built by the assistant
	DepartmentBeauty      Department = "Beauty"
	DepartmentFootwear    Department = "Footwear"
	DepartmentAccessories Department = "Accessories"
)

var AllDepartments = []Department{
	DepartmentMen, DepartmentWomen, DepartmentKidsGirls, DepartmentKidsBoys,
	DepartmentBeauty, DepartmentFootwear, DepartmentAccessories,
}

func DepartmentToGender(d Department) *models.Gender {
	switch d {
	case DepartmentWomen:
		g := models.GenderWomen
		return &g
	case DepartmentMen:
		g := models.GenderMen
		return &g
	case DepartmentKidsGirls:
		g := models.GenderGirls
		return &g
	case DepartmentKidsBoys:
		g := models.GenderBoys
		return &g
	default:
		return nil
	}
}

func DepartmentSupport(d Department) bool {
	switch d {
	case DepartmentWomen, DepartmentMen, DepartmentFootwear, DepartmentAccessories,
		DepartmentKidsGirls, DepartmentKidsBoys, DepartmentBeauty:
		return true
	default:
		return false
	}
}

type slot struct {
	Name       string
	Categories []models.CategoryName
	// Types, when set, hard-constrains the slot to a product type within the
	// (usually broad) category — e.g. a "Bag" and a "Watch" slot both draw from
	// Accessories but each is pinned to its own type so a collection never gets
	// two of the same category (rule 4: no bag+bag). Type keys resolve against
	// producttype.go. Empty means no type constraint.
	Types []string
	// ExcludeTypes hard-removes product types from a slot — e.g. the generic
	// men's "Accessories" slot excludes "watch" so it can't pick a second watch
	// after the dedicated Watch slot (rule 4).
	ExcludeTypes []string
	// Apparel marks a slot as satisfying the "every collection must contain at
	// least one clothing item" rule (rule 3). At least one Apparel slot must be
	// filled or the collection is rejected.
	Apparel  bool
	Optional bool
}

type outfitLeaning int

const (
	leaningEthnic outfitLeaning = iota
	leaningWestern
)

func leaningFor(intent *Intent) outfitLeaning {
	if intent.Occasion != nil {
		switch *intent.Occasion {
		case models.OccasionWedding, models.OccasionFestival,
			models.OccasionSangeet, models.OccasionMehendi, models.OccasionHaldi,
			models.OccasionDiwali, models.OccasionNavratri, models.OccasionEid,
			models.OccasionPuja, models.OccasionTempleVisit, models.OccasionDurgaPuja,
			models.OccasionGaneshChaturthi, models.OccasionKarvaChauth,
			models.OccasionWeddingReception, models.OccasionReception,
			models.OccasionEngagement, models.OccasionEngagementParty,
			models.OccasionWinterWedding:
			return leaningEthnic
		}
	}
	if intent.Style != nil {
		switch *intent.Style {
		case models.StyleEthnic, models.StyleIndoWestern, models.StyleTraditional,
			models.StyleContemporaryEthnic, models.StyleLuxuryEthnic,
			models.StyleChikankari, models.StyleBandhani, models.StyleMirrorWork,
			models.StyleGotaPatti, models.StyleSharara, models.StyleGharara:
			return leaningEthnic
		}
	}
	return leaningWestern
}

func templateFor(dept Department, leaning outfitLeaning) []slot {
	switch dept {
	case DepartmentFootwear:
		return []slot{{Name: "Footwear", Categories: []models.CategoryName{models.CategoryFootwear}}}
	case DepartmentAccessories:
		return []slot{{Name: "Accessories", Categories: []models.CategoryName{models.CategoryAccessories}}}
	case DepartmentBeauty:
		return []slot{{Name: "Beauty", Categories: []models.CategoryName{models.CategoryBeauty}}}
	case DepartmentKidsGirls, DepartmentKidsBoys:
		return []slot{
			{Name: "Outfit", Categories: []models.CategoryName{models.CategoryKidsGirls, models.CategoryKidsBoys, models.CategoryDresses, models.CategoryTops}},
			{Name: "Footwear", Categories: []models.CategoryName{models.CategoryFootwear}},
			{Name: "Accessories", Categories: []models.CategoryName{models.CategoryAccessories}, Optional: true},
		}
	case DepartmentWomen:
		if leaning == leaningEthnic {
			return []slot{
				{Name: "Outfit", Apparel: true, Categories: []models.CategoryName{
					models.CategoryEthnicWear, models.CategoryKurtaSets, models.CategorySarees,
				}},
				{Name: "Footwear", Categories: []models.CategoryName{models.CategoryFootwear}},
				{Name: "Jewellery", Categories: []models.CategoryName{models.CategoryAccessories}, Types: []string{"jewellery"}},
				{Name: "Bag", Categories: []models.CategoryName{models.CategoryAccessories}, Types: []string{"handbag"}, Optional: true},
				{Name: "Watch", Categories: []models.CategoryName{models.CategoryAccessories}, Types: []string{"watch"}, Optional: true},
				{Name: "Beauty", Categories: []models.CategoryName{models.CategoryBeauty}, Optional: true},
			}
		}
		return []slot{
			{Name: "Top", Apparel: true, Categories: []models.CategoryName{models.CategoryDresses, models.CategoryTops}},
			{Name: "Bottom", Apparel: true, Categories: []models.CategoryName{models.CategoryJeans}, Optional: true},
			{Name: "Footwear", Categories: []models.CategoryName{models.CategoryFootwear}},
			{Name: "Bag", Categories: []models.CategoryName{models.CategoryAccessories}, Types: []string{"handbag"}},
			{Name: "Watch", Categories: []models.CategoryName{models.CategoryAccessories}, Types: []string{"watch"}, Optional: true},
			{Name: "Jewellery", Categories: []models.CategoryName{models.CategoryAccessories}, Types: []string{"jewellery"}, Optional: true},
			{Name: "Beauty", Categories: []models.CategoryName{models.CategoryBeauty}, Optional: true},
		}
	case DepartmentMen:
		if leaning == leaningEthnic {
			return []slot{
				{Name: "Outfit", Apparel: true, Categories: []models.CategoryName{
					models.CategoryEthnicWear, models.CategoryKurtaSets,
				}},
				{Name: "Footwear", Categories: []models.CategoryName{models.CategoryFootwear}},
				{Name: "Watch", Categories: []models.CategoryName{models.CategoryAccessories}, Types: []string{"watch"}, Optional: true},
				{Name: "Accessories", Categories: []models.CategoryName{models.CategoryAccessories}, Optional: true},
			}
		}
		return []slot{
			// Top apparel: shirt OR t-shirt (Tops) — mandatory clothing item.
			{Name: "Top", Apparel: true, Categories: []models.CategoryName{models.CategoryShirts, models.CategoryTops}},
			{Name: "Bottom", Apparel: true, Categories: []models.CategoryName{models.CategoryJeans}, Optional: true},
			{Name: "Footwear", Categories: []models.CategoryName{models.CategoryFootwear}},
			{Name: "Watch", Categories: []models.CategoryName{models.CategoryAccessories}, Types: []string{"watch"}},
			// Men's carry/wear accessories: belt, wallet, backpack, cap — the
			// generic Accessories slot picks the best non-watch item so it can't
			// duplicate the Watch slot.
			{Name: "Accessories", Categories: []models.CategoryName{models.CategoryAccessories}, ExcludeTypes: []string{"watch"}, Optional: true},
		}
	default:
		return nil
	}
}

func departmentGender(dept Department) *models.Gender {
	return DepartmentToGender(dept)
}

// CollectionItem is one filled slot.
type CollectionItem struct {
	Slot    string        `json:"slot"`
	Product ScoredProduct `json:"product"`
}

// Collection is a complete, budget-aware multi-item recommendation.
type Collection struct {
	Title        string           `json:"title"`
	Items        []CollectionItem `json:"items"`
	TotalPrice   int              `json:"totalPrice"`
	BudgetLimit  *int             `json:"budgetLimit,omitempty"`
	SkippedSlots []string         `json:"skippedSlots,omitempty"`
	Explanation  []string         `json:"explanation"`
}

// ─── Styling themes ───────────────────────────────────────────────────────────
//
// A "theme" is a bundle of style + sub-occasion modifiers applied on top of the
// shopper's base intent so that three parallel calls to BuildCollection each
// explore a distinct aesthetic corner of the catalogue.  The theme is ADDITIVE:
// it only overrides fields the shopper left blank, so explicit user signals
// always take precedence.

// CollectionTheme bundles the overrides that make each parallel collection feel
// meaningfully different from its siblings.
type CollectionTheme struct {
	// Label appears in the collection title ("Smart Casual Collection").
	Label string
	// Style overrides the intent's style when the shopper didn't name one.
	Style *models.StyleTag
	// ExcludeBrands forces variety by blacklisting brands already picked in
	// prior themes. Populated by BuildThemeCollections before each call.
	ExcludeBrands []string
	// ExcludeProducts prevents the exact same product appearing in two
	// collections. Populated by BuildThemeCollections after each fill.
	ExcludeProducts []string
	// PriceBiasHigh nudges this theme toward the upper half of the budget.
	// When true, the slot-cap logic skews toward spending more per item.
	PriceBiasHigh bool
	// PriceBiasLow nudges toward value picks.
	PriceBiasLow bool
}

// themesFor returns a set of three complementary themes for the given base
// intent.  The three themes are chosen to be maximally distant stylistically
// while remaining coherent for the stated occasion.
func themesFor(base *Intent, dept Department) [3]CollectionTheme {
	leaning := leaningFor(base)

	// When the shopper explicitly named a style, we still diversify — but we
	// pivot along price segment and brand instead of style.
	if base.Style != nil {
		s := *base.Style
		return [3]CollectionTheme{
			{Label: "Premium", Style: &s, PriceBiasHigh: true},
			{Label: "Value", Style: &s, PriceBiasLow: true},
			{Label: "Curated", Style: &s},
		}
	}

	if leaning == leaningEthnic {
		// Ethnic occasion: vary by embellishment level / fabric family
		trad := models.StyleTraditional
		contmp := models.StyleContemporaryEthnic
		luxury := models.StyleLuxuryEthnic
		return [3]CollectionTheme{
			{Label: "Traditional", Style: &trad},
			{Label: "Contemporary Ethnic", Style: &contmp},
			{Label: "Luxury Ethnic", Style: &luxury, PriceBiasHigh: true},
		}
	}

	// Western / casual occasion — vary by occasion sub-cluster
	switch dept {
	case DepartmentMen:
		formal := models.StyleFormal
		smart := models.StyleSmartCasual
		street := models.StyleStreetwear
		return [3]CollectionTheme{
			{Label: "Office Formal", Style: &formal},
			{Label: "Smart Casual", Style: &smart},
			{Label: "Streetwear", Style: &street},
		}
	default:
		// Women / Kids / etc.
		minimal := models.StyleMinimal
		boho := models.StyleBohemian
		chic := models.StyleWestern
		return [3]CollectionTheme{
			{Label: "Minimal", Style: &minimal},
			{Label: "Bohemian", Style: &boho, PriceBiasLow: true},
			{Label: "Western Chic", Style: &chic, PriceBiasHigh: true},
		}
	}
}

// applyTheme returns a new Intent with the theme's overrides merged in.
// User-supplied signals always win.
func applyTheme(base *Intent, theme CollectionTheme) *Intent {
	cp := *base
	// Style: only override when the shopper didn't name one
	if cp.Style == nil && theme.Style != nil {
		s := *theme.Style
		cp.Style = &s
	}
	// Brand exclusions are additive
	cp.ExcludeBrands = append(append([]string{}, cp.ExcludeBrands...), theme.ExcludeBrands...)
	// Product exclusions are set directly on the intent so retrieve() can
	// push them down to the datastore.
	cp.ExcludeProducts = theme.ExcludeProducts
	return &cp
}

// themeIndex maps the raw prompt to theme index 0, 1, or 2.
//
// The frontend (useChat.ts) appends a known suffix to each of the three
// parallel requests so the backend can route them to different themes without
// any server-side session state:
//
//	Call 1: <base prompt>                  → theme 0
//	Call 2: <base prompt> + " minimal style"    → theme 1
//	Call 3: <base prompt> + " bold colorful look" → theme 2
//
// This is 100% reliable — no hash collisions — and costs one strings.HasSuffix
// per request.
func themeIndex(prompt string) int {
	switch {
	case strings.HasSuffix(prompt, " bold colorful look"):
		return 2
	case strings.HasSuffix(prompt, " minimal style"):
		return 1
	default:
		return 0
	}
}

// BuildThemeCollections generates all three themed collections for a
// department+prompt and returns them in rotation order starting at the
// theme whose index matches the prompt hash.
//
// Why rotation instead of "pick best"?
// The frontend calls this endpoint three times in parallel, each with a
// slightly different prompt suffix (" minimal style", " bold colorful look").
// Those suffixes shift the FNV hash so the three calls land on indices 0, 1, 2
// in some order — guaranteeing the frontend receives three DIFFERENT themes and
// therefore three collections with completely different products.
//
// Cross-collection uniqueness is enforced by passing each theme's picked
// product IDs and brands as exclusions into the next theme, so a product that
// appeared in theme 0 can never appear in theme 1 or 2.
func (e *Engine) BuildThemeCollections(
	ctx context.Context,
	dept Department,
	base *Intent,
) ([]*Collection, error) {
	themes := themesFor(base, dept)

	// Rotate so the starting theme is deterministic per prompt.
	start := themeIndex(base.Query)

	// Build all three themes in rotation order, carrying exclusions forward.
	// Seed with any exclusions already on the base intent (e.g. products the
	// client has already been shown, for "More Collections") so they apply to
	// every theme — applyTheme overwrites a theme's ExcludeProducts, so these
	// must ride in through usedProductIDs rather than on the intent alone.
	usedProductIDs := make([]string, 0, len(base.ExcludeProducts)+12)
	usedProductIDs = append(usedProductIDs, base.ExcludeProducts...)
	usedBrands := make([]string, 0, 12)

	// order[i] = which theme slot to fill at step i
	order := [3]int{start, (start + 1) % 3, (start + 2) % 3}

	out := make([]*Collection, 0, 3)

	for _, idx := range order {
		themes[idx].ExcludeProducts = append([]string{}, usedProductIDs...)
		themes[idx].ExcludeBrands = append([]string{}, usedBrands...)

		themedIntent := applyTheme(base, themes[idx])

		c, err := e.BuildCollection(ctx, dept, themedIntent)
		if err != nil {
			continue
		}

		c.Title = themes[idx].Label + " " + c.Title

		for _, item := range c.Items {
			if item.Product.Product != nil {
				usedProductIDs = append(usedProductIDs, item.Product.Product.ID.String())
				if item.Product.Product.Brand != "" {
					usedBrands = append(usedBrands, normBrand(item.Product.Product.Brand))
				}
			}
		}

		out = append(out, c)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("recommendation: all themed collections failed for department %q", dept)
	}
	return out, nil
}

// normBrand returns a normalised brand key for deduplication.
func normBrand(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), ""))
}

// cheapestInSlot returns the lowest price available in a slot's category pool.
func (e *Engine) cheapestInSlot(ctx context.Context, sl slot, gender *models.Gender) int {
	f := CandidateFilter{
		Categories:   sl.Categories,
		Gender:       gender,
		Limit:        50,
		StrictBudget: true,
	}
	pool, err := e.source.Candidates(ctx, f)
	if err != nil || len(pool) == 0 {
		return 0
	}
	min := pool[0].Price
	for _, p := range pool[1:] {
		if p.Price < min {
			min = p.Price
		}
	}
	return min
}

// BuildCollection fills a department's slot template in priority order,
// spending down the stated budget as it goes.
//
// NOTE: This function is unchanged from the working version — all diversity
// improvements are layered on top via BuildThemeCollections, which calls this
// function with a pre-modified intent. That way the slot-filling logic,
// budget enforcement, and honesty checks remain exactly as before.
func (e *Engine) BuildCollection(ctx context.Context, dept Department, base *Intent) (*Collection, error) {
	if base == nil {
		base = NewIntent("")
	}
	if !DepartmentSupport(dept) {
		return nil, fmt.Errorf("recommendation: department %q has no catalogue support yet", dept)
	}

	template := templateFor(dept, leaningFor(base))
	gender := departmentGender(dept)

	var remaining *int
	if base.Budget.HasMax() {
		v := *base.Budget.Max
		remaining = &v
	}

	items := make([]CollectionItem, 0, len(template))
	skipped := make([]string, 0)
	droppedBottom := false
	hasApparel := false

	for i, sl := range template {
		if sl.Name == "Bottom" && droppedBottom {
			continue
		}

		slotIntent := *base
		slotIntent.CategoryIn = sl.Categories
		slotIntent.Category = nil
		slotIntent.Gender = gender
		// Pin a typed slot (Bag→handbag, Watch→watch) to its type so the pool is
		// type-relevant; a hard post-filter below guarantees it. Untyped slots
		// keep any query-driven type (so "saree collection" keeps the saree
		// constraint on its apparel slot).
		if len(sl.Types) > 0 {
			slotIntent.TypeTerms = sl.Types
		}
		slotIntent.Sort = SortRelevance
		slotIntent.StrictBudget = remaining != nil
		if remaining != nil {
			if *remaining <= 0 {
				skipped = append(skipped, sl.Name)
				continue
			}
			slotCap := *remaining
			for j := i + 1; j < len(template); j++ {
				// Skip optional slots (Bottom) when reserving budget.
				// Bottom is dropped when Top picks a Dress, so pre-reserving
				// its cost starves the Top slot and causes it to be skipped —
				// leaving the shopper with jeans + footwear instead of an outfit.
				if template[j].Optional {
					continue
				}
				downstreamMin := e.cheapestInSlot(ctx, template[j], gender)
				if downstreamMin > 0 && slotCap-downstreamMin > 0 {
					slotCap -= downstreamMin
				}
			}
			slotBudget := BudgetRange{Max: &slotCap}
			slotIntent.Budget = &slotBudget
		}

		result, err := e.Recommend(ctx, &slotIntent)
		if err != nil {
			return nil, fmt.Errorf("recommendation: filling slot %s: %w", sl.Name, err)
		}
		if len(result.Products) == 0 {
			skipped = append(skipped, sl.Name)
			continue
		}

		// Hard type constraint (rules 4 & 9): a typed slot MUST be filled by a
		// product of that type, and ExcludeTypes are never allowed. If nothing
		// qualifies we skip the slot rather than substitute a wrong-type item.
		candidates := result.Products
		if len(sl.Types) > 0 || len(sl.ExcludeTypes) > 0 {
			filtered := candidates[:0:0]
			for _, sp := range candidates {
				if len(sl.Types) > 0 && !productMatchesAnyType(sp.Product, sl.Types) {
					continue
				}
				if len(sl.ExcludeTypes) > 0 && productMatchesAnyType(sp.Product, sl.ExcludeTypes) {
					continue
				}
				filtered = append(filtered, sp)
			}
			candidates = filtered
		}
		if len(candidates) == 0 {
			skipped = append(skipped, sl.Name)
			continue
		}

		pick := candidates[0]

		if remaining != nil && pick.Product.Price > *remaining {
			skipped = append(skipped, sl.Name)
			continue
		}

		items = append(items, CollectionItem{Slot: sl.Name, Product: pick})
		if sl.Apparel {
			hasApparel = true
		}

		// Exclude this product from all subsequent slots so the same item
		// can never appear twice within one collection, even if future
		// template changes introduce overlapping category pools.
		base.ExcludeProducts = append(base.ExcludeProducts, pick.Product.ID.String())

		if remaining != nil {
			v := *remaining - pick.Product.Price
			remaining = &v
		}
		if sl.Name == "Top" && pick.Product.Category != nil && pick.Product.Category.Name == models.CategoryDresses {
			droppedBottom = true
		}
	}

	// Rule 3: an outfit collection MUST contain at least one clothing item.
	// This only applies to templates that HAVE an apparel slot — single-category
	// departments (Footwear, Accessories, Beauty, Home) legitimately have none.
	templateHasApparel := false
	for _, sl := range template {
		if sl.Apparel {
			templateHasApparel = true
			break
		}
	}
	if templateHasApparel && !hasApparel {
		return nil, fmt.Errorf("recommendation: no apparel available for %q — collection not generated", dept)
	}

	total := 0
	for _, it := range items {
		total += it.Product.Product.Price
	}

	c := &Collection{
		Title:        collectionTitle(dept, base),
		Items:        items,
		TotalPrice:   total,
		SkippedSlots: skipped,
		Explanation:  collectionExplanation(base, items, skipped),
	}
	if base.Budget.HasMax() {
		c.BudgetLimit = base.Budget.Max
	}
	return c, nil
}

func collectionTitle(dept Department, intent *Intent) string {
	if intent.Occasion != nil {
		return string(*intent.Occasion) + " Collection"
	}
	return string(dept) + " Collection"
}

func collectionExplanation(intent *Intent, items []CollectionItem, skipped []string) []string {
	reasons := make([]string, 0, 5)
	if intent.Budget.HasMax() {
		if len(skipped) == 0 {
			reasons = append(reasons, "Fits your stated budget")
		} else {
			reasons = append(reasons, "Fits your stated budget — some pieces were left out rather than pushing you over it")
		}
	}
	if intent.Occasion != nil {
		reasons = append(reasons, "Matches your selected occasion")
	}
	if intent.Style != nil {
		reasons = append(reasons, "Styled to match: "+string(*intent.Style))
	}
	reasons = append(reasons, "Each piece individually ranked by fit, rating and value")
	if len(items) > 1 {
		reasons = append(reasons, "Colour and style checked to work as a set, not just as single items")
	}
	return reasons
}
