package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CartItem is one product in a user's cart.
//
// Like FavoriteProduct, ownership is a column, not a convention: every query in
// the repository filters on user_id, and the uniqueIndex over
// (user_id, product_id) is scoped to the user — so two accounts can hold the
// same product independently, and adding the same product twice bumps a
// quantity instead of creating a second row.
type CartItem struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_cart_user_product,priority:1;index" json:"userId"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`

	ProductID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_cart_user_product,priority:2" json:"productId"`
	Product   *Product  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`

	Quantity int `gorm:"not null;default:1" json:"quantity"`

	CreatedAt time.Time `gorm:"index:idx_cart_user_created,sort:desc" json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName pins the table name.
func (CartItem) TableName() string { return "cart_items" }

// BeforeCreate assigns an ID when absent.
func (c *CartItem) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.Quantity <= 0 {
		c.Quantity = 1
	}
	return nil
}

// SavedCollection is an AI-generated collection a user has saved.
//
// The collection payload is stored as a jsonb snapshot rather than as foreign
// keys to products. A collection is an arrangement the ranker produced for one
// prompt at one moment — re-deriving it later would yield a *different*
// collection, so the snapshot is what makes "saved" mean what a shopper expects.
type SavedCollection struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_saved_coll_user_name,priority:1;index" json:"userId"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`

	Name       string `gorm:"type:varchar(180);not null;uniqueIndex:idx_saved_coll_user_name,priority:2" json:"name"`
	Department string `gorm:"type:varchar(48);not null;default:'';uniqueIndex:idx_saved_coll_user_name,priority:3" json:"department"`
	Prompt     string `gorm:"type:text;not null;default:''" json:"prompt"`

	Collection JSONB `gorm:"type:jsonb;not null" json:"collection"`

	CreatedAt time.Time `gorm:"index:idx_saved_coll_user_created,sort:desc" json:"createdAt"`
}

// TableName pins the table name.
func (SavedCollection) TableName() string { return "saved_collections" }

// BeforeCreate assigns an ID when absent.
func (s *SavedCollection) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// CollectionCartItem is a whole collection a user added to their cart.
type CollectionCartItem struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_coll_cart_user_name,priority:1;index" json:"userId"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`

	Name       string `gorm:"type:varchar(180);not null;uniqueIndex:idx_coll_cart_user_name,priority:2" json:"name"`
	Department string `gorm:"type:varchar(48);not null;default:'';uniqueIndex:idx_coll_cart_user_name,priority:3" json:"department"`

	Collection JSONB `gorm:"type:jsonb;not null" json:"collection"`

	CreatedAt time.Time `gorm:"index:idx_coll_cart_user_created,sort:desc" json:"createdAt"`
}

// TableName pins the table name.
func (CollectionCartItem) TableName() string { return "collection_cart_items" }

// BeforeCreate assigns an ID when absent.
func (c *CollectionCartItem) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
