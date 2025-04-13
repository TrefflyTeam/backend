-- name: GetTags :many
SELECT * FROM tags
ORDER BY id;

-- name: AddUserTags :exec
INSERT INTO user_tags (user_id, tag_id)
SELECT @user_id, unnest(@tags::int[]);

-- name: DeleteUserTags :exec
DELETE FROM user_tags
WHERE user_id = @user_id;

-- name: AddEventTag :one
INSERT INTO event_tags (event_id, tag_id)
VALUES ($1, $2)
RETURNING event_id, tag_id;

-- name: DeleteAllEventTags :exec
DELETE FROM event_tags
WHERE event_id = $1;

-- name: GetAllUserTags :one
SELECT tags FROM user_with_tags_view WHERE id = $1;
