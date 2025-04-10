package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"treffly/apperror"
)

func (server *Server) geocode(ctx *gin.Context) {
	lat, err := strconv.ParseFloat(ctx.Query("lat"), 64)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}
	if lat > 90.0 || lat < -90.0 {
		err = fmt.Errorf("invalid lat %f", lat)
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	lon, err := strconv.ParseFloat(ctx.Query("lon"), 64)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}
	if lon > 180.0 || lon < -180.0 {
		err = fmt.Errorf("invalid lon %f", lon)
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	response, err := server.mapsClient.GetLocation(lat, lon)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	ctx.Data(http.StatusOK, "application/json", response)
}
