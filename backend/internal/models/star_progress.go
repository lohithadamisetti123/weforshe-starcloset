package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StarProgress is one shopper's unlock progress in one StarCloset department.
//
// Progress is stored per (user, department) — never as a single shared counter —
// so the three journeys stay independent: adding to the cart while shopping
// Women advances only the Women row, and Men/Kids remain untouched. The
// uniqueIndex over (user_id, department) is what makes each department's counter
// a single, upsertable row.
type StarProgress struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_star_progress_user_dept,priority:1;index" json:"userId"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`

	Department string `gorm:"type:varchar(16);not null;uniqueIndex:idx_star_progress_user_dept,priority:2" json:"department"`
	Items      int    `gorm:"not null;default:0" json:"items"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName pins the table name.
func (StarProgress) TableName() string { return "star_progress" }

// BeforeCreate assigns an ID when absent.
func (s *StarProgress) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Items < 0 {
		s.Items = 0
	}
	return nil
}
