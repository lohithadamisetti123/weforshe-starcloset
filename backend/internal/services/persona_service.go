package services

import (
	"errors"

	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/personas"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// PersonaService is the use-case layer over the persona store. It powers both
// the shopper roster (published personas) and the partner portal (full CRUD +
// publish). It holds no persistence opinion beyond the store interface.
type PersonaService struct {
	store personas.Store
	log   *zap.Logger
}

// NewPersonaService constructs a PersonaService.
func NewPersonaService(store personas.Store, log *zap.Logger) *PersonaService {
	return &PersonaService{store: store, log: log}
}

// ListPublished returns the shopper-facing roster (published only).
func (s *PersonaService) ListPublished() []dto.StylistPersona {
	return dto.NewStylistPersonas(s.store.Published())
}

// ListAll returns every persona for the partner portal (drafts included).
func (s *PersonaService) ListAll() []dto.PersonaDetail {
	return dto.NewPersonaDetails(s.store.All())
}

// Get resolves one persona in full partner detail.
func (s *PersonaService) Get(slug string) (dto.PersonaDetail, error) {
	p, ok := s.store.Get(slug)
	if !ok {
		return dto.PersonaDetail{}, utils.NewNotFound("Stylist")
	}
	return dto.NewPersonaDetail(p), nil
}

// Schema returns the onboarding form contract.
func (s *PersonaService) Schema() dto.PersonaSchemaResponse {
	return dto.NewPersonaSchema()
}

// Create adds a new draft persona.
func (s *PersonaService) Create(payload dto.OnboardPersonaPayload) (dto.PersonaDetail, error) {
	created, err := s.store.Create(payload.ToPersona())
	if err != nil {
		if errors.Is(err, personas.ErrSlugTaken) {
			return dto.PersonaDetail{}, utils.NewConflict("A stylist with that name already exists")
		}
		return dto.PersonaDetail{}, utils.NewBadRequest("Could not create the stylist", err)
	}
	s.log.Info("persona created", zap.String("slug", created.Slug))
	return dto.NewPersonaDetail(created), nil
}

// Update replaces the mutable fields of an existing persona.
func (s *PersonaService) Update(slug string, payload dto.OnboardPersonaPayload) (dto.PersonaDetail, error) {
	updated, err := s.store.Update(slug, payload.ToPersona())
	if err != nil {
		if errors.Is(err, personas.ErrNotFound) {
			return dto.PersonaDetail{}, utils.NewNotFound("Stylist")
		}
		return dto.PersonaDetail{}, utils.NewBadRequest("Could not update the stylist", err)
	}
	s.log.Info("persona updated", zap.String("slug", updated.Slug))
	return dto.NewPersonaDetail(updated), nil
}

// Publish flips a persona to published so shoppers can see it.
func (s *PersonaService) Publish(slug string) (dto.PersonaDetail, error) {
	published, err := s.store.Publish(slug)
	if err != nil {
		if errors.Is(err, personas.ErrNotFound) {
			return dto.PersonaDetail{}, utils.NewNotFound("Stylist")
		}
		return dto.PersonaDetail{}, utils.NewInternal(err)
	}
	s.log.Info("persona published", zap.String("slug", published.Slug))
	return dto.NewPersonaDetail(published), nil
}
