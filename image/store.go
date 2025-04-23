package image

import (
	"io"
)

type Store interface {
	Upload(file io.Reader, filename string) (string, error)
	Get(filename string) (io.ReadCloser, string, error)
	Delete(filename string) error
}
