-- name: GetPost :one
SELECT
  *
FROM
  posts
WHERE
  id = sqlc.arg ('id')
LIMIT
  1;

-- name: GetPostByFilename :one
SELECT
  *
FROM
  posts
WHERE
  filename = sqlc.arg ('filename')
LIMIT
  1;

-- name: GetAllPosts :many
SELECT
  *
FROM
  posts
WHERE
    pos_by_id (id) > sqlc.arg ('page_size')::bigint * sqlc.arg ('page_num')::bigint
AND pos_by_id (id) <= sqlc.arg ('page_size')::bigint * (1 + sqlc.arg ('page_num')::bigint);

-- name: CreatePost :one
INSERT INTO
  posts (filename, deletion_key, hash)
VALUES
  (
    sqlc.arg ('filename'),
    sqlc.arg ('delkey'),
    sqlc.arg ('hash')
  ) RETURNING *;

-- name: UpdatePost :one
UPDATE posts
SET
  filename = coalesce(sqlc.narg ('filename'), filename),
  deletion_key = coalesce(sqlc.narg ('deletion_key'), deletion_key),
  hash = coalesce(sqlc.narg ('hash'), hash),
  status = coalesce(sqlc.narg ('status'), status),
  updated_at = now ()
WHERE
  id = sqlc.arg ('id') RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE
  id = sqlc.arg ('id');
