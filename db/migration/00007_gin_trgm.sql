-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION pg_trgm;
CREATE INDEX event_name_trgm_idx ON events USING GIN (name gin_trgm_ops);
CREATE INDEX event_description_trgm_idx ON events USING GIN (description gin_trgm_ops)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX event_name_trgm_idx;
DROP INDEX event_description_trgm_idx;
-- +goose StatementEnd
