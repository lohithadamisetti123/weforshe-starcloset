package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is an account. Guests are real rows too (IsGuest = true), which keeps
// every downstream foreign key non-nullable and means history and favourites
// need exactly one code path regardless of whether the visitor has signed up.
type User struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name  string    `gorm:"type:varchar(120);not null" json:"name"`
	Email string    `gorm:"type:varchar(180);uniqueIndex;not null" json:"email"`

	// PasswordHash is bcrypt output. The json:"-" tag is the last line of
	// defence: no handler serialises models.User directly, but if one ever did,
	// this must not leak.
	PasswordHash string `gorm:"type:varchar(120);not null" json:"-"`

	Role Role `gorm:"type:varchar(16);not null;default:'User'" json:"role"`

	// IsGuest marks anonymous browser-scoped identities minted from the
	// X-Session-Id header.
	IsGuest bool `gorm:"not null;default:false;index" json:"isGuest"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName pins the table name.
func (User) TableName() string { return "users" }

// BeforeCreate assigns an ID when absent.
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// IsAdmin reports whether the user holds the Admin role.
func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }
