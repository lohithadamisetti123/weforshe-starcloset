package dto

import (
	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// RegisterRequest is the body of POST /api/v1/auth/register.
type RegisterRequest struct {
	Name  string `json:"name" binding:"required" validate:"required,min=2,max=120" example:"Asha Menon"`
	Email string `json:"email" binding:"required" validate:"required,email,max=180" example:"asha@example.com"`
	// Password has a floor of 8 characters and no composition rules. Forcing a
	// symbol and a digit measurably pushes people toward "Password1!" and its
	// cousins; length is the property that actually matters.
	Password string `json:"password" binding:"required" validate:"required,min=8,max=72" example:"correct-horse-battery"`
}

// LoginRequest is the body of POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required" validate:"required,email" example:"asha@example.com"`
	Password string `json:"password" binding:"required" validate:"required,min=1" example:"correct-horse-battery"`
}

// RefreshRequest is the body of POST /api/v1/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required" validate:"required"`
}

// User is the wire shape of an account. There is deliberately no password
// field of any kind, hashed or otherwise.
type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsGuest   bool   `json:"isGuest"`
	CreatedAt string `json:"createdAt"`
}

// NewUser maps a user model.
func NewUser(u *models.User) User {
	if u == nil {
		return User{}
	}
	return User{
		ID:        u.ID.String(),
		Name:      u.Name,
		Email:     u.Email,
		Role:      string(u.Role),
		IsGuest:   u.IsGuest,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// AuthResponse is returned by register, login and refresh.
type AuthResponse struct {
	User   User            `json:"user"`
	Tokens utils.TokenPair `json:"tokens"`
}
