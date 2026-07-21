package models

import "strings"

// This file is the SINGLE SOURCE OF TRUTH for the catalogue vocabulary.
// Every constant below must stay byte-identical to the corresponding member
// of the TypeScript union in src/types/index.ts and the CHECK constraints in
// the migrations. They are the same string travelling three different paths.

// ---------------------------------------------------------------------------
// Gender
// ---------------------------------------------------------------------------

type Gender string

const (
	GenderWomen  Gender = "Women"
	GenderMen    Gender = "Men"
	GenderGirls  Gender = "Girls"
	GenderBoys   Gender = "Boys"
	GenderUnisex Gender = "Unisex"
)

var AllGenders = []Gender{
	GenderWomen, GenderMen, GenderGirls, GenderBoys, GenderUnisex,
}

// ---------------------------------------------------------------------------
// Occasion
// ---------------------------------------------------------------------------

type Occasion string

const (
	// Wedding & ceremonies
	OccasionWedding          Occasion = "Wedding"
	OccasionWeddingReception Occasion = "Wedding Reception"
	OccasionReception        Occasion = "Reception"
	OccasionEngagement       Occasion = "Engagement"
	OccasionEngagementParty  Occasion = "Engagement Party"
	OccasionSangeet          Occasion = "Sangeet"
	OccasionMehendi          Occasion = "Mehendi"
	OccasionHaldi            Occasion = "Haldi"
	OccasionWinterWedding    Occasion = "Winter Wedding"

	// Indian festivals
	OccasionFestival        Occasion = "Festival"
	OccasionDiwali          Occasion = "Diwali"
	OccasionNavratri        Occasion = "Navratri"
	OccasionKarvaChauth     Occasion = "Karva Chauth"
	OccasionEid             Occasion = "Eid"
	OccasionChristmas       Occasion = "Christmas"
	OccasionNewYear         Occasion = "New Year"
	OccasionPongal          Occasion = "Pongal"
	OccasionOnam            Occasion = "Onam"
	OccasionDurgaPuja       Occasion = "Durga Puja"
	OccasionGaneshChaturthi Occasion = "Ganesh Chaturthi"
	OccasionPuja            Occasion = "Puja"
	OccasionTempleVisit     Occasion = "Temple Visit"
	OccasionSpringFestival  Occasion = "Spring Festival"
	OccasionFestiveDinner   Occasion = "Festive Dinner"
	OccasionCulturalEvent   Occasion = "Cultural Event"

	// Work & professional
	OccasionOffice          Occasion = "Office"
	OccasionOfficeParty     Occasion = "Office Party"
	OccasionBusinessMeeting Occasion = "Business Meeting"
	OccasionInterview       Occasion = "Interview"
	OccasionCorporateEvent  Occasion = "Corporate Event"
	OccasionConference      Occasion = "Conference"

	// Education
	OccasionCollege    Occasion = "College"
	OccasionSchool     Occasion = "School"
	OccasionGraduation Occasion = "Graduation"
	OccasionFarewell   Occasion = "Farewell"
	OccasionFreshers   Occasion = "Freshers"

	// Everyday
	OccasionCasual         Occasion = "Casual"
	OccasionDailyWear      Occasion = "Daily Wear"
	OccasionTravel         Occasion = "Travel"
	OccasionSummerVacation Occasion = "Summer Vacation"
	OccasionWeekendOuting  Occasion = "Weekend Outing"
	OccasionWeekendBrunch  Occasion = "Weekend Brunch"
	OccasionBrunch         Occasion = "Brunch"
	OccasionPicnic         Occasion = "Picnic"
	OccasionShopping       Occasion = "Shopping"
	OccasionStaycation     Occasion = "Staycation"
	OccasionWorkFromHome   Occasion = "Work From Home"
	OccasionLounge         Occasion = "Lounge"
	OccasionBeach          Occasion = "Beach"
	OccasionResort         Occasion = "Resort"

	// Social & celebrations
	OccasionParty             Occasion = "Party"
	OccasionCocktailParty     Occasion = "Cocktail Party"
	OccasionBirthdayParty     Occasion = "Birthday Party"
	OccasionDinnerDate        Occasion = "Dinner Date"
	OccasionDateNight         Occasion = "Date Night"
	OccasionCafeDate          Occasion = "Cafe Date"
	OccasionNightOut          Occasion = "Night Out"
	OccasionFamilyGathering   Occasion = "Family Gathering"
	OccasionHousewarming      Occasion = "Housewarming"
	OccasionBabyShower        Occasion = "Baby Shower"
	OccasionAnniversary       Occasion = "Anniversary"
	OccasionAnniversaryDinner Occasion = "Anniversary Dinner"

	// Fitness & sport
	OccasionGym     Occasion = "Gym"
	OccasionSports  Occasion = "Sports"
	OccasionRunning Occasion = "Running"
	OccasionYoga    Occasion = "Yoga"

	// Kids-specific
	OccasionKidsParty Occasion = "Kids Party"
	OccasionPlaytime  Occasion = "Playtime"

	// Self-care & home
	OccasionSelfCare Occasion = "Self Care"
)

var AllOccasions = []Occasion{
	OccasionWedding, OccasionWeddingReception, OccasionReception,
	OccasionEngagement, OccasionEngagementParty,
	OccasionSangeet, OccasionMehendi, OccasionHaldi, OccasionWinterWedding,
	OccasionFestival, OccasionDiwali, OccasionNavratri, OccasionKarvaChauth,
	OccasionEid, OccasionChristmas, OccasionNewYear, OccasionPongal, OccasionOnam,
	OccasionDurgaPuja, OccasionGaneshChaturthi, OccasionPuja, OccasionTempleVisit,
	OccasionSpringFestival, OccasionFestiveDinner, OccasionCulturalEvent,
	OccasionOffice, OccasionOfficeParty, OccasionBusinessMeeting,
	OccasionInterview, OccasionCorporateEvent, OccasionConference,
	OccasionCollege, OccasionSchool, OccasionGraduation, OccasionFarewell, OccasionFreshers,
	OccasionCasual, OccasionDailyWear, OccasionTravel, OccasionSummerVacation,
	OccasionWeekendOuting, OccasionWeekendBrunch, OccasionBrunch,
	OccasionPicnic, OccasionShopping, OccasionStaycation,
	OccasionWorkFromHome, OccasionLounge, OccasionBeach, OccasionResort,
	OccasionParty, OccasionCocktailParty, OccasionBirthdayParty,
	OccasionDinnerDate, OccasionDateNight, OccasionCafeDate, OccasionNightOut,
	OccasionFamilyGathering, OccasionHousewarming, OccasionBabyShower,
	OccasionAnniversary, OccasionAnniversaryDinner,
	OccasionGym, OccasionSports, OccasionRunning, OccasionYoga,
	OccasionKidsParty, OccasionPlaytime,
	OccasionSelfCare,
}

// ---------------------------------------------------------------------------
// StyleTag
// ---------------------------------------------------------------------------

type StyleTag string

const (
	// Classic base styles
	StyleEthnic      StyleTag = "Ethnic"
	StyleWestern     StyleTag = "Western"
	StyleFormal      StyleTag = "Formal"
	StyleStreetwear  StyleTag = "Streetwear"
	StyleMinimal     StyleTag = "Minimal"
	StyleBohemian    StyleTag = "Bohemian"
	StyleAthleisure  StyleTag = "Athleisure"
	StyleIndoWestern StyleTag = "Indo-Western"

	// Indian ethnic extensions
	StyleTraditional        StyleTag = "Traditional"
	StyleContemporaryEthnic StyleTag = "Contemporary Ethnic"
	StyleLuxuryEthnic       StyleTag = "Luxury Ethnic"
	StyleFestive            StyleTag = "Festive"
	StyleElegant            StyleTag = "Elegant"
	StyleCasual             StyleTag = "Casual"

	// Fit / silhouette
	StyleSmartCasual    StyleTag = "Smart Casual"
	StyleBusinessCasual StyleTag = "Business Casual"
	StyleOversized      StyleTag = "Oversized"
	StyleSlimFit        StyleTag = "Slim Fit"
	StyleRegularFit     StyleTag = "Regular Fit"
	StyleRelaxedFit     StyleTag = "Relaxed Fit"

	// Print & embellishment
	StyleFloral      StyleTag = "Floral"
	StylePrinted     StyleTag = "Printed"
	StyleGraphic     StyleTag = "Graphic"
	StyleEmbroidered StyleTag = "Embroidered"
	StyleChikankari  StyleTag = "Chikankari"
	StyleBandhani    StyleTag = "Bandhani"
	StyleMirrorWork  StyleTag = "Mirror Work"
	StyleGotaPatti   StyleTag = "Gota Patti"

	// Weave / fabric style
	StyleJacquard  StyleTag = "Jacquard"
	StyleLucknowi  StyleTag = "Lucknowi"
	StyleOrganza   StyleTag = "Organza"
	StyleSharara   StyleTag = "Sharara"
	StyleGharara   StyleTag = "Gharara"
	StyleAngrakha  StyleTag = "Angrakha"
	StyleChanderi  StyleTag = "Chanderi"
	StyleGeorgette StyleTag = "Georgette"
	StyleChiffon   StyleTag = "Chiffon"
	StyleVelvet    StyleTag = "Velvet"
	StyleSilk      StyleTag = "Silk"
	StyleCotton    StyleTag = "Cotton"
	StyleLinen     StyleTag = "Linen"
	StyleDenim     StyleTag = "Denim"
	StyleLeather   StyleTag = "Leather"
	StyleSuede     StyleTag = "Suede"

	// Aesthetic registers
	StyleVintage      StyleTag = "Vintage"
	StyleClassic      StyleTag = "Classic"
	StyleModern       StyleTag = "Modern"
	StyleLuxury       StyleTag = "Luxury"
	StylePremium      StyleTag = "Premium"
	StyleMinimalist   StyleTag = "Minimalist"
	StyleResort       StyleTag = "Resort"
	StyleLounge       StyleTag = "Lounge"
	StyleSporty       StyleTag = "Sporty"
	StyleBoho         StyleTag = "Boho"
	StyleContemporary StyleTag = "Contemporary"

	// Home & decor aesthetics
	StyleTraditionalDecor StyleTag = "Traditional Decor"
	StyleIndustrial       StyleTag = "Industrial"
	StyleNatural          StyleTag = "Natural"
	StyleMatte            StyleTag = "Matte"
	StyleGlossy           StyleTag = "Glossy"
	StyleScandinavian     StyleTag = "Scandinavian"
)

// AllStyles is the controlled vocabulary used for CHECK constraints and
// the normalised dataset. Only these 7 values should appear in the DB.
// The additional constants above remain for affinity-matrix keys and
// alias lookups — they are never stored as style values themselves.
var AllStyles = []StyleTag{
	StyleEthnic,
	StyleWestern,
	StyleFormal,
	StyleCasual,
	StyleIndoWestern,
	StyleAthleisure,
}

// ---------------------------------------------------------------------------
// CategoryName
// ---------------------------------------------------------------------------

type CategoryName string

const (
	// Apparel
	CategoryKurtaSets  CategoryName = "Kurta Sets"
	CategorySarees     CategoryName = "Sarees"
	CategoryDresses    CategoryName = "Dresses"
	CategoryShirts     CategoryName = "Shirts"
	CategoryTops       CategoryName = "Tops"
	CategoryJeans      CategoryName = "Jeans"
	CategoryBlazers    CategoryName = "Blazers"
	CategoryEthnicWear CategoryName = "Ethnic Wear"

	// Departments / verticals
	CategoryWomen          CategoryName = "Women"
	CategoryMen            CategoryName = "Men"
	CategoryKidsGirls      CategoryName = "Kids Girls"
	CategoryKidsBoys       CategoryName = "Kids Boys"
	CategoryFootwear       CategoryName = "Footwear"
	CategoryAccessories    CategoryName = "Accessories"
	CategoryBeauty         CategoryName = "Beauty"
	CategoryHomeEssentials CategoryName = "Home Essentials"
)

var AllCategories = []CategoryName{
	CategoryKurtaSets, CategorySarees, CategoryDresses, CategoryShirts,
	CategoryTops, CategoryJeans, CategoryBlazers, CategoryEthnicWear,
	CategoryWomen, CategoryMen, CategoryKidsGirls, CategoryKidsBoys,
	CategoryFootwear, CategoryAccessories, CategoryBeauty, CategoryHomeEssentials,
}

// Apparel-only categories — used by the collection builder to scope slot
// filling to garment categories rather than department-level ones.
var AllApparelCategories = []CategoryName{
	CategoryKurtaSets, CategorySarees, CategoryDresses, CategoryShirts,
	CategoryTops, CategoryJeans, CategoryBlazers, CategoryEthnicWear,
}

// ---------------------------------------------------------------------------
// Role
// ---------------------------------------------------------------------------

type Role string

const (
	RoleUser  Role = "User"
	RoleAdmin Role = "Admin"
)

var AllRoles = []Role{RoleUser, RoleAdmin}

// ---------------------------------------------------------------------------
// HistoryStatus
// ---------------------------------------------------------------------------

type HistoryStatus string

const (
	HistoryStatusCompleted HistoryStatus = "Completed"
	HistoryStatusSaved     HistoryStatus = "Saved"
	HistoryStatusPurchased HistoryStatus = "Purchased"
)

var AllHistoryStatuses = []HistoryStatus{
	HistoryStatusCompleted, HistoryStatusSaved, HistoryStatusPurchased,
}

// ---------------------------------------------------------------------------
// Case-insensitive parsing — every parse function uses the same index built
// in init() so there is exactly one place to add an alias.
// ---------------------------------------------------------------------------

func normalizeKey(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

var (
	genderIndex   = map[string]Gender{}
	occasionIndex = map[string]Occasion{}
	styleIndex    = map[string]StyleTag{}
	categoryIndex = map[string]CategoryName{}
	roleIndex     = map[string]Role{}
	statusIndex   = map[string]HistoryStatus{}
)

func init() {
	for _, g := range AllGenders {
		genderIndex[normalizeKey(string(g))] = g
	}
	for _, o := range AllOccasions {
		occasionIndex[normalizeKey(string(o))] = o
	}
	for _, s := range AllStyles {
		styleIndex[normalizeKey(string(s))] = s
	}
	for _, c := range AllCategories {
		categoryIndex[normalizeKey(string(c))] = c
	}
	for _, r := range AllRoles {
		roleIndex[normalizeKey(string(r))] = r
	}
	for _, s := range AllHistoryStatuses {
		statusIndex[normalizeKey(string(s))] = s
	}

	// ── Style aliases ──────────────────────────────────────────────────────
	// These map every synonym and fabric name the extractor might encounter
	// to one of the 7 canonical style values.
	styleIndex["traditional"] = StyleEthnic
	styleIndex["ethnic traditional"] = StyleEthnic
	styleIndex["contemporary ethnic"] = StyleEthnic
	styleIndex["luxury ethnic"] = StyleEthnic
	styleIndex["festive"] = StyleEthnic
	styleIndex["chikankari"] = StyleEthnic
	styleIndex["bandhani"] = StyleEthnic
	styleIndex["mirror work"] = StyleEthnic
	styleIndex["gota patti"] = StyleEthnic
	styleIndex["embroidered"] = StyleEthnic
	styleIndex["jacquard"] = StyleEthnic
	styleIndex["organza"] = StyleEthnic
	styleIndex["chanderi"] = StyleEthnic
	styleIndex["georgette"] = StyleEthnic
	styleIndex["sharara"] = StyleEthnic
	styleIndex["gharara"] = StyleEthnic
	styleIndex["angrakha"] = StyleEthnic
	styleIndex["lucknowi"] = StyleEthnic
	styleIndex["anarkali"] = StyleEthnic
	styleIndex["elegant"] = StyleFormal
	styleIndex["classic"] = StyleFormal
	styleIndex["luxury"] = StyleFormal
	styleIndex["premium"] = StyleFormal
	styleIndex["velvet"] = StyleFormal
	styleIndex["silk"] = StyleFormal
	styleIndex["business casual"] = StyleFormal
	styleIndex["slim fit"] = StyleFormal
	styleIndex["smart casual"] = StyleCasual
	styleIndex["minimal"] = StyleCasual
	styleIndex["minimalist"] = StyleCasual
	styleIndex["streetwear"] = StyleCasual
	styleIndex["bohemian"] = StyleCasual
	styleIndex["boho"] = StyleCasual
	styleIndex["floral"] = StyleCasual
	styleIndex["printed"] = StyleCasual
	styleIndex["graphic"] = StyleCasual
	styleIndex["resort"] = StyleCasual
	styleIndex["linen"] = StyleCasual
	styleIndex["cotton"] = StyleCasual
	styleIndex["denim"] = StyleCasual
	styleIndex["chiffon"] = StyleCasual
	styleIndex["suede"] = StyleCasual
	styleIndex["leather"] = StyleCasual
	styleIndex["vintage"] = StyleCasual
	styleIndex["modern"] = StyleCasual
	styleIndex["lounge"] = StyleCasual
	styleIndex["contemporary"] = StyleCasual
	styleIndex["oversized"] = StyleCasual
	styleIndex["regular fit"] = StyleCasual
	styleIndex["relaxed fit"] = StyleCasual
	styleIndex["relaxed"] = StyleCasual
	styleIndex["sporty"] = StyleAthleisure
	styleIndex["sport"] = StyleAthleisure
	styleIndex["athletic"] = StyleAthleisure
	styleIndex["gym"] = StyleAthleisure
	styleIndex["party"] = StyleFormal
	styleIndex["indo western"] = StyleIndoWestern
	styleIndex["indo-western"] = StyleIndoWestern

	// ── Occasion aliases ────────────────────────────────────────────────────
	// Hinglish
	occasionIndex["shaadi"] = OccasionWedding
	occasionIndex["shadi"] = OccasionWedding
	occasionIndex["vivah"] = OccasionWedding
	occasionIndex["nikah"] = OccasionWedding
	occasionIndex["roka"] = OccasionEngagement
	occasionIndex["sangeet night"] = OccasionSangeet
	occasionIndex["mehandi"] = OccasionMehendi
	occasionIndex["haldi ceremony"] = OccasionHaldi
	occasionIndex["garba"] = OccasionNavratri
	occasionIndex["dandiya"] = OccasionNavratri
	occasionIndex["karwa chauth"] = OccasionKarvaChauth
	occasionIndex["deepavali"] = OccasionDiwali
	occasionIndex["ganesh puja"] = OccasionGaneshChaturthi
	occasionIndex["durga"] = OccasionDurgaPuja
	// English shorthands
	occasionIndex["date-night"] = OccasionDateNight
	occasionIndex["datenight"] = OccasionDateNight
	occasionIndex["dinner"] = OccasionDinnerDate
	occasionIndex["birthday"] = OccasionBirthdayParty
	occasionIndex["nye"] = OccasionNewYear
	occasionIndex["x-mas"] = OccasionChristmas
	occasionIndex["xmas"] = OccasionChristmas
	occasionIndex["wfh"] = OccasionWorkFromHome
	occasionIndex["work from home"] = OccasionWorkFromHome
	occasionIndex["cocktail"] = OccasionCocktailParty
	occasionIndex["brunch"] = OccasionBrunch
	occasionIndex["freshers"] = OccasionFreshers
	occasionIndex["fresher"] = OccasionFreshers
	occasionIndex["convocation"] = OccasionGraduation
	occasionIndex["sports"] = OccasionSports
	occasionIndex["running"] = OccasionRunning
	occasionIndex["yoga"] = OccasionYoga
	occasionIndex["gym"] = OccasionGym
	occasionIndex["lounge"] = OccasionLounge
	occasionIndex["beach"] = OccasionBeach
	occasionIndex["resort"] = OccasionResort
	occasionIndex["staycation"] = OccasionStaycation
	occasionIndex["self care"] = OccasionSelfCare
	occasionIndex["spa"] = OccasionSelfCare
	occasionIndex["kids party"] = OccasionKidsParty
	occasionIndex["playtime"] = OccasionPlaytime

	// ── Style aliases ────────────────────────────────────────────────────────
	styleIndex["indo western"] = StyleIndoWestern
	styleIndex["indowestern"] = StyleIndoWestern
	styleIndex["fusion"] = StyleIndoWestern
	styleIndex["traditional"] = StyleTraditional
	styleIndex["chikan"] = StyleChikankari
	styleIndex["chikankari work"] = StyleChikankari
	styleIndex["mirror"] = StyleMirrorWork
	styleIndex["gota"] = StyleGotaPatti
	styleIndex["bandhej"] = StyleBandhani
	styleIndex["printed"] = StylePrinted
	styleIndex["floral print"] = StyleFloral
	styleIndex["embroidered"] = StyleEmbroidered
	styleIndex["zari"] = StyleEmbroidered
	styleIndex["georgette saree"] = StyleGeorgette
	styleIndex["chiffon saree"] = StyleChiffon
	styleIndex["velvet"] = StyleVelvet
	styleIndex["contemporary"] = StyleContemporary
	styleIndex["luxury"] = StyleLuxury
	styleIndex["premium"] = StylePremium
	styleIndex["festive"] = StyleFestive
	styleIndex["elegant"] = StyleElegant
	styleIndex["smart casual"] = StyleSmartCasual
	styleIndex["business casual"] = StyleBusinessCasual
	styleIndex["oversized"] = StyleOversized
	styleIndex["slim fit"] = StyleSlimFit
	styleIndex["regular fit"] = StyleRegularFit
	styleIndex["relaxed fit"] = StyleRelaxedFit
	styleIndex["boho"] = StyleBoho
	styleIndex["bohemian"] = StyleBohemian
	styleIndex["sporty"] = StyleSporty
	styleIndex["lounge"] = StyleLounge
	styleIndex["resort"] = StyleResort
	styleIndex["minimalist"] = StyleMinimalist
	styleIndex["vintage"] = StyleVintage
	styleIndex["classic"] = StyleClassic
	styleIndex["modern"] = StyleModern
	styleIndex["denim"] = StyleDenim
	styleIndex["leather"] = StyleLeather
	styleIndex["suede"] = StyleSuede
	styleIndex["silk"] = StyleSilk
	styleIndex["cotton"] = StyleCotton
	styleIndex["linen"] = StyleLinen

	// ── Gender aliases ───────────────────────────────────────────────────────
	genderIndex["woman"] = GenderWomen
	genderIndex["female"] = GenderWomen
	genderIndex["ladies"] = GenderWomen
	genderIndex["lady"] = GenderWomen
	genderIndex["her"] = GenderWomen
	genderIndex["man"] = GenderMen
	genderIndex["male"] = GenderMen
	genderIndex["gentleman"] = GenderMen
	genderIndex["him"] = GenderMen
	genderIndex["girl"] = GenderGirls
	genderIndex["kids girls"] = GenderGirls
	genderIndex["girl's"] = GenderGirls
	genderIndex["boy"] = GenderBoys
	genderIndex["kids boys"] = GenderBoys
	genderIndex["boy's"] = GenderBoys
	genderIndex["kids"] = GenderGirls // default kids to girls, extractor corrects with context

	// ── Category aliases ─────────────────────────────────────────────────────
	categoryIndex["saree"] = CategorySarees
	categoryIndex["sari"] = CategorySarees
	categoryIndex["kurta set"] = CategoryKurtaSets
	categoryIndex["kurta"] = CategoryKurtaSets
	categoryIndex["kurti"] = CategoryKurtaSets
	categoryIndex["anarkali"] = CategoryKurtaSets
	categoryIndex["salwar suit"] = CategoryKurtaSets
	categoryIndex["ethnic"] = CategoryEthnicWear
	categoryIndex["lehenga"] = CategoryEthnicWear
	categoryIndex["sherwani"] = CategoryEthnicWear
	categoryIndex["shoes"] = CategoryFootwear
	categoryIndex["sneakers"] = CategoryFootwear
	categoryIndex["heels"] = CategoryFootwear
	categoryIndex["sandals"] = CategoryFootwear
	categoryIndex["footwear"] = CategoryFootwear
	categoryIndex["dress"] = CategoryDresses
	categoryIndex["shirt"] = CategoryShirts
	categoryIndex["top"] = CategoryTops
	categoryIndex["blazer"] = CategoryBlazers
	categoryIndex["jewellery"] = CategoryAccessories
	categoryIndex["jewelry"] = CategoryAccessories
	categoryIndex["bag"] = CategoryAccessories
	categoryIndex["watch"] = CategoryAccessories
	categoryIndex["beauty"] = CategoryBeauty
	categoryIndex["makeup"] = CategoryBeauty
	categoryIndex["skincare"] = CategoryBeauty
	categoryIndex["home"] = CategoryHomeEssentials
	categoryIndex["home essentials"] = CategoryHomeEssentials
	categoryIndex["decor"] = CategoryHomeEssentials
	categoryIndex["kids girls"] = CategoryKidsGirls
	categoryIndex["kids boys"] = CategoryKidsBoys
	categoryIndex["girls"] = CategoryKidsGirls
	categoryIndex["boys"] = CategoryKidsBoys
}

func ParseGender(s string) (Gender, bool) {
	v, ok := genderIndex[normalizeKey(s)]
	return v, ok
}

func ParseOccasion(s string) (Occasion, bool) {
	v, ok := occasionIndex[normalizeKey(s)]
	return v, ok
}

func ParseStyle(s string) (StyleTag, bool) {
	v, ok := styleIndex[normalizeKey(s)]
	return v, ok
}

func ParseCategory(s string) (CategoryName, bool) {
	v, ok := categoryIndex[normalizeKey(s)]
	return v, ok
}

func ParseRole(s string) (Role, bool) {
	v, ok := roleIndex[normalizeKey(s)]
	return v, ok
}

func ParseHistoryStatus(s string) (HistoryStatus, bool) {
	v, ok := statusIndex[normalizeKey(s)]
	return v, ok
}

// IsKidsGender returns true when the gender indicates a child's product.
func IsKidsGender(g Gender) bool {
	return g == GenderGirls || g == GenderBoys
}

// GenderToDisplayDept returns the department-level display label for a gender,
// used by the guided assistant flow's first-step category cards.
func GenderToDisplayDept(g Gender) string {
	switch g {
	case GenderWomen:
		return "Women"
	case GenderMen:
		return "Men"
	case GenderGirls:
		return "Kids Girls"
	case GenderBoys:
		return "Kids Boys"
	default:
		return string(g)
	}
}
