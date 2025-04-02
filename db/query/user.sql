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
SELECT
    u.id,
    u.username,
    u.email,
    u.created_at,
    COALESCE(
            JSON_AGG(
                    json_build_object('id', t.id, 'name', t.name)
                        ORDER BY t.name
            ) FILTER (WHERE t.id IS NOT NULL),
            '[]'::JSON
    ) AS tags
FROM users u
         LEFT JOIN user_tags ut ON u.id = ut.user_id
         LEFT JOIN tags t ON ut.tag_id = t.id
WHERE u.id = $1
GROUP BY u.id;