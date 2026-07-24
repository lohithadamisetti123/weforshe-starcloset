package personas

import "strings"

// StarCloset V2 organises the stylist experience into three independent shopping
// departments, each with its own ordered roster and its own unlock thresholds.
// The same stylist can appear in more than one department at a different
// threshold — Shah Rukh Khan opens both Women and Men, while Ranveer Singh
// unlocks at 4 items in Women but 2 in Men — so the unlock threshold lives on the
// roster entry, not on the persona. Personas remain a shared pool keyed by slug.
//
// This is configuration, not logic: adding a department or re-ordering a roster
// is an edit to the table below, and the recommendation engine is never touched.

// RosterEntry places a persona at a department-specific unlock threshold.
type RosterEntry struct {
	Slug          string
	UnlockAtItems int
}

// StarDepartment is one shopping journey.
type StarDepartment struct {
	Key   string // stable identity: "Women" | "Men" | "Kids"
	Label string
	Emoji string
	// CollectionsDepartment is the engine department Women/Men map to directly.
	// It is empty for Kids: the studio resolves "Kids - Girls" / "Kids - Boys"
	// from the chosen stylist's gender, so a Princess styles girls and a
	// Superhero styles boys — the engine's department support is unchanged.
	CollectionsDepartment string
	Stylists              []RosterEntry
}

// starDepartments is the V2 roster configuration.
var starDepartments = []StarDepartment{
	{
		Key: "Women", Label: "Women", Emoji: "👩", CollectionsDepartment: "Women",
		Stylists: []RosterEntry{
			{"alia-bhatt", 0},
			{"sonam-kapoor", 0},
			{"deepika-padukone", 2},
			{"kiara-advani", 4},
			{"priyanka-chopra", 6},
		},
	},
	{
		Key: "Men", Label: "Men", Emoji: "👨", CollectionsDepartment: "Men",
		Stylists: []RosterEntry{
			{"shahrukh-khan", 0},
			{"hrithik-roshan", 0},
			{"ranveer-singh", 2},
			{"vicky-kaushal", 4},
			{"virat-kohli", 6},
		},
	},
	{
		Key: "Kids", Label: "Kids", Emoji: "🧒", CollectionsDepartment: "",
		Stylists: []RosterEntry{
			{"princess-stylist", 0},
			{"superhero-stylist", 0},
			{"cartoon-stylist", 2},
			{"party-stylist", 4},
			{"playtime-stylist", 6},
		},
	},
}

// StarDepartments returns the department roster configuration.
func StarDepartments() []StarDepartment { return starDepartments }

// IsStarDepartment reports whether key names a real department (case-insensitive).
func IsStarDepartment(key string) bool {
	for _, d := range starDepartments {
		if strings.EqualFold(d.Key, key) {
			return true
		}
	}
	return false
}

// NormalizeDepartment returns the canonical department key for a case-insensitive
// input, or "" if unknown.
func NormalizeDepartment(key string) string {
	for _, d := range starDepartments {
		if strings.EqualFold(d.Key, key) {
			return d.Key
		}
	}
	return ""
}
