-- +goose Up
-- +goose StatementBegin
CREATE TABLE "users"
(
    "id"            INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    "username"      varchar(50)         NOT NULL,
    "email"         varchar(100) UNIQUE NOT NULL,
    "password_hash" text                NOT NULL,
    "created_at"    timestamptz         NOT NULL DEFAULT (now()),
    "is_admin"      boolean             NOT NULL DEFAULT false
);

CREATE TABLE "sessions"
(
    "uuid"          UUID PRIMARY KEY,
    "user_id"       integer     NOT NULL,
    "refresh_token" text        NOT NULL,
    "expires_at"    timestamptz NOT NULL,
    "is_blocked"    boolean     NOT NULL DEFAULT false,
    "created_at"    timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX "idx_sessions_user_id" ON "sessions" ("user_id");

ALTER TABLE "sessions"
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
DROP TABLE sessions;

-- +goose StatementEnd
