package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"strconv"
	"time"
	"treffly/apperror"
	db "treffly/db/sqlc"
)

const (
	defaultLat = "51.660781"
	defaultLon = "39.200296"
)

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
	OwnerID     int32          `json:"owner_id"`
	Tags        []db.Tag       `json:"tags"`
}

func newEventTxResponse(event db.EventTxResult) eventResponse {
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

type createEventResponse struct {
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

func newCreateEventResponse(event db.EventTxResult) createEventResponse {
	return createEventResponse{
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

	ctx.JSON(http.StatusOK, newEventTxResponse(event))
}


type eventsListResponse struct {
	Events []eventResponse `json:"events"`
}

func newEventsListResponse(events []db.EventWithTagsView) eventsListResponse {
	response := make([]eventResponse, 0, len(events))
	for _, event := range events {
		response = append(response, newEventResponse(event))
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

func newEventResponse(event db.EventWithTagsView) eventResponse {
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
		OwnerID:     event.OwnerID,
		Tags:        event.Tags,
	}
}

func (server *Server) getEvent(ctx *gin.Context) {
	eventID, err := getEventID(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	event, err := server.store.GetEvent(ctx, eventID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, newEventResponse(event))
}

type updateEventRequest struct {
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

type updateEventResponse struct {
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

func newUpdateEventResponse(event db.EventTxResult) updateEventResponse {
	return updateEventResponse{
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

func (server *Server) updateEvent(ctx *gin.Context) {
	var req updateEventRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	eventID, err := getEventID(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	userID := getUserIDFromContextPayload(ctx)

	event, err := server.store.GetEvent(ctx, eventID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	if event.OwnerID != userID {
		ctx.Error(apperror.Forbidden.WithCause(err))
		return
	}

	arg := db.UpdateEventTxParams{
		EventID:     eventID,
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Address:     req.Address,
		Date:        req.Date,
		IsPrivate:   req.IsPrivate,
		Tags:        req.Tags,
	}

	eventUpdated, err := server.store.UpdateEventTx(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusOK, newEventTxResponse(eventUpdated))
}

func (server *Server) deleteEvent(ctx *gin.Context) {
	eventID, err := getEventID(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	userID := getUserIDFromContextPayload(ctx)

	event, err := server.store.GetEvent(ctx, eventID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	if event.OwnerID != userID {
		ctx.Error(apperror.Forbidden.WithCause(err))
		return
	}

	err = server.store.DeleteEvent(ctx, eventID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{})
}

func getEventID(ctx *gin.Context) (int32, error) {
	eventIDStr := ctx.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return 0, err
	}

	return int32(eventID), nil
}
