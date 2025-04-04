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
SELECT * FROM event_with_tags_view
WHERE id = $1;

-- name: ListEvents :many
SELECT * FROM event_with_tags_view
WHERE ST_DWithin(
              ST_MakePoint(e.longitude, e.latitude)::GEOGRAPHY,
              ST_MakePoint(@user_lon::numeric, @user_lat::numeric)::GEOGRAPHY,
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