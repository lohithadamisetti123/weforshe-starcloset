package dto

import (
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/personas"
)

// This file is the wire contract for the StarCloset V2 department journey: the
// department rosters (GET /starcloset/departments) and per-department unlock
// progress (GET/POST /starcloset/progress). It mirrors the frontend
// src/types/starcloset.ts.

// StarStylist is one stylist card within a department roster. UnlockAtItems is
// the DEPARTMENT-specific threshold (from the roster entry), not the persona's
// own field — the same stylist can unlock at different points in different
// departments.
type StarStylist struct {
	Slug           string   `json:"slug"`
	DisplayName    string   `json:"displayName"`
	Tagline        string   `json:"tagline"`
	Greeting       string   `json:"greeting"`
	PortraitURL    string   `json:"portraitUrl"`
	AccentFrom     string   `json:"accentFrom"`
	AccentTo       string   `json:"accentTo"`
	UnlockAtItems  int      `json:"unlockAtItems"`
	SamplePrompts  []string `json:"samplePrompts"`
	StyleSignature string   `json:"styleSignature"`
	DefaultGender  string   `json:"defaultGender"`
}

// StarDepartmentResponse is one shopping journey with its resolved roster.
type StarDepartmentResponse struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Emoji string `json:"emoji"`
	// CollectionsDepartment is the engine department for Women/Men; empty for
	// Kids, where the studio derives Kids - Girls / Kids - Boys from the chosen
	// stylist's gender.
	CollectionsDepartment string        `json:"collectionsDepartment"`
	Stylists              []StarStylist `json:"stylists"`
}

// ProgressResponse carries a shopper's unlock counter for every department. All
// departments are always present, defaulting to 0, so the client never has to
// distinguish "no row yet" from "zero progress".
type ProgressResponse struct {
	Progress map[string]int `json:"progress"`
}

// IncrementProgressRequest advances one department's counter. Delta defaults to
// 1 (a single cart-add) when omitted or non-positive.
type IncrementProgressRequest struct {
	Department string `json:"department" binding:"required" validate:"required"`
	Delta      int    `json:"delta,omitempty"`
}

// NewStarDepartments resolves every department's roster against the persona
// store, dropping any slug that is missing or unpublished so the client only
// ever receives stylists it can actually use.
func NewStarDepartments(store personas.Store) []StarDepartmentResponse {
	depts := personas.StarDepartments()
	out := make([]StarDepartmentResponse, 0, len(depts))
	for _, d := range depts {
		stylists := make([]StarStylist, 0, len(d.Stylists))
		for _, entry := range d.Stylists {
			p, ok := store.Get(entry.Slug)
			if !ok || !p.IsPublished() {
				continue
			}
			stylists = append(stylists, StarStylist{
				Slug:           p.Slug,
				DisplayName:    p.DisplayName,
				Tagline:        p.Tagline,
				Greeting:       p.Greeting,
				PortraitURL:    p.PortraitURL,
				AccentFrom:     p.AccentFrom,
				AccentTo:       p.AccentTo,
				UnlockAtItems:  entry.UnlockAtItems,
				SamplePrompts:  nonNil(p.SamplePrompts),
				StyleSignature: p.StyleSignature,
				DefaultGender:  p.DefaultGender,
			})
		}
		out = append(out, StarDepartmentResponse{
			Key:                   d.Key,
			Label:                 d.Label,
			Emoji:                 d.Emoji,
			CollectionsDepartment: d.CollectionsDepartment,
			Stylists:              stylists,
		})
	}
	return out
}

// NewProgressMap projects the stored rows onto a complete department→items map,
// seeding every known department at 0 first so the response is exhaustive.
func NewProgressMap(rows []*models.StarProgress) map[string]int {
	m := make(map[string]int, len(personas.StarDepartments()))
	for _, d := range personas.StarDepartments() {
		m[d.Key] = 0
	}
	for _, r := range rows {
		m[r.Department] = r.Items
	}
	return m
}
