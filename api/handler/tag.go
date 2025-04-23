package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	tagservice "treffly/api/service/tag"
	"treffly/apperror"
	db "treffly/db/sqlc"
)

type TagHandler struct {
	service *tagservice.Service
}

func NewTagHandler(service *tagservice.Service) *TagHandler {
	return &TagHandler{
		service: service,
	}
}

type tagResponse struct {
	Tags []db.Tag `json:"tags"`
}

func newTagResponse(tags []db.Tag) tagResponse {
	return tagResponse{tags}
}

func (h *TagHandler) GetTags(ctx *gin.Context) {
	tags, err := h.service.GetTags(ctx)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, newTagResponse(tags))
}
