// Package personas is the source of truth for StarCloset's celebrity stylist
// personas. A persona is three layers of pure data:
//
//  1. Identity  — how the stylist presents to a shopper (name, portrait, accent
//     gradient, greeting, unlock threshold, sample prompts).
//  2. Style DNA — the signature the recommendation engine reads. It is merged
//     into the shopper's PersonalizationContext *before* ranking, so the engine
//     stays entirely rule-based and explainable — no LLM ever picks products.
//  3. Voice     — the deterministic template material the persona-voice module
//     uses to speak in the stylist's register. No external AI is involved.
//
// The package holds no I/O opinion beyond a small file-backed store, and imports
// nothing from the recommendation engine: personas are data, and the mapping
// from Style DNA onto engine inputs lives in the services layer (applyPersona).
package personas

// Status is a persona's publication state. Only published personas are offered
// to shoppers; drafts are visible only in the partner portal.
type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
)

// Voice is Layer 3 — everything the deterministic persona-voice generator needs
// to speak as this stylist. No field is ever sent to an external model.
type Voice struct {
	// VoiceTone is a short descriptor of the register ("warm and playful").
	VoiceTone string `json:"voiceTone"`
	// GreetingTemplates open a reply. One is chosen deterministically per prompt.
	GreetingTemplates []string `json:"greetingTemplates"`
	// Catchphrases are the stylist's signature sign-offs, appended occasionally.
	Catchphrases []string `json:"catchphrases"`
	// StyleRules are human-readable styling principles, surfaced in copy.
	StyleRules []string `json:"styleRules"`
	// Avoid lists what this persona steers away from (informational, for copy).
	Avoid []string `json:"avoid"`
	// ExampleResponses are illustrative replies shown in the partner preview.
	ExampleResponses []string `json:"exampleResponses"`
	// ResponseTemplates are the body of a reply, with {placeholders} the voice
	// module fills: {occasion} {count} {budget} {department} {highlight} {pieces}.
	ResponseTemplates []string `json:"responseTemplates"`
}

// Persona is a complete celebrity stylist across all three layers.
type Persona struct {
	// ── Layer 1 — Identity ──────────────────────────────────────────────────
	DisplayName   string   `json:"displayName"`
	Slug          string   `json:"slug"`
	PartnerName   string   `json:"partnerName"`
	Tagline       string   `json:"tagline"`
	Greeting      string   `json:"greeting"`
	PortraitURL   string   `json:"portraitUrl"`
	AccentFrom    string   `json:"accentFrom"`
	AccentTo      string   `json:"accentTo"`
	UnlockAtItems int      `json:"unlockAtItems"`
	SamplePrompts []string `json:"samplePrompts"`
	Status        Status   `json:"status"`

	// ── Layer 2 — Style DNA ─────────────────────────────────────────────────
	StyleSignature  string   `json:"styleSignature"`
	PreferredStyles []string `json:"preferredStyles"`
	PreferredBrands []string `json:"preferredBrands"`
	PreferredColors []string `json:"preferredColors"`
	// StyleOverride, when set and the shopper named no style, biases the engine's
	// style signal toward this persona's aesthetic. It is a soft nudge: an
	// explicit shopper style always wins.
	StyleOverride string `json:"styleOverride"`
	// DefaultGender scopes direct-chat requests when the shopper implied none.
	// Collection building takes gender from the chosen department instead.
	DefaultGender string `json:"defaultGender"`

	// ── Layer 3 — Voice ─────────────────────────────────────────────────────
	Voice Voice `json:"voice"`
}

// IsPublished reports whether the persona is shopper-visible.
func (p Persona) IsPublished() bool { return p.Status == StatusPublished }
