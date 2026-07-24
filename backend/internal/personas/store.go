package personas

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Store is the persona repository. It is deliberately narrow: the assistant and
// the partner portal talk to this interface, never to the file underneath.
type Store interface {
	// All returns every persona (drafts included), unlock-order first.
	All() []Persona
	// Published returns only shopper-visible personas, unlock-order first.
	Published() []Persona
	// Get resolves one persona by slug.
	Get(slug string) (Persona, bool)
	// Create adds a new persona. It fails if the slug already exists.
	Create(p Persona) (Persona, error)
	// Update replaces the mutable fields of an existing persona.
	Update(slug string, p Persona) (Persona, error)
	// Publish flips a persona's status to published.
	Publish(slug string) (Persona, error)
}

// ErrSlugTaken is returned when creating a persona whose slug already exists.
var ErrSlugTaken = errors.New("personas: slug already exists")

// ErrNotFound is returned when no persona matches a slug.
var ErrNotFound = errors.New("personas: persona not found")

// fileStore is a JSON-file-backed, in-memory, concurrency-safe Store.
//
// On construction it loads backend/data/personas.json; if that file is absent it
// seeds from the embedded default and writes it, so a fresh clone boots with the
// five launch stylists and zero manual setup. Writes (partner create/update/
// publish) persist back to the same file. If the file cannot be written the
// change still lives in memory for the session — the demo keeps working even on
// a read-only checkout.
type fileStore struct {
	mu    sync.RWMutex
	byID  map[string]Persona
	order []string // slugs in insertion order
	path  string   // resolved file path, "" if persistence is unavailable
}

// NewFileStore builds a Store backed by the JSON file at path. A missing file is
// seeded from the embedded default and written.
func NewFileStore(path string) (Store, error) {
	s := &fileStore{byID: map[string]Persona{}, path: path}

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if lerr := s.loadBytes(data); lerr != nil {
			return nil, fmt.Errorf("personas: parsing %s: %w", path, lerr)
		}
	case errors.Is(err, os.ErrNotExist):
		// First boot: seed from the embedded default and persist it.
		if lerr := s.loadBytes(seedJSON); lerr != nil {
			return nil, fmt.Errorf("personas: parsing embedded seed: %w", lerr)
		}
		_ = s.persist() // best-effort; in-memory store works regardless
	default:
		return nil, fmt.Errorf("personas: reading %s: %w", path, err)
	}
	return s, nil
}

// DefaultStore resolves the conventional data path and builds the file store.
// It tries the process CWD (`data/personas.json`, the case when the server runs
// from backend/) and the repo-root layout (`backend/data/personas.json`), then
// falls back to seeding a fresh file at the first candidate.
func DefaultStore() (Store, error) {
	if env := strings.TrimSpace(os.Getenv("STARCLOSET_PERSONAS_PATH")); env != "" {
		return NewFileStore(env)
	}
	candidates := []string{
		filepath.Join("data", "personas.json"),
		filepath.Join("backend", "data", "personas.json"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return NewFileStore(c)
		}
	}
	// None exist yet — seed the conventional one.
	return NewFileStore(candidates[0])
}

func (s *fileStore) loadBytes(data []byte) error {
	var list []Persona
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	s.byID = make(map[string]Persona, len(list))
	s.order = s.order[:0]
	for _, p := range list {
		p.normalize()
		if p.Slug == "" {
			continue
		}
		if _, dup := s.byID[p.Slug]; !dup {
			s.order = append(s.order, p.Slug)
		}
		s.byID[p.Slug] = p
	}
	return nil
}

// persist writes the current set to disk atomically (temp file + rename). It is
// a no-op when no path is configured.
func (s *fileStore) persist() error {
	if s.path == "" {
		return nil
	}
	list := s.snapshot()
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(s.path); dir != "" {
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return mkErr
		}
	}
	tmp := s.path + ".tmp"
	if werr := os.WriteFile(tmp, data, 0o644); werr != nil {
		return werr
	}
	return os.Rename(tmp, s.path)
}

// snapshot returns personas in insertion order (caller holds no lock; used
// internally where the lock is already held or ordering is stable).
func (s *fileStore) snapshot() []Persona {
	out := make([]Persona, 0, len(s.order))
	for _, slug := range s.order {
		if p, ok := s.byID[slug]; ok {
			out = append(out, p)
		}
	}
	return out
}

// sortedByUnlock returns personas ordered by unlock threshold then display name,
// which is the order the roster is presented in.
func sortedByUnlock(list []Persona) []Persona {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].UnlockAtItems != list[j].UnlockAtItems {
			return list[i].UnlockAtItems < list[j].UnlockAtItems
		}
		return list[i].DisplayName < list[j].DisplayName
	})
	return list
}

func (s *fileStore) All() []Persona {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return sortedByUnlock(s.snapshot())
}

func (s *fileStore) Published() []Persona {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Persona, 0, len(s.order))
	for _, p := range s.snapshot() {
		if p.IsPublished() {
			out = append(out, p)
		}
	}
	return sortedByUnlock(out)
}

func (s *fileStore) Get(slug string) (Persona, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.byID[normalizeSlug(slug)]
	return p, ok
}

func (s *fileStore) Create(p Persona) (Persona, error) {
	p.normalize()
	if p.Slug == "" {
		p.Slug = Slugify(p.DisplayName)
	}
	if p.Slug == "" {
		return Persona{}, errors.New("personas: a display name or slug is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byID[p.Slug]; exists {
		return Persona{}, ErrSlugTaken
	}
	if p.Status == "" {
		p.Status = StatusDraft
	}
	s.byID[p.Slug] = p
	s.order = append(s.order, p.Slug)
	if err := s.persist(); err != nil {
		return p, nil // kept in memory; persistence is best-effort
	}
	return p, nil
}

func (s *fileStore) Update(slug string, in Persona) (Persona, error) {
	slug = normalizeSlug(slug)
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.byID[slug]
	if !ok {
		return Persona{}, ErrNotFound
	}
	in.normalize()
	// Identity: slug is immutable; status is changed only through Publish.
	in.Slug = existing.Slug
	if in.Status == "" {
		in.Status = existing.Status
	}
	s.byID[slug] = in
	_ = s.persist()
	return in, nil
}

func (s *fileStore) Publish(slug string) (Persona, error) {
	slug = normalizeSlug(slug)
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.byID[slug]
	if !ok {
		return Persona{}, ErrNotFound
	}
	p.Status = StatusPublished
	s.byID[slug] = p
	_ = s.persist()
	return p, nil
}
