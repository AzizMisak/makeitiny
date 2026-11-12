-- name: GetLinkByAlias :one
SELECT * FROM links
WHERE alias = $1
LIMIT 1;

-- name: IncrementLinkClicks :exec
UPDATE links
SET clicks = clicks + 1
WHERE alias = $1;

-- name: CreateLog :one
INSERT INTO logs (
    link_id,
    ip_address,
    user_agent,
    device_type,
    browser,
    os,
    referer,
    accept_language
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
