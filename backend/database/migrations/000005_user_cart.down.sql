-- Revert migration 000005 — drop the per-user cart and collection stores.
BEGIN;

DROP TABLE IF EXISTS collection_cart_items;
DROP TABLE IF EXISTS saved_collections;
DROP TABLE IF EXISTS cart_items;

COMMIT;
