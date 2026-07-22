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

// BuildThemeCollections generates all three themed collections for a
// department+prompt in a single call, and returns them in a fixed order.
//
// Cross-collection uniqueness is enforced by passing each theme's picked
// product IDs and brands as exclusions into the next theme, so a product that
// appeared in theme 0 can never appear in theme 1 or 2. That guarantee is the
// entire reason this function exists: the client used to fire three parallel
// requests to the singular endpoint to get variety, but each of those calls
// was stateless, so the de-duplication was computed and then discarded and the
// three responses were free to overlap. Building all three together is what
// makes "three genuinely different collections" true rather than aspirational.
//
// Ordering note: an earlier design rotated the starting theme based on a
// suffix the client appended to the prompt (" minimal style", " bold colorful
// look"), so that three separate stateless calls would land on three different
// themes. The client no longer sends those suffixes — it makes one call and
// receives the whole set — so the rotation had become dead code that always
// started at 0. It has been removed rather than left to mislead: the order is
// now plainly 0, 1, 2, and every theme is returned regardless.
func (e *Engine) BuildThemeCollections(
	ctx context.Context,
	dept Department,
	base *Intent,
) ([]*Collection, error) {
	themes := themesFor(base, dept)

	// Build all three themes in order, carrying exclusions forward.
	// Seed with any exclusions already on the base intent (e.g. products the
	// client has already been shown, for "More Collections") so they apply to
	// every theme — applyTheme overwrites a theme's ExcludeProducts, so these
	// must ride in through usedProductIDs rather than on the intent alone.
	usedProductIDs := make([]string, 0, len(base.ExcludeProducts)+12)
	usedProductIDs = append(usedProductIDs, base.ExcludeProducts...)
	usedBrands := make([]string, 0, 12)

	// order[i] = which theme slot to fill at step i. Fixed rather than rotated:
	// exclusions accumulate along this order, so a stable order also makes the
	// result reproducible for a given prompt.
	order := [3]int{0, 1, 2}

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

// cheapestInSlot returns the lowest price available in a slot's candidate pool,
// used to reserve budget for downstream mandatory slots.
func (e *Engine) cheapestInSlot(ctx context.Context, spec SlotSpec, gender *models.Gender) int {
	pool, err := e.source.Candidates(ctx, CandidateFilter{
		Categories:   spec.Categories,
		Gender:       gender,
		Limit:        60,
		StrictBudget: true,
	})
	if err != nil || len(pool) == 0 {
		return 0
	}
	best := 0
	for _, p := range pool {
		if !genderPermits(gender, p) || !specAcceptsProduct(spec, p) {
			continue
		}
		if best == 0 || p.Price < best {
			best = p.Price
		}
	}
	return best
}

// specAcceptsProduct is the hard admissibility test for a slot.
//
// This single predicate is what makes the collection's structural guarantees
// hold: a candidate must occupy one of the slot's garment roles, and must
// satisfy any pinned product type. Everything else in the pipeline is scoring;
// this is the part that cannot be outvoted by a high rating.
func specAcceptsProduct(spec SlotSpec, p *models.Product) bool {
	if p == nil {
		return false
	}
	role := roleOf(p)
	roleOK := false
	for _, r := range spec.Roles {
		if r == role {
			roleOK = true
			break
		}
	}
	if !roleOK {
		return false
	}
	if len(spec.Types) > 0 && !productMatchesAnyType(p, spec.Types) {
		return false
	}
	return true
}

// consumeRoles marks the roles a chosen product occupies.
//
// A one-piece consumes Top and Bottom as well as OnePiece: a saree or a dress
// *is* the top and the bottom, so the slots for those must not also be filled.
// This is the rule that stops a skirt being stacked on a dress.
func consumeRoles(consumed map[GarmentRole]bool, role GarmentRole) {
	consumed[role] = true
	if role == RoleOnePiece {
		consumed[RoleTop] = true
		consumed[RoleBottom] = true
	}
}

// slotIsAvailable reports whether any of a slot's roles is still unfilled.
func slotIsAvailable(consumed map[GarmentRole]bool, spec SlotSpec) bool {
	for _, r := range spec.Roles {
		if !consumed[r] {
			return true
		}
	}
	return false
}

// BuildCollection assembles one outfit for a department by filling a slot
// template in order, spending down the stated budget as it goes.
//
// The pipeline per slot is: retrieve a wide pool → apply hard constraints
// (gender, garment role, product type, budget) → score with the existing
// relevance scorer and personalization boosts → add an outfit-compatibility
// bonus computed against the pieces already chosen → take the argmax.
//
// Three guarantees hold structurally rather than by convention:
//
//   - Exactly one item per garment role. Roles are consumed as they are filled,
//     and a one-piece consumes Top and Bottom, so no collection can contain two
//     bottoms (the trouser-and-skirt bug) or a dress plus a skirt.
//   - Gender is never violated. Every candidate passes genderPermits, applied
//     after retrieval as well as in SQL, because retrieval only filters when an
//     intent carries a gender and not every path sets one.
//   - The budget is never exceeded. Slots are skipped rather than overspent.
func (e *Engine) BuildCollection(ctx context.Context, dept Department, base *Intent) (*Collection, error) {
	if base == nil {
		base = NewIntent("")
	}
	if !DepartmentSupport(dept) {
		return nil, fmt.Errorf("recommendation: department %q has no catalogue support yet", dept)
	}

	template, tmpl := resolveTemplate(dept, base)
	if len(template) == 0 {
		return nil, fmt.Errorf("recommendation: no slot template for department %q", dept)
	}

	// Gender is resolved once, from the department, and applied to every slot.
	// When the department is gender-agnostic (Footwear, Accessories, Beauty) we
	// fall back to whatever the prompt asked for, so "women's heels" stays
	// women's even though the department itself carries no gender.
	gender := departmentGender(dept)
	if gender == nil && base.Gender != nil {
		g := *base.Gender
		gender = &g
	}

	var remaining *int
	if base.Budget.HasMax() {
		v := *base.Budget.Max
		remaining = &v
	}

	items := make([]CollectionItem, 0, len(template))
	skipped := make([]string, 0)
	consumed := make(map[GarmentRole]bool, 8)
	excluded := append([]string{}, base.ExcludeProducts...)
	var anchor *models.Product
	hasApparel := false

	for i, spec := range template {
		// A role already filled (e.g. Bottom after a dress) is not a skip — the
		// slot simply does not apply to this outfit.
		if !slotIsAvailable(consumed, spec) {
			continue
		}
		if remaining != nil && *remaining <= 0 {
			skipped = append(skipped, spec.Name)
			continue
		}

		// Reserve enough budget for the remaining mandatory slots so an
		// expensive top cannot starve the shoes.
		slotCap := 0
		if remaining != nil {
			slotCap = *remaining
			for j := i + 1; j < len(template); j++ {
				if template[j].Optional || !slotIsAvailable(consumed, template[j]) {
					continue
				}
				if floor := e.cheapestInSlot(ctx, template[j], gender); floor > 0 && slotCap-floor > 0 {
					slotCap -= floor
				}
			}
		}

		pick := e.pickForSlot(ctx, spec, tmpl, base, gender, excluded, slotCap, anchor, items)
		if pick == nil {
			skipped = append(skipped, spec.Name)
			continue
		}
		if remaining != nil && pick.Product.Price > *remaining {
			skipped = append(skipped, spec.Name)
			continue
		}

		items = append(items, CollectionItem{Slot: spec.Name, Product: *pick})
		role := roleOf(pick.Product)
		consumeRoles(consumed, role)
		if role.IsApparel() {
			hasApparel = true
			if anchor == nil {
				// The first apparel piece anchors the look: everything after it
				// is chosen to go with it.
				anchor = pick.Product
			}
		}
		excluded = append(excluded, pick.Product.ID.String())

		if remaining != nil {
			v := *remaining - pick.Product.Price
			remaining = &v
		}
	}

	// An outfit collection must contain at least one clothing item. Templates
	// that have no apparel slot at all (Footwear, Accessories, Beauty) are
	// exempt — they are browses, not outfits.
	templateHasApparel := false
	for _, spec := range template {
		for _, r := range spec.Roles {
			if r.IsApparel() {
				templateHasApparel = true
				break
			}
		}
		if templateHasApparel {
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

// pickForSlot retrieves, filters, scores and selects the single best product
// for one slot. It returns nil when nothing admissible is available, which the
// caller records as a skipped slot rather than filling with a wrong-role item.
func (e *Engine) pickForSlot(
	ctx context.Context,
	spec SlotSpec,
	tmpl OccasionTemplate,
	base *Intent,
	gender *models.Gender,
	excluded []string,
	slotCap int,
	anchor *models.Product,
	picked []CollectionItem,
) *ScoredProduct {
	filter := CandidateFilter{
		Categories:      spec.Categories,
		Gender:          gender,
		ExcludeColors:   base.ExcludeColors,
		ExcludeStyles:   base.ExcludeStyles,
		ExcludeBrands:   base.ExcludeBrands,
		ExcludeProducts: excluded,
		// A wide pool matters here: the role filter below discards most of it
		// (a Bottom slot pulls from categories that are mostly tops), so a
		// TopN-sized pool would routinely leave a slot empty.
		Limit:        240,
		StrictBudget: slotCap > 0,
	}
	if slotCap > 0 {
		ceiling := slotCap
		filter.MaxPrice = &ceiling
	}

	pool, err := e.source.Candidates(ctx, filter)
	if err != nil || len(pool) == 0 {
		return nil
	}

	// Hard constraints. Nothing below this point can rescue a candidate that
	// fails here — that is what makes these guarantees rather than preferences.
	admissible := make([]*models.Product, 0, len(pool))
	for _, p := range pool {
		if p == nil {
			continue
		}
		if !genderPermits(gender, p) {
			continue
		}
		if !specAcceptsProduct(spec, p) {
			continue
		}
		if slotCap > 0 && p.Price > slotCap {
			continue
		}
		admissible = append(admissible, p)
	}
	if len(admissible) == 0 {
		return nil
	}

	// Score with the existing relevance scorer so occasion, style, colour,
	// budget fit, rating and popularity all keep their established meaning,
	// then layer the shopper's personalization boosts on top.
	slotIntent := *base
	slotIntent.CategoryIn = spec.Categories
	slotIntent.Category = nil
	slotIntent.Gender = gender
	slotIntent.Sort = SortRelevance
	if len(spec.Types) > 0 {
		slotIntent.TypeTerms = spec.Types
	}
	if slotCap > 0 {
		slotBudget := BudgetRange{Max: &slotCap}
		slotIntent.Budget = &slotBudget
	}

	scored := e.scorer.ScoreAll(&slotIntent, admissible)
	ApplyPersonalization(scored, &slotIntent)
	if len(scored) == 0 {
		return nil
	}

	// Outfit compatibility decides between candidates the scorer rates
	// similarly — which is what turns a ranked list into a styled look.
	bestIdx := 0
	bestScore := -1.0
	for i := range scored {
		if scored[i].Product == nil {
			continue
		}
		s := scored[i].Score + compatibilityBonus(scored[i].Product, spec, tmpl, anchor, picked)
		if s > bestScore {
			bestScore, bestIdx = s, i
		}
	}
	if bestScore < 0 {
		return nil
	}

	pick := scored[bestIdx]
	pick.Score = clamp01(bestScore)
	pick.Rank = 1
	pick.FinalScore = pick.Score
	return &pick
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
