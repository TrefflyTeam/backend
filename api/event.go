package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"strconv"
	"strings"
	"treffly/api/dto/event"
	"treffly/apperror"
	db "treffly/db/sqlc"
)

const (
	defaultLat = "51.660781"
	defaultLon = "39.200296"
)

func (server *Server) createEvent(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	var req eventdto.CreateEvent
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

	ctx.JSON(http.StatusOK, eventdto.ConvertEvent(event))
}

func (server *Server) listEvents(ctx *gin.Context) {
	lat, lon, err := getUserLocation(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	searchTerm := ctx.Query("keywords")
	tags := ctx.Query("tags")
	date := ctx.Query("dateWithin")
	tagIDs, err := parseTagIDs(tags)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	arg := db.ListEventsParams{
		UserLon: lon,
		UserLat: lat,
		SearchTerm: searchTerm,
		TagIds: tagIDs,
		DateRange: date,
	}

	var events []db.EventRow
	rows, err := server.store.ListEvents(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	events = db.ConvertToEventRow(rows)

	ctx.JSON(http.StatusOK, eventdto.NewEventsList(events))
}

func parseTagIDs(tagsStr string) ([]int32, error) {
	if tagsStr == "" {
		return []int32{}, nil
	}

	strIDs := strings.Split(tagsStr, ",")
	result := make([]int32, 0, len(strIDs))

	for _, strID := range strIDs {
		cleaned := strings.TrimSpace(strID)
		if cleaned == "" {
			continue
		}

		id, err := strconv.ParseInt(cleaned, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid tag ID: %s", strID)
		}

		result = append(result, int32(id))
	}

	return result, nil
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
	userID := getUserIDFromSoftAuth(ctx)

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

	isOwner := row.OwnerID == userID

	arg := db.IsParticipantParams{
		UserID: userID,
		EventID: eventID,
	}

	isParticipant, err := server.store.IsParticipant(ctx, arg)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	resp := eventdto.NewEventByID(eventdto.ConvertEvent(row), isOwner, isParticipant)

	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) updateEvent(ctx *gin.Context) {
	var req eventdto.UpdateEvent
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

	ctx.JSON(http.StatusOK, eventdto.ConvertEvent(eventUpdated))
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

	event, err := server.store.GetEvent(ctx, eventID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	argPart := db.IsParticipantParams{
		EventID: eventID,
		UserID:  userID,
	}

	isParticipant, err := server.store.IsParticipant(ctx, argPart)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	isOwner := event.OwnerID == userID

	resp := eventdto.NewEventByID(eventdto.ConvertEvent(event), isOwner, isParticipant)

	ctx.JSON(http.StatusOK, resp)
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

	event, err := server.store.GetEvent(ctx, eventID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	argPart := db.IsParticipantParams{
		EventID: eventID,
		UserID:  userID,
	}

	isParticipant, err := server.store.IsParticipant(ctx, argPart)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	isOwner := event.OwnerID == userID

	resp := eventdto.NewEventByID(eventdto.ConvertEvent(event), isOwner, isParticipant)

	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) getHomeEvents(ctx *gin.Context) {
	userID := getUserIDFromSoftAuth(ctx)

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
	if userID > 0 {
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

	resp := eventdto.NewGetHomeEvents(premiumEvents, recommendedEvents, latestEvents, popularEvents)

	ctx.JSON(http.StatusOK, resp)
}

func getUserIDFromSoftAuth(ctx *gin.Context) int32 {
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		userIDStr = -1
	}

	userID, ok := userIDStr.(int32)
	if !ok {
		userID = -1
	}

	return userID
}

func (server *Server) getCurrentUserUpcomingEvents(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	var events []db.EventRow
	rows, err := server.store.GetUpcomingUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	events = db.ConvertToEventRow(rows)

	ctx.JSON(http.StatusOK, eventdto.ConvertEvents(events))
}

func (server *Server) getCurrentUserPastEvents(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	var events []db.EventRow
	rows, err := server.store.GetPastUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}
	events = db.ConvertToEventRow(rows)

	ctx.JSON(http.StatusOK, eventdto.ConvertEvents(events))
}

func (server *Server) getCurrentUserOwnedEvents(ctx *gin.Context) {
	userID := getUserIDFromContextPayload(ctx)

	var events []db.EventRow
	rows, err := server.store.GetOwnedUserEvents(ctx, userID)
	if err != nil {
		ctx.Error(apperror.WrapDBError(err))
		return
	}

	events = db.ConvertToEventRow(rows)

	ctx.JSON(http.StatusOK, eventdto.ConvertEvents(events))
}