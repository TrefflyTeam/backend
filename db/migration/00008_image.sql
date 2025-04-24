-- +goose Up
-- +goose StatementBegin
CREATE TABLE images (
                        id UUID PRIMARY KEY,
                        path TEXT NOT NULL,
                        mime_type TEXT NOT NULL
);

CREATE INDEX ON images (mime_type);

ALTER TABLE events
    ADD COLUMN image_id UUID;

ALTER TABLE events
    ADD FOREIGN KEY (image_id) REFERENCES images (id) ON DELETE CASCADE;

ALTER TABLE users
    ADD COLUMN image_id UUID;

ALTER TABLE users
    ADD FOREIGN KEY (image_id) REFERENCES images (id) ON DELETE CASCADE;

CREATE INDEX idx_events_image_id ON events(image_id);
CREATE INDEX idx_users_image_id ON users(image_id);

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
     WHERE eu.event_id = e.id) AS participants_count,
     i_event.path AS event_image_path,
     i_user.path AS user_image_path
FROM events e
         LEFT JOIN event_tags et ON e.id = et.event_id
         LEFT JOIN tags t ON et.tag_id = t.id
         LEFT JOIN users u ON e.owner_id = u.id
         LEFT JOIN images i_event ON e.image_id = i_event.id
         LEFT JOIN images i_user ON e.image_id = i_user.id
GROUP BY
    e.id,
    u.username,
    i_event.path,
    i_user.path;

CREATE OR REPLACE VIEW user_with_tags_view AS
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
    ) AS tags,
    i.path AS image_path
FROM users u
         LEFT JOIN user_tags ut ON u.id = ut.user_id
         LEFT JOIN tags t ON ut.tag_id = t.id
         LEFT JOIN images i ON u.image_id = i.id
GROUP BY u.id, i.path;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
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

CREATE VIEW user_with_tags_view AS
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
GROUP BY u.id;

ALTER TABLE events DROP COLUMN image_id;
ALTER TABLE users DROP COLUMN image_id;
-- +goose StatementEnd
