-- Revert migration 000002 — restore the original 8-value constraints.
-- WARNING: any products using the new vocabulary values will fail this revert.

BEGIN;

ALTER TABLE products
    DROP CONSTRAINT IF EXISTS chk_products_gender,
    DROP CONSTRAINT IF EXISTS chk_products_occasion,
    DROP CONSTRAINT IF EXISTS chk_products_style;

ALTER TABLE products
    ADD CONSTRAINT chk_products_gender   CHECK (gender  IN ('Women','Men','Unisex')),
    ADD CONSTRAINT chk_products_occasion CHECK (occasion IN (
        'Wedding','Office','College','Party','Travel','Festival','Casual','Date Night'
    )),
    ADD CONSTRAINT chk_products_style    CHECK (style IN (
        'Ethnic','Western','Formal','Streetwear','Minimal','Bohemian','Athleisure','Indo-Western'
    ));

ALTER TABLE categories
    DROP CONSTRAINT IF EXISTS chk_categories_name;

ALTER TABLE categories
    ADD CONSTRAINT chk_categories_name CHECK (name IN (
        'Kurta Sets','Sarees','Dresses','Shirts','Tops','Jeans',
        'Blazers','Ethnic Wear','Footwear','Accessories'
    ));

COMMIT;
