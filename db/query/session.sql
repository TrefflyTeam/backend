-- name: CreateSession :exec
INSERT INTO sessions (
                      uuid,
                      user_id,
                      refresh_token,
                      expires_at,
                      is_blocked
) VALUES (
          $1, $2, $3, $4, $5
         ) RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE uuid = $1 LIMIT 1;

-- name: UpdateSession :exec
UPDATE sessions
SET uuid = sqlc.arg(new_uuid), refresh_token = sqlc.arg(refresh_token), expires_at = sqlc.arg(expires_at)
WHERE uuid = sqlc.arg(old_uuid)
RETURNING *;