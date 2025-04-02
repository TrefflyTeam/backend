package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"treffly/apperror"
	db "treffly/db/sqlc"
)

type tagResponse struct {
	tags []db.Tag
}

func newTagResponse(tags []db.Tag) tagResponse {
	return tagResponse{tags}
}

func (server *Server) getTags(ctx *gin.Context) {
	tags, err := server.store.GetTags(ctx)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, newTagResponse(tags))
}
