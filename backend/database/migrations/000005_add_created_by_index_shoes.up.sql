-- 1ファイル1文: CONCURRENTLY はトランザクション外で実行（golang-migrate の postgres ドライバ仕様）
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shoes_created_by ON shoes (created_by);
