package recommendation

import (
	"math"
	"strings"
	"unicode"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// Free-text relevance, scored with BM25 over the candidate pool.
//
// The attribute signals cover what the shopper said *structurally* — occasion,
// budget, colour. They cannot cover "chikankari", "anarkali", "wrap dress" or
// "camp collar", which live only in the title and description. Those are often
// the most specific thing in the prompt, so the engine needs a text signal that
// is properly weighted rather than a naive substring check.
//
// BM25 is the right tool: it saturates term frequency (a title repeating
// "saree" five times is not five times as relevant), normalises for document
// length (a short title should not be penalised against a long description),
// and weights rare terms above common ones via IDF ("chikankari" discriminates,
// "cotton" does not). Substring matching gets every one of those wrong.

const (
	// bm25K1 controls term-frequency saturation. 1.2 is the standard default
	// and is well suited to short documents like product copy.
	bm25K1 = 1.2
	// bm25B controls length normalisation. 0.75 is the standard default.
	bm25B = 0.75
	// titleBoost is the multiplier applied to term frequency in the title.
	// A term in the product name is a far stronger signal of aboutness than the
	// same term buried in marketing copy.
	titleBoost = 2.5
	// brandBoost is the multiplier for terms appearing in the brand name.
	brandBoost = 2.0
	// fuzzyThreshold is the maximum edit distance tolerated for a near-miss
	// token match, scaled by token length.
	fuzzyCredit = 0.6
)

// stopWords are terms with no discriminative power in this domain. "outfit",
// "look", "wear" and "need" appear in half the prompts and in a good share of
// product descriptions; leaving them in would let a wall of marketing copy
// out-rank a genuinely relevant product.
var stopWords = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "i": {}, "me": {}, "my": {}, "we": {}, "our": {},
	"you": {}, "your": {}, "is": {}, "am": {}, "are": {}, "was": {}, "be": {},
	"to": {}, "of": {}, "in": {}, "on": {}, "at": {}, "for": {}, "with": {},
	"and": {}, "or": {}, "but": {}, "if": {}, "then": {}, "than": {}, "that": {},
	"this": {}, "it": {}, "its": {}, "as": {}, "by": {}, "from": {}, "up": {},
	"out": {}, "so": {}, "some": {}, "any": {}, "can": {}, "will": {}, "would": {},
	"should": {}, "could": {}, "want": {}, "need": {}, "needs": {}, "looking": {},
	"look": {}, "find": {}, "show": {}, "suggest": {}, "recommend": {}, "give": {},
	"get": {}, "please": {}, "something": {}, "anything": {}, "outfit": {},
	"outfits": {}, "wear": {}, "buy": {}, "shop": {}, "under": {}, "below": {},
	"over": {}, "above": {}, "around": {}, "about": {}, "rs": {}, "inr": {},
	"rupees": {}, "budget": {}, "price": {}, "cost": {}, "good": {}, "nice": {},
	"best": {}, "great": {}, "help": {}, "hey": {}, "hi": {}, "hello": {},
}

// Tokenize splits text into lowercase alphanumeric tokens with stop words and
// single characters removed.
func Tokenize(s string) []string {
	fields := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if len(f) < 2 {
			continue
		}
		if _, stop := stopWords[f]; stop {
			continue
		}
		out = append(out, stem(f))
	}
	return out
}

// ExtractKeywords returns the deduplicated content terms of a prompt, preserving
// order of first appearance.
func ExtractKeywords(s string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, 8)
	for _, t := range Tokenize(s) {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

// stem applies a deliberately conservative suffix strip. A full Porter stemmer
// would be overkill and would happily mangle domain nouns — "sarees"/"saree"
// and "dresses"/"dress" are the cases that actually matter here, and an
// aggressive stemmer risks collapsing "kurta" and "kurti", which are different
// garments.
func stem(w string) string {
	switch {
	case len(w) > 4 && strings.HasSuffix(w, "ies"):
		return w[:len(w)-3] + "y"
	case len(w) > 4 && strings.HasSuffix(w, "sses"):
		return w[:len(w)-2]
	case len(w) > 4 && strings.HasSuffix(w, "ses"):
		return w[:len(w)-2]
	case len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") &&
		!strings.HasSuffix(w, "us") && !strings.HasSuffix(w, "is"):
		return w[:len(w)-1]
	default:
		return w
	}
}

// docTerms holds the pre-tokenized, boosted term frequencies of one product.
type docTerms struct {
	tf     map[string]float64
	length float64
}

// TextIndex is a BM25 index over a candidate pool, rebuilt per request.
//
// Rebuilding per request is a conscious trade. The pool is bounded (a few
// hundred products) so the build is microseconds, and it keeps the index
// perfectly consistent with the candidates actually being ranked — no cache
// invalidation, no staleness after a catalogue write. If the catalogue grew by
// two orders of magnitude this would move to a persistent index or Postgres
// full-text search behind the same interface.
type TextIndex struct {
	docs   map[string]*docTerms
	df     map[string]int
	n      float64
	avgLen float64
}

// NewTextIndex builds a BM25 index over the given products.
func NewTextIndex(products []*models.Product) *TextIndex {
	idx := &TextIndex{
		docs: make(map[string]*docTerms, len(products)),
		df:   make(map[string]int),
		n:    float64(len(products)),
	}
	var totalLen float64

	for _, p := range products {
		if p == nil {
			continue
		}
		d := &docTerms{tf: make(map[string]float64, 24)}

		addField(d, p.Title, titleBoost)
		addField(d, p.Brand, brandBoost)
		addField(d, p.Description, 1.0)
		addField(d, p.Color, 1.5)
		addField(d, string(p.Occasion), 1.0)
		addField(d, string(p.Style), 1.0)
		if p.Category != nil {
			addField(d, string(p.Category.Name), 1.5)
		}

		idx.docs[p.ID.String()] = d
		totalLen += d.length

		for term := range d.tf {
			idx.df[term]++
		}
	}

	if idx.n > 0 {
		idx.avgLen = totalLen / idx.n
	}
	if idx.avgLen == 0 {
		idx.avgLen = 1
	}
	return idx
}

func addField(d *docTerms, text string, boost float64) {
	for _, t := range Tokenize(text) {
		d.tf[t] += boost
		d.length++
	}
}

// idf returns the inverse document frequency of a term, using the BM25+ variant
// with a +1 inside the log so the value can never go negative for terms present
// in more than half the corpus.
func (idx *TextIndex) idf(term string) float64 {
	df := float64(idx.df[term])
	return math.Log(1 + (idx.n-df+0.5)/(df+0.5))
}

// Score returns the BM25 relevance of a product to the query terms, normalised
// to [0,1].
//
// Normalisation is against the query's own theoretical maximum rather than a
// fixed constant: BM25 scores are unbounded and depend on corpus statistics, so
// a fixed divisor would make the text signal's real weight vary by query and
// silently break the weighting scheme.
func (idx *TextIndex) Score(productID string, queryTerms []string) float64 {
	if len(queryTerms) == 0 {
		return 0
	}
	d, ok := idx.docs[productID]
	if !ok {
		return 0
	}

	var score, maxPossible float64
	for _, term := range queryTerms {
		idf := idx.idf(term)
		maxPossible += idf * (bm25K1 + 1)

		tf, hit := d.tf[term]
		if !hit {
			// Tolerate a typo or an unstemmed variant at reduced credit.
			// "wedidng" and "kurtas" should not score zero.
			if t, dist := idx.nearestTerm(d, term); dist > 0 {
				tf = t * fuzzyCredit
			}
		}
		if tf <= 0 {
			continue
		}

		norm := 1 - bm25B + bm25B*(d.length/idx.avgLen)
		score += idf * (tf * (bm25K1 + 1)) / (tf + bm25K1*norm)
	}

	if maxPossible <= 0 {
		return 0
	}
	return clamp01(score / maxPossible)
}

// nearestTerm looks for a token in the document within a small edit distance of
// term, returning its term frequency. Bounded by a length guard so this stays
// cheap on the pool sizes involved.
func (idx *TextIndex) nearestTerm(d *docTerms, term string) (float64, int) {
	if len(term) < 5 {
		// Short tokens are too easy to false-positive on: "red" is one edit
		// from "bed", "led" and "rec".
		return 0, 0
	}
	tolerance := 1
	if len(term) >= 8 {
		tolerance = 2
	}
	for docTerm, tf := range d.tf {
		if abs(len(docTerm)-len(term)) > tolerance {
			continue
		}
		if levenshteinWithin(docTerm, term, tolerance) {
			return tf, tolerance
		}
	}
	return 0, 0
}

// levenshteinWithin reports whether the edit distance between a and b is at
// most limit. It bails out early rather than computing the full distance, since
// the answer is only ever used as a boolean.
func levenshteinWithin(a, b string, limit int) bool {
	if a == b {
		return true
	}
	if abs(len(a)-len(b)) > limit {
		return false
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		rowMin := curr[0]
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = minInt(prev[j]+1, minInt(curr[j-1]+1, prev[j-1]+cost))
			if curr[j] < rowMin {
				rowMin = curr[j]
			}
		}
		if rowMin > limit {
			return false
		}
		prev, curr = curr, prev
	}
	return prev[len(b)] <= limit
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
