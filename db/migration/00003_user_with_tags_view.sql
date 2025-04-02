-- +goose Up
-- +goose StatementBegin
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW user_with_tags_view;
-- +goose StatementEnd
