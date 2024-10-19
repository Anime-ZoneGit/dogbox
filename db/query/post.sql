-- name: GetPost :one
SELECT * FROM posts
WHERE id = $1 LIMIT 1;

-- name: CreatePost :one
INSERT INTO posts (
    filename, deletion_key, hash
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdatePost :one
UPDATE posts
SET filename = coalesce(sqlc.narg('filename'), filename),
    deletion_key = coalesce(sqlc.narg('deletion_key'), deletion_key),
    hash = coalesce(sqlc.narg('hash'), hash),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = $1;
