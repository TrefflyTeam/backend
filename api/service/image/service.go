package imageservice

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"treffly/image"
	"treffly/util"
)

type Service struct {
	imageStore image.Store
	config util.Config
}

func New(imageStore image.Store, config util.Config) *Service {
	return &Service{
		imageStore: imageStore,
		config: config,
	}
}

func (s *Service) Upload(ctx *gin.Context, objType string, objID int32) (string, error) {
	file, header, err := ctx.Request.FormFile("image")
	if err != nil {
		return "", err
	}

	defer file.Close()

	if header.Size > 5<<20 {
		return "", errors.New("file too large")
	}

	if !isValidImageType(header) {
		return "", errors.New("invalid image type")
	}

	filename := filepath.Join(objType, fmt.Sprintf("%d%s", objID, filepath.Ext(header.Filename)))
	filename, err = s.imageStore.Upload(file, filename)
	if err != nil {
		return "", err
	}

	url := imageURL(s.config.Environment, s.config.Domain, filename)

	return url, nil
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

func imageURL(env, domain, path string) string {
	protocol := "http"
	if env == "production" {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s/%s", protocol, domain, path)

	return url
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
