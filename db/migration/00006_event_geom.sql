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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events DROP COLUMN geom;

DROP TRIGGER trg_events_geom ON events;

DROP FUNCTION update_geom;
-- +goose StatementEnd
