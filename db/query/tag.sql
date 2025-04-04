-- name: GetTags :many
SELECT * FROM tags
ORDER BY id;

-- name: AddUserTag :one
INSERT INTO user_tags (user_id, tag_id)
VALUES ($1, $2)
RETURNING user_id, tag_id;

-- name: DeleteUserTag :exec
DELETE FROM user_tags
WHERE user_id = $1 AND tag_id = $2;

-- name: AddEventTag :one
INSERT INTO event_tags (event_id, tag_id)
VALUES ($1, $2)
RETURNING event_id, tag_id;

-- name: DeleteAllEventTags :exec
DELETE FROM event_tags
WHERE event_id = $1;

-- name: GetAllUserTags :one
SELECT tags FROM user_with_tags_view WHERE id = $1;
