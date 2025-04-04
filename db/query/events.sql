-- name: CreateEvent :one
INSERT INTO events (
    name,
    description,
    capacity,
    latitude,
    longitude,
    address,
    date,
    owner_id,
    is_private
) VALUES (
             @name,
             @description,
             @capacity,
             @latitude,
             @longitude,
             @address,
             @date,
             @owner_id,
             @is_private
         )
    RETURNING *;

-- name: GetEvent :one
SELECT
    id,
    name,
    description,
    capacity,
    latitude,
    longitude,
    address,
    date,
    owner_id,
    is_private,
    is_premium,
    created_at
FROM events
WHERE id = @id
LIMIT 1;

-- name: ListEvents :many
SELECT
    id,
    name,
    description,
    capacity,
    latitude,
    longitude,
    address,
    date,
    owner_id,
    is_private,
    is_premium,
    created_at
FROM events
WHERE ST_DWithin(
    ST_MakePoint(longitude, latitude)::GEOGRAPHY,
    ST_MakePoint(@user_lon, @user_lat)::GEOGRAPHY,
    100000
    )
ORDER BY id;

-- name: UpdateEvent :one
UPDATE events
SET
    name = @name,
    description = @description,
    capacity = @capacity,
    latitude = @latitude,
    longitude = @longitude,
    address = @address,
    date = @date,
    is_private = @is_private
WHERE id = @id
    RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE id = $1;