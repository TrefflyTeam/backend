-- name: CreateImage :one
INSERT INTO images (
    id,
    path
) VALUES (
             @id,
             @path
         )
RETURNING id, path;