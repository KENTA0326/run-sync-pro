-- 段階1: カラム追加 + バックフィル（索引は 000005/000006 で CONCURRENTLY）
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '60s';

ALTER TABLE shoes ADD COLUMN IF NOT EXISTS created_by BIGINT;
ALTER TABLE training_logs ADD COLUMN IF NOT EXISTS created_by BIGINT;

UPDATE shoes SET created_by = user_id WHERE created_by IS NULL;
UPDATE training_logs SET created_by = user_id WHERE created_by IS NULL;
