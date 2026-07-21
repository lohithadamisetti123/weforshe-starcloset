package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatHistory is one prompt the user sent the assistant, plus a denormalised
// summary of what came back.
//
// The denormalisation (budget label, occasion, product count) is intentional:
// the history list is read on every visit to /history and must render without
// touching the recommendation items table. The authoritative detail still hangs
// off RecommendationID.
type ChatHistory struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_history_user_created,priority:1" json:"userId"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`

	Prompt   string `gorm:"type:text;not null" json:"prompt"`
	Response string `gorm:"type:text" json:"response"`

	// BudgetLabel is the display string ("₹5,000", "flexible budget") rather
	// than a number, because that is exactly what the frontend renders and
	// re-deriving the formatting in two places invites drift.
	BudgetLabel string `gorm:"type:varchar(48);not null;default:''" json:"budgetLabel"`
	Occasion    string `gorm:"type:varchar(32);not null;default:''" json:"occasion"`
	Style       string `gorm:"type:varchar(32);not null;default:''" json:"style"`

	ProductsRecommended int           `gorm:"not null;default:0" json:"productsRecommended"`
	Status              HistoryStatus `gorm:"type:varchar(16);not null;default:'Completed'" json:"status"`

	RecommendationID *uuid.UUID      `gorm:"type:uuid;index" json:"recommendationId,omitempty"`
	Recommendation   *Recommendation `gorm:"foreignKey:RecommendationID;constraint:OnDelete:SET NULL" json:"-"`

	CreatedAt time.Time      `gorm:"index:idx_history_user_created,priority:2,sort:desc" json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName pins the table name.
func (ChatHistory) TableName() string { return "chat_history" }

// BeforeCreate assigns an ID when absent.
func (c *ChatHistory) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.Status == "" {
		c.Status = HistoryStatusCompleted
	}
	return nil
}
