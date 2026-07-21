package recommendation

import (
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// Product-type filtering closes the gap between a broad category and what a
// shopper actually asked for.
//
// The catalogue keeps watches, bags and jewellery all under one "Accessories"
// category, and footwear under one "Footwear" category. That is fine for
// storage but useless for a query like "watch": category extraction resolves it
// to Accessories, and every accessory then competes on rating, so the shopper
// asking for a watch is shown the highest-rated necklace. The same is true for
// "sneakers" vs "heels", or "clutch" vs "backpack".
//
// The fix is a canonical product-type vocabulary. A type named in the query
// (watch, necklace, clutch, sneakers, saree…) narrows the candidate pool to
// products that both live in that type's parent category AND whose title tokens
// identify it. The category constraint is what keeps "sneakers" from matching a
// mesh-strap watch or a canvas bag; the title match is what separates a watch
// from a necklace inside Accessories. Whole-word token matching (not substring)
// keeps "ring" from matching "earring".
//
// It is a soft narrowing: if nothing in the pool matches (e.g. a "saree" term
// reaching the footwear slot of a collection), the original pool is returned,
// so a slot is never emptied and cross-category collections still fill.

type productType struct {
	categories []models.CategoryName // parent categories a match must belong to
	terms      []string              // title tokens that identify the type
}

// productTypes maps a canonical type key to its parent category set and the
// title tokens that identify it. The same tokens serve as query triggers.
var productTypes = map[string]productType{
	// ── Accessories sub-types ────────────────────────────────────────────
	"watch":      {[]models.CategoryName{models.CategoryAccessories}, []string{"watch"}},
	"necklace":   {[]models.CategoryName{models.CategoryAccessories}, []string{"necklace", "choker", "pendant", "haram"}},
	"earrings":   {[]models.CategoryName{models.CategoryAccessories}, []string{"earring", "jhumka", "chandbali", "stud", "hoop", "huggie"}},
	"ring":       {[]models.CategoryName{models.CategoryAccessories}, []string{"ring"}},
	"bracelet":   {[]models.CategoryName{models.CategoryAccessories}, []string{"bracelet", "kada", "cuff"}},
	"bangles":    {[]models.CategoryName{models.CategoryAccessories}, []string{"bangle"}},
	"jewellery":  {[]models.CategoryName{models.CategoryAccessories}, []string{"necklace", "earring", "ring", "bracelet", "bangle", "choker", "pendant", "jhumka", "chandbali", "tikka", "maang", "nath", "passa", "kundan", "polki", "haram", "jewellery", "jewelry"}},
	"handbag":    {[]models.CategoryName{models.CategoryAccessories}, []string{"bag", "handbag", "tote", "clutch", "satchel", "hobo", "bucket", "sling", "crossbody", "backpack", "purse", "pouch", "vanity", "saddle", "shopper", "shoulder"}},
	"clutch":     {[]models.CategoryName{models.CategoryAccessories}, []string{"clutch"}},
	"backpack":   {[]models.CategoryName{models.CategoryAccessories}, []string{"backpack"}},
	"belt":       {[]models.CategoryName{models.CategoryAccessories}, []string{"belt"}},
	"sunglasses": {[]models.CategoryName{models.CategoryAccessories}, []string{"sunglasses", "sunglass", "shades", "aviator"}},
	"cap":        {[]models.CategoryName{models.CategoryAccessories}, []string{"cap"}},
	"scarf":      {[]models.CategoryName{models.CategoryAccessories}, []string{"scarf", "stole"}},

	// ── Footwear sub-types ───────────────────────────────────────────────
	"sneakers": {[]models.CategoryName{models.CategoryFootwear}, []string{"sneaker", "trainer", "running", "walking", "canvas", "slip-on", "mesh", "platform", "dad"}},
	"heels":    {[]models.CategoryName{models.CategoryFootwear}, []string{"heel", "stiletto", "pump"}},
	"boots":    {[]models.CategoryName{models.CategoryFootwear}, []string{"boot", "chelsea", "riding"}},
	"loafers":  {[]models.CategoryName{models.CategoryFootwear}, []string{"loafer", "oxford", "moccasin"}},
	"sandals":  {[]models.CategoryName{models.CategoryFootwear}, []string{"sandal", "wedge", "espadrille", "kolhapuri", "strappy", "slingback"}},
	"flats":    {[]models.CategoryName{models.CategoryFootwear}, []string{"flat", "ballet", "jutti", "mule"}},

	// ── Apparel — redundant with category, but helps free-text chat where
	//    category extraction is unsure. Each is pinned to its own category. ──
	"saree":   {[]models.CategoryName{models.CategorySarees}, []string{"saree", "sari", "banarasi"}},
	"lehenga": {[]models.CategoryName{models.CategoryEthnicWear}, []string{"lehenga"}},
	"kurta":   {[]models.CategoryName{models.CategoryKurtaSets}, []string{"kurta", "kurti", "anarkali"}},
	"dress":   {[]models.CategoryName{models.CategoryDresses}, []string{"dress", "gown", "frock"}},
	"shirt":   {[]models.CategoryName{models.CategoryShirts}, []string{"shirt"}},
	"jeans":   {[]models.CategoryName{models.CategoryJeans}, []string{"jeans", "denim"}},
	"blazer":  {[]models.CategoryName{models.CategoryBlazers}, []string{"blazer"}},
}

// tokenize splits a string into lowercase alphanumeric word tokens.
func tokenize(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
}

// tokenMatches reports whether tokens a and b name the same thing, tolerant of
// simple singular/plural ("watch"/"watches", "heel"/"heels"). It is deliberately
// not a substring test — that is what makes "ring" match "earring", the bug this
// file exists to avoid.
func tokenMatches(a, b string) bool {
	a, b = strings.ToLower(a), strings.ToLower(b)
	if a == b {
		return true
	}
	return a == b+"s" || b == a+"s" || a == b+"es" || b == a+"es"
}

func anyTokenMatches(tokens []string, term string) bool {
	for _, t := range tokens {
		if tokenMatches(t, term) {
			return true
		}
	}
	return false
}

// ExtractTypeTerms returns the canonical product-type keys named in the query
// (e.g. ["watch"] or ["saree"]). Empty when the query names no product type.
// The name is retained for compatibility; the values are type keys, resolved
// against productTypes by FilterByType.
func ExtractTypeTerms(query string) []string {
	qt := tokenize(query)
	if len(qt) == 0 {
		return nil
	}
	var out []string
	for key, pt := range productTypes {
		for _, term := range pt.terms {
			if anyTokenMatches(qt, term) {
				out = append(out, key)
				break
			}
		}
	}
	return out
}

// productMatchesType reports whether a product satisfies a type: its category is
// one of the type's parents (when known) and a title token identifies it.
func productMatchesType(p *models.Product, pt productType) bool {
	if p == nil {
		return false
	}
	if p.Category != nil && len(pt.categories) > 0 {
		ok := false
		for _, c := range pt.categories {
			if p.Category.Name == c {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	title := tokenize(p.Title)
	for _, term := range pt.terms {
		if anyTokenMatches(title, term) {
			return true
		}
	}
	return false
}

// productMatchesAnyType reports whether a product matches any of the given
// canonical type keys (e.g. ["handbag"], ["watch"]). Used by the collection
// builder to hard-constrain a slot to a product type so a collection never
// contains two items of the same category (rule 4) and never substitutes a
// different type when the requested one is unavailable (rule 9).
func productMatchesAnyType(p *models.Product, keys []string) bool {
	for _, k := range keys {
		if pt, ok := productTypes[k]; ok && productMatchesType(p, pt) {
			return true
		}
	}
	return false
}

// FilterByType narrows a candidate pool to products matching any requested
// product type. Soft: if no product matches (the type is irrelevant to this
// pool, e.g. "saree" against a footwear slot), the original pool is returned so
// the caller never ends up empty.
func FilterByType(pool []*models.Product, intent *Intent) []*models.Product {
	if intent == nil || len(intent.TypeTerms) == 0 || len(pool) == 0 {
		return pool
	}
	types := make([]productType, 0, len(intent.TypeTerms))
	for _, key := range intent.TypeTerms {
		if pt, ok := productTypes[key]; ok {
			types = append(types, pt)
		}
	}
	if len(types) == 0 {
		return pool
	}
	filtered := make([]*models.Product, 0, len(pool))
	for _, p := range pool {
		for _, pt := range types {
			if productMatchesType(p, pt) {
				filtered = append(filtered, p)
				break
			}
		}
	}
	if len(filtered) == 0 {
		return pool
	}
	return filtered
}
