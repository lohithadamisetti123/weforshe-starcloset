-- Migration 000005: per-user cart, saved collections, and collection cart.
--
-- Why this exists: these three stores previously lived in browser localStorage
-- under a single global key each, so every account sharing a browser shared one
-- cart and one wishlist. Ownership is not a client concern — it is enforced
-- here, by a user_id column on every row, and in the repositories, which scope
-- every read *and* every write by the authenticated user id.
--
-- Design notes mirror favorite_products (000001-000003):
--  1. user_id is a CASCADE foreign key, so deleting an account removes its cart
--     and saved collections with it — no orphans.
--  2. A unique index over (user_id, ...) is what makes "add" idempotent at the
--     database level instead of in a racy read-then-write, and it is scoped to
--     the user so two accounts can hold the same product/collection
--     independently.
--  3. Collections are stored as a jsonb snapshot of the exact payload the UI
--     renders. A collection is an AI-generated arrangement of products, not a
--     catalogue entity, and re-deriving one would produce a *different*
--     collection; the snapshot is what makes a saved look reproducible.

BEGIN;

-- ---------------------------------------------------------------------------
-- cart_items — a user's product cart.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cart_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    product_id UUID        NOT NULL,
    quantity   INTEGER     NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_cart_items_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_cart_items_product
        FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
    -- One row per product per user; adding again bumps the quantity.
    CONSTRAINT uq_cart_items_user_product UNIQUE (user_id, product_id),
    CONSTRAINT chk_cart_items_quantity CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_cart_items_user_created
    ON cart_items (user_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- saved_collections — a user's saved AI collections (the "Saved" page).
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS saved_collections (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    name       VARCHAR(180) NOT NULL,
    department VARCHAR(48)  NOT NULL DEFAULT '',
    prompt     TEXT         NOT NULL DEFAULT '',
    collection JSONB        NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_saved_collections_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    -- Scoped to the user, so two accounts can each save the same collection.
    CONSTRAINT uq_saved_collections_user_name_dept UNIQUE (user_id, name, department)
);

CREATE INDEX IF NOT EXISTS idx_saved_collections_user_created
    ON saved_collections (user_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- collection_cart_items — whole collections added to a user's cart.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS collection_cart_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    name       VARCHAR(180) NOT NULL,
    department VARCHAR(48)  NOT NULL DEFAULT '',
    collection JSONB        NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_collection_cart_items_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_collection_cart_user_name_dept UNIQUE (user_id, name, department)
);

CREATE INDEX IF NOT EXISTS idx_collection_cart_user_created
    ON collection_cart_items (user_id, created_at DESC);

COMMIT;
