// Package ai turns a shopper's sentence into a structured recommendation.Intent.
//
// Two extractors live here and they are not redundant:
//
//   - RuleExtractor is deterministic, offline, and answers in microseconds.
//   - GeminiExtractor understands paraphrase, implication and context.
//
// The service runs both and merges the results (merge.go). The rules own the
// facts they can prove — a number in the prompt is a number, not a guess — and
// Gemini owns everything requiring judgement. Neither is a fallback for the
// other in the degraded sense; the rules are a full-strength extractor that
// also happens to keep the product working when the API key is missing, the
// quota is exhausted or Gemini is having an outage.
package ai

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
)

// Vocabulary carries the catalogue-specific terms the extractor matches
// against. It is injected rather than hardcoded so that adding a brand to the
// catalogue makes it searchable without a code change.
type Vocabulary struct {
	Brands []string
	Colors []string
}

// RuleExtractor implements deterministic intent extraction.
type RuleExtractor struct {
	vocab Vocabulary
}

// NewRuleExtractor returns an extractor bound to a catalogue vocabulary.
func NewRuleExtractor(v Vocabulary) *RuleExtractor {
	return &RuleExtractor{vocab: v}
}

// ---------------------------------------------------------------------------
// Lexicons
//
// The Hinglish entries are the point of this table, not decoration. "Shaadi",
// "sangeet", "mehendi", "haldi" and "roka" are how Indian shoppers describe an
// Indian wedding, and a fashion assistant called Bharat AI that only understands
// the English word "wedding" has failed at the one thing its name promises. An
// English-only lexicon would send every one of those prompts down the
// no-occasion path and return a generic shortlist.
// ---------------------------------------------------------------------------

var occasionLexicon = map[models.Occasion][]string{
	// Wedding cluster — every Hinglish ceremony word maps here first; the
	// engine then resolves to the exact sub-occasion via the longer alias table.
	models.OccasionWedding:          {"wedding", "shaadi", "shadi", "marriage", "bridal", "bride", "groom", "vivah", "phera"},
	models.OccasionWeddingReception: {"wedding reception", "reception party"},
	models.OccasionReception:        {"reception"},
	models.OccasionEngagement:       {"engagement", "roka", "sagai"},
	models.OccasionEngagementParty:  {"engagement party"},
	models.OccasionSangeet:          {"sangeet", "sangeet night"},
	models.OccasionMehendi:          {"mehendi", "mehandi", "mehndi"},
	models.OccasionHaldi:            {"haldi", "haldi ceremony"},
	models.OccasionWinterWedding:    {"winter wedding"},

	// Indian festivals
	models.OccasionFestival:        {"festival", "festive", "utsav"},
	models.OccasionDiwali:          {"diwali", "deepavali", "diwali party"},
	models.OccasionNavratri:        {"navratri", "garba", "dandiya"},
	models.OccasionKarvaChauth:     {"karva chauth", "karwa chauth"},
	models.OccasionEid:             {"eid", "eid ul fitr", "eid ul adha"},
	models.OccasionChristmas:       {"christmas", "xmas", "x-mas", "christmas party"},
	models.OccasionNewYear:         {"new year", "nye", "new years eve"},
	models.OccasionPongal:          {"pongal", "sankranti"},
	models.OccasionOnam:            {"onam"},
	models.OccasionDurgaPuja:       {"durga puja", "durga", "navami"},
	models.OccasionGaneshChaturthi: {"ganesh chaturthi", "ganesh", "ganpati"},
	models.OccasionPuja:            {"puja", "pooja", "havan"},
	models.OccasionTempleVisit:     {"temple", "temple visit", "mandir"},
	models.OccasionSpringFestival:  {"spring festival", "holi", "basant"},
	models.OccasionFestiveDinner:   {"festive dinner"},
	models.OccasionCulturalEvent:   {"cultural event", "cultural", "folk"},

	// Work
	models.OccasionOffice:          {"office", "work", "workwear", "boardroom", "workplace", "9 to 5"},
	models.OccasionOfficeParty:     {"office party", "work party"},
	models.OccasionBusinessMeeting: {"meeting", "business meeting", "client meeting", "presentation"},
	models.OccasionInterview:       {"interview", "job interview"},
	models.OccasionCorporateEvent:  {"corporate event", "corporate", "company event"},
	models.OccasionConference:      {"conference", "seminar", "summit"},

	// Education
	models.OccasionCollege:    {"college", "campus", "university", "class", "lecture", "hostel", "viva"},
	models.OccasionSchool:     {"school", "school function"},
	models.OccasionGraduation: {"graduation", "convocation", "degree ceremony"},
	models.OccasionFarewell:   {"farewell", "farewell party"},
	models.OccasionFreshers:   {"freshers", "fresher", "fresher party", "induction"},

	// Everyday
	models.OccasionCasual:         {"casual", "everyday", "regular", "day out", "chill"},
	models.OccasionDailyWear:      {"daily wear", "daily", "routine"},
	models.OccasionTravel:         {"travel", "vacation", "trip", "holiday", "flight", "airport", "getaway", "roadtrip", "road trip", "tourist", "sightseeing", "trek"},
	models.OccasionSummerVacation: {"summer vacation", "summer trip", "summer holiday"},
	models.OccasionWeekendOuting:  {"weekend outing", "weekend trip"},
	models.OccasionWeekendBrunch:  {"weekend brunch"},
	models.OccasionBrunch:         {"brunch", "sunday brunch"},
	models.OccasionPicnic:         {"picnic"},
	models.OccasionShopping:       {"shopping", "mall"},
	models.OccasionStaycation:     {"staycation"},
	models.OccasionWorkFromHome:   {"work from home", "wfh", "remote work"},
	models.OccasionLounge:         {"lounge", "loungewear", "lazy day", "errands"},
	models.OccasionBeach:          {"beach", "beachside", "seaside", "goa"},
	models.OccasionResort:         {"resort", "resort wear"},

	// Social
	models.OccasionParty:             {"party", "celebration", "dj", "club", "clubbing"},
	models.OccasionCocktailParty:     {"cocktail party", "cocktail"},
	models.OccasionBirthdayParty:     {"birthday party", "birthday", "bday"},
	models.OccasionDinnerDate:        {"dinner date", "dinner"},
	models.OccasionDateNight:         {"date night", "date", "romantic"},
	models.OccasionCafeDate:          {"cafe date", "cafe", "coffee date"},
	models.OccasionNightOut:          {"night out"},
	models.OccasionFamilyGathering:   {"family gathering", "family function", "family"},
	models.OccasionHousewarming:      {"housewarming", "griha pravesh"},
	models.OccasionBabyShower:        {"baby shower"},
	models.OccasionAnniversary:       {"anniversary"},
	models.OccasionAnniversaryDinner: {"anniversary dinner"},

	// Fitness
	models.OccasionGym:     {"gym", "workout", "training", "fitness"},
	models.OccasionSports:  {"sports", "sport", "match", "game"},
	models.OccasionRunning: {"running", "jogging", "marathon", "run"},
	models.OccasionYoga:    {"yoga", "pilates", "meditation"},

	// Kids
	models.OccasionKidsParty: {"kids party", "children party", "kiddie party"},
	models.OccasionPlaytime:  {"playtime", "play", "playground", "outdoor play"},

	// Self-care
	models.OccasionSelfCare: {"self care", "spa", "me time", "pamper"},
}

var styleLexicon = map[models.StyleTag][]string{
	models.StyleEthnic:             {"ethnic", "desi", "saree", "sari", "kurta", "kurti", "lehenga", "anarkali", "salwar", "churidar", "sherwani", "dhoti", "banarasi", "dupatta"},
	models.StyleWestern:            {"western", "dress", "gown", "skirt", "jeans", "denim"},
	models.StyleFormal:             {"formal", "professional", "blazer", "suit", "tailored", "sharp"},
	models.StyleStreetwear:         {"street", "streetwear", "urban", "hoodie", "sneaker", "baggy", "hypebeast", "drip"},
	models.StyleMinimal:            {"minimal", "simple", "clean", "understated", "basic", "plain", "quiet"},
	models.StyleBohemian:           {"boho", "bohemian", "flowy", "breezy", "hippie", "beachy"},
	models.StyleAthleisure:         {"athleisure", "active", "activewear", "track"},
	models.StyleIndoWestern:        {"indo-western", "indo western", "fusion", "indowestern", "nehru jacket", "bandhgala"},
	models.StyleTraditional:        {"traditional"},
	models.StyleContemporaryEthnic: {"contemporary ethnic"},
	models.StyleLuxuryEthnic:       {"luxury ethnic"},
	models.StyleFestive:            {"festive"},
	models.StyleElegant:            {"elegant", "graceful"},
	models.StyleCasual:             {"casual"},
	models.StyleSmartCasual:        {"smart casual"},
	models.StyleBusinessCasual:     {"business casual"},
	models.StyleOversized:          {"oversized", "baggy", "boxy"},
	models.StyleSlimFit:            {"slim fit", "slim"},
	models.StyleRegularFit:         {"regular fit", "regular"},
	models.StyleRelaxedFit:         {"relaxed fit", "relaxed"},
	models.StyleFloral:             {"floral", "floral print", "flower print"},
	models.StylePrinted:            {"printed", "print"},
	models.StyleGraphic:            {"graphic", "graphic tee", "graphic print"},
	models.StyleEmbroidered:        {"embroidered", "embroidery", "zari"},
	models.StyleChikankari:         {"chikankari", "chikan", "lucknowi chikankari"},
	models.StyleBandhani:           {"bandhani", "bandhej", "tie-dye"},
	models.StyleMirrorWork:         {"mirror work", "mirror", "shisha"},
	models.StyleGotaPatti:          {"gota patti", "gota"},
	models.StyleJacquard:           {"jacquard"},
	models.StyleLucknowi:           {"lucknowi"},
	models.StyleOrganza:            {"organza"},
	models.StyleSharara:            {"sharara"},
	models.StyleGharara:            {"gharara"},
	models.StyleAngrakha:           {"angrakha"},
	models.StyleChanderi:           {"chanderi"},
	models.StyleGeorgette:          {"georgette"},
	models.StyleChiffon:            {"chiffon"},
	models.StyleVelvet:             {"velvet"},
	models.StyleSilk:               {"silk"},
	models.StyleCotton:             {"cotton"},
	models.StyleLinen:              {"linen"},
	models.StyleDenim:              {"denim"},
	models.StyleLeather:            {"leather"},
	models.StyleSuede:              {"suede"},
	models.StyleVintage:            {"vintage", "retro"},
	models.StyleClassic:            {"classic"},
	models.StyleModern:             {"modern"},
	models.StyleLuxury:             {"luxury", "luxe"},
	models.StylePremium:            {"premium"},
	models.StyleMinimalist:         {"minimalist"},
	models.StyleResort:             {"resort wear", "resort"},
	models.StyleLounge:             {"lounge", "loungewear"},
	models.StyleSporty:             {"sporty", "sport", "sports", "gym", "workout", "running", "yoga", "jogger", "track"},
	models.StyleBoho:               {"boho"},
	models.StyleContemporary:       {"contemporary"},
	models.StyleTraditionalDecor:   {"traditional decor"},
	models.StyleIndustrial:         {"industrial"},
	models.StyleNatural:            {"natural", "organic"},
	models.StyleMatte:              {"matte"},
	models.StyleGlossy:             {"glossy", "shiny"},
	models.StyleScandinavian:       {"scandinavian", "scandi", "nordic"},
}

var genderLexicon = map[models.Gender][]string{
	models.GenderWomen: {
		"women", "woman", "female", "her", "she", "ladies", "lady",
		"wife", "girlfriend", "mom", "mother", "sister", "womens", "women's", "bride",
	},
	models.GenderMen: {
		"men", "man", "male", "him", "he", "guy", "guys",
		"gentleman", "husband", "boyfriend", "dad", "father", "brother",
		"mens", "men's", "groom",
	},
	models.GenderGirls: {
		"girl", "girls", "kids girl", "kids girls", "baby girl", "little girl",
		"daughter", "my daughter", "girl child",
	},
	models.GenderBoys: {
		"boy", "boys", "kids boy", "kids boys", "baby boy", "little boy",
		"son", "my son", "boy child",
	},
	models.GenderUnisex: {"unisex", "anyone", "any gender", "gender neutral", "kids"},
}

var categoryLexicon = map[models.CategoryName][]string{
	models.CategoryKurtaSets:      {"kurta set", "kurta sets", "kurta", "kurti", "anarkali", "salwar suit", "suit set"},
	models.CategorySarees:         {"saree", "sarees", "sari", "saris", "banarasi"},
	models.CategoryDresses:        {"dress", "dresses", "gown", "gowns", "frock", "midi", "maxi"},
	models.CategoryShirts:         {"shirt", "shirts", "formal shirt", "casual shirt"},
	models.CategoryTops:           {"top", "tops", "tee", "t-shirt", "tshirt", "tees", "blouse", "crop top", "skirt"},
	models.CategoryJeans:          {"jeans", "denim", "trouser", "trousers", "pants", "chinos", "joggers", "leggings"},
	models.CategoryBlazers:        {"blazer", "blazers", "jacket", "coat"},
	models.CategoryEthnicWear:     {"ethnic wear", "sherwani", "lehenga", "nehru jacket", "bandhgala", "dhoti"},
	models.CategoryFootwear:       {"footwear", "shoes", "shoe", "sneakers", "heels", "sandals", "loafers", "boots", "flats", "kolhapuris", "juttis", "mojaris"},
	models.CategoryAccessories:    {"accessories", "accessory", "watch", "watches", "bag", "belt", "jewellery", "jewelry", "earrings", "clutch", "necklace", "bracelet", "wallet", "handbag", "backpack", "sunglasses", "cap", "scarf"},
	models.CategoryBeauty:         {"beauty", "makeup", "lipstick", "foundation", "skincare", "serum", "moisturizer", "perfume", "fragrance", "shampoo", "conditioner"},
	models.CategoryKidsGirls:      {"kids girls", "girls clothing", "girls wear", "baby girl"},
	models.CategoryKidsBoys:       {"kids boys", "boys clothing", "boys wear", "baby boy"},
	models.CategoryWomen:          {"women", "women's"},
	models.CategoryMen:            {"men", "men's"},
}

// negationCues introduce an exclusion. Handling these is not a nicety: a
// shopper who writes "no black" and is shown black has been actively
// contradicted, which reads worse than showing nothing relevant at all.
var negationCues = []string{
	"not ", "no ", "without ", "avoid ", "except ", "other than ", "anything but ",
	"don't want ", "dont want ", "do not want ", "skip ", "nothing ", "besides ",
	"apart from ", "excluding ", "hate ",
}

var sortLexicon = map[recommendation.SortHint][]string{
	recommendation.SortPriceAsc:  {"cheapest", "cheap", "lowest price", "least expensive", "affordable", "budget friendly", "budget-friendly", "economical", "price low to high"},
	recommendation.SortPriceDesc: {"most expensive", "priciest", "premium", "luxury", "high end", "high-end", "price high to low", "splurge"},
	recommendation.SortRating:    {"best rated", "top rated", "highest rated", "best reviewed", "most loved", "popular", "bestseller", "best selling"},
	recommendation.SortDiscount:  {"biggest discount", "best deal", "best deals", "most discounted", "on sale", "sale", "highest discount", "best offer", "offers"},
	recommendation.SortNewest:    {"newest", "latest", "new arrivals", "just in", "recent"},
}

// ---------------------------------------------------------------------------
// Budget parsing
// ---------------------------------------------------------------------------

var (
	// rangeRe catches "between 2000 and 5000", "2000 to 5000", "₹2000-₹5000".
	rangeRe = regexp.MustCompile(`(?i)(?:between\s+)?(?:₹|rs\.?|inr\s*)?\s*([0-9][0-9,]*\.?[0-9]*)\s*(k|thousand|lakh|lakhs|l)?\s*(?:-|–|—|to|and)\s*(?:₹|rs\.?|inr\s*)?\s*([0-9][0-9,]*\.?[0-9]*)\s*(k|thousand|lakh|lakhs|l)?`)

	// upperRe catches "under ₹3000", "below 3k", "within 2500", "upto 4,000".
	upperRe = regexp.MustCompile(`(?i)(?:under|below|less\s+than|lesser\s+than|within|upto|up\s+to|max(?:imum)?|no\s+more\s+than|not\s+more\s+than|cheaper\s+than|around|about|approx(?:imately)?|near|~)\s*(?:₹|rs\.?|inr)?\s*([0-9][0-9,]*\.?[0-9]*)\s*(k|thousand|lakh|lakhs|l)?`)

	// lowerRe catches "above 2000", "at least 1500", "more than ₹999".
	lowerRe = regexp.MustCompile(`(?i)(?:above|over|more\s+than|greater\s+than|at\s+least|min(?:imum)?|starting\s+(?:at|from)|from)\s*(?:₹|rs\.?|inr)?\s*([0-9][0-9,]*\.?[0-9]*)\s*(k|thousand|lakh|lakhs|l)?`)

	// bareRe catches a currency-marked or standalone amount with no qualifier.
	bareRe = regexp.MustCompile(`(?i)(?:₹|rs\.?|inr)\s*([0-9][0-9,]*\.?[0-9]*)\s*(k|thousand|lakh|lakhs|l)?|\b([0-9][0-9,]{2,}|[0-9]+(?:\.[0-9]+)?\s*(?:k|thousand|lakh|lakhs))\b`)
)

// maskBudgetSpan blanks out whatever the budget regexes matched, replacing
// each character with a space rather than deleting it — so surrounding word
// positions don't shift and two words on either side of the match never get
// spliced together into a new accidental token.
//
// Checked in the same priority order ExtractBudget itself uses (range, then
// upper, then lower): a range phrase like "1500 and 3000" would otherwise
// also partially satisfy upperRe or lowerRe on its own, and masking the wrong
// span first could leave the other regex's connector word exposed.
func maskBudgetSpan(lower string) string {
	if loc := rangeRe.FindStringIndex(lower); loc != nil {
		return blankSpan(lower, loc)
	}
	if loc := upperRe.FindStringIndex(lower); loc != nil {
		return blankSpan(lower, loc)
	}
	if loc := lowerRe.FindStringIndex(lower); loc != nil {
		return blankSpan(lower, loc)
	}
	return lower
}

// blankSpan replaces s[loc[0]:loc[1]] with spaces.
func blankSpan(s string, loc []int) string {
	b := []byte(s)
	for i := loc[0]; i < loc[1]; i++ {
		b[i] = ' '
	}
	return string(b)
}

// qualitativeBudgets map vague price language onto concrete envelopes. The
// numbers are catalogue-relative: this catalogue runs roughly ₹550–₹4,700, so
// "cheap" meaning "under ₹1,500" is a real statement about this inventory
// rather than an abstract one.
var qualitativeBudgets = map[string]recommendation.BudgetRange{
	"cheap":           {Max: intPtr(1500)},
	"budget":          {Max: intPtr(1500)},
	"affordable":      {Max: intPtr(2000)},
	"inexpensive":     {Max: intPtr(1500)},
	"economical":      {Max: intPtr(1500)},
	"pocket friendly": {Max: intPtr(1500)},
	"low budget":      {Max: intPtr(1200)},
	"premium":         {Min: intPtr(3000)},
	"luxury":          {Min: intPtr(3500)},
	"high end":        {Min: intPtr(3500)},
	"splurge":         {Min: intPtr(3000)},
	"expensive":       {Min: intPtr(3000)},
}

func intPtr(v int) *int { return &v }

// parseAmount converts a captured number and optional multiplier into rupees.
func parseAmount(num, mult string) (int, bool) {
	clean := strings.ReplaceAll(strings.TrimSpace(num), ",", "")
	if clean == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(clean, 64)
	if err != nil || f <= 0 {
		return 0, false
	}

	switch strings.ToLower(strings.TrimSpace(mult)) {
	case "k", "thousand":
		f *= 1000
	case "lakh", "lakhs", "l":
		f *= 100000
	}

	// Reject values that cannot plausibly be a clothing budget. Without this,
	// "wedding outfit for my 25 year old sister" reads 25 as ₹25 and the whole
	// shortlist collapses to the cheapest item in the catalogue. Bounds beat
	// cleverness here.
	if f < 100 || f > 10000000 {
		return 0, false
	}
	return int(f + 0.5), true
}

// ExtractBudget parses a price envelope out of free text. Exported because the
// merge step trusts this over Gemini's arithmetic, and because it carries
// enough edge cases to deserve its own tests.
func ExtractBudget(prompt string) *recommendation.BudgetRange {
	lower := strings.ToLower(prompt)

	// Ranges are checked first: "between 2000 and 5000" contains a bare number
	// that the other patterns would happily misread in isolation.
	if m := rangeRe.FindStringSubmatch(lower); m != nil {
		lo, okLo := parseAmount(m[1], m[2])
		hi, okHi := parseAmount(m[3], m[4])
		if okLo && okHi {
			if lo > hi {
				lo, hi = hi, lo
			}
			return &recommendation.BudgetRange{Min: &lo, Max: &hi}
		}
	}

	br := &recommendation.BudgetRange{}
	if m := upperRe.FindStringSubmatch(lower); m != nil {
		if v, ok := parseAmount(m[1], m[2]); ok {
			br.Max = &v
		}
	}
	if m := lowerRe.FindStringSubmatch(lower); m != nil {
		if v, ok := parseAmount(m[1], m[2]); ok {
			br.Min = &v
		}
	}
	if !br.IsZero() {
		if br.HasMin() && br.HasMax() && *br.Min > *br.Max {
			br.Min, br.Max = br.Max, br.Min
		}
		return br
	}

	// A bare amount with no qualifier: "wedding outfit 3000" almost always
	// means a ceiling, which is also how the shipped frontend's mock read it.
	if m := bareRe.FindStringSubmatch(lower); m != nil {
		num, mult := m[1], m[2]
		if num == "" && m[3] != "" {
			// The alternate branch captures number and unit together.
			raw := strings.TrimSpace(m[3])
			for _, suffix := range []string{"lakhs", "lakh", "thousand", "k"} {
				if strings.HasSuffix(raw, suffix) {
					num = strings.TrimSpace(strings.TrimSuffix(raw, suffix))
					mult = suffix
					break
				}
			}
			if num == "" {
				num = raw
			}
		}
		if v, ok := parseAmount(num, mult); ok {
			return &recommendation.BudgetRange{Max: &v}
		}
	}

	// Qualitative language, only when no number was found.
	for term, envelope := range qualitativeBudgets {
		if strings.Contains(lower, term) {
			e := envelope
			return &e
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Extraction
// ---------------------------------------------------------------------------

// ExtractIntent implements Extractor. The context is accepted for interface
// conformance and ignored: rule extraction is pure CPU work measured in
// microseconds, so there is nothing meaningful to cancel.
func (r *RuleExtractor) ExtractIntent(_ context.Context, prompt string) (*recommendation.Intent, error) {
	return r.Extract(prompt), nil
}

// Name implements Extractor.
func (r *RuleExtractor) Name() string { return "rules" }

// Extract reads an Intent out of a prompt using rules alone.
func (r *RuleExtractor) Extract(prompt string) *recommendation.Intent {
	intent := recommendation.NewIntent(prompt)
	lower := " " + strings.ToLower(strings.TrimSpace(prompt)) + " "

	intent.Budget = ExtractBudget(prompt)

	if occ, ok := matchOccasion(lower); ok {
		intent.Occasion = &occ
	}
	if st, ok := matchStyle(lower); ok {
		intent.Style = &st
	}
	if g, ok := matchGender(lower); ok {
		intent.Gender = &g
	}
	if c, ok := matchCategory(lower); ok {
		intent.Category = &c
	}

	// Product type (watch, necklace, sneakers, saree…) narrows within a broad
	// category where the category alone can't tell sub-types apart.
	intent.TypeTerms = recommendation.ExtractTypeTerms(prompt)

	// Brand and colour matching run against a copy of the text with any
	// numeric-range phrase blanked out — see maskBudgetSpan. Without this,
	// "between 1500 and 3000" hands the standalone word "and" to the brand
	// matcher, which is indistinguishable from a shopper naming the real
	// catalogue brand "AND". Whole-word matching alone cannot catch this: the
	// word genuinely is a whole word here, just not a brand mention. Every
	// other extractor keeps reading the unmasked text — occasion, style,
	// category and the keywords field all have legitimate reasons to see the
	// numbers and connector words a budget phrase contains.
	brandColorText := maskBudgetSpan(lower)
	intent.Colors, intent.ExcludeColors = r.matchColors(brandColorText)
	intent.Brands, intent.ExcludeBrands = r.matchBrands(brandColorText)
	intent.ExcludeStyles = matchExcludedStyles(lower)

	// A style cannot be both wanted and excluded. The exclusion wins: it is the
	// more specific statement, and the positive match is usually an incidental
	// keyword hit from the same phrase ("not too formal").
	if intent.Style != nil {
		for _, ex := range intent.ExcludeStyles {
			if ex == *intent.Style {
				intent.Style = nil
				break
			}
		}
	}

	if s, ok := matchSort(lower); ok {
		intent.Sort = s
	}

	intent.Keywords = recommendation.ExtractKeywords(prompt)
	intent.Source = recommendation.SourceRules
	intent.Confidence = ruleConfidence(intent)

	return intent
}

// isNegated reports whether term appears inside a negation window in text.
//
// The window is deliberately short. "No black, I want a red saree for the
// wedding" must negate black without also negating "red", "saree" or
// "wedding" — a whole-sentence negation scope would swallow the entire brief.
func isNegated(text, term string) bool {
	idx := strings.Index(text, term)
	for idx >= 0 {
		start := idx - 24
		if start < 0 {
			start = 0
		}
		window := text[start:idx]
		for _, cue := range negationCues {
			if strings.Contains(window, cue) {
				return true
			}
		}
		next := strings.Index(text[idx+len(term):], term)
		if next < 0 {
			break
		}
		idx = idx + len(term) + next
	}
	return false
}

// matchLexicon finds the best non-negated hit across a lexicon.
//
// "Best" is the longest matching phrase, not the first. Map iteration in Go is
// randomised, so a first-match rule would make extraction non-deterministic;
// and longest-match is correct anyway, since "date night" must beat "date" and
// "indo western" must beat "western".
func matchLexicon[T comparable](text string, lex map[T][]string) (T, bool) {
	var best T
	bestLen := 0
	found := false

	for key, terms := range lex {
		for _, term := range terms {
			padded := term
			if !strings.Contains(text, padded) {
				continue
			}
			if isNegated(text, padded) {
				continue
			}
			if len(term) > bestLen {
				best, bestLen, found = key, len(term), true
			}
		}
	}
	return best, found
}

func matchOccasion(text string) (models.Occasion, bool) {
	return matchLexicon(text, occasionLexicon)
}

func matchStyle(text string) (models.StyleTag, bool) {
	return matchLexicon(text, styleLexicon)
}

func matchCategory(text string) (models.CategoryName, bool) {
	return matchLexicon(text, categoryLexicon)
}

// matchGender resolves the audience, resisting a subtle trap: possessive
// pronouns and relationship nouns describe who the clothes are *for*, which is
// usually not the speaker. "A gift for my wife" is Women; "my husband's
// birthday" is Men. Matching on the longest term handles this correctly since
// the relationship noun is longer than the stray pronoun.
func matchGender(text string) (models.Gender, bool) {
	return matchLexicon(text, genderLexicon)
}

func matchExcludedStyles(text string) []models.StyleTag {
	var out []models.StyleTag
	for style, terms := range styleLexicon {
		for _, term := range terms {
			if strings.Contains(text, term) && isNegated(text, term) {
				out = append(out, style)
				break
			}
		}
	}
	return out
}

func matchSort(text string) (recommendation.SortHint, bool) {
	hint, ok := matchLexicon(text, sortLexicon)
	return hint, ok
}

// matchColors splits colour mentions into wanted and excluded.
func (r *RuleExtractor) matchColors(text string) (wanted, excluded []string) {
	seen := make(map[string]struct{})

	consider := func(term string) {
		if term == "" || len(term) < 3 {
			return
		}
		if !strings.Contains(text, term) {
			return
		}
		if _, dup := seen[term]; dup {
			return
		}
		seen[term] = struct{}{}
		if isNegated(text, term) {
			excluded = append(excluded, term)
		} else {
			wanted = append(wanted, term)
		}
	}

	// Family words first ("red", "blue"), then the catalogue's own colour names
	// ("Bottle Green"), which are what a shopper echoing the site would type.
	for family := range colorFamilyTerms() {
		consider(family)
	}
	for _, c := range r.vocab.Colors {
		consider(strings.ToLower(c))
	}
	return wanted, excluded
}

// matchBrands splits brand mentions into wanted and excluded.
func (r *RuleExtractor) matchBrands(text string) (wanted, excluded []string) {
	strippedText := stripPunct(text)

	for _, b := range r.vocab.Brands {
		term := strings.ToLower(strings.TrimSpace(b))
		if term == "" {
			continue
		}
		strippedTerm := stripPunct(term)

		// Single-word brand names are matched as whole words only, never as
		// a substring. Several of the catalogue's real brands are also
		// common English words: "AND" collides with the conjunction in
		// "between 1500 and 3000", "Only" with the adverb, "W" with the
		// letter in "with"/"wedding". A substring match on any of these is a
		// false positive waiting to happen, and it isn't a rare edge case —
		// "and" in particular shows up in most budget-range phrasings.
		//
		// Multi-word brands ("Van Heusen", "Jack & Jones") don't carry this
		// risk: the odds of one appearing as an accidental substring inside
		// unrelated text are effectively zero, so substring matching there
		// stays both safe and more forgiving of how a shopper actually types.
		var matched bool
		if strings.Contains(strippedTerm, " ") {
			matched = strings.Contains(text, term) || strings.Contains(strippedText, strippedTerm)
		} else {
			matched = containsWord(strippedText, strippedTerm)
		}
		if !matched {
			continue
		}

		if isNegated(text, term) {
			excluded = append(excluded, b)
		} else {
			wanted = append(wanted, b)
		}
	}
	return wanted, excluded
}

func containsWord(text, word string) bool {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return !('a' <= r && r <= 'z') && !('0' <= r && r <= '9')
	})
	for _, f := range fields {
		if f == word {
			return true
		}
	}
	return false
}

func stripPunct(s string) string {
	var b strings.Builder
	for _, r := range s {
		if ('a' <= r && r <= 'z') || ('0' <= r && r <= '9') || r == ' ' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// colorFamilyTerms exposes the recommendation package's colour family keys.
func colorFamilyTerms() map[string]struct{} {
	families := []string{
		"red", "maroon", "pink", "blue", "navy", "green", "teal", "black",
		"white", "grey", "gray", "brown", "beige", "nude", "khaki", "gold",
		"silver", "yellow", "orange", "purple", "pastel",
	}
	out := make(map[string]struct{}, len(families))
	for _, f := range families {
		out[f] = struct{}{}
	}
	return out
}

// ruleConfidence estimates how much of the prompt the rules actually explained.
//
// This is a self-assessment, and it needs to be a sober one: a confidently
// wrong extractor is worse than a hesitant one, because the ranker amplifies
// whatever it is handed. Confidence rises with the number of constraints
// recognised and falls when a long prompt yielded little.
func ruleConfidence(i *recommendation.Intent) float64 {
	specified := i.SpecifiedCount()
	conf := 0.35 + 0.14*float64(specified)

	words := len(strings.Fields(i.Query))
	if words > 12 && specified <= 1 {
		// A long sentence the rules barely understood is precisely where an
		// LLM earns its place, so say so honestly.
		conf -= 0.15
	}
	if words <= 3 && specified >= 1 {
		// A short, on-vocabulary prompt ("festival wear") is about as
		// unambiguous as this gets.
		conf += 0.1
	}

	if conf < 0.2 {
		conf = 0.2
	}
	if conf > 0.9 {
		// The rules never claim certainty: they cannot know what they missed.
		conf = 0.9
	}
	return conf
}
