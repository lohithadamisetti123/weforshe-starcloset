-- Migration 000002: expand vocabulary to match vocabulary.go v2
--
-- Rules:
--  1. Never edit 000001. This migration is additive on top of it.
--  2. DROP CONSTRAINT + ADD CONSTRAINT is the only portable way to widen an
--     IN-list CHECK in PostgreSQL (ALTER TABLE ... ALTER CONSTRAINT does not
--     exist for CHECK). We do it inside a single transaction so either
--     everything succeeds or nothing changes.
--  3. The VARCHAR columns that store occasion / style are widened first
--     because some new values ("Wedding Reception", "Ganesh Chaturthi",
--     "Traditional Decor") are longer than the old VARCHAR(24).

BEGIN;

-- ── products table ─────────────────────────────────────────────────────────

ALTER TABLE products
    ALTER COLUMN gender    TYPE VARCHAR(16),
    ALTER COLUMN occasion  TYPE VARCHAR(64),
    ALTER COLUMN style     TYPE VARCHAR(64);

ALTER TABLE products
    DROP CONSTRAINT IF EXISTS chk_products_gender,
    DROP CONSTRAINT IF EXISTS chk_products_occasion,
    DROP CONSTRAINT IF EXISTS chk_products_style;

ALTER TABLE products
    ADD CONSTRAINT chk_products_gender CHECK (gender IN (
        'Women','Men','Girls','Boys','Unisex'
    )),
    ADD CONSTRAINT chk_products_occasion CHECK (occasion IN (
        'Wedding','Wedding Reception','Reception',
        'Engagement','Engagement Party',
        'Sangeet','Mehendi','Haldi','Winter Wedding',
        'Festival','Diwali','Navratri','Karva Chauth',
        'Eid','Christmas','New Year','Pongal','Onam',
        'Durga Puja','Ganesh Chaturthi','Puja','Temple Visit',
        'Spring Festival','Festive Dinner','Cultural Event',
        'Office','Office Party','Business Meeting',
        'Interview','Corporate Event','Conference',
        'College','School','Graduation','Farewell','Freshers',
        'Casual','Daily Wear','Travel','Summer Vacation',
        'Weekend Outing','Weekend Brunch','Brunch',
        'Picnic','Shopping','Staycation',
        'Work From Home','Lounge','Beach','Resort',
        'Party','Cocktail Party','Birthday Party',
        'Dinner Date','Date Night','Cafe Date','Night Out',
        'Family Gathering','Housewarming','Baby Shower',
        'Anniversary','Anniversary Dinner',
        'Gym','Sports','Running','Yoga',
        'Kids Party','Playtime',
        'Self Care'
    )),
    ADD CONSTRAINT chk_products_style CHECK (style IN (
        'Ethnic','Western','Formal','Streetwear',
        'Minimal','Bohemian','Athleisure','Indo-Western',
        'Traditional','Contemporary Ethnic','Luxury Ethnic',
        'Festive','Elegant','Casual',
        'Smart Casual','Business Casual',
        'Oversized','Slim Fit','Regular Fit','Relaxed Fit',
        'Floral','Printed','Graphic','Embroidered',
        'Chikankari','Bandhani','Mirror Work','Gota Patti',
        'Jacquard','Lucknowi','Organza','Sharara','Gharara','Angrakha',
        'Chanderi','Georgette','Chiffon','Velvet','Silk',
        'Cotton','Linen','Denim','Leather','Suede',
        'Vintage','Classic','Modern','Luxury','Premium',
        'Minimalist','Resort','Lounge','Sporty','Boho','Contemporary',
        'Traditional Decor','Industrial','Natural','Matte','Glossy','Scandinavian'
    ));

-- ── categories table ───────────────────────────────────────────────────────

ALTER TABLE categories
    DROP CONSTRAINT IF EXISTS chk_categories_name;

ALTER TABLE categories
    ADD CONSTRAINT chk_categories_name CHECK (name IN (
        'Kurta Sets','Sarees','Dresses','Shirts','Tops','Jeans',
        'Blazers','Ethnic Wear',
        'Women','Men','Kids Girls','Kids Boys',
        'Footwear','Accessories','Beauty','Home Essentials'
    ));

-- ── recommendations table ──────────────────────────────────────────────────

ALTER TABLE recommendations
    ALTER COLUMN occasion TYPE VARCHAR(64),
    ALTER COLUMN style    TYPE VARCHAR(64);

-- ── chat_history table ─────────────────────────────────────────────────────

ALTER TABLE chat_history
    ALTER COLUMN occasion TYPE VARCHAR(64),
    ALTER COLUMN style    TYPE VARCHAR(64);

-- Rebuild the hot composite index — occasion column is now wider so the old
-- index definition still works but we drop+recreate to pick up any planner
-- statistics that aged out.
DROP INDEX IF EXISTS idx_products_gender_occasion_price;
CREATE INDEX idx_products_gender_occasion_price
    ON products (gender, occasion, price) WHERE deleted_at IS NULL;

COMMIT;
