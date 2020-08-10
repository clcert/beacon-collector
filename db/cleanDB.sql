DELETE FROM events;
DELETE FROM external_values;
ALTER SEQUENCE events_id_seq RESTART WITH 1;
ALTER SEQUENCE external_values_id_seq RESTART WITH 1;