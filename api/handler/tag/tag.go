package tag

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"treffly/api/models"
	"treffly/apperror"
)

type getter interface {
	GetTags(ctx context.Context) ([]models.Tag, error)
}

type Handler struct {
	getter getter
}

func NewTagHandler(getter getter) *Handler {
	return &Handler{
		getter: getter,
	}
}

type tagResponse struct {
	Tags []models.Tag `json:"tags"`
}

func newTagResponse(tags []models.Tag) tagResponse {
	return tagResponse{tags}
}

func (h *Handler) GetTags(ctx *gin.Context) {
	tags, err := h.getter.GetTags(ctx)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, newTagResponse(tags))
}
