-- +goose Up
-- +goose StatementBegin

-- Paket düşürüldüğünde yeni paketin link sınırını aşan (en eski) linkler askıya alınır.
-- Askıdaki link yönlendirmeye devam eder; plan_grace_until geçince kalıcı olarak silinir.
ALTER TABLE links
    ADD COLUMN plan_grace_until TIMESTAMPTZ NULL;

CREATE INDEX idx_links_plan_grace_until ON links (plan_grace_until) WHERE plan_grace_until IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_links_plan_grace_until;
ALTER TABLE links
    DROP COLUMN IF EXISTS plan_grace_until;

-- +goose StatementEnd
