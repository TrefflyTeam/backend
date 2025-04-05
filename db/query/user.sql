-- name: CreateUser :one
INSERT INTO users (username,
                   email,
                   password_hash)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET username = $2
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id=$1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserWithTags :one
SELECT * FROM user_with_tags_view WHERE id = $1;

-- name: SubscribeToEvent :exec
INSERT INTO event_user (user_id, event_id)
VALUES ($1, $2);

-- name: UnsubscribeFromEvent :exec
DELETE FROM event_user
WHERE user_id = $1 AND event_id = $2;