package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"strings"
	imageservice "treffly/api/service/image"
	"treffly/apperror"
)

type ImageHandler struct {
	imageService *imageservice.Service
}

func NewImageHandler(imageService *imageservice.Service) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

func (h *ImageHandler) Get(ctx *gin.Context) {
	path := ctx.Param("path")

	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "invalid path"})
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
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}
