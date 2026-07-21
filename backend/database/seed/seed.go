// Package seed populates the catalogue.
//
// Seeding is idempotent and safe to run on every boot. That property is what
// makes `docker compose up` a one-command demo: the container starts, the
// schema migrates, the catalogue fills, and a second `up` changes nothing.
// A seeder that duplicates rows on re-run would make the compose file a
// single-use artefact.
package seed

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/myntra/bharat-ai-backend/internal/models"
	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// productsJSON is the catalogue fixture, embedded into the binary.
//
// It is a direct port of the frontend's `src/data/products.ts`, with the
// original 40 products preserved verbatim — same titles, prices, ratings and
// image seeds — plus 12 added to reach the brief's 50 and to fill thin spots in
// the occasion mix (Men's ethnic, College, Festival). Keeping the originals
// byte-identical means the backend-served grid looks exactly like the mocked
// one, which makes the cutover a genuine A/B rather than a leap of faith.
//
//go:embed products.json
var productsJSON []byte

// seedProduct is the fixture's shape.
type seedProduct struct {
	UUID        string  `json:"uuid"`
	Slug        string  `json:"id"`
	Title       string  `json:"title"`
	Brand       string  `json:"brand"`
	Category    string  `json:"category"`
	Price       int     `json:"price"`
	MRP         int     `json:"mrp"`
	Discount    int     `json:"discount"`
	Rating      float64 `json:"rating"`
	RatingCount int     `json:"ratingCount"`
	Image       string  `json:"image"`
	Gender      string  `json:"gender"`
	Occasion    string  `json:"occasion"`
	Style       string  `json:"style"`
	Color       string  `json:"color"`
	Description string  `json:"description"`
}

// categoryDescriptions gives each garment family a line of merchandising copy.
var categoryDescriptions = map[models.CategoryName]string{
	// Apparel
	models.CategoryKurtaSets:  "Coordinated kurta sets, from everyday cotton to occasion-ready anarkalis.",
	models.CategorySarees:     "Handloom, silk and contemporary drapes for weddings and festivals.",
	models.CategoryDresses:    "Western silhouettes for parties, dates and day-to-day wear.",
	models.CategoryShirts:     "Formal and casual shirting for work, travel and weekends.",
	models.CategoryTops:       "Tees, blouses, crop tops and skirts for layering and everyday wear.",
	models.CategoryJeans:      "Denim, chinos, trousers and joggers across every fit.",
	models.CategoryBlazers:    "Tailored jackets and blazers that sharpen a look instantly.",
	models.CategoryEthnicWear: "Sherwanis, lehengas and Indo-Western pieces for the big occasions.",
	// Verticals / departments
	models.CategoryWomen:          "Complete women's fashion — clothing, footwear and accessories.",
	models.CategoryMen:            "Complete men's fashion — clothing, footwear and accessories.",
	models.CategoryKidsGirls:      "Dresses, sets and accessories for girls.",
	models.CategoryKidsBoys:       "Shirts, trousers and accessories for boys.",
	models.CategoryFootwear:       "Sneakers, heels, loafers and sandals to finish the outfit.",
	models.CategoryAccessories:    "Watches, bags, belts and jewellery — the details that carry a look.",
	models.CategoryBeauty:         "Skincare, makeup and fragrances for everyday and special occasions.",
	models.CategoryHomeEssentials: "Bedding, décor, kitchenware and organizers for every home.",
}

// Seeder populates the database.
type Seeder struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewSeeder constructs a Seeder.
func NewSeeder(db *gorm.DB, log *zap.Logger) *Seeder {
	return &Seeder{db: db, log: log}
}

// Run seeds categories, products and the demo accounts.
//
// Everything happens in one transaction: a partially seeded catalogue — say,
// categories inserted and products failed — would leave the app in a state
// where the grid renders empty and nothing indicates why.
func (s *Seeder) Run(ctx context.Context) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		categories, err := s.seedCategories(tx)
		if err != nil {
			return err
		}
		if err := s.seedProducts(tx, categories); err != nil {
			return err
		}
		return s.seedUsers(tx)
	})
}

// seedCategories inserts the garment families and returns a name->ID lookup.
func (s *Seeder) seedCategories(tx *gorm.DB) (map[models.CategoryName]uuid.UUID, error) {
	lookup := make(map[models.CategoryName]uuid.UUID, len(models.AllCategories))

	for _, name := range models.AllCategories {
		cat := models.Category{
			// A deterministic UUID derived from the name, so re-seeding — and
			// seeding a fresh environment — always yields the same ID. Random
			// IDs here would make the product fixture's category references
			// unstable across environments.
			ID:          deterministicUUID("category:" + string(name)),
			Name:        name,
			Slug:        utils.Slugify(string(name)),
			Description: categoryDescriptions[name],
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"slug", "description", "updated_at"}),
		}).Create(&cat).Error; err != nil {
			return nil, fmt.Errorf("seed: upserting category %q: %w", name, err)
		}
		lookup[name] = cat.ID
	}

	s.log.Info("categories seeded", zap.Int("count", len(lookup)))
	return lookup, nil
}

// seedProducts inserts the catalogue.
func (s *Seeder) seedProducts(tx *gorm.DB, categories map[models.CategoryName]uuid.UUID) error {
	var fixtures []seedProduct
	if err := json.Unmarshal(productsJSON, &fixtures); err != nil {
		return fmt.Errorf("seed: parsing products.json: %w", err)
	}

	products := make([]models.Product, 0, len(fixtures))
	for _, f := range fixtures {
		// The fixture is validated rather than trusted. A typo'd enum here
		// would otherwise fail later against a CHECK constraint with an error
		// that names a column, not the product that caused it.
		category, ok := models.ParseCategory(f.Category)
		if !ok {
			return fmt.Errorf("seed: product %q has unknown category %q", f.Slug, f.Category)
		}
		categoryID, ok := categories[category]
		if !ok {
			return fmt.Errorf("seed: product %q references unseeded category %q", f.Slug, category)
		}
		gender, ok := models.ParseGender(f.Gender)
		if !ok {
			return fmt.Errorf("seed: product %q has unknown gender %q", f.Slug, f.Gender)
		}
		occasion, ok := models.ParseOccasion(f.Occasion)
		if !ok {
			return fmt.Errorf("seed: product %q has unknown occasion %q", f.Slug, f.Occasion)
		}
		style, ok := models.ParseStyle(f.Style)
		if !ok {
			return fmt.Errorf("seed: product %q has unknown style %q", f.Slug, f.Style)
		}
		id, err := uuid.Parse(f.UUID)
		if err != nil {
			return fmt.Errorf("seed: product %q has invalid uuid: %w", f.Slug, err)
		}
		if f.MRP < f.Price {
			return fmt.Errorf("seed: product %q has mrp below price", f.Slug)
		}

		p := models.Product{
			ID:          id,
			Slug:        f.Slug,
			Title:       f.Title,
			Brand:       f.Brand,
			Description: f.Description,
			Image:       f.Image,
			CategoryID:  categoryID,
			Price:       f.Price,
			MRP:         f.MRP,
			Discount:    f.Discount,
			Rating:      f.Rating,
			RatingCount: f.RatingCount,
			Gender:      gender,
			Occasion:    occasion,
			Style:       style,
			Color:       f.Color,
		}
		// Populate search_doc and re-derive discount exactly as a live write
		// would, so seeded rows are indistinguishable from application-written
		// ones. BeforeSave does not fire for the batch insert below.
		_ = p.BeforeSave(tx)
		products = append(products, p)
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"title", "brand", "description", "image", "category_id", "price",
			"mrp", "discount", "rating", "rating_count", "gender", "occasion",
			"style", "color", "search_doc", "updated_at",
		}),
	}).CreateInBatches(&products, 50).Error; err != nil {
		return fmt.Errorf("seed: upserting products: %w", err)
	}

	s.log.Info("products seeded", zap.Int("count", len(products)))
	return nil
}

// demoUsers are the accounts the Postman collection and the README log in with.
var demoUsers = []struct {
	Name     string
	Email    string
	Password string
	Role     models.Role
}{
	{"Asha Menon", "asha@myntra.dev", "bharat-ai-demo", models.RoleUser},
	{"Admin", "admin@myntra.dev", "bharat-ai-admin", models.RoleAdmin},
}

// seedUsers inserts the demo accounts.
//
// These exist so a reviewer can exercise the authenticated routes without
// registering first. The passwords are weak, published in the README, and
// entirely appropriate for a seeded demo — but the ENV guard in cmd/server
// keeps AutoSeed off in production, which is what stops them from ever being
// real credentials.
//
// This is a plain check-then-insert rather than an ON CONFLICT upsert on
// purpose. users.email is backed by a *partial* unique index
// (idx_users_email_active, WHERE deleted_at IS NULL — see
// database/migrations) so that a soft-deleted account doesn't block
// re-registration, and Postgres will only accept a partial index as an ON
// CONFLICT arbiter if the statement's own WHERE clause repeats that
// predicate exactly. That's a sharp edge worth avoiding entirely in a
// once-at-boot seeder: there's no concurrent writer racing this function, so
// the tiny check-then-insert window is a non-issue, and the SQL stays
// obviously correct instead of depending on exact arbiter-matching syntax.
func (s *Seeder) seedUsers(tx *gorm.DB) error {
	for _, d := range demoUsers {
		var existing models.User
		err := tx.Where("email = ?", d.Email).First(&existing).Error
		if err == nil {
			// Already seeded on a previous boot. Left alone rather than
			// upserted, so an operator's password change survives a restart.
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("seed: checking for existing user %s: %w", d.Email, err)
		}

		hash, err := utils.HashPassword(d.Password)
		if err != nil {
			return fmt.Errorf("seed: hashing password for %s: %w", d.Email, err)
		}

		user := models.User{
			ID:           deterministicUUID("user:" + d.Email),
			Name:         d.Name,
			Email:        d.Email,
			PasswordHash: hash,
			Role:         d.Role,
		}
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("seed: creating user %s: %w", d.Email, err)
		}
	}

	s.log.Info("demo users seeded", zap.Int("count", len(demoUsers)))
	return nil
}

// seedNamespace is the UUIDv5 namespace for all derived IDs. It is a fixed
// random UUID; its only requirement is that it never changes, because changing
// it would re-key the entire catalogue.
var seedNamespace = uuid.MustParse("6f9b2c1e-0a4d-5e7f-8b3a-1d2c3e4f5a6b")

// deterministicUUID derives a stable UUIDv5 from a name.
func deterministicUUID(name string) uuid.UUID {
	return uuid.NewSHA1(seedNamespace, []byte(name))
}
