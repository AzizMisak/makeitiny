-- +goose Up
-- +goose StatementBegin

-- 1. RBAC: Roller (roles) ve Kullanıcı-Rol İlişkisi (users_roles)
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (id, name, description) VALUES
    (1, 'user', 'Standart platform kullanıcısı'),
    (2, 'admin', 'Sistem yöneticisi'),
    (3, 'owner', 'Sistemin tek sahibi (Owner)')
ON CONFLICT (name) DO NOTHING;

SELECT setval('roles_id_seq', 3, true);

CREATE TABLE users_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_users_roles_role_id ON users_roles(role_id);

-- Sistemde 'owner' (role_id = 3) rolüne sahip yalnızca 1 kullanıcı olabilir
CREATE UNIQUE INDEX idx_single_owner_role ON users_roles(role_id) WHERE role_id = 3;

-- 2. Üyelik Paketleri (plans)
CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    price_monthly NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    max_links INT NOT NULL DEFAULT 50,
    allow_custom_alias BOOLEAN NOT NULL DEFAULT FALSE,
    allow_password BOOLEAN NOT NULL DEFAULT FALSE,
    allow_qr BOOLEAN NOT NULL DEFAULT FALSE,
    allow_ads BOOLEAN NOT NULL DEFAULT FALSE,
    analytics_days INT NOT NULL DEFAULT 7,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO plans (id, slug, name, price_monthly, max_links, allow_custom_alias, allow_password, allow_qr, allow_ads, analytics_days) VALUES
    (1, 'free', 'Ücretsiz Paket', 0.00, 50, FALSE, TRUE, FALSE, FALSE, 7),
    (2, 'pro', 'Pro Paket', 99.00, 1000, TRUE, TRUE, TRUE, TRUE, 90),
    (3, 'enterprise', 'Kurumsal / Sınırsız', 299.00, -1, TRUE, TRUE, TRUE, TRUE, 365)
ON CONFLICT (slug) DO NOTHING;

SELECT setval('plans_id_seq', 3, true);

-- 3. Kullanıcılar (users) Tablosu Güncellemeleri
ALTER TABLE users
    ADD COLUMN plan_id INT NOT NULL DEFAULT 1 REFERENCES plans(id),
    ADD COLUMN plan_expires_at TIMESTAMPTZ NULL,
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- 4. Linkler (links) Tablosu Güncellemeleri
ALTER TABLE links
    ALTER COLUMN user_id DROP NOT NULL,
    ADD COLUMN public_id VARCHAR(16) NOT NULL DEFAULT substr(replace(gen_random_uuid()::text, '-', ''), 1, 16) UNIQUE,
    ADD COLUMN title TEXT NOT NULL DEFAULT '',
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN expires_at TIMESTAMPTZ NULL,
    ADD COLUMN max_clicks BIGINT NULL,
    ADD COLUMN ad_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN security_status TEXT NOT NULL DEFAULT 'unchecked',
    ADD COLUMN security_checked_at TIMESTAMPTZ NULL,
    ADD COLUMN qr_code_path TEXT NOT NULL DEFAULT '',
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- 5. Loglar (logs) Tablosu Güncellemeleri
ALTER TABLE logs
    ADD COLUMN country_code TEXT NOT NULL DEFAULT '';

-- 6. Mevcut Veritabanını Temizle ve Tek Owner Hesabını (admin@admin.com / admin) Oluştur
TRUNCATE TABLE logs, links, users_roles, users RESTART IDENTITY CASCADE;

INSERT INTO users (name, email, password_hash, plan_id, is_active)
VALUES (
    'admin',
    'admin@admin.com',
    '$argon2id$v=19$m=65536,t=1,p=4$xVbnRrqcz7its0aDZ7Kw8A$c6EzTA1fp2J3BRDITdgLjb11PEfcrhLfYyt5vHHHDQY',
    3,
    TRUE
);

INSERT INTO users_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.name = 'owner'
WHERE u.email = 'admin@admin.com';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE logs
    DROP COLUMN IF EXISTS country_code;

ALTER TABLE links
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS qr_code_path,
    DROP COLUMN IF EXISTS security_checked_at,
    DROP COLUMN IF EXISTS security_status,
    DROP COLUMN IF EXISTS ad_enabled,
    DROP COLUMN IF EXISTS max_clicks,
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS public_id;

DELETE FROM links WHERE user_id IS NULL;
ALTER TABLE links
    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE users
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS plan_expires_at,
    DROP COLUMN IF EXISTS plan_id;

DROP INDEX IF EXISTS idx_single_owner_role;
DROP TABLE IF EXISTS users_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS plans;

-- +goose StatementEnd
