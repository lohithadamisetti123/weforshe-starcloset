-- StarCloset RBAC: introduce the PARTNER role.
--
-- Stylist partners may manage personas through the partner portal but hold no
-- other elevated access. The users.role CHECK is widened to admit 'Partner'
-- alongside the existing 'User' and 'Admin'. This is additive: every existing
-- row remains valid, and nothing outside the persona portal changes.

BEGIN;

ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;
ALTER TABLE users ADD CONSTRAINT chk_users_role
    CHECK (role IN ('User', 'Partner', 'Admin'));

COMMIT;
