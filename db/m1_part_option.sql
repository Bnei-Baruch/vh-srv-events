BEGIN;

ALTER TABLE participation_option
ADD description TEXT,
ADD content JSON;

COMMIT;