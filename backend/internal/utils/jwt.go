package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType distinguishes access from refresh tokens.
//
// Without this, a refresh token — which is long-lived by design — would be
// accepted as an access token, quietly turning a 7-day credential into a 7-day
// session and defeating the point of short access expiry.
type TokenType string

// Token types.
const (
	TokenAccess  TokenType = "access"
	TokenRefresh TokenType = "refresh"
)

// Claims is the JWT payload.
type Claims struct {
	UserID  uuid.UUID `json:"uid"`
	Email   string    `json:"email"`
	Role    string    `json:"role"`
	IsGuest bool      `json:"guest"`
	Type    TokenType `json:"typ"`
	jwt.RegisteredClaims
}

// Token errors.
var (
	ErrTokenInvalid  = errors.New("utils: token is invalid")
	ErrTokenExpired  = errors.New("utils: token has expired")
	ErrTokenWrongUse = errors.New("utils: token used for the wrong purpose")
)

// TokenManager issues and verifies JWTs.
type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
}

// NewTokenManager constructs a TokenManager.
func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		issuer:     "myntra-bharat-ai",
	}
}

// TokenPair is an access/refresh pair plus the access token's lifetime.
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int    `json:"expiresIn"`
}

// Generate issues a fresh token pair for a user.
func (m *TokenManager) Generate(userID uuid.UUID, email, role string, isGuest bool) (*TokenPair, error) {
	access, err := m.sign(userID, email, role, isGuest, TokenAccess, m.accessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := m.sign(userID, email, role, isGuest, TokenRefresh, m.refreshTTL)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresIn:    int(m.accessTTL.Seconds()),
	}, nil
}

func (m *TokenManager) sign(userID uuid.UUID, email, role string, isGuest bool, typ TokenType, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:  userID,
		Email:   email,
		Role:    role,
		IsGuest: isGuest,
		Type:    typ,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("utils: signing token: %w", err)
	}
	return signed, nil
}

// Verify parses and validates a token, checking that it is of the expected type.
func (m *TokenManager) Verify(tokenString string, expected TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		// Pin the algorithm. Without this check a token signed with "none",
		// or one signed with HMAC against a public key, is accepted — the
		// classic JWT algorithm-confusion attack.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("utils: unexpected signing method %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	if claims.Type != expected {
		return nil, ErrTokenWrongUse
	}
	return claims, nil
}
