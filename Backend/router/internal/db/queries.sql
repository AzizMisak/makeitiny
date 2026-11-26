-- name: GetLinkByAlias :one
SELECT
    l.id,
    l.public_id,
    COALESCE(l.user_id, 0)::BIGINT AS user_id,
    l.alias,
    l.target_url,
    l.password_hash,
    l.clicks,
    l.created_at,
    l.wait_enabled,
    l.wait_seconds,
    l.title,
    (l.is_active AND COALESCE(u.is_active OR (u.banned_until IS NOT NULL AND u.banned_until <= NOW()), TRUE)) AS is_active,
    l.expires_at,
    l.max_clicks,
    l.ad_enabled,
    l.security_status,
    l.security_checked_at,
    l.qr_code_path,
    l.is_anonymous,
    l.updated_at
FROM links l
LEFT JOIN users u ON u.id = l.user_id
WHERE l.alias = $1
LIMIT 1;

-- name: IncrementLinkClicks :execrows
UPDATE links
SET clicks = clicks + 1
WHERE alias = $1
  AND (max_clicks IS NULL OR max_clicks <= 0 OR clicks + $2 < max_clicks);

-- name: CreateLog :one
INSERT INTO logs (
    link_id,
    link_alias,
    link_public_id,
    ip_address,
    user_agent,
    device_type,
    browser,
    os,
    referer,
    accept_language,
    country_code,
    event_type
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;
