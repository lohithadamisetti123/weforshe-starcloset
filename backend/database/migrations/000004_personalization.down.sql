-- Revert migration 000004 — drop the personalization tables.
BEGIN;

DROP TABLE IF EXISTS past_orders;
DROP TABLE IF EXISTS user_profiles;

COMMIT;
