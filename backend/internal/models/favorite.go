package models

import (
	"time"

	"github.com/google/uuid"
)

// FavoriteProduct is a user's saved product.
//
// There is no soft delete here on purpose: un-favouriting is a genuine removal,
// and a uniqueIndex over (user_id, product_id) is what makes re-favouriting
// idempotent at the database level rather than in a racy read-then-write.
type FavoriteProduct struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_fav_user_product,priority:1;index" json:"userId"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`

	ProductID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_fav_user_product,priority:2" json:"productId"`
	Product   *Product  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`

	CreatedAt time.Time `gorm:"index:idx_fav_user_created,sort:desc" json:"createdAt"`
}

// TableName pins the table name.
func (FavoriteProduct) TableName() string { return "favorite_products" }
