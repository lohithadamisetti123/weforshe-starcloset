package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Product is the catalogue row. Field names are DB-facing; the wire shape the
// frontend consumes is produced by dto.NewProduct so that storage concerns and
// the public contract can evolve independently.
type Product struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	// Slug is the stable human-readable handle ("prod-001") carried over from
	// the frontend fixture. It keeps seeding idempotent and makes support
	// tickets legible; it is never exposed to the client.
	Slug string `gorm:"type:varchar(32);uniqueIndex;not null" json:"slug"`

	Title       string `gorm:"type:varchar(180);not null;index" json:"title"`
	Brand       string `gorm:"type:varchar(80);not null;index" json:"brand"`
	Description string `gorm:"type:text;not null" json:"description"`
	Image       string `gorm:"type:text;not null" json:"image"`

	CategoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"categoryId"`
	Category   *Category `gorm:"foreignKey:CategoryID;constraint:OnDelete:RESTRICT" json:"category,omitempty"`

	// Price is the selling price in whole rupees. Money is stored as an integer
	// because paise are never fractional here and float rupees would eventually
	// produce a ₹2,999.9999999 on someone's screen.
	Price int `gorm:"not null;index" json:"price"`
	// MRP is the pre-discount list price in whole rupees.
	MRP int `gorm:"column:mrp;not null" json:"mrp"`
	// Discount is a whole percentage, derived from Price and MRP at seed time.
	Discount int `gorm:"not null;index" json:"discount"`

	Rating      float64 `gorm:"not null;index" json:"rating"`
	RatingCount int     `gorm:"not null" json:"ratingCount"`

	Gender   Gender   `gorm:"type:varchar(16);not null;index" json:"gender"`
	Occasion Occasion `gorm:"type:varchar(24);not null;index" json:"occasion"`
	Style    StyleTag `gorm:"type:varchar(24);not null;index" json:"style"`
	Color    string   `gorm:"type:varchar(48);not null;index" json:"color"`

	// SearchDoc is the pre-lowercased concatenation of the searchable fields.
	// It is maintained on write so that the text scorer and the SQL LIKE
	// prefilter never have to lower() a column at query time.
	SearchDoc string `gorm:"type:text;not null" json:"-"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName pins the table name so a future rename of the Go type cannot
// silently orphan the migrations.
func (Product) TableName() string { return "products" }

// BeforeSave assigns an ID when absent and keeps derived columns consistent.
// Doing this in a hook rather than at every call site means a product can never
// be written with a stale SearchDoc or an out-of-band discount.
func (p *Product) BeforeSave(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.SearchDoc = BuildSearchDoc(p.Title, p.Brand, p.Color, p.Description,
		string(p.Occasion), string(p.Style), string(p.Gender))
	if p.MRP > 0 {
		p.Discount = int(float64(p.MRP-p.Price)/float64(p.MRP)*100.0 + 0.5)
	}
	return nil
}

// EffectiveDiscount returns the discount percentage, recomputed defensively for
// rows that predate the BeforeSave hook.
func (p *Product) EffectiveDiscount() int {
	if p.MRP <= 0 || p.Price >= p.MRP {
		return 0
	}
	return int(float64(p.MRP-p.Price)/float64(p.MRP)*100.0 + 0.5)
}
