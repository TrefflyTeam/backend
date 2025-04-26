package image

import (
	"io"
	"mime"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	BasePath string
}

func NewLocalStorage(basePath string) (LocalStorage, error) {
	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		return LocalStorage{}, err
	}
	err = os.MkdirAll(filepath.Join(basePath, "user"), os.ModePerm)
	if err != nil {
		return LocalStorage{}, err
	}
	err = os.MkdirAll(filepath.Join(basePath, "event"), os.ModePerm)
	if err != nil {
		return LocalStorage{}, err
	}
	return LocalStorage{BasePath: basePath}, nil
}

func (s LocalStorage) Upload(file io.Reader, filename string) (string, error) {
	path := filepath.Join(s.BasePath, filename)
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}
	return filename, nil
}

func (s LocalStorage) Get(filename string) (io.ReadCloser, string, error) {
	path := filepath.Join(s.BasePath, filename)
	_, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}

	mimeType := mime.TypeByExtension(filepath.Ext(path))
	return file, mimeType, nil
}

func (s LocalStorage) Delete(filename string) error {
	path := filepath.Join(s.BasePath, filename)
	return os.Remove(path)
}
