BEGIN;

-- Drop the phone_number column if it exists
DO $$
BEGIN
    BEGIN
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'participant'
            AND column_name = 'phone_number'
        ) THEN
            ALTER TABLE participant DROP COLUMN IF EXISTS phone_number;
        END IF;
    EXCEPTION
        WHEN undefined_table THEN NULL;
        WHEN undefined_column THEN NULL;
        WHEN invalid_table_definition THEN NULL;
    END;
END$$;

COMMIT;
