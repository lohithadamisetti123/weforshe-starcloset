package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// UserRepository is the account data access contract.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	EnsureGuest(ctx context.Context, sessionID string) (*models.User, error)
}

type userRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewUserRepository constructs a UserRepository.
func NewUserRepository(db *gorm.DB, log *zap.Logger) UserRepository {
	return &userRepository{db: db, log: log}
}

// Create inserts a user, translating the unique-email violation into a 409.
//
// The check-then-insert alternative has a race: two simultaneous registrations
// for the same address both see "available" and one dies on the constraint with
// a 500. Letting the database arbitrate and mapping its verdict is both correct
// and one round trip cheaper.
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		if isUniqueViolation(err) {
			return utils.NewConflict("An account with this email already exists")
		}
		return fmt.Errorf("repositories: creating user: %w", err)
	}
	return nil
}

// GetByID fetches a user.
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFound("User")
		}
		return nil, fmt.Errorf("repositories: fetching user %s: %w", id, err)
	}
	return &u, nil
}

// GetByEmail fetches a user by email address.
//
// Email is normalised to lower case on both write and read. Without that,
// "Asha@myntra.dev" and "asha@myntra.dev" are two accounts, and the one the
// shopper can log into is whichever they happened to type first.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.WithContext(ctx).First(&u, "email = ?", normalizeEmail(email)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFound("User")
		}
		return nil, fmt.Errorf("repositories: fetching user by email: %w", err)
	}
	return &u, nil
}

// EmailExists reports whether an address is taken.
func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("email = ?", normalizeEmail(email)).Count(&n).Error; err != nil {
		return false, fmt.Errorf("repositories: checking email: %w", err)
	}
	return n > 0, nil
}

// EnsureGuest returns the guest account for a browser session, creating it on
// first sight.
//
// The frontend ships no login screen, so without this every assistant request
// would either 401 or write to a shared anonymous bucket where one shopper sees
// another's history. A guest is a real row with a derived, stable ID: history
// and favourites work immediately, and the account can later be claimed by a
// real registration without migrating anything.
//
// The insert is an idempotent upsert rather than a read-then-write because two
// tabs opening at once is the normal case, not the edge case.
func (r *userRepository) EnsureGuest(ctx context.Context, sessionID string) (*models.User, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, utils.NewBadRequest("A session identifier is required for guest access", nil)
	}

	id := guestUUID(sessionID)

	var existing models.User
	err := r.db.WithContext(ctx).First(&existing, "id = ?", id).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("repositories: looking up guest: %w", err)
	}

	guest := &models.User{
		ID:      id,
		Name:    "Guest",
		Email:   fmt.Sprintf("guest-%s@session.local", id.String()),
		IsGuest: true,
		Role:    models.RoleUser,
		// Guests never authenticate with a password, so there is no password to
		// hash. The column is NOT NULL, and a sentinel that bcrypt can never
		// produce a match for is safer than an empty string that a future
		// CheckPassword change might treat as valid.
		PasswordHash: "!guest-no-login",
	}

	if err := r.db.WithContext(ctx).Create(guest).Error; err != nil {
		if isUniqueViolation(err) {
			// Lost the race against a concurrent request from another tab. The
			// other one won; its row is the answer.
			return r.GetByID(ctx, id)
		}
		return nil, fmt.Errorf("repositories: creating guest: %w", err)
	}

	r.log.Debug("guest session created", zap.String("userId", id.String()))
	return guest, nil
}

// guestNamespace derives guest IDs from session IDs. Deriving rather than
// randomising means the same browser returns to the same history after a
// restart, with no server-side session table to keep.
var guestNamespace = uuid.MustParse("2b0f4a71-6c5e-5d38-9a1b-7e4c2f8d3a90")

func guestUUID(sessionID string) uuid.UUID {
	return uuid.NewSHA1(guestNamespace, []byte(sessionID))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
