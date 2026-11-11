-- name: GetLinkByAlias :one
SELECT * FROM links
WHERE alias = $1
LIMIT 1;

-- name: IncrementLinkClicks :exec
UPDATE links
SET clicks = clicks + 1
WHERE alias = $1;
