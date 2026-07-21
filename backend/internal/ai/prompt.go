package ai

import (
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/myntra/bharat-ai-backend/internal/models"
)

// The prompt is part of the system's contract, so it lives in version control
// next to the code that depends on it rather than in a config blob somewhere.
//
// Three rules shape it:
//
//  1. The model is an *extractor*, not a stylist. It never picks products,
//     never writes copy, and never sees the catalogue. Its only job is turning
//     a sentence into fields. Ranking stays in code, where it is testable and
//     explainable; a model that recommended products directly could not be
//     unit-tested, audited by a merchandiser, or made to justify itself.
//  2. It must answer with the exact enum members the catalogue uses, because
//     anything else silently matches nothing.
//  3. It must say "null" rather than guess. An invented occasion is worse than
//     a missing one — the ranker acts on what it is given, so a confident
//     hallucination becomes six wrong products, whereas a null gracefully
//     falls back to quality-led ranking.

const systemInstruction = `You are the natural-language understanding layer of Myntra Bharat AI, a fashion shopping assistant for Indian shoppers.

Your ONLY job is to convert a shopper's message into a structured JSON object describing what they are asking for. You do not recommend products. You do not write prose. You do not chat.

CRITICAL RULES:
1. Use ONLY the exact enum values listed below. Never invent new values. Never reword them.
2. If the shopper did not express something, return null for it. Do NOT guess or infer beyond what is reasonably implied. A null is always better than a wrong value.
3. Budget is in Indian Rupees as a plain integer (no symbols, no commas, no decimals). "3k" is 3000. "2 lakh" is 200000.
4. "under X" / "below X" / "within X" sets budget.max only.
   "above X" / "at least X" sets budget.min only.
   "between X and Y" sets both.
   A bare amount with no qualifier (e.g. "wedding outfit 3000") sets budget.max.
5. Distinguish who the clothes are FOR, not who is speaking. "gift for my wife" is Women. "something for my husband" is Men.
6. Handle Indian context natively: shaadi/sangeet/mehendi/haldi/baraat/roka are Wedding. Diwali/Navratri/Garba/Puja/Onam/Pongal/Eid/Holi/Karwa Chauth are Festival. Kurta/saree/lehenga/anarkali/sherwani are Ethnic.
7. Negations go in the exclude* arrays, never in the positive fields. "not black" means excludeColors:["black"], NOT colors:["black"].
8. Colours may be everyday words ("red", "pastel") — do not force them to catalogue names.
9. confidence reflects how sure you are that you read the shopper correctly, from 0.0 to 1.0. Be honest: a vague message deserves a low score.

ENUMS:
occasion: "Wedding" | "Office" | "College" | "Party" | "Travel" | "Festival" | "Casual" | "Date Night" | null
style: "Ethnic" | "Western" | "Formal" | "Streetwear" | "Minimal" | "Bohemian" | "Athleisure" | "Indo-Western" | null
gender: "Women" | "Men" | "Unisex" | null
category: "Kurta Sets" | "Sarees" | "Dresses" | "Shirts" | "Tops" | "Jeans" | "Blazers" | "Ethnic Wear" | "Footwear" | "Accessories" | null
sort: "relevance" | "price_asc" | "price_desc" | "rating" | "discount" | "newest"

Respond with a single JSON object and nothing else.`

// fewShotExamples teach the boundaries that prose instructions consistently
// fail to hold on their own: the difference between speaker and recipient, that
// a bare number is a ceiling, that negation must not leak into positive fields,
// and — most importantly — that returning nulls for a vague prompt is the
// correct answer rather than a failure.
const fewShotExamples = `EXAMPLES:

Message: "I need a wedding outfit under ₹3000"
{"budget":{"min":null,"max":3000},"occasion":"Wedding","style":"Ethnic","gender":null,"category":null,"colors":[],"brands":[],"excludeColors":[],"excludeStyles":[],"excludeBrands":[],"sort":"relevance","confidence":0.9,"reasoning":"Explicit occasion and ceiling; wedding implies ethnic in Indian context."}

Message: "something for my husband's office, nothing formal, budget 2500"
{"budget":{"min":null,"max":2500},"occasion":"Office","style":null,"gender":"Men","category":null,"colors":[],"brands":[],"excludeColors":[],"excludeStyles":["Formal"],"excludeBrands":[],"sort":"relevance","confidence":0.85,"reasoning":"Recipient is the husband, so Men. 'nothing formal' is an exclusion, not a style."}

Message: "sangeet ke liye kuch achha dikha do, red ya maroon, 5k tak"
{"budget":{"min":null,"max":5000},"occasion":"Wedding","style":"Ethnic","gender":null,"category":null,"colors":["red","maroon"],"brands":[],"excludeColors":[],"excludeStyles":[],"excludeBrands":[],"sort":"relevance","confidence":0.85,"reasoning":"Sangeet is a wedding function; 5k tak is a 5000 ceiling."}

Message: "cheapest running joggers, not Nike"
{"budget":null,"occasion":"Casual","style":"Athleisure","gender":null,"category":"Jeans","colors":[],"brands":[],"excludeColors":[],"excludeStyles":[],"excludeBrands":["Nike"],"sort":"price_asc","confidence":0.8,"reasoning":"'cheapest' is an explicit sort; brand exclusion captured."}

Message: "show me something nice"
{"budget":null,"occasion":null,"style":null,"gender":null,"category":null,"colors":[],"brands":[],"excludeColors":[],"excludeStyles":[],"excludeBrands":[],"sort":"relevance","confidence":0.2,"reasoning":"No extractable constraints; guessing would be worse than admitting this."}

Message: "diwali kurta set for mom between 2000 and 4000"
{"budget":{"min":2000,"max":4000},"occasion":"Festival","style":"Ethnic","gender":"Women","category":"Kurta Sets","colors":[],"brands":[],"excludeColors":[],"excludeStyles":[],"excludeBrands":[],"sort":"relevance","confidence":0.95,"reasoning":"Diwali is Festival; mom is Women; explicit range."}`

// buildUserPrompt wraps the shopper's message with the examples.
//
// The examples ride on the user turn rather than the system instruction on
// purpose: models weight recent context more heavily than a distant system
// preamble, and this task's failure modes are all pattern-matching failures
// that the examples directly correct.
func buildUserPrompt(prompt string) string {
	var b strings.Builder
	b.WriteString(fewShotExamples)
	b.WriteString("\n\nNow extract from this message.\n\nMessage: ")
	b.WriteString(fmt.Sprintf("%q", strings.TrimSpace(prompt)))
	b.WriteString("\n")
	return b.String()
}

// intentSchema is the response schema handed to the API.
//
// This is belt-and-braces alongside the prompt: the schema makes the JSON shape
// a hard guarantee at the API level rather than a request the model may ignore,
// and the enum constraints stop whole classes of hallucination before the
// response is even generated. The prompt still restates them because the
// schema alone gives the model no idea *why* those are the options.
func intentSchema() *genai.Schema {
	str := func(desc string, enum ...string) *genai.Schema {
		s := &genai.Schema{Type: genai.TypeString, Description: desc, Nullable: true}
		if len(enum) > 0 {
			s.Enum = enum
		}
		return s
	}
	strArray := func(desc string) *genai.Schema {
		return &genai.Schema{
			Type:        genai.TypeArray,
			Description: desc,
			Items:       &genai.Schema{Type: genai.TypeString},
		}
	}

	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"budget": {
				Type:        genai.TypeObject,
				Description: "Price envelope in whole Indian Rupees. Null when not mentioned.",
				Nullable:    true,
				Properties: map[string]*genai.Schema{
					"min": {Type: genai.TypeInteger, Description: "Lower bound in rupees, or null.", Nullable: true},
					"max": {Type: genai.TypeInteger, Description: "Upper bound in rupees, or null.", Nullable: true},
				},
			},
			"occasion": str("The event the shopper is dressing for.",
				string(models.OccasionWedding), string(models.OccasionOffice),
				string(models.OccasionCollege), string(models.OccasionParty),
				string(models.OccasionTravel), string(models.OccasionFestival),
				string(models.OccasionCasual), string(models.OccasionDateNight)),
			"style": str("The aesthetic register requested.",
				string(models.StyleEthnic), string(models.StyleWestern),
				string(models.StyleFormal), string(models.StyleStreetwear),
				string(models.StyleMinimal), string(models.StyleBohemian),
				string(models.StyleAthleisure), string(models.StyleIndoWestern)),
			"gender": str("Who the clothes are for, not who is asking.",
				string(models.GenderWomen), string(models.GenderMen), string(models.GenderUnisex)),
			"category": str("The garment family requested.",
				string(models.CategoryKurtaSets), string(models.CategorySarees),
				string(models.CategoryDresses), string(models.CategoryShirts),
				string(models.CategoryTops), string(models.CategoryJeans),
				string(models.CategoryBlazers), string(models.CategoryEthnicWear),
				string(models.CategoryFootwear), string(models.CategoryAccessories)),
			"colors":        strArray("Colours the shopper wants, in everyday words."),
			"brands":        strArray("Brands the shopper explicitly named."),
			"excludeColors": strArray("Colours the shopper ruled out."),
			"excludeStyles": strArray("Styles the shopper ruled out, using the style enum."),
			"excludeBrands": strArray("Brands the shopper ruled out."),
			"sort": str("An explicit ordering the shopper asked for.",
				"relevance", "price_asc", "price_desc", "rating", "discount", "newest"),
			"confidence": {
				Type:        genai.TypeNumber,
				Description: "How confident you are in this reading, 0.0 to 1.0. Be honest.",
			},
			"reasoning": {
				Type:        genai.TypeString,
				Description: "One short sentence explaining the extraction. For debugging only.",
			},
		},
		Required: []string{"occasion", "style", "gender", "category", "confidence"},
	}
}
