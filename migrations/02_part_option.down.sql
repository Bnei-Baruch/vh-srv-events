BEGIN;
DO
$$
BEGIN
    -- Check if the columns exist before removing them
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'participation_option'
          AND column_name = 'description'
    )
    THEN
        ALTER TABLE participation_option DROP COLUMN description;
    END IF;
    
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'participation_option'
          AND column_name = 'content'
    )
    THEN
        ALTER TABLE participation_option DROP COLUMN content;
    END IF;
END;
$$;
COMMIT;
