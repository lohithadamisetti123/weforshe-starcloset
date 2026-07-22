package recommendation

import (
	"github.com/myntra/bharat-ai-backend/internal/models"
)

// ─── Garment roles ───────────────────────────────────────────────────────────
//
// A garment ROLE is the part of an outfit a product occupies: top, bottom,
// one-piece, layer, footwear, bag, and so on. It is the unit the collection
// builder reasons about, and it exists because the catalogue's categories are
// too coarse to build an outfit from.
//
// Concretely: the catalogue has no "Trousers" or "Skirts" category. Trousers
// live under Jeans, skirts under Tops/Dresses. A slot template written against
// categories therefore cannot express "exactly one bottom" — which is precisely
// how an Office collection ended up holding both a trouser and a skirt. Roles
// fix that at the root: both classify as RoleBottom, the builder fills each role
// at most once, and the conflict becomes structurally impossible rather than
// something a rule has to catch after the fact.
//
// Classification reads the title first (most specific, and the only place a
// "trouser" is distinguishable from a "skirt") and falls back to the category.

type GarmentRole string

const (
	RoleTop       GarmentRole = "Top"
	RoleBottom    GarmentRole = "Bottom"
	RoleOnePiece  GarmentRole = "OnePiece" // dress, saree, lehenga, kurta set, jumpsuit
	RoleLayer     GarmentRole = "Layer"    // blazer, jacket, shrug, waistcoat
	RoleFootwear  GarmentRole = "Footwear"
	RoleBag       GarmentRole = "Bag"
	RoleWatch     GarmentRole = "Watch"
	RoleJewellery GarmentRole = "Jewellery"
	RoleBelt      GarmentRole = "Belt"
	RoleEyewear   GarmentRole = "Eyewear"
	RoleHeadwear  GarmentRole = "Headwear"
	RoleScarf     GarmentRole = "Scarf"
	RoleBeauty    GarmentRole = "Beauty"
	RoleUnknown   GarmentRole = "Unknown"
)

// IsApparel reports whether a role is a clothing item (as opposed to footwear,
// an accessory or beauty). At least one apparel role must be filled for an
// outfit collection to be honest about being an outfit.
func (r GarmentRole) IsApparel() bool {
	switch r {
	case RoleTop, RoleBottom, RoleOnePiece, RoleLayer:
		return true
	}
	return false
}

// roleTitleTokens maps a role to the title tokens that identify it. Order
// matters at classification time: the most specific roles are tested first, so
// a "shirt dress" resolves to OnePiece rather than Top, and a "denim jacket"
// resolves to Layer rather than Bottom.
var roleTitleTokens = []struct {
	role   GarmentRole
	tokens []string
}{
	// One-piece garments first: they subsume a top AND a bottom, so a
	// misclassification here is the most expensive one to get wrong.
	{RoleOnePiece, []string{
		"dress", "gown", "frock", "jumpsuit", "romper", "playsuit", "dungaree",
		"saree", "sari", "lehenga", "anarkali", "sherwani", "abaya", "kaftan",
	}},
	// Layers before tops: a "blazer" and a "denim jacket" are layers.
	{RoleLayer, []string{
		"blazer", "jacket", "coat", "overcoat", "shrug", "cardigan", "waistcoat",
		"nehru", "bomber", "parka", "sweatshirt", "hoodie", "sweater", "pullover",
	}},
	// Bottoms before tops: "trouser", "skirt", "palazzo" are all one role, which
	// is what makes trouser+skirt impossible in one collection.
	{RoleBottom, []string{
		"trouser", "pant", "chino", "jeans", "denim", "jegging", "jogger",
		"trackpant", "skirt", "short", "legging", "palazzo", "culotte",
		"churidar", "salwar", "patiala", "dhoti", "sharara", "gharara", "pyjama",
	}},
	{RoleTop, []string{
		"shirt", "blouse", "tee", "tshirt", "top", "tank", "camisole", "cami",
		"crop", "polo", "tunic", "kurti", "kurta", "peplum", "bodysuit", "choli",
	}},
	// Accessory and footwear roles.
	{RoleWatch, []string{"watch", "chronograph"}},
	{RoleJewellery, []string{
		"necklace", "earring", "jhumka", "chandbali", "stud", "hoop", "ring",
		"bracelet", "bangle", "kada", "choker", "pendant", "tikka", "maang",
		"nath", "passa", "kundan", "polki", "haram", "jewellery", "jewelry", "anklet",
	}},
	{RoleBag, []string{
		"bag", "handbag", "tote", "clutch", "satchel", "hobo", "bucket", "sling",
		"crossbody", "backpack", "purse", "pouch", "wallet", "shopper",
	}},
	{RoleBelt, []string{"belt"}},
	{RoleEyewear, []string{"sunglasses", "sunglass", "shades", "aviator", "eyewear"}},
	{RoleHeadwear, []string{"cap", "hat", "beanie", "turban"}},
	{RoleScarf, []string{"scarf", "stole", "dupatta", "shawl"}},
	{RoleFootwear, []string{
		"shoe", "sneaker", "trainer", "heel", "stiletto", "pump", "boot",
		"loafer", "oxford", "moccasin", "sandal", "wedge", "espadrille",
		"kolhapuri", "flat", "ballet", "jutti", "mule", "slipper", "flipflop", "derby",
	}},
}

// roleForCategory is the fallback when a title says nothing useful.
func roleForCategory(c models.CategoryName) GarmentRole {
	switch c {
	case models.CategorySarees, models.CategoryDresses, models.CategoryKurtaSets,
		models.CategoryEthnicWear:
		return RoleOnePiece
	case models.CategoryShirts, models.CategoryTops:
		return RoleTop
	case models.CategoryJeans:
		return RoleBottom
	case models.CategoryBlazers:
		return RoleLayer
	case models.CategoryFootwear:
		return RoleFootwear
	case models.CategoryAccessories:
		// Deliberately not guessed: an untyped accessory could be a bag, a
		// watch or a necklace, and guessing wrong would let two of the same
		// kind into one collection. RoleUnknown never satisfies a typed slot.
		return RoleUnknown
	case models.CategoryBeauty:
		return RoleBeauty
	}
	return RoleUnknown
}

// roleOf classifies a product into its garment role.
func roleOf(p *models.Product) GarmentRole {
	if p == nil {
		return RoleUnknown
	}
	title := tokenize(p.Title)
	for _, entry := range roleTitleTokens {
		for _, term := range entry.tokens {
			if anyTokenMatches(title, term) {
				return entry.role
			}
		}
	}
	if p.Category != nil {
		return roleForCategory(p.Category.Name)
	}
	return RoleUnknown
}

// ─── Formality ───────────────────────────────────────────────────────────────
//
// Formality is a 0..1 axis used to keep an outfit internally coherent: a
// stylist does not pair an oxford shirt with running shoes. It is derived from
// the product's title and style rather than stored, so it works on the existing
// catalogue with no schema change.

var formalTokens = []string{
	"formal", "oxford", "derby", "brogue", "loafer", "blazer", "suit", "tuxedo",
	"trouser", "chino", "pump", "stiletto", "heel", "tie", "waistcoat", "sheath",
}

var casualTokens = []string{
	"sneaker", "trainer", "jeans", "denim", "tee", "tshirt", "hoodie", "short",
	"jogger", "trackpant", "flipflop", "slipper", "canvas", "cargo", "graphic",
	"sweatshirt", "beanie", "cap",
}

// formalityOf returns roughly how dressy a product is, in [0,1]. 0.5 is the
// neutral middle for pieces that read either way.
func formalityOf(p *models.Product) float64 {
	if p == nil {
		return 0.5
	}
	score := 0.5

	switch p.Style {
	case models.StyleFormal:
		score = 0.9
	case models.StyleSmartCasual:
		score = 0.7
	case models.StyleCasual, models.StyleStreetwear, models.StyleAthleisure:
		score = 0.25
	case models.StyleTraditional, models.StyleEthnic, models.StyleLuxuryEthnic:
		score = 0.75
	}

	title := tokenize(p.Title)
	for _, t := range formalTokens {
		if anyTokenMatches(title, t) {
			score += 0.15
			break
		}
	}
	for _, t := range casualTokens {
		if anyTokenMatches(title, t) {
			score -= 0.15
			break
		}
	}
	return clamp01(score)
}

// ─── Gender enforcement ──────────────────────────────────────────────────────
//
// Gender is a hard constraint, applied as a predicate over every candidate
// *after* retrieval as well as in the SQL WHERE clause. The redundancy is the
// point: retrieval only filters when intent.Gender is set, and it is not set on
// every path (a free-text prompt whose gender the extractor could not read, or
// a gender-agnostic department). A shopper seeing menswear in a women's
// collection is the single most obvious way for this product to look broken, so
// the builder never trusts that an upstream filter ran.

// genderPermits reports whether a product may appear in a collection built for
// `want`.
//
//   - Women accept Women + Unisex
//   - Men accept Men + Unisex
//   - Girls/Boys accept only that kid gender, plus Unisex *only* inside a kids
//     category — an adult unisex watch is not a kids product, and "kids only
//     for kids" has to mean that literally.
//   - A nil `want` (a gender-agnostic department like Footwear) permits
//     everything, because there is no shopper gender to be inconsistent with.
func genderPermits(want *models.Gender, p *models.Product) bool {
	if want == nil {
		return true
	}
	if p == nil {
		return false
	}
	switch *want {
	case models.GenderWomen:
		return p.Gender == models.GenderWomen || p.Gender == models.GenderUnisex
	case models.GenderMen:
		return p.Gender == models.GenderMen || p.Gender == models.GenderUnisex
	case models.GenderGirls, models.GenderBoys:
		if p.Gender == *want {
			return true
		}
		if p.Gender == models.GenderUnisex && p.Category != nil {
			switch p.Category.Name {
			case models.CategoryKidsGirls, models.CategoryKidsBoys:
				return true
			}
		}
		return false
	case models.GenderUnisex:
		return true
	}
	return true
}
