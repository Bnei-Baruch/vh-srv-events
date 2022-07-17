BEGIN;

ALTER TABLE event
ADD archive_link TEXT;

COMMIT;