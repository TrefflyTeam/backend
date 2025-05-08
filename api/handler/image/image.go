package image

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"strings"
	"treffly/apperror"
)

type imageService interface {
	Get(path string) (io.ReadCloser, string, error)
}

type Handler struct {
	imageService imageService
}

func NewImageHandler(imageService imageService) *Handler {
	return &Handler{
		imageService: imageService,
	}
}

func (h *Handler) Get(ctx *gin.Context) {
	path := ctx.Param("path")

	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		err := fmt.Errorf("invalid path: %s", path)
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	reader, mimeType, err := h.imageService.Get(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ctx.Error(apperror.NotFound.WithCause(err))
			return
		}
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}
	defer reader.Close()

	ctx.Header("Content-Type", mimeType)

	_, err = io.Copy(ctx.Writer, reader)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}
}
