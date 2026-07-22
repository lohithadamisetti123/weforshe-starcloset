package recommendation

import (
	"github.com/myntra/bharat-ai-backend/internal/models"
)

// ─── Occasion templates ──────────────────────────────────────────────────────
//
// A template is a stylist's brief expressed as data: which garment roles make up
// this outfit, in which order they should be chosen, which are mandatory, and
// what footwear and accessories belong with them.
//
// Templates are configuration, not code. Adding an occasion or retuning one is
// an edit to the table at the bottom of this file — no branching logic to
// touch, which is what keeps the builder itself small and testable.
//
// Two invariants make the "one item per slot" guarantee structural rather than
// aspirational:
//
//  1. Every slot declares the garment ROLES it accepts, and the builder rejects
//     any candidate whose role is not among them. A skirt (RoleBottom) can
//     therefore never land in a Top slot, whatever the catalogue's category
//     says.
//  2. A role, once filled, is consumed. A one-piece additionally consumes Top
//     and Bottom, because a saree or a dress *is* the top and the bottom.

// SlotSpec describes one position in an outfit.
type SlotSpec struct {
	// Name is what the UI shows ("Shirt", "Bottom", "Footwear").
	Name string
	// Roles are the garment roles allowed to fill this slot. Hard constraint.
	Roles []GarmentRole
	// Categories is the retrieval allow-list. It is deliberately wider than the
	// role: skirts and trousers are filed under different categories, so a
	// Bottom slot pulls from several and lets the role filter do the deciding.
	Categories []models.CategoryName
	// Types hard-constrains the slot to product types (watch, handbag), so an
	// Accessories pool cannot hand a Watch slot a necklace.
	Types []string
	// PreferTypes are soft: matching candidates get a styling bonus. This is
	// how footwear rules are expressed — an office look prefers loafers and
	// heels over sneakers without banning anything and emptying the slot.
	PreferTypes []string
	// Optional slots may be skipped without the collection being incomplete.
	Optional bool
	// Formality is the target dressiness for this position, 0..1.
	Formality float64
}

// OccasionTemplate is a full brief for one occasion, per gender lane.
type OccasionTemplate struct {
	Key string
	// Formality is the occasion's overall dressiness, used to keep the pieces
	// coherent with one another.
	Formality float64
	// Styles are the aesthetics that suit this occasion. Used as a soft scoring
	// preference, never a filter.
	Styles []models.StyleTag
	Women  []SlotSpec
	Men    []SlotSpec
}

// Shared category groups. A Bottom slot draws from several categories because
// the catalogue has no dedicated trousers/skirts category — the role filter is
// what guarantees only bottoms survive.
var (
	catsTop      = []models.CategoryName{models.CategoryShirts, models.CategoryTops}
	catsBottom   = []models.CategoryName{models.CategoryJeans, models.CategoryTops, models.CategoryDresses}
	catsTopOrOne = []models.CategoryName{models.CategoryShirts, models.CategoryTops, models.CategoryDresses}
	catsLayer    = []models.CategoryName{models.CategoryBlazers}
	catsShoes    = []models.CategoryName{models.CategoryFootwear}
	catsAcc      = []models.CategoryName{models.CategoryAccessories}
	catsEthnicW  = []models.CategoryName{models.CategorySarees, models.CategoryEthnicWear, models.CategoryKurtaSets, models.CategoryDresses}
	catsEthnicM  = []models.CategoryName{models.CategoryEthnicWear, models.CategoryKurtaSets}
)

// Reusable slot builders keep the template table readable and consistent.

func slotShoes(formality float64, prefer ...string) SlotSpec {
	return SlotSpec{Name: "Footwear", Roles: []GarmentRole{RoleFootwear}, Categories: catsShoes,
		PreferTypes: prefer, Formality: formality}
}

func slotWatch(optional bool) SlotSpec {
	return SlotSpec{Name: "Watch", Roles: []GarmentRole{RoleWatch}, Categories: catsAcc,
		Types: []string{"watch"}, Optional: optional, Formality: 0.6}
}

func slotBag(name string, optional bool, types ...string) SlotSpec {
	if len(types) == 0 {
		types = []string{"handbag"}
	}
	return SlotSpec{Name: name, Roles: []GarmentRole{RoleBag}, Categories: catsAcc,
		Types: types, Optional: optional, Formality: 0.5}
}

func slotJewellery(optional bool) SlotSpec {
	return SlotSpec{Name: "Jewellery", Roles: []GarmentRole{RoleJewellery}, Categories: catsAcc,
		Types: []string{"jewellery"}, Optional: optional, Formality: 0.7}
}

func slotBelt(optional bool) SlotSpec {
	return SlotSpec{Name: "Belt", Roles: []GarmentRole{RoleBelt}, Categories: catsAcc,
		Types: []string{"belt"}, Optional: optional, Formality: 0.7}
}

// slotAccessory is the catch-all men's accessory slot. It excludes the roles
// that have dedicated slots so it can never duplicate them.
func slotAccessory(optional bool) SlotSpec {
	return SlotSpec{Name: "Accessories", Roles: []GarmentRole{RoleBelt, RoleEyewear, RoleHeadwear, RoleBag, RoleScarf},
		Categories: catsAcc, Optional: optional, Formality: 0.5}
}

func slotLayer(name string, formality float64) SlotSpec {
	return SlotSpec{Name: name, Roles: []GarmentRole{RoleLayer}, Categories: catsLayer,
		Optional: true, Formality: formality}
}

// occasionTemplates is the configuration table. Twelve briefs, each defining
// mandatory apparel, optional accessories and footwear preferences.
var occasionTemplates = map[string]OccasionTemplate{

	"office": {
		Key: "office", Formality: 0.85,
		Styles: []models.StyleTag{models.StyleFormal, models.StyleSmartCasual, models.StyleBusinessCasual},
		Women: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.85},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.85},
			slotLayer("Blazer", 0.9),
			slotShoes(0.8, "heels", "loafers", "flats"),
			slotBag("Bag", false, "handbag"),
			slotWatch(true),
		},
		Men: []SlotSpec{
			// Named for the garment rather than the position: a men's office
			// look is a shirt and a trouser, and saying so reads like a stylist.
			{Name: "Shirt", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.9},
			{Name: "Trouser", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.9},
			slotLayer("Blazer", 0.95),
			slotShoes(0.9, "loafers", "boots"),
			slotWatch(false),
			slotAccessory(true),
		},
	},

	"business-meeting": {
		Key: "business-meeting", Formality: 0.95,
		Styles: []models.StyleTag{models.StyleFormal, models.StyleClassic},
		Women: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.95},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.95},
			slotLayer("Blazer", 0.95),
			slotShoes(0.9, "heels", "loafers"),
			slotBag("Bag", false, "handbag"),
			slotWatch(true),
		},
		Men: []SlotSpec{
			{Name: "Shirt", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.95},
			{Name: "Trouser", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.95},
			slotLayer("Blazer", 1.0),
			slotShoes(0.95, "loafers"),
			slotWatch(false),
			slotBelt(true),
		},
	},

	"casual": {
		Key: "casual", Formality: 0.3,
		Styles: []models.StyleTag{models.StyleCasual, models.StyleMinimal, models.StyleWestern},
		Women: []SlotSpec{
			// Accepts a dress: a one-piece covers the top half, and the builder
			// then drops the Bottom slot rather than stacking a skirt on it.
			{Name: "Top", Roles: []GarmentRole{RoleTop, RoleOnePiece}, Categories: catsTopOrOne, Formality: 0.3},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.3},
			slotShoes(0.25, "sneakers", "flats", "sandals"),
			slotBag("Bag", true, "handbag"),
			slotWatch(true),
		},
		Men: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.3},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.3},
			slotShoes(0.25, "sneakers"),
			slotWatch(true),
			slotAccessory(true),
		},
	},

	"college": {
		Key: "college", Formality: 0.3,
		Styles: []models.StyleTag{models.StyleCasual, models.StyleStreetwear, models.StyleWestern},
		Women: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleTop, RoleOnePiece}, Categories: catsTopOrOne, Formality: 0.3},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.3},
			slotShoes(0.2, "sneakers", "flats"),
			slotBag("Backpack", true, "backpack", "handbag"),
			slotWatch(true),
		},
		Men: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.3},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.3},
			slotShoes(0.2, "sneakers"),
			slotBag("Backpack", true, "backpack"),
			slotWatch(true),
		},
	},

	"wedding": {
		Key: "wedding", Formality: 0.9,
		Styles: []models.StyleTag{models.StyleTraditional, models.StyleEthnic, models.StyleLuxuryEthnic},
		Women: []SlotSpec{
			{Name: "Outfit", Roles: []GarmentRole{RoleOnePiece}, Categories: catsEthnicW, Formality: 0.95},
			slotShoes(0.8, "heels", "flats", "sandals"),
			slotJewellery(false),
			slotBag("Clutch", true, "clutch", "handbag"),
		},
		Men: []SlotSpec{
			{Name: "Outfit", Roles: []GarmentRole{RoleOnePiece, RoleTop}, Categories: catsEthnicM, Formality: 0.95},
			slotShoes(0.85, "loafers", "flats"),
			slotWatch(true),
			slotAccessory(true),
		},
	},

	"reception": {
		Key: "reception", Formality: 0.9,
		Styles: []models.StyleTag{models.StyleLuxuryEthnic, models.StyleIndoWestern, models.StyleElegant},
		Women: []SlotSpec{
			{Name: "Outfit", Roles: []GarmentRole{RoleOnePiece}, Categories: catsEthnicW, Formality: 0.95},
			slotShoes(0.85, "heels", "sandals"),
			slotJewellery(false),
			slotBag("Clutch", true, "clutch"),
		},
		Men: []SlotSpec{
			// A men's reception look is western-formal more often than ethnic.
			{Name: "Shirt", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.9},
			{Name: "Trouser", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.9},
			slotLayer("Blazer", 0.95),
			slotShoes(0.9, "loafers"),
			slotWatch(true),
		},
	},

	"party": {
		Key: "party", Formality: 0.6,
		Styles: []models.StyleTag{models.StyleWestern, models.StyleElegant, models.StyleLuxury},
		Women: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleOnePiece, RoleTop}, Categories: catsTopOrOne, Formality: 0.65},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.65},
			slotShoes(0.7, "heels", "sandals"),
			slotBag("Clutch", true, "clutch", "handbag"),
			slotJewellery(true),
		},
		Men: []SlotSpec{
			{Name: "Shirt", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.6},
			{Name: "Trouser", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.6},
			slotLayer("Jacket", 0.7),
			slotShoes(0.65, "loafers", "sneakers"),
			slotWatch(true),
		},
	},

	"travel": {
		Key: "travel", Formality: 0.25,
		Styles: []models.StyleTag{models.StyleCasual, models.StyleAthleisure, models.StyleResort},
		Women: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleTop, RoleOnePiece}, Categories: catsTopOrOne, Formality: 0.25},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.25},
			slotShoes(0.2, "sneakers", "flats"),
			slotBag("Backpack", true, "backpack", "handbag"),
			{Name: "Sunglasses", Roles: []GarmentRole{RoleEyewear}, Categories: catsAcc,
				Types: []string{"sunglasses"}, Optional: true, Formality: 0.3},
		},
		Men: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.25},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.25},
			slotShoes(0.2, "sneakers"),
			slotBag("Backpack", true, "backpack"),
			slotWatch(true),
		},
	},

	"date": {
		Key: "date", Formality: 0.55,
		Styles: []models.StyleTag{models.StyleSmartCasual, models.StyleElegant, models.StyleMinimal},
		Women: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleOnePiece, RoleTop}, Categories: catsTopOrOne, Formality: 0.6},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.6},
			slotShoes(0.6, "heels", "flats", "sandals"),
			slotBag("Bag", true, "clutch", "handbag"),
			slotJewellery(true),
		},
		Men: []SlotSpec{
			{Name: "Shirt", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.6},
			{Name: "Trouser", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.6},
			slotShoes(0.6, "loafers", "sneakers"),
			slotWatch(true),
		},
	},

	"festival": {
		Key: "festival", Formality: 0.8,
		Styles: []models.StyleTag{models.StyleFestive, models.StyleEthnic, models.StyleTraditional},
		Women: []SlotSpec{
			{Name: "Outfit", Roles: []GarmentRole{RoleOnePiece}, Categories: catsEthnicW, Formality: 0.85},
			slotShoes(0.7, "flats", "sandals", "heels"),
			slotJewellery(false),
			slotBag("Clutch", true, "clutch", "handbag"),
		},
		Men: []SlotSpec{
			{Name: "Outfit", Roles: []GarmentRole{RoleOnePiece, RoleTop}, Categories: catsEthnicM, Formality: 0.85},
			slotShoes(0.75, "loafers", "flats"),
			slotWatch(true),
		},
	},

	"brunch": {
		Key: "brunch", Formality: 0.45,
		Styles: []models.StyleTag{models.StyleSmartCasual, models.StyleMinimal, models.StyleFloral},
		Women: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleOnePiece, RoleTop}, Categories: catsTopOrOne, Formality: 0.45},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.45},
			slotShoes(0.4, "flats", "sandals", "sneakers"),
			slotBag("Bag", true, "handbag"),
			{Name: "Sunglasses", Roles: []GarmentRole{RoleEyewear}, Categories: catsAcc,
				Types: []string{"sunglasses"}, Optional: true, Formality: 0.4},
		},
		Men: []SlotSpec{
			{Name: "Shirt", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.45},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.45},
			slotShoes(0.45, "loafers", "sneakers"),
			slotWatch(true),
		},
	},

	"weekend": {
		Key: "weekend", Formality: 0.3,
		Styles: []models.StyleTag{models.StyleCasual, models.StyleRelaxedFit, models.StyleWestern},
		Women: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleTop, RoleOnePiece}, Categories: catsTopOrOne, Formality: 0.3},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.3},
			slotShoes(0.25, "sneakers", "sandals", "flats"),
			slotBag("Bag", true, "handbag"),
			slotWatch(true),
		},
		Men: []SlotSpec{
			{Name: "Top", Roles: []GarmentRole{RoleTop}, Categories: catsTop, Formality: 0.3},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: catsBottom, Formality: 0.3},
			slotShoes(0.25, "sneakers"),
			slotWatch(true),
			slotAccessory(true),
		},
	},
}

// occasionTemplateKey maps the catalogue's fine-grained occasion vocabulary onto
// the twelve briefs above. Anything unmapped falls back to leaning-based
// defaults, so a new occasion constant degrades gracefully instead of erroring.
func occasionTemplateKey(o models.Occasion) (string, bool) {
	switch o {
	case models.OccasionOffice, models.OccasionOfficeParty:
		return "office", true
	case models.OccasionBusinessMeeting:
		return "business-meeting", true
	case models.OccasionCasual, models.OccasionDailyWear:
		return "casual", true
	case models.OccasionCollege, models.OccasionSchool, models.OccasionFreshers,
		models.OccasionFarewell, models.OccasionGraduation:
		return "college", true
	case models.OccasionWedding, models.OccasionSangeet, models.OccasionMehendi,
		models.OccasionHaldi, models.OccasionEngagement, models.OccasionEngagementParty,
		models.OccasionWinterWedding:
		return "wedding", true
	case models.OccasionReception, models.OccasionWeddingReception:
		return "reception", true
	case models.OccasionParty, models.OccasionCocktailParty,
		models.OccasionBirthdayParty, models.OccasionNightOut:
		return "party", true
	case models.OccasionTravel, models.OccasionSummerVacation:
		return "travel", true
	case models.OccasionDateNight, models.OccasionDinnerDate, models.OccasionCafeDate:
		return "date", true
	case models.OccasionFestival, models.OccasionDiwali, models.OccasionNavratri,
		models.OccasionEid, models.OccasionPuja, models.OccasionDurgaPuja,
		models.OccasionGaneshChaturthi, models.OccasionKarvaChauth,
		models.OccasionTempleVisit:
		return "festival", true
	case models.OccasionBrunch, models.OccasionWeekendBrunch:
		return "brunch", true
	case models.OccasionWeekendOuting:
		return "weekend", true
	}
	return "", false
}

// defaultTemplateKey is used when the shopper named no occasion: an ethnic
// leaning gets the festival brief, a western one the casual brief.
func defaultTemplateKey(leaning outfitLeaning) string {
	if leaning == leaningEthnic {
		return "festival"
	}
	return "casual"
}

// resolveTemplate picks the slot plan for a department and intent.
//
// Single-category departments (Footwear, Accessories, Beauty) are not outfits —
// they are one-slot browses, and are returned as such so the existing endpoints
// keep behaving exactly as before.
func resolveTemplate(dept Department, base *Intent) ([]SlotSpec, OccasionTemplate) {
	switch dept {
	case DepartmentFootwear:
		return []SlotSpec{{Name: "Footwear", Roles: []GarmentRole{RoleFootwear}, Categories: catsShoes, Formality: 0.5}},
			OccasionTemplate{Key: "footwear", Formality: 0.5}
	case DepartmentAccessories:
		return []SlotSpec{{Name: "Accessories", Roles: []GarmentRole{
			RoleBag, RoleWatch, RoleJewellery, RoleBelt, RoleEyewear, RoleHeadwear, RoleScarf, RoleUnknown,
		}, Categories: catsAcc, Formality: 0.5}}, OccasionTemplate{Key: "accessories", Formality: 0.5}
	case DepartmentBeauty:
		return []SlotSpec{{Name: "Beauty", Roles: []GarmentRole{RoleBeauty, RoleUnknown},
				Categories: []models.CategoryName{models.CategoryBeauty}, Formality: 0.5}},
			OccasionTemplate{Key: "beauty", Formality: 0.5}
	case DepartmentKidsGirls, DepartmentKidsBoys:
		return []SlotSpec{
			{Name: "Outfit", Roles: []GarmentRole{RoleOnePiece, RoleTop}, Categories: []models.CategoryName{
				models.CategoryKidsGirls, models.CategoryKidsBoys, models.CategoryDresses, models.CategoryTops,
			}, Formality: 0.4},
			{Name: "Bottom", Roles: []GarmentRole{RoleBottom}, Categories: []models.CategoryName{
				models.CategoryKidsGirls, models.CategoryKidsBoys, models.CategoryJeans,
			}, Optional: true, Formality: 0.4},
			slotShoes(0.3, "sneakers", "sandals", "flats"),
			{Name: "Accessories", Roles: []GarmentRole{RoleBag, RoleHeadwear, RoleEyewear},
				Categories: catsAcc, Optional: true, Formality: 0.3},
		}, OccasionTemplate{Key: "kids", Formality: 0.4}
	}

	key := ""
	if base != nil && base.Occasion != nil {
		if k, ok := occasionTemplateKey(*base.Occasion); ok {
			key = k
		}
	}
	if key == "" {
		key = defaultTemplateKey(leaningFor(base))
	}

	tmpl, ok := occasionTemplates[key]
	if !ok {
		tmpl = occasionTemplates["casual"]
	}

	if dept == DepartmentMen {
		return tmpl.Men, tmpl
	}
	return tmpl.Women, tmpl
}
