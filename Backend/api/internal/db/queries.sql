-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: AssignRoleByName :exec
INSERT INTO users_roles (user_id, role_id)
SELECT $1, id FROM roles WHERE name = $2
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: ClearUserNonOwnerRoles :exec
DELETE FROM users_roles
WHERE user_id = $1 AND role_id <> 3;

-- name: GetUserRoles :many
SELECT r.name
FROM roles r
JOIN users_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1
ORDER BY r.id ASC;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: UpdateUserPassword :execrows
UPDATE users
SET password_hash = $2,
    updated_at = NOW()
WHERE email = $1;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = $1;

-- name: GetPlanByID :one
SELECT id, slug, name, price_monthly::FLOAT8 AS price_monthly, max_links, allow_custom_alias, allow_password, allow_qr, allow_ads, analytics_days, created_at
FROM plans
WHERE id = $1
LIMIT 1;

-- name: GetPlanBySlug :one
SELECT id, slug, name, price_monthly::FLOAT8 AS price_monthly, max_links, allow_custom_alias, allow_password, allow_qr, allow_ads, analytics_days, created_at
FROM plans
WHERE slug = $1
LIMIT 1;

-- name: ListPlans :many
SELECT id, slug, name, price_monthly::FLOAT8 AS price_monthly, max_links, allow_custom_alias, allow_password, allow_qr, allow_ads, analytics_days, created_at
FROM plans
ORDER BY id ASC;

-- name: UpdatePlan :one
UPDATE plans
SET name = $2,
    price_monthly = $3,
    max_links = $4,
    allow_custom_alias = $5,
    allow_password = $6,
    allow_qr = $7,
    allow_ads = $8,
    analytics_days = $9
WHERE id = $1
RETURNING id, slug, name, price_monthly::FLOAT8 AS price_monthly, max_links, allow_custom_alias, allow_password, allow_qr, allow_ads, analytics_days, created_at;

-- name: CountLinksByUser :one
SELECT COUNT(*)::BIGINT FROM links
WHERE user_id = $1;

-- name: CreateLink :one
INSERT INTO links (
    public_id,
    user_id,
    alias,
    target_url,
    password_hash,
    wait_enabled,
    wait_seconds,
    title,
    is_active,
    expires_at,
    max_clicks,
    ad_enabled
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: ListLinksByUser :many
SELECT * FROM links
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListLinksByUserPaginated :many
SELECT * FROM links
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: GetUserLinkStats :one
SELECT
    COUNT(*)::BIGINT AS total_links,
    COALESCE(SUM(clicks), 0)::BIGINT AS total_clicks,
    COUNT(*) FILTER (WHERE password_hash IS NOT NULL AND password_hash <> '')::BIGINT AS locked_links
FROM links
WHERE user_id = $1;

-- name: ListTopLinksByUser :many
SELECT * FROM links
WHERE user_id = $1
ORDER BY clicks DESC, created_at DESC, id DESC
LIMIT $2;

-- name: GetLinkByIDAndUser :one
SELECT * FROM links
WHERE public_id = $1 AND user_id = $2
LIMIT 1;

-- name: GetLinkByPublicID :one
SELECT * FROM links
WHERE public_id = $1
LIMIT 1;

-- name: GetLinkByAlias :one
SELECT * FROM links
WHERE alias = $1
LIMIT 1;

-- name: AliasExists :one
SELECT EXISTS(
    SELECT 1 FROM links WHERE alias = $1
);

-- name: UpdateLink :one
UPDATE links
SET alias = $3,
    target_url = $4,
    password_hash = $5,
    wait_enabled = $6,
    wait_seconds = $7,
    title = $8,
    is_active = $9,
    expires_at = $10,
    max_clicks = $11,
    ad_enabled = $12,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteLink :execrows
DELETE FROM links
WHERE id = $1 AND user_id = $2;

-- name: ListLogsByLink :many
SELECT * FROM logs
WHERE link_id = $1
ORDER BY clicked_at DESC;

-- name: CountLogsByLink :one
SELECT COUNT(*)::BIGINT FROM logs
WHERE link_id = $1 AND clicked_at >= $2;

-- name: ListLogsByLinkPaginated :many
SELECT * FROM logs
WHERE link_id = $1 AND clicked_at >= $2
ORDER BY clicked_at DESC, id DESC
LIMIT $3 OFFSET $4;

-- name: GetPlatformStats :one
SELECT
    (SELECT COUNT(*)::BIGINT FROM users) AS total_users,
    (SELECT COUNT(*) FILTER (WHERE is_active = TRUE)::BIGINT FROM users) AS active_users,
    COUNT(*)::BIGINT AS total_links,
    COUNT(*) FILTER (WHERE is_active = TRUE)::BIGINT AS active_links,
    COUNT(*) FILTER (WHERE security_status = 'blocked' OR security_status = 'malicious')::BIGINT AS blocked_links,
    COALESCE(SUM(clicks), 0)::BIGINT AS total_clicks
FROM links;

-- name: CountAdminUsers :one
SELECT COUNT(*)::BIGINT
FROM users
WHERE ($1::text = '' OR email ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%');

-- name: ListAdminUsersPaginated :many
SELECT *
FROM users
WHERE ($1::text = '' OR email ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%')
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: AdminUpdateUser :one
UPDATE users
SET plan_id = $2,
    plan_expires_at = $3,
    is_active = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountAdminLinks :one
SELECT COUNT(*)::BIGINT
FROM links l
LEFT JOIN users u ON u.id = l.user_id
WHERE ($1::text = '' OR l.alias ILIKE '%' || $1 || '%' OR l.target_url ILIKE '%' || $1 || '%' OR l.title ILIKE '%' || $1 || '%' OR COALESCE(u.email, '') ILIKE '%' || $1 || '%')
  AND (
    $2::text = '' OR $2::text = 'all'
    OR ($2::text = 'active' AND l.is_active = TRUE AND l.security_status <> 'blocked' AND l.security_status <> 'malicious')
    OR ($2::text = 'inactive' AND l.is_active = FALSE)
    OR ($2::text = 'blocked' AND (l.security_status = 'blocked' OR l.security_status = 'malicious'))
  );

-- name: ListAdminLinksPaginated :many
SELECT
    l.*,
    COALESCE(u.email, '') AS user_email,
    COALESCE(u.name, '') AS user_name
FROM links l
LEFT JOIN users u ON u.id = l.user_id
WHERE ($1::text = '' OR l.alias ILIKE '%' || $1 || '%' OR l.target_url ILIKE '%' || $1 || '%' OR l.title ILIKE '%' || $1 || '%' OR COALESCE(u.email, '') ILIKE '%' || $1 || '%')
  AND (
    $2::text = '' OR $2::text = 'all'
    OR ($2::text = 'active' AND l.is_active = TRUE AND l.security_status <> 'blocked' AND l.security_status <> 'malicious')
    OR ($2::text = 'inactive' AND l.is_active = FALSE)
    OR ($2::text = 'blocked' AND (l.security_status = 'blocked' OR l.security_status = 'malicious'))
  )
ORDER BY l.created_at DESC, l.id DESC
LIMIT $3 OFFSET $4;

-- name: AdminUpdateLink :one
UPDATE links
SET alias = $2,
    target_url = $3,
    password_hash = $4,
    wait_enabled = $5,
    wait_seconds = $6,
    title = $7,
    is_active = $8,
    expires_at = $9,
    max_clicks = $10,
    ad_enabled = $11,
    security_status = $12,
    security_checked_at = $13,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: AdminDeleteLink :execrows
DELETE FROM links
WHERE id = $1;
