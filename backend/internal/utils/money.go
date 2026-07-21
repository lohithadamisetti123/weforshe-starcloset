package utils

import (
	"strconv"
	"strings"
)

// FormatINR renders whole rupees in the Indian digit grouping system —
// ₹1,00,000 rather than ₹100,000.
//
// Go's stdlib has no locale-aware number formatting and golang.org/x/text
// would be a heavy dependency for one function, so this is hand-rolled. The
// rule: the last three digits are grouped together, then every two digits
// thereafter. This must match the frontend's
// `toLocaleString('en-IN')` output exactly, since both render the same value in
// the same screen.
func FormatINR(amount int) string {
	return "₹" + GroupINR(amount)
}

// GroupINR applies Indian digit grouping without the currency symbol.
func GroupINR(amount int) string {
	neg := amount < 0
	if neg {
		amount = -amount
	}
	s := strconv.Itoa(amount)

	var out string
	if len(s) <= 3 {
		out = s
	} else {
		last3 := s[len(s)-3:]
		rest := s[:len(s)-3]

		var groups []string
		for len(rest) > 2 {
			groups = append([]string{rest[len(rest)-2:]}, groups...)
			rest = rest[:len(rest)-2]
		}
		if len(rest) > 0 {
			groups = append([]string{rest}, groups...)
		}
		out = strings.Join(groups, ",") + "," + last3
	}

	if neg {
		return "-" + out
	}
	return out
}
