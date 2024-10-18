-- name: GetFile :one
SELECT * FROM posts
WHERE identifier = $1 LIMIT 1;

-- name: UploadFile :one
INSERT INTO posts (
    filename, identifier, uploaddate, deletetoken
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteFile :exec
DELETE FROM posts
WHERE identifier = $1;
