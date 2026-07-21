package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PastOrder is one line of a shopper's purchase history.
//
// For the MVP the rows are mock (see services.ProfileService and the
// past_order repository's mock generator), but the shape is the production one:
// a denormalised snapshot of what was bought, plus an optional soft link to the
// catalogue product it was derived from.
//
// Why the snapshot instead of just a product_id: real Myntra order history
// references products that may no longer be sold, may predate this catalogue,
// or may never have existed in it. Storing title/brand/category/price on the
// order row means importing that history is a data load, not a schema change —
// and the personalization signals (which brands and categories a shopper has
// actually bought) read straight off these columns without a join.
type PastOrder struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_past_orders_user_ordered,priority:1" json:"userId"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`

	// ProductID is a soft link (nullable, ON DELETE SET NULL). Present for mock
	// orders derived from the catalogue; absent for imported real orders that
	// don't map to a current product.
	ProductID *uuid.UUID `gorm:"type:uuid" json:"productId,omitempty"`
	Product   *Product   `gorm:"foreignKey:ProductID;constraint:OnDelete:SET NULL" json:"-"`

	Title    string `gorm:"type:varchar(180);not null" json:"title"`
	Brand    string `gorm:"type:varchar(80);not null;default:''" json:"brand"`
	Category string `gorm:"type:varchar(48);not null;default:''" json:"category"`
	Image    string `gorm:"type:text;not null;default:''" json:"image"`
	Price    int    `gorm:"not null;default:0" json:"price"`

	OrderedAt time.Time `gorm:"index:idx_past_orders_user_ordered,priority:2,sort:desc" json:"orderedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// TableName pins the table name.
func (PastOrder) TableName() string { return "past_orders" }

// BeforeCreate assigns an ID and a default order time when absent.
func (o *PastOrder) BeforeCreate(_ *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	if o.OrderedAt.IsZero() {
		o.OrderedAt = time.Now()
	}
	return nil
}
