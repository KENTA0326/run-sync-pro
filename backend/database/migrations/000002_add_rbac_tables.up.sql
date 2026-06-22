SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '60s';

CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (resource, action)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions (resource, action);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions (permission_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles (role_id);

INSERT INTO roles (name, description) VALUES
    ('viewer', '閲覧のみ'),
    ('editor', '作成・更新・削除が可能'),
    ('admin', '全権限')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (resource, action, description) VALUES
    ('shoe', 'read', 'シューズ閲覧'),
    ('shoe', 'write', 'シューズ作成・削除'),
    ('training_log', 'read', '走行ログ閲覧'),
    ('training_log', 'write', '走行ログ作成・インポート'),
    ('analysis', 'read', '分析結果閲覧'),
    ('user_profile', 'read', '自分のプロフィール閲覧'),
    ('*', '*', '全リソース全操作')
ON CONFLICT (resource, action) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON
    (r.name = 'viewer' AND (p.resource, p.action) IN (('shoe', 'read'), ('training_log', 'read'), ('analysis', 'read'), ('user_profile', 'read')))
    OR
    (r.name = 'editor' AND (p.resource, p.action) IN (('shoe', 'read'), ('shoe', 'write'), ('training_log', 'read'), ('training_log', 'write'), ('analysis', 'read'), ('user_profile', 'read')))
    OR
    (r.name = 'admin' AND (p.resource, p.action) = ('*', '*'))
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.name = 'viewer'
ON CONFLICT (user_id, role_id) DO NOTHING;
