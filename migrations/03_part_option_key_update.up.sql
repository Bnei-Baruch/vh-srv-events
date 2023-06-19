BEGIN;

-- Drop the primary key constraint if it exists
DO $$BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'participation_status'::regclass
        AND conname = 'participation_status_pkey'
    ) THEN
        ALTER TABLE participation_status DROP CONSTRAINT participation_status_pkey;
    END IF;
END$$;

-- Add the primary key constraint if it does not exist
DO $$BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'participation_status'::regclass
        AND conname = 'participation_status_pkey'
    ) THEN
        ALTER TABLE participation_status ADD PRIMARY KEY (participant_id, event_id);
    END IF;
END$$;

COMMIT;
