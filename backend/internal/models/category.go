package models

import (
	"time"

	"github.com/google/uuid"
)

// Category is a garment family ("Kurta Sets", "Footwear", ...). It is a real
// table rather than a string column on Product so the catalogue can carry
// per-category metadata and so the FK guarantees referential integrity.
type Category struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Name        CategoryName `gorm:"type:varchar(48);uniqueIndex;not null" json:"name"`
	Slug        string       `gorm:"type:varchar(48);uniqueIndex;not null" json:"slug"`
	Description string       `gorm:"type:text" json:"description"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// TableName pins the table name.
func (Category) TableName() string { return "categories" }
