-- StarCloset V2: per-department unlock progress.
--
-- Each shopper carries an independent progress counter for every department
-- (Women / Men / Kids). Unlocking a stylist in one department must never affect
-- another, so progress is stored one row per (user, department) rather than a
-- single shared counter. The recommendation engine and cart are untouched — this
-- table only records how far a shopper has progressed in each journey.

BEGIN;

CREATE TABLE IF NOT EXISTS star_progress (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    department  VARCHAR(16) NOT NULL,
    items       INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_star_progress_items CHECK (items >= 0),
    CONSTRAINT chk_star_progress_department CHECK (department IN ('Women', 'Men', 'Kids')),
    CONSTRAINT uq_star_progress_user_dept UNIQUE (user_id, department)
);

CREATE INDEX IF NOT EXISTS idx_star_progress_user ON star_progress (user_id);

COMMIT;
