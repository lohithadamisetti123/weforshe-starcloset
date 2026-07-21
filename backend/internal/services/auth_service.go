// Package services holds the application's business logic.
//
// Services orchestrate repositories, the AI layer and the recommendation engine.
// They own transactions and policy; handlers own HTTP and nothing else. The
// division matters because it is what lets the assistant pipeline — the part of
// this system worth testing hardest — be exercised without a router, a request,
// or a socket.
package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/repositories"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// AuthService handles registration, login and token issuance.
type AuthService struct {
	users  repositories.UserRepository
	tokens *utils.TokenManager
	log    *zap.Logger
}

// NewAuthService constructs an AuthService.
func NewAuthService(users repositories.UserRepository, tokens *utils.TokenManager, log *zap.Logger) *AuthService {
	return &AuthService{users: users, tokens: tokens, log: log}
}

// Register creates an account and returns it with a fresh token pair.
//
// Registering logs you in. Making a shopper who just typed their password type
// it again to get a token is ceremony with no security benefit — they proved
// possession of the credentials by choosing them a millisecond ago.
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Name:         strings.TrimSpace(req.Name),
		Email:        email,
		PasswordHash: hash,
		Role:         models.RoleUser,
	}

	// The duplicate-email check lives in the repository, where the unique index
	// arbitrates. A check here would be a race with a nicer error message.
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	pair, err := s.tokens.Generate(user.ID, user.Email, string(user.Role), user.IsGuest)
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	s.log.Info("user registered", zap.String("userId", user.ID.String()))

	return &dto.AuthResponse{
		User:   dto.NewUser(user),
		Tokens: *pair,
	}, nil
}

// Login authenticates a shopper.
//
// Both the unknown-email and wrong-password paths return the same error. The
// alternative — "no account with that email" — is a free account-enumeration
// oracle: an attacker learns which addresses are registered without guessing a
// single password. The marginal UX gain does not buy that back.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) && appErr.Status == 404 {
			return nil, utils.NewUnauthorized("Incorrect email or password")
		}
		return nil, err
	}

	if user.IsGuest {
		// A guest row has a sentinel hash that bcrypt can never match, so this
		// branch is belt-and-braces — but it makes the intent explicit rather
		// than resting on a string constant staying unguessable.
		return nil, utils.NewUnauthorized("Incorrect email or password")
	}

	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		return nil, utils.NewUnauthorized("Incorrect email or password")
	}

	pair, err := s.tokens.Generate(user.ID, user.Email, string(user.Role), user.IsGuest)
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	s.log.Info("user logged in", zap.String("userId", user.ID.String()))

	return &dto.AuthResponse{
		User:   dto.NewUser(user),
		Tokens: *pair,
	}, nil
}

// Refresh exchanges a refresh token for a new pair.
//
// The user is re-read rather than trusted from the token's claims. A token
// minted an hour ago asserts a role that may since have changed, and a deleted
// account's refresh token would otherwise keep minting valid access tokens
// until it expired on its own schedule.
func (s *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error) {
	claims, err := s.tokens.Verify(req.RefreshToken, utils.TokenRefresh)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrTokenExpired):
			return nil, utils.NewUnauthorized("Your session has expired. Please sign in again")
		case errors.Is(err, utils.ErrTokenWrongUse):
			return nil, utils.NewUnauthorized("That is not a refresh token")
		default:
			return nil, utils.NewUnauthorized("Invalid refresh token")
		}
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, utils.NewUnauthorized("Your account is no longer active")
	}

	pair, err := s.tokens.Generate(user.ID, user.Email, string(user.Role), user.IsGuest)
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	return &dto.AuthResponse{
		User:   dto.NewUser(user),
		Tokens: *pair,
	}, nil
}

// Profile returns the authenticated shopper's account.
func (s *AuthService) Profile(ctx context.Context, userID uuid.UUID) (*dto.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	u := dto.NewUser(user)
	return &u, nil
}

// EnsureGuest resolves a browser session to a guest account, minting one on
// first contact.
//
// This is what lets the shipped frontend — which has no sign-in screen — use
// the assistant, keep a history and save favourites. The guest is a real user
// row, so every downstream query is ordinary user-scoped code with no
// anonymous special case threaded through it.
func (s *AuthService) EnsureGuest(ctx context.Context, sessionID string) (*models.User, *utils.TokenPair, error) {
	user, err := s.users.EnsureGuest(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	pair, err := s.tokens.Generate(user.ID, user.Email, string(user.Role), true)
	if err != nil {
		return nil, nil, utils.NewInternal(err)
	}
	return user, pair, nil
}
