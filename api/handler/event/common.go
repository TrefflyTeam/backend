package event

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"mime/multipart"
	"strconv"
	"treffly/api/common"
)

type IDParser struct{}

func (p *IDParser) ParseEventID(ctx *gin.Context) (int32, error) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	return int32(id), err
}

func (p *IDParser) GetUserID(c *gin.Context) int32 {
	return common.GetUserIDFromContextPayload(c)
}

type BaseHandler struct {
	idParser  *IDParser
}

type ImageService interface {
	Upload(file multipart.File, header *multipart.FileHeader, objType string, id string) (path string, err error)
	Delete(path string) error
	GetDBImageByEventID(ctx context.Context, eventID int32) (uuid.UUID, string, error)
}