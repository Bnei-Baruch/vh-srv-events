BEGIN;
DO
$$
BEGIN
    ALTER TABLE IF EXISTS participation_option ADD COLUMN IF NOT EXISTS description TEXT;
    ALTER TABLE IF EXISTS participation_option ADD COLUMN IF NOT EXISTS content JSON;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END;
$$;
COMMIT;
