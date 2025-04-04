package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"time"
	"treffly/apperror"
	db "treffly/db/sqlc"
)

const (
	defaultLat = "51.660781"
	defaultLon = "39.200296"
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

type eventResponse struct {
	ID          int32          `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Capacity    int32          `json:"capacity"`
	Latitude    pgtype.Numeric `json:"latitude"`
	Longitude   pgtype.Numeric `json:"longitude"`
	Address     string         `json:"address"`
	Date        time.Time      `json:"date"`
	IsPrivate   bool           `json:"is_private"`
	IsPremium   bool           `json:"is_premium"`
	CreatedAt   time.Time      `json:"created_at"`
	Tags        []db.Tag       `json:"tags"`
}

func newCreateEventResponse(event db.CreateEventTxResult) eventResponse {
	return eventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		Capacity:    event.Capacity,
		Latitude:    event.Latitude,
		Longitude:   event.Longitude,
		Address:     event.Address,
		Date:        event.Date,
		IsPrivate:   event.IsPrivate,
		IsPremium:   event.IsPremium,
		CreatedAt:   event.CreatedAt,
		Tags:        event.Tags,
	}
}

func (server *Server) createEvent(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	var req createEventRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	arg := db.CreateEventTxParams{
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Address:     req.Address,
		Date:        req.Date,
		IsPrivate:   req.IsPrivate,
		OwnerID:     userID,
		Tags:        req.Tags,
	}

	event, err := server.store.CreateEventTx(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, newCreateEventResponse(event))
}

type eventsListResponse struct {
	Events []db.EventWithTagsView `json:"events"`
}

func newEventsListResponse(events []db.EventWithTagsView) eventsListResponse {
	var response []db.EventWithTagsView
	for _, event := range events {
		response = append(response, event)
	}
	return eventsListResponse{Events: response}
}

func (server *Server) listEvents(ctx *gin.Context) {
	lat, lon, err := getUserLocation(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	arg := db.ListEventsParams{
		UserLon: lon,
		UserLat: lat,
	}

	events, err := server.store.ListEvents(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, newEventsListResponse(events))
}

func getUserLocation(ctx *gin.Context) (pgtype.Numeric, pgtype.Numeric, error) {
	latStr, err := ctx.Cookie("user_lat")
	if err != nil {
		latStr = defaultLat
	}

	lonStr, err := ctx.Cookie("user_lon")
	if err != nil {
		lonStr = defaultLon
	}

	latNum := pgtype.Numeric{}
	if err := latNum.Scan(latStr); err != nil {
		return pgtype.Numeric{}, pgtype.Numeric{}, fmt.Errorf("invalid latitude: %v", err)
	}

	lonNum := pgtype.Numeric{}
	if err := lonNum.Scan(lonStr); err != nil {
		return pgtype.Numeric{}, pgtype.Numeric{}, fmt.Errorf("invalid longitude: %v", err)
	}

	return latNum, lonNum, nil
}
