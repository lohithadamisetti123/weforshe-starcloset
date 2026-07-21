-- Migration 000004: personalization — user profiles and (mock) past orders.
--
-- Design notes:
--  1. Additive on top of 000001–000003. Never edits an earlier migration.
--  2. Both tables are user-scoped and cascade-delete with the owning user row,
--     exactly like favorite_products, so a deleted account leaves nothing
--     behind. Guests are real user rows, so this works for them too.
--  3. The array-valued preferences (styles/colors/brands/categories) are stored
--     as jsonb rather than as normalised join tables. They are read back whole
--     to build a scoring context, never queried by element, and their shape is
--     expected to grow (sizes, fits, fabrics) — the same trade the project
--     already makes for the recommendation intent and score breakdowns.
--  4. past_orders keeps a denormalised snapshot (title/brand/category/image/
--     price) *and* an optional product_id. The snapshot is what makes the row
--     survive being swapped for real Myntra purchase history later, where an
--     order need not map to a current catalogue row; the product_id is a soft
--     link (ON DELETE SET NULL) used only to enrich the demo.

BEGIN;

-- ---------------------------------------------------------------------------
-- user_profiles — one row per user, holding onboarding answers + preferences.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_profiles (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID        NOT NULL,
    name                 VARCHAR(120) NOT NULL DEFAULT '',
    gender               VARCHAR(16),
    city                 VARCHAR(120) NOT NULL DEFAULT '',
    budget_min           INTEGER,
    budget_max           INTEGER,
    favorite_styles      JSONB       NOT NULL DEFAULT '[]'::jsonb,
    favorite_colors      JSONB       NOT NULL DEFAULT '[]'::jsonb,
    favorite_brands      JSONB       NOT NULL DEFAULT '[]'::jsonb,
    favorite_categories  JSONB       NOT NULL DEFAULT '[]'::jsonb,
    onboarded            BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_profiles_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    -- One profile per user. The unique index is what makes the profile write
    -- an idempotent upsert (ON CONFLICT (user_id)) rather than a racy
    -- read-then-insert.
    CONSTRAINT uq_user_profiles_user UNIQUE (user_id),
    -- Mirror the gender vocabulary CHECK the products/users columns use, so a
    -- profile can never hold a value the rest of the system can't render.
    CONSTRAINT chk_user_profiles_gender
        CHECK (gender IS NULL OR gender IN ('Women', 'Men', 'Girls', 'Boys', 'Unisex')),
    CONSTRAINT chk_user_profiles_budget
        CHECK (
            (budget_min IS NULL OR budget_min >= 0) AND
            (budget_max IS NULL OR budget_max >= 0) AND
            (budget_min IS NULL OR budget_max IS NULL OR budget_max >= budget_min)
        )
);

-- ---------------------------------------------------------------------------
-- past_orders — a user's purchase history. Mock for the MVP; the schema is the
-- real one so replacing the seed with imported Myntra orders is a data change,
-- not a migration.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS past_orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL,
    product_id  UUID,
    title       VARCHAR(180) NOT NULL,
    brand       VARCHAR(80)  NOT NULL DEFAULT '',
    category    VARCHAR(48)  NOT NULL DEFAULT '',
    image       TEXT         NOT NULL DEFAULT '',
    price       INTEGER      NOT NULL DEFAULT 0,
    ordered_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_past_orders_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    -- Soft link: an order keeps its snapshot even if the catalogue row it was
    -- derived from is later removed. This is why real order history (which may
    -- reference products no longer sold) drops straight in.
    CONSTRAINT fk_past_orders_product
        FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_past_orders_user_ordered
    ON past_orders (user_id, ordered_at DESC);

COMMIT;
