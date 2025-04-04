package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"time"
	"treffly/apperror"
	db "treffly/db/sqlc"
)

type createEventRequest struct {
	Name        string         `json:"name" binding:"required,event_name,min=5,max=50"`
	Description string         `json:"description" binding:"required,min=50,max=1000"`
	Capacity    int32          `json:"capacity" binding:"required,min=1,max=500"`
	Latitude    pgtype.Numeric `json:"latitude" binding:"required,latitude"`
	Longitude   pgtype.Numeric `json:"longitude" binding:"required,longitude"`
	Address     string         `json:"address" binding:"required"`
	Date        time.Time      `json:"date" binding:"required,date"`
	IsPrivate   bool           `json:"is_private" binding:"boolean"`
	Tags        []int32        `json:"tags" binding:"required,min=1,max=3,dive,required,positive"`
}

func (server *Server) createEvent(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	var req createEventRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	arg := db.CreateEventParams{
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Address:     req.Address,
		Date:        req.Date,
		IsPrivate:   req.IsPrivate,
		OwnerID:     userID,
	}

	event, err := server.store.CreateEvent(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, event)
}
