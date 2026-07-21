package models

import "strings"

// BuildSearchDoc concatenates and lowercases the parts of a product that are
// worth matching free text against. Kept in its own file because both the model
// hook and the seeder call it, and because the text ranker's notion of "what is
// searchable" must stay identical to what is indexed.
func BuildSearchDoc(parts ...string) string {
	var b strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i > 0 && b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strings.ToLower(p))
	}
	return b.String()
}
