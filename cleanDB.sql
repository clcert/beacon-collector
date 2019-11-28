DELETE FROM events_collected;
DELETE FROM external_events;
ALTER SEQUENCE events_id_seq RESTART WITH 1;
ALTER SEQUENCE external_events_id_seq RESTART WITH 1;