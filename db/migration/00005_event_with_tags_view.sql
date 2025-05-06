-- +goose Up
-- +goose StatementBegin
CREATE VIEW event_with_tags_view AS
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
    ) AS tags
FROM events e
         LEFT JOIN event_tags et ON e.id = et.event_id
         LEFT JOIN tags t ON et.tag_id = t.id
GROUP BY e.id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW event_with_tags_view;
-- +goose StatementEnd
