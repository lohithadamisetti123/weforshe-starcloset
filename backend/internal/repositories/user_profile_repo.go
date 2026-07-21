package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// UserProfileRepository is the data-access contract for shopper profiles.
type UserProfileRepository interface {
	// Get returns a user's profile, or (nil, nil) when they have none yet.
	// "No profile" is an ordinary state (not onboarded), not an error — the
	// caller decides what to do with the absence, so this does not manufacture
	// a 404 the way a required-resource lookup would.
	Get(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error)
	// Upsert creates or fully replaces a user's profile. Idempotent on user_id.
	Upsert(ctx context.Context, profile *models.UserProfile) (*models.UserProfile, error)
}

type userProfileRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewUserProfileRepository constructs a UserProfileRepository.
func NewUserProfileRepository(db *gorm.DB, log *zap.Logger) UserProfileRepository {
	return &userProfileRepository{db: db, log: log}
}

func (r *userProfileRepository) Get(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error) {
	var p models.UserProfile
	err := r.db.WithContext(ctx).First(&p, "user_id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("repositories: fetching profile for %s: %w", userID, err)
	}
	return &p, nil
}

// Upsert writes the profile, replacing the existing one for this user if any.
//
// The unique index on user_id turns this into a single atomic statement via
// ON CONFLICT, so a shopper editing their profile can never end up with two
// rows and there is no read-then-write race between two tabs.
func (r *userProfileRepository) Upsert(ctx context.Context, profile *models.UserProfile) (*models.UserProfile, error) {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		// Replace the mutable columns; id/user_id/created_at are left as-is on
		// the existing row. updated_at is bumped by GORM's autoUpdateTime.
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "gender", "city", "budget_min", "budget_max",
			"favorite_styles", "favorite_colors", "favorite_brands",
			"favorite_categories", "onboarded", "updated_at",
		}),
	}).Create(profile).Error
	if err != nil {
		if isForeignKeyViolation(err) {
			// The user_id FK failed — the account referenced does not exist.
			return nil, fmt.Errorf("repositories: profile references unknown user: %w", err)
		}
		return nil, fmt.Errorf("repositories: upserting profile: %w", err)
	}

	// Read back so the caller gets the canonical row (with the real created_at
	// and any DB-side defaults) rather than the in-memory struct, which on an
	// update carries a stale/absent created_at.
	return r.Get(ctx, profile.UserID)
}
