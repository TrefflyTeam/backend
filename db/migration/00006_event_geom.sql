-- +goose Up
-- +goose StatementBegin
ALTER TABLE events
    ADD COLUMN geom GEOGRAPHY(POINT);

UPDATE events
SET geom = ST_MakePoint(longitude, latitude);

CREATE INDEX idx_events_geom ON events USING GIST(geom);

CREATE OR REPLACE FUNCTION update_geom()
RETURNS TRIGGER AS $$
BEGIN
  NEW.geom = ST_MakePoint(NEW.longitude, NEW.latitude);
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_events_geom
    BEFORE INSERT OR UPDATE OF longitude, latitude ON events
    FOR EACH ROW
    EXECUTE FUNCTION update_geom();

CREATE OR REPLACE VIEW event_with_tags_view AS
SELECT
    e.id,
    e.name,
    e.description,
    e.capacity,
    e.latitude,
    e.longitude,
    e.address,
    e.date,
    e.owner_id,
    e.is_private,
    e.is_premium,
    e.created_at,
    COALESCE(
            JSON_AGG(
                    json_build_object('id', t.id, 'name', t.name)
                        ORDER BY t.name
            ) FILTER (WHERE t.id IS NOT NULL),
            '[]'::JSON
    ) AS tags,
    e.geom,
    u.username AS owner_username,
    (SELECT COUNT(*)
     FROM event_user eu
     WHERE eu.event_id = e.id) AS participants_count
FROM events e
         LEFT JOIN event_tags et ON e.id = et.event_id
         LEFT JOIN tags t ON et.tag_id = t.id
         LEFT JOIN users u ON e.owner_id = u.id
GROUP BY
    e.id,
    u.username;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER trg_events_geom ON events;

DROP FUNCTION update_geom;
DROP VIEW event_with_tags_view;
ALTER TABLE events DROP COLUMN geom;

-- +goose StatementEnd
