BEGIN;

ALTER TABLE participation_status DROP CONSTRAINT participation_status_pkey;
ALTER TABLE participation_status ADD PRIMARY KEY (participant_id, event_id);

COMMIT;