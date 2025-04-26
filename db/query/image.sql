-- name: CreateImage :one
INSERT INTO images (
    id,
    path
) VALUES (
             @id,
             @path
         )
RETURNING id, path;

-- name: GetImageByEventID :one
SELECT i.id, i.path
FROM images i LEFT JOIN events e ON e.image_id = i.id
WHERE e.id = @id;

-- name: DeleteImage :exec
DELETE FROM images
WHERE id = @id;