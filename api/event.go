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

func newEventTxResponse(event db.EventResponse) eventResponse {
	return eventResponse{
		ID:            event.ID,
		Name:          event.Name,
		Description:   event.Description,
		Capacity:         event.Capacity,
		Latitude:         event.Latitude,
		Longitude:        event.Longitude,
		Address:          event.Address,
		Date:             event.Date,
		IsPrivate:        event.IsPrivate,
		IsPremium:        event.IsPremium,
		CreatedAt:        event.CreatedAt,
		Tags:             event.Tags,
		OwnerUsername:    event.OwnerUsername,
		ParticipantCount: int32(event.ParticipantsCount),
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

func newCreateEventResponse(event db.EventResponse) createEventResponse {
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

func newEventsListResponse(events []db.EventResponse) eventsListResponse {
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

	var events []db.EventResponse
	rows, err := server.store.ListEvents(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	for _, row := range rows {
		events = append(events, db.ConvertRowToEvent(row))
	}

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

func newEventResponse(event db.EventResponse) eventResponse {
	return eventResponse{
		ID:               event.ID,
		Name:             event.Name,
		Description:      event.Description,
		Capacity:         event.Capacity,
		Latitude:         event.Latitude,
		Longitude:        event.Longitude,
		Address:          event.Address,
		Date:             event.Date,
		IsPrivate:        event.IsPrivate,
		IsPremium:        event.IsPremium,
		CreatedAt:        event.CreatedAt,
		Tags:             event.Tags,
		OwnerUsername:    event.OwnerUsername,
		ParticipantCount: int32(event.ParticipantsCount),
	}
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
	event := db.ConvertRowToEvent(row)

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

func newUpdateEventResponse(event db.EventResponse) updateEventResponse {
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
	premium []db.EventResponse,
	recommended []db.EventResponse,
	latest []db.EventResponse,
	popular []db.EventResponse,
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

	var premiumEvents []db.EventResponse
	rowsPremium, err := server.store.GetPremiumEvents(ctx)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	for _, row := range rowsPremium {
		premiumEvents = append(premiumEvents, db.ConvertRowToEvent(row))
	}

	var recommendedEvents []db.EventResponse

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
		for _, row := range rows {
			recommendedEvents = append(recommendedEvents, db.ConvertRowToEvent(row))
		}
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

		for _, row := range rows {
			recommendedEvents = append(recommendedEvents, db.ConvertRowToEvent(row))
		}
	}

	var latestEvents []db.EventResponse
	rowsLatest, err := server.store.GetLatestEvents(ctx)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	for _, row := range rowsLatest {
		latestEvents = append(latestEvents, db.ConvertRowToEvent(row))
	}

	var popularEvents []db.EventResponse
	rowsPopular, err := server.store.GetPopularEvents(ctx)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	for _, row := range rowsPopular {
		popularEvents = append(popularEvents, db.ConvertRowToEvent(row))
	}

	resp := newGetHomeEventsResponse(premiumEvents, recommendedEvents, latestEvents, popularEvents)

	ctx.JSON(http.StatusOK, resp)
}
