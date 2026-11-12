-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

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
INSERT INTO links (user_id, alias, target_url, password_hash, wait_enabled, wait_seconds)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListLinksByUser :many
SELECT * FROM links
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetLinkByIDAndUser :one
SELECT * FROM links
WHERE id = $1 AND user_id = $2
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
