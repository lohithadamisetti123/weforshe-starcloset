-- Revert the PARTNER role. Any Partner accounts must be reassigned before this
-- runs, or the re-added CHECK will reject them.

BEGIN;

UPDATE users SET role = 'User' WHERE role = 'Partner';

ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;
ALTER TABLE users ADD CONSTRAINT chk_users_role
    CHECK (role IN ('User', 'Admin'));

COMMIT;
