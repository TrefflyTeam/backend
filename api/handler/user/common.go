package user

import (
	"context"
	"github.com/google/uuid"
	"mime/multipart"
)

type imageService interface {
	Upload(file multipart.File, header *multipart.FileHeader, objType string, id string) (path string, err error)
	Delete(path string) error
	GetDBImageByUserID(ctx context.Context, userID int32) (uuid.UUID, string, error)
}