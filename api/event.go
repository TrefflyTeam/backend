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
	ID               int32          `json:"id"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Capacity         int32          `json:"capacity"`
	Latitude         pgtype.Numeric `json:"latitude"`
	Longitude        pgtype.Numeric `json:"longitude"`
	Address          string         `json:"address"`
	Date             time.Time      `json:"date"`
	IsPrivate        bool           `json:"is_private"`
	IsPremium        bool           `json:"is_premium"`
	CreatedAt        time.Time      `json:"created_at"`
	OwnerUsername    string         `json:"owner_username"`
	Tags             []db.Tag       `json:"tags"`
	ParticipantCount int32          `json:"participant_count"`
}


type createEventRequest struct {
	Name        string         `json:"name" binding:"required,event_name,min=5,max=50"`
	Description string         `json:"description" binding:"required,min=50,max=1000"`
	Capacity    int32          `json:"capacity" binding:"required,min=1,max=500"`
	Latitude    pgtype.Numeric `json:"latitude" binding:"required,latitude"`
	Longitude   pgtype.Numeric `json:"longitude" binding:"required,longitude"`
	Address     string         `json:"address" binding:"required"`
	Date        time.Time      `json:"date" binding:"required,date,future"`
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

	ctx.JSON(http.StatusOK, convertEvent(event))
}

type eventsListResponse struct {
	Events []eventResponse `json:"events"`
}

func newEventsListResponse(events []db.EventRow) eventsListResponse {
	return eventsListResponse{Events: convertEvents(events)}
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

	var events []db.EventRow
	rows, err := server.store.ListEvents(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	events = db.ConvertToEventRow(rows)

	ctx.JSON(http.StatusOK, newEventsListResponse(events))
}

func getUserLocation(ctx *gin.Context) (lat pgtype.Numeric, lon pgtype.Numeric, err error) {
	latStr, err := ctx.Cookie("user_lat")
	if err != nil {
		latStr = defaultLat
	}

	lonStr, err := ctx.Cookie("user_lon")
	if err != nil {
		lonStr = defaultLon
	}

	if err := lat.Scan(latStr); err != nil {
		return pgtype.Numeric{}, pgtype.Numeric{}, fmt.Errorf("invalid latitude: %v", err)
	}

	if err := lon.Scan(lonStr); err != nil {
		return pgtype.Numeric{}, pgtype.Numeric{}, fmt.Errorf("invalid longitude: %v", err)
	}

	return lat, lon, nil
}

func (server *Server) getEvent(ctx *gin.Context) {
	eventID, err := getEventID(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	row, err := server.store.GetEvent(ctx, eventID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}


	ctx.JSON(http.StatusOK, convertEvent(row))
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

	ctx.JSON(http.StatusOK, convertEvent(eventUpdated))
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

func (server *Server) subscribeCurrentUserToEvent(ctx *gin.Context) {
	eventID, err := getEventID(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	userID := getUserIDFromContextPayload(ctx)

	arg := db.SubscribeToEventParams{
		EventID: eventID,
		UserID:  userID,
	}

	err = server.store.SubscribeToEvent(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{})
}

func (server *Server) unsubscribeCurrentUserFromEvent(ctx *gin.Context) {
	eventID, err := getEventID(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	userID := getUserIDFromContextPayload(ctx)

	arg := db.UnsubscribeFromEventParams{
		EventID: eventID,
		UserID:  userID,
	}

	err = server.store.UnsubscribeFromEvent(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{})
}

type getHomeEventsResponse struct {
	Premium     []eventResponse `json:"premium"`
	Recommended []eventResponse `json:"recommended"`
	Latest      []eventResponse `json:"latest"`
	Popular     []eventResponse `json:"popular"`
}

func newGetHomeEventsResponse(
	premium []db.EventRow,
	recommended []db.EventRow,
	latest []db.EventRow,
	popular []db.EventRow,
) getHomeEventsResponse {
	return getHomeEventsResponse{
		Premium:     convertEvents(premium),
		Recommended: convertEvents(recommended),
		Latest:      convertEvents(latest),
		Popular:     convertEvents(popular),
	}
}

func (server *Server) getHomeEvents(ctx *gin.Context) {
	userID, ok := ctx.Request.Context().Value("user_id").(int32)

	lat, lon, err := getUserLocation(ctx)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	var premiumEvents []db.EventRow
	premiumRows, err := server.store.GetPremiumEvents(ctx)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	premiumEvents = db.ConvertToEventRow(premiumRows)

	var recommendedEvents []db.EventRow
	if ok {
		arg := db.GetUserRecommendedEventsParams{
			UserID:  userID,
			UserLat: lat,
			UserLon: lon,
		}

		rows, err := server.store.GetUserRecommendedEvents(ctx, arg)
		if err != nil {
			ctx.Error(apperror.WrapDBError(err))
			return
		}

		recommendedEvents = db.ConvertToEventRow(rows)
	} else {
		arg := db.GetGuestRecommendedEventsParams{
			UserLat: lat,
			UserLon: lon,
		}

		rows, err := server.store.GetGuestRecommendedEvents(ctx, arg)
		if err != nil {
			ctx.Error(apperror.WrapDBError(err))
			return
		}

		recommendedEvents = db.ConvertToEventRow(rows)

	}

	var latestEvents []db.EventRow
	rowsLatest, err := server.store.GetLatestEvents(ctx)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	latestEvents = db.ConvertToEventRow(rowsLatest)

	var popularEvents []db.EventRow
	rowsPopular, err := server.store.GetPopularEvents(ctx)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	popularEvents = db.ConvertToEventRow(rowsPopular)

	resp := newGetHomeEventsResponse(premiumEvents, recommendedEvents, latestEvents, popularEvents)

	ctx.JSON(http.StatusOK, resp)
}
