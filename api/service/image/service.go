package imageservice

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	db "treffly/db/sqlc"
	"treffly/image"
	"treffly/util"
)

type Service struct {
	imageStore image.Store
	store      db.Store
	config     util.Config
}

func New(imageStore image.Store, config util.Config, store db.Store) *Service {
	return &Service{
		imageStore: imageStore,
		config:     config,
		store:      store,
	}
}

func (s *Service) Upload(file multipart.File, header *multipart.FileHeader, objType string, id string) (string, error) {
	defer file.Close()

	if header.Size > 5<<20 {
		return "", errors.New("file too large")
	}

	if !isValidImageType(header) {
		return "", errors.New("invalid image type")
	}

	filename := filepath.Join(objType, fmt.Sprintf("%s%s", id, filepath.Ext(header.Filename)))
	_, err := s.imageStore.Upload(file, filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (s *Service) Get(path string) (io.ReadCloser, string, error) {
	reader, mimeType, err := s.imageStore.Get(path)
	if err != nil {
		return nil, "", err
	}

	return reader, mimeType, nil
}

func (s *Service) Delete(path string) error {
	return s.imageStore.Delete(path)
}

func isValidImageType(header *multipart.FileHeader) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/jpg":  true,
	}

	file, _ := header.Open()
	defer file.Close()

	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return false
	}

	mimeType := http.DetectContentType(buffer)
	return allowedTypes[mimeType]
}


func (s *Service) GetDBImageByEventID(ctx context.Context, eventID int32) (uuid.UUID, string, error) {
	img, err := s.store.GetImageByEventID(ctx, eventID)
	if err != nil {
		return uuid.Nil, "", err
	}

	return img.ID, img.Path, err
}

func (s *Service) GetDBImageByUserID(ctx context.Context, userID int32) (uuid.UUID, string, error) {
	img, err := s.store.GetImageByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, "", err
	}

	return img.ID, img.Path, err
}