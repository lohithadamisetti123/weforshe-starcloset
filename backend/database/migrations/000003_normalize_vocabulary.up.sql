-- Migration 003: Normalize style vocabulary to 6 canonical values.
--
-- The 60+ style variants in migration 002 produced poor recommendations
-- because the scorer couldn't distinguish ethnic wedding wear from casual
-- dresses. This migration:
--   1. Maps every product's style to one of 6 canonical values
--   2. Tightens the CHECK constraint to enforce the clean vocabulary
--
-- The mapping is:
--   Ethnic    ← Traditional, Contemporary Ethnic, Luxury Ethnic, Festive,
--               Chikankari, Bandhani, Mirror Work, Gota Patti, Embroidered,
--               Jacquard, Organza, Chanderi, Georgette, Sharara, Gharara,
--               Angrakha, Lucknowi
--   Formal    ← Elegant, Classic, Luxury, Premium, Velvet, Silk,
--               Business Casual, Slim Fit, Party
--   Casual    ← Minimal, Minimalist, Smart Casual, Streetwear, Bohemian,
--               Boho, Floral, Printed, Graphic, Resort, Linen, Cotton, Denim,
--               Chiffon, Suede, Leather, Vintage, Modern, Lounge, Contemporary,
--               Oversized, Regular Fit, Relaxed Fit, Natural, Matte, Glossy,
--               Scandinavian, Industrial, Traditional Decor, Sporty (→ Athleisure)

BEGIN;

-- Step 1: Widen the column temporarily to allow all existing values
ALTER TABLE products ALTER COLUMN style TYPE VARCHAR(128);

-- Step 2: Drop the old constraint
ALTER TABLE products DROP CONSTRAINT IF EXISTS chk_products_style;

-- Step 3: Remap all non-canonical style values
UPDATE products SET style = 'Ethnic' WHERE style IN (
    'Traditional','Contemporary Ethnic','Luxury Ethnic','Festive',
    'Chikankari','Bandhani','Mirror Work','Gota Patti','Embroidered',
    'Jacquard','Lucknowi','Organza','Sharara','Gharara','Angrakha',
    'Chanderi','Georgette'
);

UPDATE products SET style = 'Formal' WHERE style IN (
    'Elegant','Classic','Luxury','Premium','Velvet','Silk',
    'Business Casual','Slim Fit','Party'
);

UPDATE products SET style = 'Casual' WHERE style IN (
    'Minimal','Minimalist','Smart Casual','Streetwear','Bohemian','Boho',
    'Floral','Printed','Graphic','Resort','Linen','Cotton','Denim',
    'Chiffon','Suede','Leather','Vintage','Modern','Lounge','Contemporary',
    'Oversized','Regular Fit','Relaxed Fit','Natural','Matte','Glossy',
    'Scandinavian','Industrial','Traditional Decor'
);

UPDATE products SET style = 'Athleisure' WHERE style IN (
    'Sporty'
);

-- Catch-all: anything still unrecognised defaults to Casual
UPDATE products SET style = 'Casual'
WHERE style NOT IN ('Ethnic','Western','Formal','Casual','Indo-Western','Athleisure');

-- Step 4: Re-add the tight CHECK constraint
ALTER TABLE products
    ADD CONSTRAINT chk_products_style CHECK (style IN (
        'Ethnic','Western','Formal','Casual','Indo-Western','Athleisure'
    ));

-- Step 5: Trim column back
ALTER TABLE products ALTER COLUMN style TYPE VARCHAR(64);

COMMIT;
