-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT,
    email TEXT NOT NULL,
    password TEXT,
    role INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

-- シューズテーブル
CREATE TABLE IF NOT EXISTS shoes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    brand TEXT,
    model TEXT,
    purchase_date DATE NOT NULL,
    total_distance DOUBLE PRECISION DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_shoes_user_id ON shoes (user_id);
CREATE INDEX IF NOT EXISTS idx_shoes_deleted_at ON shoes (deleted_at);

-- 走行ログテーブル
CREATE TABLE IF NOT EXISTS training_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    training_date DATE NOT NULL,
    distance DOUBLE PRECISION NOT NULL,
    duration INTEGER NOT NULL,
    pace TEXT,
    memo TEXT,
    kind INTEGER NOT NULL,
    shoe_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_training_logs_user_id ON training_logs (user_id);
CREATE INDEX IF NOT EXISTS idx_training_logs_shoe_id ON training_logs (shoe_id);
CREATE INDEX IF NOT EXISTS idx_training_logs_deleted_at ON training_logs (deleted_at);
