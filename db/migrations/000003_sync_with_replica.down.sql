BEGIN;

-- Remove published column from event table
ALTER TABLE event
DROP COLUMN IF EXISTS published;

COMMIT;

