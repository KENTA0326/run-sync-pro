SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '60s';

ALTER TABLE training_logs DROP COLUMN IF EXISTS created_by;
ALTER TABLE shoes DROP COLUMN IF EXISTS created_by;
