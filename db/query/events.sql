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
    RETURNING id, name, description, capacity, latitude, longitude,
    address, date, owner_id, is_private, is_premium, created_at;

-- name: GetEvent :one
SELECT * FROM event_with_tags_view
WHERE id = $1;

-- name: ListEvents :many
SELECT * FROM event_with_tags_view
WHERE ST_DWithin(
              ST_MakePoint(longitude, latitude)::GEOGRAPHY,
              ST_MakePoint(@user_lon::numeric, @user_lat::numeric)::GEOGRAPHY,
              100000
      )
ORDER BY id;

-- name: UpdateEvent :exec
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
WHERE id = @id;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE id = $1;

-- name: GetPremiumEvents :many
SELECT *
FROM events
WHERE is_premium = TRUE
  AND date > NOW()
ORDER BY created_at DESC
    LIMIT 6;

-- name: GetLatestEvents :many
SELECT *
FROM events
WHERE date > NOW()
ORDER BY created_at DESC
    LIMIT 6;

-- name: GetPopularEvents :many
SELECT
    e.*,
    COUNT(eu.event_id) AS participants_count
FROM
    events e
        LEFT JOIN
    event_user eu ON e.id = eu.event_id
WHERE
    e.date > NOW()
GROUP BY
    e.id
ORDER BY
    participants_count DESC,
    e.created_at DESC
    LIMIT 6;

-- name: GetUserRecommendedEvents :many
WITH user_tags AS (
    SELECT (tag->>'id')::INT AS tag_id
    FROM user_with_tags_view,
         json_array_elements(tags) AS tag
    WHERE user_with_tags_view.id = @user_id
)
SELECT
    e.*,
    COUNT(et.tag_id) AS matched_tags,
    ST_Distance(
            e.geom,
            ST_MakePoint(@user_lon::numeric, @user_lat::numeric)::GEOGRAPHY
    ) AS distance
FROM events e
         LEFT JOIN event_tags et
                   ON e.id = et.event_id
                       AND et.tag_id IN (SELECT tag_id FROM user_tags)
WHERE
    e.date > NOW() AND
    ST_DWithin(e.geom, ST_MakePoint(@user_lon::numeric, @user_lat::numeric)::GEOGRAPHY, 100000)
GROUP BY e.id
ORDER BY
    matched_tags DESC,
    e.created_at DESC,
    distance ASC
    LIMIT 6;

-- name: GetGuestRecommendedEvents :many
SELECT *
FROM events
WHERE
    date > NOW() AND
    ST_DWithin(geom, ST_MakePoint(@user_lon::numeric, @user_lat::numeric)::GEOGRAPHY, 100000)
ORDER BY
    ST_Distance(geom, ST_MakePoint(@user_lon::numeric, @user_lat::numeric)::GEOGRAPHY) ASC,
    created_at DESC
    LIMIT 6;