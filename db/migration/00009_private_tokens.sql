-- +goose Up
-- +goose StatementBegin
CREATE TABLE event_tokens (
                              event_id     INTEGER NOT NULL,
                              token        TEXT NOT NULL,
                              created_at   timestamptz NOT NULL DEFAULT NOW(),
                              expires_at   timestamptz NOT NULL,
                              PRIMARY KEY (event_id, token)
);

ALTER TABLE "event_tokens" ADD FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON DELETE CASCADE;

CREATE INDEX idx_event_tokens_token ON event_tokens(token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE event_tokens;
-- +goose StatementEnd
