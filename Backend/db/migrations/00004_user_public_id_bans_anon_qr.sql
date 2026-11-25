-- +goose Up
-- +goose StatementBegin

-- 1. Kullanıcılar: 16 haneli public_id ve ban/ceza alanları
ALTER TABLE users
    ADD COLUMN public_id VARCHAR(16),
    ADD COLUMN ban_reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN banned_at TIMESTAMPTZ NULL,
    ADD COLUMN banned_until TIMESTAMPTZ NULL;

UPDATE users
SET public_id = substr(replace(gen_random_uuid()::text, '-', ''), 1, 16)
WHERE public_id IS NULL OR public_id = '';

ALTER TABLE users
    ALTER COLUMN public_id SET NOT NULL,
    ADD CONSTRAINT users_public_id_unique UNIQUE (public_id);

CREATE INDEX idx_users_public_id ON users(public_id);

-- Defense-in-depth: Veritabanı seviyesinde Owner hesabının silinmesini engelle
CREATE OR REPLACE FUNCTION prevent_owner_delete()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM users_roles
        WHERE user_id = OLD.id AND role_id = 3
    ) THEN
        RAISE EXCEPTION 'owner account cannot be deleted';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_owner_delete
    BEFORE DELETE ON users
    FOR EACH ROW
    EXECUTE FUNCTION prevent_owner_delete();

-- 2. Linkler: Anonim link desteği
ALTER TABLE links
    ADD COLUMN is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN created_by_ip TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_links_anon_expires ON links (expires_at) WHERE is_anonymous = TRUE;

-- 3. Loglar: Anonim link silinse bile tıklama loglarının korunması (ON DELETE SET NULL)
ALTER TABLE logs
    ADD COLUMN link_alias TEXT NOT NULL DEFAULT '',
    ADD COLUMN link_public_id VARCHAR(16) NOT NULL DEFAULT '',
    ADD COLUMN event_type TEXT NOT NULL DEFAULT 'redirect';

UPDATE logs lg
SET link_alias = l.alias,
    link_public_id = l.public_id
FROM links l
WHERE lg.link_id = l.id AND (lg.link_alias = '' OR lg.link_public_id = '');

ALTER TABLE logs
    ALTER COLUMN link_id DROP NOT NULL,
    DROP CONSTRAINT IF EXISTS logs_link_id_fkey,
    ADD CONSTRAINT logs_link_id_fkey FOREIGN KEY (link_id) REFERENCES links(id) ON DELETE SET NULL;

-- 4. Refresh Token tablosu (Kısa ömürlü Access JWT + Rotating HttpOnly Refresh Token)
CREATE TABLE refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_agent TEXT NOT NULL DEFAULT '',
    ip_address TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS refresh_tokens;

DELETE FROM logs WHERE link_id IS NULL;
ALTER TABLE logs
    DROP CONSTRAINT IF EXISTS logs_link_id_fkey,
    ALTER COLUMN link_id SET NOT NULL,
    ADD CONSTRAINT logs_link_id_fkey FOREIGN KEY (link_id) REFERENCES links(id) ON DELETE CASCADE,
    DROP COLUMN IF EXISTS event_type,
    DROP COLUMN IF EXISTS link_public_id,
    DROP COLUMN IF EXISTS link_alias;

DROP INDEX IF EXISTS idx_links_anon_expires;
ALTER TABLE links
    DROP COLUMN IF EXISTS created_by_ip,
    DROP COLUMN IF EXISTS is_anonymous;

DROP TRIGGER IF EXISTS trg_prevent_owner_delete ON users;
DROP FUNCTION IF EXISTS prevent_owner_delete();

DROP INDEX IF EXISTS idx_users_public_id;
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_public_id_unique,
    DROP COLUMN IF EXISTS banned_until,
    DROP COLUMN IF EXISTS banned_at,
    DROP COLUMN IF EXISTS ban_reason,
    DROP COLUMN IF EXISTS public_id;

-- +goose StatementEnd
