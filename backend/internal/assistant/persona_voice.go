// Package assistant holds the deterministic persona-voice generator.
//
// StarCloset's stylist replies are assembled from a persona's own template
// material — greetings, response templates, catchphrases — with a few
// placeholders filled from the collection the rule-based engine produced. There
// is NO call to any external model here: the same input always yields the same
// reply, which is what makes the stylist voice testable and free of the failure
// modes (latency, hallucination, cost) an LLM would introduce.
//
// The engine decides *what* to recommend; this module decides *how the stylist
// says it*. Those concerns never mix — a persona can never conjure a product the
// engine did not rank.
package assistant

import (
	"hash/fnv"
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/personas"
)

// VoiceInput is the collection context the voice module narrates. Every field is
// optional; the generator degrades to sensible copy when something is missing.
type VoiceInput struct {
	Prompt        string   // the shopper's brief — also the determinism seed
	OccasionLabel string   // e.g. "sangeet", "everyday"
	BudgetLabel   string   // e.g. "₹8,000"
	Department    string   // e.g. "Women"
	ItemCount     int      // number of pieces styled across the collections
	Highlight     string   // headline product title
	Pieces        []string // short "Slot: Brand" descriptions, best first
}

// Reply composes a stylist reply in the persona's voice. It returns "" when no
// persona is supplied, so callers can treat an absent persona as "no reply" and
// fall back to their existing generic copy.
func Reply(p *personas.Persona, in VoiceInput) string {
	if p == nil {
		return ""
	}
	seed := seedOf(in.Prompt)

	var parts []string
	if g := pick(p.Voice.GreetingTemplates, seed); g != "" {
		parts = append(parts, ensureSentence(g))
	} else if p.Greeting != "" {
		parts = append(parts, ensureSentence(p.Greeting))
	}

	if b := body(p, in, seed); b != "" {
		parts = append(parts, b)
	}

	// A catchphrase lands on roughly half of replies (seed-driven, deterministic)
	// so the voice has personality without every reply ending the same way.
	if cp := pick(p.Voice.Catchphrases, seed>>3); cp != "" && seed%2 == 0 {
		parts = append(parts, ensureSentence(cp))
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

// body fills a response template, or builds a graceful default when the persona
// declares none.
func body(p *personas.Persona, in VoiceInput, seed uint32) string {
	tmpl := pick(p.Voice.ResponseTemplates, seed)
	if tmpl == "" {
		return defaultBody(in)
	}
	return fill(tmpl, in)
}

// fill substitutes {placeholders} in a template and cleans up the result.
func fill(tmpl string, in VoiceInput) string {
	r := strings.NewReplacer(
		"{occasion}", fallback(in.OccasionLabel, "your occasion"),
		"{count}", countWord(in.ItemCount),
		"{budget}", fallback(in.BudgetLabel, "your budget"),
		"{department}", fallback(in.Department, "your wardrobe"),
		"{highlight}", fallback(in.Highlight, "lead piece"),
		"{pieces}", strings.Join(in.Pieces, ", "),
	)
	return tidy(r.Replace(tmpl))
}

// defaultBody is used when a persona has no response templates of its own.
func defaultBody(in VoiceInput) string {
	occ := fallback(in.OccasionLabel, "you")
	b := strings.Builder{}
	b.WriteString("I've put together ")
	b.WriteString(countWord(in.ItemCount))
	b.WriteString(" looks for ")
	b.WriteString(occ)
	if in.Highlight != "" {
		b.WriteString(" — starting with the ")
		b.WriteString(in.Highlight)
	}
	if in.BudgetLabel != "" {
		b.WriteString(", all within ")
		b.WriteString(in.BudgetLabel)
	}
	b.WriteString(".")
	return tidy(b.String())
}

// ── helpers ──────────────────────────────────────────────────────────────────

func seedOf(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(s))))
	return h.Sum32()
}

// pick chooses one entry deterministically from seed. Returns "" for an empty
// slice.
func pick(list []string, seed uint32) string {
	if len(list) == 0 {
		return ""
	}
	return list[int(seed)%len(list)]
}

func countWord(n int) string {
	switch {
	case n <= 0:
		return "a few"
	case n == 1:
		return "one"
	default:
		return itoa(n)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func fallback(s, alt string) string {
	if strings.TrimSpace(s) == "" {
		return alt
	}
	return s
}

// ensureSentence guarantees terminal punctuation so joined fragments read well.
func ensureSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// A trailing ellipsis or dash is a deliberate lead-in (greetings often end
	// "— ") and must flow into the next fragment, not gain a full stop.
	if strings.HasSuffix(s, "…") || strings.HasSuffix(s, "—") {
		return s
	}
	switch s[len(s)-1] {
	case '.', '!', '?', ':', '-':
		return s
	}
	return s + "."
}

// tidy collapses doubled spaces and fixes spacing left by an empty placeholder.
func tidy(s string) string {
	s = strings.ReplaceAll(s, " ,", ",")
	s = strings.ReplaceAll(s, "  ", " ")
	s = strings.ReplaceAll(s, " .", ".")
	return strings.TrimSpace(s)
}
