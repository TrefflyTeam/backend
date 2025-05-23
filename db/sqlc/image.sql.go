// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: image.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createImage = `-- name: CreateImage :one
INSERT INTO images (
    id,
    path
) VALUES (
             $1,
             $2
         )
RETURNING id, path
`

type CreateImageParams struct {
	ID   uuid.UUID `json:"id"`
	Path string    `json:"path"`
}

func (q *Queries) CreateImage(ctx context.Context, arg CreateImageParams) (Image, error) {
	row := q.db.QueryRow(ctx, createImage, arg.ID, arg.Path)
	var i Image
	err := row.Scan(&i.ID, &i.Path)
	return i, err
}

const deleteImage = `-- name: DeleteImage :exec
DELETE FROM images
WHERE id = $1
`

func (q *Queries) DeleteImage(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteImage, id)
	return err
}

const getImageByEventID = `-- name: GetImageByEventID :one
SELECT i.id, i.path
FROM images i LEFT JOIN events e ON e.image_id = i.id
WHERE e.id = $1
`

func (q *Queries) GetImageByEventID(ctx context.Context, id int32) (Image, error) {
	row := q.db.QueryRow(ctx, getImageByEventID, id)
	var i Image
	err := row.Scan(&i.ID, &i.Path)
	return i, err
}

const getImageByUserID = `-- name: GetImageByUserID :one
SELECT i.id, i.path
FROM images i LEFT JOIN users u ON u.image_id = i.id
WHERE u.id = $1
`

func (q *Queries) GetImageByUserID(ctx context.Context, id int32) (Image, error) {
	row := q.db.QueryRow(ctx, getImageByUserID, id)
	var i Image
	err := row.Scan(&i.ID, &i.Path)
	return i, err
}
