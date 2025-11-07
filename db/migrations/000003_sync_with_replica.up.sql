BEGIN;

-- Add published column to event table to match replica
ALTER TABLE event
ADD COLUMN IF NOT EXISTS published BOOLEAN DEFAULT true;

COMMIT;

