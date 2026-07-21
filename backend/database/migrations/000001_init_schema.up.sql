-- Myntra Bharat AI — initial schema.
--
-- The vocabulary CHECK constraints below intentionally mirror the frontend's
-- TypeScript union types. They are the last line of defence against a bad seed
-- or a rogue insert putting an unrenderable value like 'Semi-Formal' into a
-- column the UI switches on.

BEGIN;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
-- pg_trgm powers fuzzy ILIKE search on titles without a sequential scan.
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ---------------------------------------------------------------------------
-- users
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(120) NOT NULL,
    email         VARCHAR(180) NOT NULL,
    password_hash VARCHAR(120) NOT NULL,
    role          VARCHAR(16)  NOT NULL DEFAULT 'User',
    is_guest      BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    CONSTRAINT chk_users_role CHECK (role IN ('User', 'Admin'))
);

-- Partial unique index: a soft-deleted account must not block re-registration
-- with the same address, which a plain UNIQUE would.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active
    ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_is_guest  ON users (is_guest);

-- ---------------------------------------------------------------------------
-- categories
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(48) NOT NULL UNIQUE,
    slug        VARCHAR(48) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_categories_name CHECK (name IN (
        'Kurta Sets', 'Sarees', 'Dresses', 'Shirts', 'Tops',
        'Jeans', 'Blazers', 'Ethnic Wear', 'Footwear', 'Accessories'
    ))
);

-- ---------------------------------------------------------------------------
-- products
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS products (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug         VARCHAR(32)  NOT NULL UNIQUE,
    title        VARCHAR(180) NOT NULL,
    brand        VARCHAR(80)  NOT NULL,
    description  TEXT         NOT NULL,
    image        TEXT         NOT NULL,
    category_id  UUID         NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    price        INTEGER      NOT NULL,
    mrp          INTEGER      NOT NULL,
    discount     INTEGER      NOT NULL DEFAULT 0,
    rating       DOUBLE PRECISION NOT NULL DEFAULT 0,
    rating_count INTEGER      NOT NULL DEFAULT 0,
    gender       VARCHAR(16)  NOT NULL,
    occasion     VARCHAR(24)  NOT NULL,
    style        VARCHAR(24)  NOT NULL,
    color        VARCHAR(48)  NOT NULL,
    search_doc   TEXT         NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,

    -- Money is stored in whole rupees and can never be negative. MRP must be at
    -- least the selling price, or "discount" would be a markup.
    CONSTRAINT chk_products_price    CHECK (price > 0),
    CONSTRAINT chk_products_mrp      CHECK (mrp >= price),
    CONSTRAINT chk_products_discount CHECK (discount BETWEEN 0 AND 100),
    CONSTRAINT chk_products_rating   CHECK (rating BETWEEN 0 AND 5),
    CONSTRAINT chk_products_gender   CHECK (gender IN ('Women', 'Men', 'Unisex')),
    CONSTRAINT chk_products_occasion CHECK (occasion IN (
        'Wedding', 'Office', 'College', 'Party', 'Travel', 'Festival', 'Casual', 'Date Night'
    )),
    CONSTRAINT chk_products_style CHECK (style IN (
        'Ethnic', 'Western', 'Formal', 'Streetwear', 'Minimal', 'Bohemian', 'Athleisure', 'Indo-Western'
    ))
);

CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products (deleted_at);
CREATE INDEX IF NOT EXISTS idx_products_category   ON products (category_id);
CREATE INDEX IF NOT EXISTS idx_products_brand      ON products (brand);
CREATE INDEX IF NOT EXISTS idx_products_price      ON products (price);
CREATE INDEX IF NOT EXISTS idx_products_rating     ON products (rating DESC);
CREATE INDEX IF NOT EXISTS idx_products_discount   ON products (discount DESC);

-- The engine's single hottest access pattern is "candidates for a gender and
-- occasion, cheapest-relevant first". A composite index covers it in one scan.
CREATE INDEX IF NOT EXISTS idx_products_gender_occasion_price
    ON products (gender, occasion, price) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_products_style_gender
    ON products (style, gender) WHERE deleted_at IS NULL;

-- Trigram index for fuzzy free-text prefilters ("wedidng kurtaa").
CREATE INDEX IF NOT EXISTS idx_products_search_doc_trgm
    ON products USING GIN (search_doc gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_products_title_trgm
    ON products USING GIN (title gin_trgm_ops);

-- ---------------------------------------------------------------------------
-- recommendations
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS recommendations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    prompt       TEXT        NOT NULL,
    summary      TEXT        NOT NULL DEFAULT '',
    reason       TEXT        NOT NULL DEFAULT '',
    budget_label VARCHAR(48) NOT NULL DEFAULT '',
    occasion     VARCHAR(32) NOT NULL DEFAULT '',
    style        VARCHAR(32) NOT NULL DEFAULT '',
    source       VARCHAR(16) NOT NULL DEFAULT 'rules',
    intent       JSONB,
    latency_ms   INTEGER     NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    CONSTRAINT chk_recommendations_source CHECK (source IN ('rules', 'gemini', 'hybrid'))
);

CREATE INDEX IF NOT EXISTS idx_recommendations_deleted_at ON recommendations (deleted_at);
CREATE INDEX IF NOT EXISTS idx_rec_user_created
    ON recommendations (user_id, created_at DESC) WHERE deleted_at IS NULL;
-- GIN over the extracted intent, so relevance questions ("what do people ask
-- for that we have no stock for?") are a query rather than a data export.
CREATE INDEX IF NOT EXISTS idx_recommendations_intent ON recommendations USING GIN (intent);

CREATE TABLE IF NOT EXISTS recommendation_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recommendation_id UUID    NOT NULL REFERENCES recommendations (id) ON DELETE CASCADE,
    product_id        UUID    NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    rank              INTEGER NOT NULL,
    score             DOUBLE PRECISION NOT NULL DEFAULT 0,
    match_percent     INTEGER NOT NULL DEFAULT 0,
    breakdown         JSONB,
    reasons           JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_rec_items_rank  CHECK (rank > 0),
    CONSTRAINT chk_rec_items_match CHECK (match_percent BETWEEN 0 AND 100),
    CONSTRAINT uq_rec_items_rank UNIQUE (recommendation_id, rank)
);

CREATE INDEX IF NOT EXISTS idx_rec_items_recommendation ON recommendation_items (recommendation_id);
CREATE INDEX IF NOT EXISTS idx_rec_items_product        ON recommendation_items (product_id);

-- ---------------------------------------------------------------------------
-- chat_history
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS chat_history (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    prompt               TEXT        NOT NULL,
    response             TEXT        NOT NULL DEFAULT '',
    budget_label         VARCHAR(48) NOT NULL DEFAULT '',
    occasion             VARCHAR(32) NOT NULL DEFAULT '',
    style                VARCHAR(32) NOT NULL DEFAULT '',
    products_recommended INTEGER     NOT NULL DEFAULT 0,
    status               VARCHAR(16) NOT NULL DEFAULT 'Completed',
    recommendation_id    UUID REFERENCES recommendations (id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ,
    CONSTRAINT chk_chat_history_status CHECK (status IN ('Completed', 'Saved', 'Purchased'))
);

CREATE INDEX IF NOT EXISTS idx_chat_history_deleted_at ON chat_history (deleted_at);
CREATE INDEX IF NOT EXISTS idx_history_user_created
    ON chat_history (user_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_chat_history_recommendation ON chat_history (recommendation_id);

-- ---------------------------------------------------------------------------
-- favorite_products
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS favorite_products (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Makes re-favouriting idempotent in the database rather than in a racy
    -- read-then-write in the service layer.
    CONSTRAINT uq_fav_user_product UNIQUE (user_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_fav_user_created ON favorite_products (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_fav_product      ON favorite_products (product_id);

COMMIT;
