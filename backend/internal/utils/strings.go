package utils

import (
	"strings"
	"unicode"
)

// Truncate shortens s to at most n runes, appending an ellipsis when it had to
// cut. Rune-aware because prompts routinely contain the ₹ sign and Devanagari,
// and slicing those by byte produces mojibake.
func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return strings.TrimRight(string(r[:n-1]), " ") + "…"
}

// Slugify converts a display name into a URL-safe slug ("Kurta Sets" ->
// "kurta-sets").
func Slugify(s string) string {
	var b strings.Builder
	lastDash := true
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// JoinWithAnd renders a list as readable prose: "a", "a and b", "a, b and c".
func JoinWithAnd(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
	}
}

// LowerFirst lowercases the first rune only, for splicing a label into the
// middle of a sentence.
func LowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
