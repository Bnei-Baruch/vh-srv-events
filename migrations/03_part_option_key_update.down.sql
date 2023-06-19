BEGIN;

-- Drop the primary key constraint if it exists
DO $$
BEGIN
    BEGIN
        IF EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conrelid = 'participation_status'::regclass
            AND conname = 'participation_status_pkey'
        ) THEN
            ALTER TABLE participation_status DROP CONSTRAINT IF EXISTS participation_status_pkey;
        END IF;
    EXCEPTION
        WHEN undefined_table THEN NULL;
        WHEN undefined_object THEN NULL;
        WHEN invalid_table_definition THEN NULL;
    END;
END$$;

COMMIT;
