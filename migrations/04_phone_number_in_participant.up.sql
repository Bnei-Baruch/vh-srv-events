BEGIN;
DO
$$
BEGIN
    BEGIN
        ALTER TABLE participant ADD COLUMN IF NOT EXISTS phone_number TEXT;
    EXCEPTION
        WHEN duplicate_column THEN NULL;
    END;
END;
$$;
COMMIT;
