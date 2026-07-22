package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/recommendation"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
)

// mockOrderCount is how many purchase-history lines a newly-onboarded shopper is
// given, so past-order personalization has something to work with in the demo.
const mockOrderCount = 4

// ProfileService owns shopper profiles and their (mock) purchase history.
//
// It composes three repositories: the profile store, the order store, and the
// product catalogue (used only to synthesise believable mock orders from real
// products). The mock-order generation is quarantined to this service and gated
// behind a "does this user already have orders?" check, so swapping in real
// Myntra order history later means deleting EnsureMockOrders and importing rows
// through the same PastOrderRepository — no other code moves.
type ProfileService struct {
	profiles repositories.UserProfileRepository
	orders   repositories.PastOrderRepository
	products repositories.ProductRepository
	log      *zap.Logger
}

// NewProfileService constructs a ProfileService.
func NewProfileService(
	profiles repositories.UserProfileRepository,
	orders repositories.PastOrderRepository,
	products repositories.ProductRepository,
	log *zap.Logger,
) *ProfileService {
	return &ProfileService{profiles: profiles, orders: orders, products: products, log: log}
}

// Get returns a shopper's profile, or a zero not-onboarded profile if none.
func (s *ProfileService) Get(ctx context.Context, userID uuid.UUID) (dto.ProfileResponse, error) {
	p, err := s.profiles.Get(ctx, userID)
	if err != nil {
		return dto.ProfileResponse{}, err
	}
	return dto.NewProfileResponse(p), nil
}

// Status reports whether the shopper has completed onboarding.
func (s *ProfileService) Status(ctx context.Context, userID uuid.UUID) (dto.OnboardingStatusResponse, error) {
	p, err := s.profiles.Get(ctx, userID)
	if err != nil {
		return dto.OnboardingStatusResponse{}, err
	}
	return dto.OnboardingStatusResponse{Onboarded: p != nil && p.Onboarded}, nil
}

// Update creates or replaces the profile. When the submission completes
// onboarding, it lazily seeds mock purchase history so the shopper immediately
// has a "past orders" signal to personalize on.
func (s *ProfileService) Update(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) (dto.ProfileResponse, error) {
	// Preserve a prior onboarded=true across edits that don't resend the flag.
	existing, err := s.profiles.Get(ctx, userID)
	if err != nil {
		return dto.ProfileResponse{}, err
	}
	onboarded := existing != nil && existing.Onboarded
	if req.Onboarded != nil {
		onboarded = *req.Onboarded
	}

	profile := &models.UserProfile{
		UserID:             userID,
		Name:               req.Name,
		City:               req.City,
		BudgetMin:          req.BudgetMin,
		BudgetMax:          req.BudgetMax,
		FavoriteStyles:     models.ToJSONB(orEmptyStrings(req.FavoriteStyles)),
		FavoriteColors:     models.ToJSONB(orEmptyStrings(req.FavoriteColors)),
		FavoriteBrands:     models.ToJSONB(orEmptyStrings(req.FavoriteBrands)),
		FavoriteCategories: models.ToJSONB(orEmptyStrings(req.FavoriteCategories)),
		FavoriteOccasions:  models.ToJSONB(orEmptyStrings(req.FavoriteOccasions)),
		Sizes:              models.ToJSONB(orEmptyStrings(req.Sizes)),
		Onboarded:          onboarded,
	}
	if req.Gender != "" {
		if g, ok := models.ParseGender(req.Gender); ok {
			profile.Gender = g
		}
	}

	saved, err := s.profiles.Upsert(ctx, profile)
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	// Seed mock orders once, on first completed onboarding. Failure here must
	// not fail the profile save — the profile is the shopper's answer; the mock
	// history is a demo convenience layered on top.
	if saved != nil && saved.Onboarded {
		if err := s.ensureMockOrders(ctx, userID, saved); err != nil {
			s.log.Warn("seeding mock past orders failed",
				zap.String("userId", userID.String()), zap.Error(err))
		}
	}

	return dto.NewProfileResponse(saved), nil
}

// Orders returns a shopper's purchase history, newest first.
func (s *ProfileService) Orders(ctx context.Context, userID uuid.UUID) ([]dto.PastOrderResponse, error) {
	orders, err := s.orders.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewPastOrders(orders), nil
}

// ensureMockOrders gives a shopper a believable purchase history the first time
// they onboard, drawn from real catalogue products that fit their profile so
// the past-order boost has something meaningful to reinforce.
//
// It is idempotent: if the shopper already has orders it does nothing, so a
// profile edit never duplicates history. Real order history replaces this
// wholesale — the point of generating from the catalogue is only that a fresh
// demo account isn't empty.
func (s *ProfileService) ensureMockOrders(ctx context.Context, userID uuid.UUID, profile *models.UserProfile) error {
	count, err := s.orders.Count(ctx, userID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	picks := s.pickMockProducts(ctx, profile)
	if len(picks) == 0 {
		return nil
	}

	now := time.Now()
	orders := make([]*models.PastOrder, 0, len(picks))
	for i, p := range picks {
		pid := p.ID
		category := ""
		if p.Category != nil {
			category = string(p.Category.Name)
		}
		orders = append(orders, &models.PastOrder{
			UserID:    userID,
			ProductID: &pid,
			Title:     p.Title,
			Brand:     p.Brand,
			Category:  category,
			Image:     p.Image,
			Price:     p.Price,
			// Spread orders back in time so the history reads like real activity
			// rather than a single bulk insert.
			OrderedAt: now.AddDate(0, 0, -(i+1)*17),
		})
	}
	return s.orders.CreateMany(ctx, orders)
}

// pickMockProducts selects up to mockOrderCount catalogue products that suit the
// profile, favouring the shopper's stated categories and gender and spreading
// across distinct categories so the mock history looks like a real wardrobe.
func (s *ProfileService) pickMockProducts(ctx context.Context, profile *models.UserProfile) []*models.Product {
	filter := recommendation.CandidateFilter{Limit: 60}
	if profile.Gender != "" {
		g := profile.Gender
		filter.Gender = &g
	}

	pool, err := s.products.Candidates(ctx, filter)
	if err != nil || len(pool) == 0 {
		return nil
	}

	preferredCats := map[string]bool{}
	for _, c := range profile.CategoriesList() {
		if cat, ok := models.ParseCategory(c); ok {
			preferredCats[string(cat)] = true
		}
	}

	// Two passes: first take products in preferred categories (one per
	// category), then fill the remainder with other categories — always one per
	// category so the four orders don't collapse into four of the same thing.
	picks := make([]*models.Product, 0, mockOrderCount)
	usedCat := map[string]bool{}

	take := func(preferOnly bool) {
		for _, p := range pool {
			if len(picks) >= mockOrderCount {
				return
			}
			if p.Category == nil {
				continue
			}
			cat := string(p.Category.Name)
			if usedCat[cat] {
				continue
			}
			if preferOnly && len(preferredCats) > 0 && !preferredCats[cat] {
				continue
			}
			picks = append(picks, p)
			usedCat[cat] = true
		}
	}
	take(true)
	take(false)
	return picks
}

// orEmptyStrings guarantees a non-nil slice so an omitted preference list is
// stored as `[]`, not `null`.
func orEmptyStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
