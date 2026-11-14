-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: AssignRoleByName :exec
INSERT INTO users_roles (user_id, role_id)
SELECT $1, id FROM roles WHERE name = $2
ON CONFLICT (user_id, role_id) DO NOTHING;

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
SET password_hash = $2
WHERE email = $1;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = $1;

-- name: CreateLink :one
INSERT INTO links (public_id, user_id, alias, target_url, password_hash, wait_enabled, wait_seconds)
VALUES ($1, $2, $3, $4, $5, $6, $7)
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
    wait_seconds = $7
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
WHERE link_id = $1;

-- name: ListLogsByLinkPaginated :many
SELECT * FROM logs
WHERE link_id = $1
ORDER BY clicked_at DESC, id DESC
LIMIT $2 OFFSET $3;
