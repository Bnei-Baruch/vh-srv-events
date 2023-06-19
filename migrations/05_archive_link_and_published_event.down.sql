BEGIN;

-- Drop the archive_link column if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'event'
        AND column_name = 'archive_link'
    ) THEN
        ALTER TABLE event DROP COLUMN archive_link;
    END IF;
EXCEPTION
    WHEN undefined_table THEN NULL;
    WHEN undefined_column THEN NULL;
    WHEN invalid_table_definition THEN NULL;
END$$;

-- Drop the published column if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'event'
        AND column_name = 'published'
    ) THEN
        ALTER TABLE event DROP COLUMN published;
    END IF;
EXCEPTION
    WHEN undefined_table THEN NULL;
    WHEN undefined_column THEN NULL;
    WHEN invalid_table_definition THEN NULL;
END$$;

COMMIT;
