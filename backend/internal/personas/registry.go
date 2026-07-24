package personas

import (
	_ "embed"
	"strings"
)

// seedJSON is the default roster of five launch stylists, embedded into the
// binary so a fresh clone always has something to publish even before the
// data/personas.json file exists. The file, once written, is authoritative; this
// is only the first-boot seed.
//
//go:embed seed_personas.json
var seedJSON []byte

// Slugify turns a display name into a URL-safe slug ("Shah Rukh Khan" →
// "shah-rukh-khan"). Non-alphanumeric runs collapse to a single hyphen.
func Slugify(name string) string {
	var b strings.Builder
	lastHyphen := false
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen && b.Len() > 0 {
				b.WriteRune('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// normalizeSlug lower-cases and trims a slug for lookups.
func normalizeSlug(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// normalize cleans a persona in place: trims strings, normalises the slug, and
// drops empty entries from the string-slice fields so downstream code never has
// to guard against blank preferences or templates.
func (p *Persona) normalize() {
	p.Slug = normalizeSlug(p.Slug)
	p.DisplayName = strings.TrimSpace(p.DisplayName)
	p.PartnerName = strings.TrimSpace(p.PartnerName)
	p.Tagline = strings.TrimSpace(p.Tagline)
	p.Greeting = strings.TrimSpace(p.Greeting)
	p.PortraitURL = strings.TrimSpace(p.PortraitURL)
	p.StyleSignature = strings.TrimSpace(p.StyleSignature)
	p.StyleOverride = strings.TrimSpace(p.StyleOverride)
	p.DefaultGender = strings.TrimSpace(p.DefaultGender)

	p.SamplePrompts = compact(p.SamplePrompts)
	p.PreferredStyles = compact(p.PreferredStyles)
	p.PreferredBrands = compact(p.PreferredBrands)
	p.PreferredColors = compact(p.PreferredColors)

	p.Voice.GreetingTemplates = compact(p.Voice.GreetingTemplates)
	p.Voice.Catchphrases = compact(p.Voice.Catchphrases)
	p.Voice.StyleRules = compact(p.Voice.StyleRules)
	p.Voice.Avoid = compact(p.Voice.Avoid)
	p.Voice.ExampleResponses = compact(p.Voice.ExampleResponses)
	p.Voice.ResponseTemplates = compact(p.Voice.ResponseTemplates)

	if p.UnlockAtItems < 0 {
		p.UnlockAtItems = 0
	}
}

// compact trims each string and drops the blanks.
func compact(in []string) []string {
	if len(in) == 0 {
		return in
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}
