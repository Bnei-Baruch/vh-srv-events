BEGIN;

ALTER TABLE event
ADD archive_link TEXT,
ADD published BOOLEAN DEFAULT true;

COMMIT;