-- +goose Up
-- +goose StatementBegin
ALTER TABLE links
    ADD COLUMN wait_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN wait_seconds INTEGER NOT NULL DEFAULT 0;

CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    ip_address TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    device_type TEXT NOT NULL DEFAULT '',
    browser TEXT NOT NULL DEFAULT '',
    os TEXT NOT NULL DEFAULT '',
    referer TEXT NOT NULL DEFAULT '',
    accept_language TEXT NOT NULL DEFAULT '',
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_logs_link_id ON logs(link_id, clicked_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS logs;

ALTER TABLE links
    DROP COLUMN IF EXISTS wait_seconds,
    DROP COLUMN IF EXISTS wait_enabled;
-- +goose StatementEnd
