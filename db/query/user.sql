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
SET username = $2,
    image_id = $3
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

-- name: SubscribeToEvent :one
WITH check_capacity AS (
    SELECT
        COUNT(*) AS participants,
        e.capacity
    FROM events e
             LEFT JOIN event_user eu ON e.id = eu.event_id
    WHERE e.id = $2
    GROUP BY e.capacity
)
INSERT INTO event_user (user_id, event_id)
SELECT $1, $2
FROM check_capacity
WHERE participants < capacity
RETURNING (SELECT participants < capacity FROM check_capacity) AS allowed;

-- name: UnsubscribeFromEvent :exec
DELETE FROM event_user
WHERE user_id = $1 AND event_id = $2;

-- name: IsParticipant :one
SELECT EXISTS (
    SELECT 1
    FROM event_user
    WHERE event_id = $1
      AND user_id = $2
) AS is_participant;